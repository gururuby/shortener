package handler

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	usecase "github.com/gururuby/shortener/internal/domain/usecase/user"
	"github.com/gururuby/shortener/internal/handler/http/api/user/mocks"
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
		err error
		res []*usecase.UserShortURL
	}
)

func Test_GetURLs_OK(t *testing.T) {
	var (
		err  error
		body []byte
		urls []*usecase.UserShortURL
	)

	ctrl := gomock.NewController(t)
	userUC := mocks.NewMockUserUseCase(ctrl)

	urls = append(urls, &usecase.UserShortURL{ShortURL: "http://example.com/alias", OriginalURL: "https://ya.ru"})

	r := chi.NewRouter()
	h := handler{router: r, userUC: userUC}

	var tests = []struct {
		request  request
		ucOutput ucOutput
		ucInput  *userEntity.User
		name     string
		response response
	}{
		{
			name: "when success receive user urls",
			request: request{
				contentType: "application/json",
				method:      http.MethodGet,
				path:        "/api/user/urls",
			},
			response: response{
				status: http.StatusOK,
				body:   `[{"short_url":"http://example.com/alias","original_url":"https://ya.ru"}]`,
			},
			ucInput: &userEntity.User{ID: 1},
			ucOutput: ucOutput{
				res: urls,
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, tt.request.path, nil)
			req.Header.Set("Content-Type", tt.request.contentType)

			w := httptest.NewRecorder()
			userUC.EXPECT().Register(gomock.Any()).Return(tt.ucInput, nil)
			userUC.EXPECT().GetURLs(gomock.Any(), tt.ucInput).Return(tt.ucOutput.res, tt.ucOutput.err).Times(1)
			h.GetURLs()(w, req)

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

func Test_DeleteURLs_OK(t *testing.T) {
	user := &userEntity.User{ID: 1}

	ctrl := gomock.NewController(t)
	userUC := mocks.NewMockUserUseCase(ctrl)

	r := chi.NewRouter()
	h := handler{router: r, userUC: userUC}

	var tests = []struct {
		name     string
		request  request
		response response
		ucInput  []string
	}{
		{
			name: "when success delete urls",
			request: request{
				contentType: "application/json",
				method:      http.MethodDelete,
				path:        "/api/user/urls",
				body:        bytes.NewBufferString(`["alias1", "alias2"]`),
			},
			response: response{
				status: http.StatusAccepted,
			},
			ucInput: []string{"alias1", "alias2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, tt.request.path, tt.request.body)
			req.Header.Set("Content-Type", tt.request.contentType)

			w := httptest.NewRecorder()
			userUC.EXPECT().Register(gomock.Any()).Return(user, nil).AnyTimes()
			userUC.EXPECT().DeleteURLs(gomock.Any(), user, tt.ucInput).AnyTimes()
			h.DeleteURLs()(w, req)

			resp := w.Result()

			defer func() {
				err := resp.Body.Close()
				require.NoError(t, err)
			}()

			assert.Equal(t, tt.response.status, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		})
	}
}

func Test_DeleteURLs_Errors(t *testing.T) {
	var (
		err  error
		body []byte
	)

	user := &userEntity.User{ID: 1}

	ctrl := gomock.NewController(t)
	userUC := mocks.NewMockUserUseCase(ctrl)

	r := chi.NewRouter()
	h := handler{router: r, userUC: userUC}

	var tests = []struct {
		name     string
		request  request
		response response
	}{
		{
			name: "when incorrect http method was used",
			request: request{
				contentType: "application/json",
				method:      http.MethodPost,
				path:        "/api/user/urls",
				body:        bytes.NewBufferString(`["alias1", "alias2"]`),
			},
			response: response{
				body:   `{"StatusCode":405,"Error":"HTTP method POST is not allowed"}`,
				status: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "when no aliases passed",
			request: request{
				contentType: "application/json",
				method:      http.MethodDelete,
				path:        "/api/user/urls",
				body:        bytes.NewBufferString(`[]`),
			},
			response: response{
				body:   `{"StatusCode":400,"Error":"no aliases passed to delete short urls"}`,
				status: http.StatusBadRequest,
			},
		},
		{
			name: "when passed incorrect JSON",
			request: request{
				contentType: "application/json",
				method:      http.MethodDelete,
				path:        "/api/user/urls",
				body:        bytes.NewBufferString(`]`),
			},
			response: response{
				body:   `{"StatusCode":400,"Error":"invalid character ']' looking for beginning of value"}`,
				status: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, tt.request.path, tt.request.body)
			req.Header.Set("Content-Type", tt.request.contentType)
			userUC.EXPECT().Register(gomock.Any()).Return(user, nil).AnyTimes()
			w := httptest.NewRecorder()
			h.DeleteURLs()(w, req)

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
