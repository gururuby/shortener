package handler

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	"github.com/gururuby/shortener/internal/domain/usecase/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateShortURL_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	uc := mock.NewMockUseCase(ctrl)
	uc.EXPECT().CreateShortURL("http://example.com").Return("http://localhost:8080/mock_alias", nil).AnyTimes()

	r := chi.NewRouter()
	h := handler{router: r, uc: uc}

	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(`{"url":"http://example.com"}`)))
	request.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateShortURL()(w, request)

	res := w.Result()

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)

	require.NoError(t, err)

	assert.Equal(t, "{\"Result\":\"http://localhost:8080/mock_alias\"}", string(resBody))
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}
