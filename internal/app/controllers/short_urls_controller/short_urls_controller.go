package short_urls_controller

import (
	"github.com/gururuby/shortener/internal/app/models/short_url_model"
	"github.com/gururuby/shortener/internal/app/repos/short_urls_repo"
	"io"
	"net/http"
	"strings"
)

var repo = short_urls_repo.ShortUrlsRepo{
	Data: map[string]short_url_model.ShortUrl{},
}

func Create(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		baseUrl, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}

		shortUrl := repo.Create(string(baseUrl))

		w.WriteHeader(http.StatusCreated)
		_, err = io.WriteString(w, shortUrl)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func Show(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		id := strings.TrimPrefix(r.URL.Path, "/")
		baseUrl := repo.Find(id)

		w.Header().Set("Location", baseUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}

}
