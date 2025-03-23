package mocks

import "github.com/gururuby/shortener/internal/models"

type MockShortURLsRepo struct {
	Data map[string]models.ShortURL
}

type MockConfig struct {
	ServerAddress string
	PublicAddress string
}

func NewShortURLsRepo() *MockShortURLsRepo {
	return &MockShortURLsRepo{
		Data: make(map[string]models.ShortURL),
	}
}

func (repo *MockShortURLsRepo) CreateShortURL(publicAddress string, BaseURL string) string {
	shortURL := models.NewShortURL(BaseURL)
	shortURL.Alias = "mock_alias"
	repo.Data[shortURL.Alias] = shortURL

	return shortURL.AliasURL(publicAddress)
}

func (repo *MockShortURLsRepo) FindShortURL(alias string) (string, bool) {
	shortURL, ok := repo.Data[alias]

	return shortURL.BaseURL, ok
}
