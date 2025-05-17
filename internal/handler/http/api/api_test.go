package handler

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
	"github.com/gururuby/shortener/internal/handler/http/shorturl/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type (
	request struct {
		body        *bytes.Buffer
		contentType string
		method      string
		path        string
	}

	response struct {
		body   string
		status int
	}

	ucOutput struct {
		res string
		err error
	}
)

func TestCreateShortURL_OK(t *testing.T) {
	var err error
	var body []byte

	ctrl := gomock.NewController(t)
	uc := mocks.NewMockShortURLUseCase(ctrl)

	r := chi.NewRouter()
	h := handler{router: r, uc: uc}

	var tests = []struct {
		name     string
		request  request
		response response
		ucInput  string
		ucOutput ucOutput
	}{
		{
			name: "when success create short url",
			request: request{
				body:        bytes.NewBufferString(`{"url":"http://example.com"}`),
				contentType: "application/json",
				method:      http.MethodPost,
				path:        "/api/shorten",
			},
			response: response{
				status: http.StatusCreated,
				body:   `{"Result":"http://localhost:8080/mock_alias"}`,
			},
			ucInput: "http://example.com",
			ucOutput: ucOutput{
				res: "http://localhost:8080/mock_alias",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, tt.request.path, tt.request.body)
			req.Header.Set("Content-Type", tt.request.contentType)
			w := httptest.NewRecorder()
			uc.EXPECT().CreateShortURL(gomock.Any(), tt.ucInput).Return(tt.ucOutput.res, tt.ucOutput.err).Times(1)
			h.CreateShortURL()(w, req)

			resp := w.Result()

			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			assert.Equal(t, tt.response.status, resp.StatusCode)
			body, err = io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.JSONEq(t, tt.response.body, string(body))
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		})
	}
}

func TestCreateShortURL_Errors(t *testing.T) {
	var err error
	var body []byte

	ctrl := gomock.NewController(t)
	uc := mocks.NewMockShortURLUseCase(ctrl)

	r := chi.NewRouter()
	h := handler{router: r, uc: uc}

	var tests = []struct {
		name     string
		request  request
		response response
		ucInput  string
		ucOutput ucOutput
	}{
		{
			name: "when incorrect http method was used",
			request: request{
				body:        bytes.NewBufferString(`{"url":"https://example.com"}`),
				contentType: "application/json",
				method:      http.MethodGet,
				path:        "/api/shorten",
			},
			response: response{
				body:   `{"StatusCode":405,"Error":"HTTP method GET is not allowed"}`,
				status: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "when invalid json passed",
			request: request{
				body:        bytes.NewBufferString(`{{"url":"https://example.com"}`),
				contentType: "application/json",
				method:      http.MethodPost,
				path:        "/api/shorten",
			},
			response: response{
				body:   `{"StatusCode":400,"Error":"invalid character '{' looking for beginning of object key string"}`,
				status: http.StatusBadRequest,
			},
		},
		{
			name:    "when passed url is incorrect",
			ucInput: "//example.com",
			ucOutput: ucOutput{
				res: "",
				err: ucErrors.ErrShortURLInvalidSourceURL,
			},
			request: request{
				body:        bytes.NewBufferString(`{"url":"//example.com"}`),
				contentType: "application/json",
				method:      http.MethodPost,
				path:        "/api/shorten",
			},
			response: response{
				body:   `{"StatusCode":422,"Error":"invalid source URL, please specify valid URL"}`,
				status: http.StatusUnprocessableEntity,
			},
		},
		{
			name:    "when passed url is not unique",
			ucInput: "https://example.com",
			ucOutput: ucOutput{
				res: "http://localhost:8080/mock_alias",
				err: ucErrors.ErrShortURLAlreadyExist,
			},
			request: request{
				body:        bytes.NewBufferString(`{"url":"https://example.com"}`),
				contentType: "application/json",
				method:      http.MethodPost,
				path:        "/api/shorten",
			},
			response: response{
				body:   `{"Result":"http://localhost:8080/mock_alias"}`,
				status: http.StatusConflict,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, tt.request.path, tt.request.body)
			req.Header.Set("Content-Type", tt.request.contentType)
			w := httptest.NewRecorder()

			if tt.ucInput != "" {
				uc.EXPECT().CreateShortURL(gomock.Any(), tt.ucInput).Return(tt.ucOutput.res, tt.ucOutput.err).Times(1)
			}

			h.CreateShortURL()(w, req)

			resp := w.Result()

			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			assert.Equal(t, tt.response.status, resp.StatusCode)
			body, err = io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.JSONEq(t, tt.response.body, string(body))
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		})
	}
}

func TestBatchShortURLs_Errors(t *testing.T) {
	var err error
	var body []byte

	ctrl := gomock.NewController(t)
	uc := mocks.NewMockShortURLUseCase(ctrl)

	r := chi.NewRouter()
	h := handler{router: r, uc: uc}

	var tests = []struct {
		name     string
		request  request
		response response
	}{
		{
			name: "when incorrect http method was used",
			request: request{
				body:        bytes.NewBufferString(""),
				contentType: "application/json",
				method:      http.MethodGet,
				path:        "/api/shorten/batch",
			},
			response: response{
				body:   `{"StatusCode":405,"Error":"HTTP method GET is not allowed"}`,
				status: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "when invalid json passed",
			request: request{
				body:        bytes.NewBufferString(`{{"url":"https://example.com"}`),
				contentType: "application/json",
				method:      http.MethodPost,
				path:        "/api/shorten/batch",
			},
			response: response{
				body:   `{"StatusCode":400,"Error":"invalid character '{' looking for beginning of object key string"}`,
				status: http.StatusBadRequest,
			},
		},
		{
			name: "when passed empty batch",
			request: request{
				body:        bytes.NewBufferString(`[]`),
				contentType: "application/json",
				method:      http.MethodPost,
				path:        "/api/shorten/batch",
			},
			response: response{
				body:   `{"StatusCode":400,"Error":"nothing to process, empty batch"}`,
				status: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, tt.request.path, tt.request.body)
			req.Header.Set("Content-Type", tt.request.contentType)
			w := httptest.NewRecorder()
			h.BatchShortURLs()(w, req)

			resp := w.Result()

			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			assert.Equal(t, tt.response.status, resp.StatusCode)
			body, err = io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.JSONEq(t, tt.response.body, string(body))
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		})
	}
}
