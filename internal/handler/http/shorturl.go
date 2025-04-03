package handler

import (
	"fmt"
	"io"
	"net/http"
)

type shortURLUseCase interface {
	CreateShortURL(sourceURL string) (string, error)
	FindShortURL(alias string) (string, error)
}

type ShortURLHandler struct {
	useCase shortURLUseCase
}

func NewShortURLHandler(uc shortURLUseCase) *ShortURLHandler {
	return &ShortURLHandler{useCase: uc}
}

func (h *ShortURLHandler) CreateShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		reqBody, _ := io.ReadAll(r.Body)
		sourceURL := string(reqBody)
		defer r.Body.Close()

		res, err := h.useCase.CreateShortURL(sourceURL)

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

func (h *ShortURLHandler) FindShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}
		result, err := h.useCase.FindShortURL(r.URL.Path)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.Header().Set("Location", result)
		w.WriteHeader(http.StatusTemporaryRedirect)

	}
}
