package handler

import (
	"github.com/go-chi/chi/v5"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
	"github.com/gururuby/shortener/internal/handler/http/app/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPingOK(t *testing.T) {
	var err error

	ctrl := gomock.NewController(t)
	uc := mocks.NewMockAppUseCase(ctrl)
	uc.EXPECT().PingDB().Return(nil)

	r := chi.NewRouter()
	h := handler{router: r, uc: uc}

	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	h.PingDB()(w, request)

	resp := w.Result()

	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	_, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
}

func TestPingErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	uc := mocks.NewMockAppUseCase(ctrl)

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
				err: ucErrors.ErrAppDBIsNotReady,
			},
			request: request{
				method: http.MethodGet,
				path:   "/ping",
			},
			response: response{
				code:        http.StatusUnprocessableEntity,
				body:        "db is not ready\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "when request method is not allowed",
			request: request{
				method: http.MethodPost,
				path:   "/ping",
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

			uc.EXPECT().PingDB().Return(tt.useCaseRes.err).AnyTimes()

			r := chi.NewRouter()
			h := handler{router: r, uc: uc}

			req := httptest.NewRequest(tt.request.method, tt.request.path, nil)
			w := httptest.NewRecorder()

			h.PingDB()(w, req)

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
