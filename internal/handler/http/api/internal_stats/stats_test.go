package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	entity "github.com/gururuby/shortener/internal/domain/entity/stats"
	"github.com/gururuby/shortener/internal/handler/http/api/internal_stats/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
		err error
		res *entity.Stats
	}
)

func Test_GetStats_OK(t *testing.T) {
	var err error
	var body []byte

	ctrl := gomock.NewController(t)
	statsUC := mocks.NewMockStatsUseCase(ctrl)
	stats := &entity.Stats{UsersCount: 1, URLsCount: 2}

	r := chi.NewRouter()
	h := handler{router: r, statsUC: statsUC}

	var tests = []struct {
		ucOutput ucOutput
		request  request
		name     string
		response response
	}{
		{
			name: "when success receive stats from usecase",
			request: request{
				contentType: "application/json",
				method:      http.MethodGet,
				path:        "/api/internal/stats",
			},
			response: response{
				status: http.StatusOK,
				body:   `{"urls":2,"users":1}`,
			},
			ucOutput: ucOutput{
				res: stats,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, tt.request.path, nil)
			req.Header.Set("Content-Type", tt.request.contentType)
			w := httptest.NewRecorder()
			statsUC.EXPECT().GetStats(gomock.Any()).Return(tt.ucOutput.res, tt.ucOutput.err).Times(1)
			h.GetStats()(w, req)

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

func Test_GetStats_Errors(t *testing.T) {
	var err error
	var body []byte

	ctrl := gomock.NewController(t)
	statsUC := mocks.NewMockStatsUseCase(ctrl)

	r := chi.NewRouter()
	h := handler{router: r, statsUC: statsUC}

	var tests = []struct {
		ucOutput *ucOutput
		request  request
		name     string
		response response
	}{
		{
			name: "when incorrect http method was used",
			request: request{
				contentType: "application/json",
				method:      http.MethodPost,
				path:        "/api/internal/stats",
			},
			response: response{
				body:   `{"StatusCode":405,"Error":"HTTP method POST is not allowed"}`,
				status: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "when usecase returns error",
			request: request{
				contentType: "application/json",
				method:      http.MethodGet,
				path:        "/api/internal/stats",
			},
			response: response{
				body:   `{"StatusCode":422,"Error":"error"}`,
				status: http.StatusUnprocessableEntity,
			},
			ucOutput: &ucOutput{
				res: nil,
				err: errors.New("error"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, tt.request.path, nil)
			req.Header.Set("Content-Type", tt.request.contentType)
			w := httptest.NewRecorder()
			if tt.ucOutput != nil {
				statsUC.EXPECT().GetStats(gomock.Any()).Return(tt.ucOutput.res, tt.ucOutput.err).Times(1)
			}

			h.GetStats()(w, req)

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
