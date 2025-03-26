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

func TestValidRequests(t *testing.T) {
	storage := mocks.NewMockStorage()
	config := appConfig.NewConfig()
	ts := httptest.NewServer(router.NewRouter(config, storage))
	defer ts.Close()

	var specs = []struct {
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
	}
	for _, spec := range specs {
		response, result := testRequest(t, ts, spec.method, spec.path, spec.body)
		err := response.Body.Close()
		require.NoError(t, err)
		assert.Equal(t, spec.status, response.StatusCode)
		assert.Equal(t, spec.want, result)
	}
}

func TestInvalidRequests(t *testing.T) {
	storage := mocks.NewMockStorage()
	config := appConfig.NewConfig()
	ts := httptest.NewServer(router.NewRouter(config, storage))
	defer ts.Close()

	var specs = []struct {
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
			want:   "Source URL not found\n",
			status: http.StatusUnprocessableEntity,
		},
	}
	for _, spec := range specs {
		response, result := testRequest(t, ts, spec.method, spec.path, spec.body)
		err := response.Body.Close()
		require.NoError(t, err)
		assert.Equal(t, spec.status, response.StatusCode)
		if spec.want != "" {
			assert.Equal(t, spec.want, result)
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
