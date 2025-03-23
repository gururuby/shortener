package repos

import (
	"github.com/gururuby/shortener/internal/models"
	"github.com/gururuby/shortener/internal/storage"
)

type ShortURLsRepo struct {
	Data map[string]models.ShortURL
}

func NewShortURLsRepo() storage.IStorage {
	return &ShortURLsRepo{
		Data: make(map[string]models.ShortURL),
	}
}

func (repo *ShortURLsRepo) CreateShortURL(baseURL string, source string) string {
	shortURL := models.NewShortURL(source)
	repo.Data[shortURL.Alias] = shortURL

	return shortURL.AliasURL(baseURL)
}

func (repo *ShortURLsRepo) FindShortURL(alias string) (string, bool) {
	shortURL, ok := repo.Data[alias]

	return shortURL.Source, ok
}
