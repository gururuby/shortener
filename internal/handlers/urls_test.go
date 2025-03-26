package handlers

import (
	appConfig "github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	storage := mocks.NewMockStorage()
	config := appConfig.NewConfig()
	handler := URLsHandler{
		storage: storage,
		config:  config,
	}

	type response struct {
		code        int
		body        string
		contentType string
	}

	type request struct {
		method string
		path   string
		body   io.Reader
	}

	tests := []struct {
		name string
		send request
		want response
	}{
		{
			name: "when successfully created short url",
			send: request{
				method: http.MethodPost,
				body:   strings.NewReader("http://example.com"),
				path:   "/",
			},
			want: response{
				code:        http.StatusCreated,
				body:        "http://localhost:8080/mock_alias",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "when was not passed base URL",
			send: request{
				method: http.MethodPost,
				body:   strings.NewReader(""),
				path:   "/",
			},
			want: response{
				code:        http.StatusUnprocessableEntity,
				body:        "Empty source URL, please specify URL\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "when request not allowed HTTP method",
			send: request{
				method: http.MethodGet,
				body:   strings.NewReader("http://example.com"),
				path:   "/",
			},
			want: response{
				code:        http.StatusMethodNotAllowed,
				body:        "Method GET is not allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.send.method, tt.send.path, tt.send.body)
			w := httptest.NewRecorder()
			handler.Create()(w, request)

			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)

			assert.Equal(t, tt.want.body, string(resBody))
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}

func TestShow(t *testing.T) {
	storage := mocks.NewMockStorage()
	config := appConfig.NewConfig()
	handler := URLsHandler{
		storage: storage,
		config:  config,
	}

	type response struct {
		code        int
		location    string
		body        string
		contentType string
	}

	type request struct {
		method string
		path   string
	}

	tests := []struct {
		name    string
		baseURL string
		send    request
		want    response
	}{
		{
			name:    "when successfully find base url by alias",
			baseURL: "https://example.com",
			send: request{
				method: http.MethodGet,
				path:   "/mock_alias",
			},
			want: response{
				code:     http.StatusTemporaryRedirect,
				location: "https://example.com",
			},
		},
		{
			name:    "when alias was not passed",
			baseURL: "https://example.com",
			send: request{
				method: http.MethodGet,
				path:   "/",
			},
			want: response{
				code:        http.StatusUnprocessableEntity,
				body:        "Empty alias, please specify alias\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "when base URL was not found",
			baseURL: "https://example.com",
			send: request{
				method: http.MethodGet,
				path:   "/unknown",
			},
			want: response{
				code:        http.StatusUnprocessableEntity,
				body:        "Source URL not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "when request not allowed HTTP method",
			baseURL: "https://example.com",
			send: request{
				method: http.MethodPost,
				path:   "/mock_alias",
			},
			want: response{
				code:        http.StatusMethodNotAllowed,
				body:        "Method POST is not allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Save(config.BaseURL, tt.baseURL)

			request := httptest.NewRequest(tt.send.method, tt.send.path, nil)
			w := httptest.NewRecorder()
			handler.Show()(w, request)

			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)

			assert.Equal(t, tt.want.body, string(resBody))
			assert.Equal(t, tt.want.location, res.Header.Get("Location"))
		})
	}
}
