package repos

import (
	"github.com/gururuby/shortener/internal/models"
)

type ShortURLsRepo struct {
	Data map[string]models.ShortURL
}

func NewShortURLsRepo() *ShortURLsRepo {
	return &ShortURLsRepo{
		Data: make(map[string]models.ShortURL),
	}
}

func (repo *ShortURLsRepo) CreateShortURL(serverBaseURL string, BaseURL string) string {
	shortURL := models.NewShortURL(BaseURL)
	repo.Data[shortURL.Alias] = shortURL

	return shortURL.AliasURL(serverBaseURL)
}

func (repo *ShortURLsRepo) FindShortURL(alias string) (string, bool) {
	shortURL, ok := repo.Data[alias]

	return shortURL.BaseURL, ok
}
