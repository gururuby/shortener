package app

import (
	appConfig "github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/mocks"
	"github.com/gururuby/shortener/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method string, path string, body io.Reader) (*http.Response, string) {
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
	storage := mocks.NewMockShortURLsRepo()
	config := appConfig.NewConfig()
	ts := httptest.NewServer(router.NewRouter(config, storage))
	defer ts.Close()

	var testTable = []struct {
		path   string
		method string
		body   io.Reader
		want   string
		status int
	}{
		{
			path:   "/",
			method: "POST",
			body:   strings.NewReader("https://ya.ru"),
			want:   "http://localhost:8080/mock_alias",
			status: http.StatusCreated,
		},
		{
			path:   "/",
			method: "GET",
			status: http.StatusMethodNotAllowed,
		},
		{
			path:   "/unknown",
			method: "GET",
			want:   "URL was not found\n",
			status: http.StatusNotFound,
		},
	}
	for _, v := range testTable {
		resp, get := testRequest(t, ts, v.method, v.path, v.body)
		resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode)
		if v.want != "" {
			assert.Equal(t, v.want, get)
		}
	}
}
