package mocks

import "github.com/gururuby/shortener/internal/models"

type MockShortURLsRepo struct {
	Data map[string]models.ShortURL
}

type MockConfig struct {
	ServerAddress string
	BaseURL       string
}

func NewMockConfig() MockConfig {
	return MockConfig{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}
}

func NewMockShortURLsRepo() *MockShortURLsRepo {
	return &MockShortURLsRepo{
		Data: make(map[string]models.ShortURL),
	}
}

func (repo *MockShortURLsRepo) CreateShortURL(baseURL string, source string) string {
	shortURL := models.NewShortURL(source)
	shortURL.Alias = "mock_alias"
	repo.Data[shortURL.Alias] = shortURL

	return shortURL.AliasURL(baseURL)
}

func (repo *MockShortURLsRepo) FindShortURL(alias string) (string, bool) {
	shortURL, ok := repo.Data[alias]

	return shortURL.Source, ok
}
