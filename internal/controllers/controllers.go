package controllers

import (
	"github.com/gururuby/shortener/internal/storage"
	"io"
	"net/http"
	"strings"
)

func ShortURLCreate(baseURL string, storage storage.IStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			input, _ := io.ReadAll(r.Body)
			source := string(input)

			if source == "" {
				http.Error(w, "Empty base URL, please specify URL", http.StatusUnprocessableEntity)
			} else {
				ShortURL := storage.CreateShortURL(baseURL, source)

				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusCreated)
				_, err := io.WriteString(w, ShortURL)
				if err != nil {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
				}
			}
		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	}
}

func ShortURLShow(storage storage.IStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			alias := strings.TrimPrefix(r.URL.Path, "/")

			if alias == "" {
				http.Error(w, "Empty alias, please specify alias", http.StatusUnprocessableEntity)
			} else {
				baseURL, ok := storage.FindShortURL(alias)

				if !ok {
					http.Error(w, "URL was not found", http.StatusNotFound)
				}

				w.Header().Set("Location", baseURL)
				w.WriteHeader(http.StatusTemporaryRedirect)
			}

		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	}

}
