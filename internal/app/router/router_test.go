package router

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method string, body io.Reader, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	ts := httptest.NewServer(Router())
	defer ts.Close()

	var testTable = []struct {
		url    string
		method string
		body   io.Reader
		want   string
		match  string
		status int
	}{
		{
			url:    "/",
			method: "POST",
			body:   strings.NewReader("https://ya.ru"),
			match:  `\Ahttp:\/\/localhost:8080\/\w{5}\z`,
			status: http.StatusCreated,
		},
		{
			url:    "/",
			method: "GET",
			status: http.StatusMethodNotAllowed,
		},
		{
			url:    "/unknown",
			method: "GET",
			want:   "URL was not found\n",
			status: http.StatusNotFound,
		},
	}
	for _, v := range testTable {
		resp, get := testRequest(t, ts, v.method, v.body, v.url)
		assert.Equal(t, v.status, resp.StatusCode)
		if v.want != "" {
			assert.Equal(t, v.want, get)
		}

		if v.match != "" {
			assert.Regexp(t, regexp.MustCompile(v.match), get)
		}

	}
}
