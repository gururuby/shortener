package app

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

func TestAppOkRequests(t *testing.T) {
	app := NewApp()
	ts := httptest.NewServer(app.router)
	defer ts.Close()

	var tests = []struct {
		path      string
		method    string
		body      io.Reader
		wantMatch string
		status    int
	}{
		{
			path:      "/",
			method:    "POST",
			body:      strings.NewReader("https://ya.ru"),
			wantMatch: "^http://localhost:8080/\\w{5}$",
			status:    http.StatusCreated,
		},
	}
	for _, tt := range tests {
		response, result := testRequest(t, ts, tt.method, tt.path, tt.body)
		err := response.Body.Close()
		require.NoError(t, err)
		assert.Equal(t, tt.status, response.StatusCode)
		assert.Regexp(t, regexp.MustCompile(tt.wantMatch), result)
	}
}

func TestAppErrorRequests(t *testing.T) {
	app := NewApp()
	ts := httptest.NewServer(app.router)
	defer ts.Close()

	var tests = []struct {
		path   string
		method string
		body   io.Reader
		want   string
		status int
	}{
		{
			path:   "/",
			method: "GET",
			status: http.StatusMethodNotAllowed,
		},
		{
			path:   "/unknown",
			method: "GET",
			want:   "source URL not found\n",
			status: http.StatusUnprocessableEntity,
		},
	}
	for _, tt := range tests {
		response, result := testRequest(t, ts, tt.method, tt.path, tt.body)
		err := response.Body.Close()
		require.NoError(t, err)
		assert.Equal(t, tt.status, response.StatusCode)
		if tt.want != "" {
			assert.Equal(t, tt.want, result)
		}
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method string, path string, body io.Reader) (*http.Response, string) {
	request, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	response, err := ts.Client().Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	respBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return response, string(respBody)
}
