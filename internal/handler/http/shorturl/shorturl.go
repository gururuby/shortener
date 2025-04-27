//go:generate mockgen -destination=./mocks/mock.go -package=mocks . ShortURLUseCase

package handler

import (
	"fmt"
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
	CreateShortURL(sourceURL string) (string, error)
	FindShortURL(alias string) (string, error)
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
		var err error
		var reqBody []byte
		var shortURL string

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

		shortURL, err = h.uc.CreateShortURL(sourceURL)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
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
		result, err := h.uc.FindShortURL(r.URL.Path)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Location", result)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
