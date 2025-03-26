package mocks

import "github.com/gururuby/shortener/internal/models"

type MockStorage struct {
	Data map[string]models.ShortURL
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		Data: make(map[string]models.ShortURL),
	}
}

func (storage *MockStorage) Save(baseURL string, source string) (string, bool) {
	shortURL := models.NewShortURL(source)
	shortURL.Alias = "mock_alias"
	storage.Data[shortURL.Alias] = shortURL

	return shortURL.AliasURL(baseURL), true
}

func (storage *MockStorage) Find(alias string) (string, bool) {
	shortURL, ok := storage.Data[alias]

	return shortURL.Source, ok
}
