package handler

import (
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
		contentType string
		method      string
		path        string
	}

	response struct {
		body   string
		status int
	}

	ucOutput struct {
		res []*usecase.UserShortURL
		err error
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
		name     string
		request  request
		response response
		ucInput  *userEntity.User
		ucOutput ucOutput
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
			h.GetUserURLs()(w, req)

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

func Test_GetURLs_Errors(t *testing.T) {
	var (
		err  error
		body []byte
	)

	ctrl := gomock.NewController(t)
	userUC := mocks.NewMockUserUseCase(ctrl)

	r := chi.NewRouter()
	h := handler{router: r, userUC: userUC}

	var tests = []struct {
		name     string
		request  request
		response response
		ucInput  *userEntity.User
		ucOutput ucOutput
	}{
		{
			name: "when incorrect http method was used",
			request: request{
				contentType: "application/json",
				method:      http.MethodPost,
				path:        "/api/user/urls",
			},
			response: response{
				body:   `{"StatusCode":405,"Error":"HTTP method POST is not allowed"}`,
				status: http.StatusMethodNotAllowed,
			},
		},
		{
			name:    "when no urls for user",
			ucInput: &userEntity.User{ID: 1},
			ucOutput: ucOutput{
				res: []*usecase.UserShortURL{},
				err: nil,
			},
			request: request{
				contentType: "application/json",
				method:      http.MethodGet,
				path:        "/api/user/urls",
			},
			response: response{
				body:   `{}`,
				status: http.StatusNoContent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, tt.request.path, nil)
			req.Header.Set("Content-Type", tt.request.contentType)
			if tt.ucInput != nil {
				userUC.EXPECT().Register(gomock.Any()).Return(tt.ucInput, nil)
				userUC.EXPECT().GetURLs(gomock.Any(), tt.ucInput).Return(tt.ucOutput.res, tt.ucOutput.err).Times(1)
			}
			w := httptest.NewRecorder()
			h.GetUserURLs()(w, req)

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
