package handlers

import (
	"fmt"
	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/services"
	"io"
	"net/http"
)

type Storage interface {
	Save(string, string) (string, bool)
	Find(string) (string, bool)
}

type URLsHandler struct {
	storage Storage
	config  *config.Config
}

func NewURLsHandler(config *config.Config, storage Storage) *URLsHandler {
	return &URLsHandler{
		storage: storage,
		config:  config,
	}
}

func (h *URLsHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ok, message := disallowedMethod(r, http.MethodPost); !ok {
			http.Error(w, message, http.StatusMethodNotAllowed)
			return
		}

		result, ok := services.SaveURL(h.config.BaseURL, h.storage, r.Body)

		if !ok {
			http.Error(w, result, http.StatusUnprocessableEntity)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_, err := io.WriteString(w, result)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	}
}

func (h *URLsHandler) Show() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ok, message := disallowedMethod(r, http.MethodGet); !ok {
			http.Error(w, message, http.StatusMethodNotAllowed)
			return
		}

		result, ok := services.FindURL(h.storage, r.URL.Path)

		if !ok {
			http.Error(w, result, http.StatusUnprocessableEntity)
			return
		}

		w.Header().Set("Location", result)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func disallowedMethod(r *http.Request, allowedMethod string) (bool, string) {
	if r.Method != allowedMethod {
		err := fmt.Sprintf("Method %s is not allowed", r.Method)
		return false, err
	}

	return true, "ok"
}
