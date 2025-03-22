package controllers

import (
	"github.com/gururuby/shortener/internal/app/models"
	"github.com/gururuby/shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockShortURLsRepo struct {
	Data map[string]models.ShortURL
}

func NewShortURLsRepo() *MockShortURLsRepo {
	return &MockShortURLsRepo{
		Data: make(map[string]models.ShortURL),
	}
}

func (repo *MockShortURLsRepo) CreateShortURL(BaseURL string) string {
	shortURL := models.NewShortURL(BaseURL)
	shortURL.Alias = "mock_alias"
	repo.Data[shortURL.Alias] = shortURL

	return shortURL.AliasURL()
}

func (repo *MockShortURLsRepo) FindShortURL(alias string) (string, bool) {
	shortURL, ok := repo.Data[alias]

	return shortURL.BaseURL, ok
}

func TestShortURLCreate(t *testing.T) {
	mockStorage := NewShortURLsRepo()

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
		name    string
		storage storage.StorageInterface
		send    request
		want    response
	}{
		{
			name:    "when successfully created short url",
			storage: mockStorage,
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
			name:    "when was not passed base URL",
			storage: mockStorage,
			send: request{
				method: http.MethodPost,
				body:   strings.NewReader(""),
				path:   "/",
			},
			want: response{
				code:        http.StatusUnprocessableEntity,
				body:        "Empty base URL, please specify URL\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "when request not allowed HTTP method",
			storage: mockStorage,
			send: request{
				method: http.MethodGet,
				body:   strings.NewReader("http://example.com"),
				path:   "/",
			},
			want: response{
				code:        http.StatusBadRequest,
				body:        "Bad request\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.send.method, tt.send.path, tt.send.body)
			w := httptest.NewRecorder()
			ShortURLCreate(tt.storage)(w, request)

			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, _ := io.ReadAll(res.Body)

			assert.Equal(t, tt.want.body, string(resBody))
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}

func TestShortURLShow(t *testing.T) {
	mockStorage := NewShortURLsRepo()

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
		storage storage.StorageInterface
		baseUrl string
		send    request
		want    response
	}{
		{
			name:    "when successfully find base url by alias",
			storage: mockStorage,
			baseUrl: "https://example.com",
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
			storage: mockStorage,
			baseUrl: "https://example.com",
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
			storage: mockStorage,
			baseUrl: "https://example.com",
			send: request{
				method: http.MethodGet,
				path:   "/unknown",
			},
			want: response{
				code:        http.StatusNotFound,
				body:        "URL was not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "when request not allowed HTTP method",
			baseUrl: "https://example.com",
			storage: mockStorage,
			send: request{
				method: http.MethodPost,
				path:   "/mock_alias",
			},
			want: response{
				code:        http.StatusBadRequest,
				body:        "Bad request\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.CreateShortURL(tt.baseUrl)

			request := httptest.NewRequest(tt.send.method, tt.send.path, nil)
			w := httptest.NewRecorder()
			ShortURLShow(tt.storage)(w, request)

			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, _ := io.ReadAll(res.Body)

			assert.Equal(t, tt.want.body, string(resBody))

			assert.Equal(t, tt.want.location, res.Header.Get("Location"))
		})
	}
}
