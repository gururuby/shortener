package handler

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	"github.com/gururuby/shortener/internal/handler/http/shorturl/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateShortURL_Ok(t *testing.T) {
	var err error
	var body []byte

	ctrl := gomock.NewController(t)
	uc := mocks.NewMockShortURLUseCase(ctrl)
	uc.EXPECT().CreateShortURL("http://example.com").Return("http://localhost:8080/mock_alias", nil).AnyTimes()

	r := chi.NewRouter()
	h := handler{router: r, uc: uc}

	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(`{"url":"http://example.com"}`)))
	request.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateShortURL()(w, request)

	resp := w.Result()

	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)

	require.NoError(t, err)

	require.JSONEq(t, `{"Result":"http://localhost:8080/mock_alias"}`, string(body))
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}
