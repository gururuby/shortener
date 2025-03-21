package repos

import (
	"github.com/gururuby/shortener/internal/app/models"
)

type ShortURLsRepo struct {
	Data map[string]models.ShortURL
}

func NewShortURLsRepo() *ShortURLsRepo {
	return &ShortURLsRepo{
		Data: make(map[string]models.ShortURL),
	}
}

func (repo *ShortURLsRepo) CreateShortURL(BaseURL string) string {
	shortURL := models.NewShortURL(BaseURL)
	repo.Data[shortURL.Alias] = shortURL

	return shortURL.AliasURL()
}

func (repo *ShortURLsRepo) FindShortURL(alias string) string {
	shortURL := repo.Data[alias]

	return shortURL.BaseURL
}
