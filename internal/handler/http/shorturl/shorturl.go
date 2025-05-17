//go:generate mockgen -destination=./mocks/mock.go -package=mocks . ShortURLUseCase

package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/gururuby/shortener/internal/domain/entity"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
	"io"
	"net/http"
)

const (
	shortensPath = "/"
	shortenPath  = "/{alias}"
)

type Router interface {
	Post(path string, h http.HandlerFunc)
	Get(path string, h http.HandlerFunc)
}

type ShortURLUseCase interface {
	CreateShortURL(ctx context.Context, sourceURL string) (string, error)
	FindShortURL(ctx context.Context, alias string) (string, error)
	BatchShortURLs(ctx context.Context, urls []entity.BatchShortURLInput) []entity.BatchShortURLOutput
}

type handler struct {
	uc     ShortURLUseCase
	router Router
}

func Register(router Router, uc ShortURLUseCase) {
	h := handler{router: router, uc: uc}
	h.router.Get(shortenPath, h.FindShortURL())
	h.router.Post(shortensPath, h.CreateShortURL())

}

func (h *handler) CreateShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err        error
			reqBody    []byte
			shortURL   string
			statusCode = http.StatusCreated
		)

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		reqBody, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sourceURL := string(reqBody)

		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}(r.Body)

		shortURL, err = h.uc.CreateShortURL(r.Context(), sourceURL)

		if err != nil {
			if errors.Is(err, ucErrors.ErrShortURLAlreadyExist) {
				statusCode = http.StatusConflict
			} else {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(statusCode)

		_, err = io.WriteString(w, shortURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *handler) FindShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}
		result, err := h.uc.FindShortURL(r.Context(), r.URL.Path)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Location", result)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
