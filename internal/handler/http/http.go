//go:generate mockgen -destination=./mock_handler/mock.go . UseCase

package handler

import (
	"fmt"
	"github.com/gururuby/shortener/internal/infra/logger"
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

type UseCase interface {
	CreateShortURL(sourceURL string) (string, error)
	FindShortURL(alias string) (string, error)
}

type handler struct {
	uc     UseCase
	router Router
}

func Register(router Router, uc UseCase) {
	h := handler{router: router, uc: uc}
	h.router.Get(shortenPath, h.FindShortURL())
	h.router.Post(shortensPath, h.CreateShortURL())

}

func (h *handler) CreateShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		reqBody, _ := io.ReadAll(r.Body)
		sourceURL := string(reqBody)
		defer r.Body.Close()

		res, err := h.uc.CreateShortURL(sourceURL)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_, err = io.WriteString(w, res)
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
		logger.Log.Info(r.URL.Path)
		result, err := h.uc.FindShortURL(r.URL.Path)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.Header().Set("Location", result)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
