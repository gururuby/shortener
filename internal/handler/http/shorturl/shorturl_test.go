package handler

import (
	"github.com/go-chi/chi/v5"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/shorturl/errors"
	"github.com/gururuby/shortener/internal/handler/http/shorturl/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_CreateShortURL_OK(t *testing.T) {
	var err error
	var body []byte

	ctrl := gomock.NewController(t)
	uc := mocks.NewMockShortURLUseCase(ctrl)

	r := chi.NewRouter()
	h := handler{router: r, uc: uc}

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com"))
	uc.EXPECT().CreateShortURL(req.Context(), "http://example.com").Return("http://localhost:8080/mock_alias", nil).AnyTimes()

	w := httptest.NewRecorder()
	h.CreateShortURL()(w, req)

	resp := w.Result()

	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:8080/mock_alias", string(body))
	assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
}

func Test_CreateShortURL_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	uc := mocks.NewMockShortURLUseCase(ctrl)

	type request struct {
		method string
		path   string
		body   string
	}

	type response struct {
		code        int
		body        string
		contentType string
	}

	type useCaseResult struct {
		res string
		err error
	}

	tests := []struct {
		name       string
		request    request
		response   response
		useCaseRes useCaseResult
	}{
		{
			name: "when use case returns some error",
			useCaseRes: useCaseResult{
				res: "",
				err: ucErrors.ErrShortURLInvalidSourceURL,
			},
			request: request{
				method: http.MethodPost,
				body:   "http://example.com",
				path:   "/",
			},
			response: response{
				code:        http.StatusUnprocessableEntity,
				body:        "invalid source URL, please specify valid URL\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "when use case conflict error",
			useCaseRes: useCaseResult{
				res: "http://localhost:8080/mock_alias",
				err: ucErrors.ErrShortURLAlreadyExist,
			},
			request: request{
				method: http.MethodPost,
				body:   "https://example.com",
				path:   "/",
			},
			response: response{
				code:        http.StatusConflict,
				body:        "http://localhost:8080/mock_alias",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "when request method is not allowed",
			request: request{
				method: http.MethodGet,
				path:   "/",
			},
			response: response{
				code:        http.StatusMethodNotAllowed,
				body:        "HTTP method GET is not allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			var body []byte

			r := chi.NewRouter()
			h := handler{router: r, uc: uc}

			req := httptest.NewRequest(tt.request.method, tt.request.path, strings.NewReader(tt.request.body))
			uc.EXPECT().CreateShortURL(req.Context(), tt.request.body).Return(tt.useCaseRes.res, tt.useCaseRes.err).AnyTimes()

			w := httptest.NewRecorder()
			h.CreateShortURL()(w, req)

			resp := w.Result()

			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			assert.Equal(t, tt.response.code, resp.StatusCode)

			body, err = io.ReadAll(resp.Body)

			require.NoError(t, err)

			assert.Equal(t, tt.response.body, string(body))
			assert.Equal(t, tt.response.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

func Test_FindShortURL_OK(t *testing.T) {
	var err error

	ctrl := gomock.NewController(t)
	uc := mocks.NewMockShortURLUseCase(ctrl)

	r := chi.NewRouter()
	h := handler{router: r, uc: uc}

	req := httptest.NewRequest(http.MethodGet, "/some_alias", nil)
	uc.EXPECT().FindShortURL(req.Context(), "/some_alias").Return("https://ya.ru", nil)

	w := httptest.NewRecorder()
	h.FindShortURL()(w, req)

	resp := w.Result()

	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()

	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)

	_, err = io.ReadAll(resp.Body)

	require.NoError(t, err)
	assert.Equal(t, "https://ya.ru", resp.Header.Get("Location"))
}

func Test_FindShortURLErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	uc := mocks.NewMockShortURLUseCase(ctrl)

	type request struct {
		method string
		path   string
	}

	type response struct {
		code        int
		body        string
		contentType string
	}

	type useCaseResult struct {
		res string
		err error
	}

	tests := []struct {
		name       string
		request    request
		response   response
		useCaseRes useCaseResult
	}{
		{
			name: "when use case returns some error",
			useCaseRes: useCaseResult{
				res: "",
				err: ucErrors.ErrShortURLEmptyAlias,
			},
			request: request{
				method: http.MethodGet,
				path:   "/alias1",
			},
			response: response{
				code:        http.StatusUnprocessableEntity,
				body:        "empty alias, please specify alias\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "when request method is not allowed",
			request: request{
				method: http.MethodPost,
				path:   "/alias2",
			},
			response: response{
				code:        http.StatusMethodNotAllowed,
				body:        "HTTP method POST is not allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			var body []byte

			r := chi.NewRouter()
			h := handler{router: r, uc: uc}

			req := httptest.NewRequest(tt.request.method, tt.request.path, nil)
			uc.EXPECT().FindShortURL(req.Context(), tt.request.path).Return(tt.useCaseRes.res, tt.useCaseRes.err).AnyTimes()

			w := httptest.NewRecorder()

			h.FindShortURL()(w, req)

			resp := w.Result()

			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			assert.Equal(t, tt.response.code, resp.StatusCode)

			body, err = io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.response.body, string(body))
		})
	}
}
