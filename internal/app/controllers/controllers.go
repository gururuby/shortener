package controllers

import (
	"github.com/gururuby/shortener/internal/app/repos"
	"io"
	"net/http"
	"strings"
)

var repo = repos.NewShortURLsRepo()

func ShortURLCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		BaseURL, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}

		ShortURL := repo.CreateShortURL(string(BaseURL))

		w.WriteHeader(http.StatusCreated)
		_, err = io.WriteString(w, ShortURL)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func ShortURLShow(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		id := strings.TrimPrefix(r.URL.Path, "/")
		BaseURL := repo.FindShortURL(id)

		w.Header().Set("Location", BaseURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}

}
