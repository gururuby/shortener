package storages

import (
	"github.com/gururuby/shortener/internal/models"
	"github.com/gururuby/shortener/internal/utils"
)

const (
	maxGenerationAttempts = 10
	aliasLength           = 5
	saveError             = "Cannot save data. Short URL with the same alias already exists"
)

type MemoryStorage struct {
	Data map[string]models.ShortURL
}

func (storage *MemoryStorage) Save(baseURL string, sourceURL string) (result string, ok bool) {
	for i := 1; i < maxGenerationAttempts; i++ {
		alias := utils.GenerateRandomString(aliasLength)
		_, exist := storage.Data[alias]
		if !exist {
			newShortURL := models.ShortURL{Source: sourceURL, Alias: alias}
			storage.Data[alias] = newShortURL
			return newShortURL.AliasURL(baseURL), true
		}
	}

	return saveError, false
}

func (storage *MemoryStorage) Find(alias string) (string, bool) {
	shortURL, ok := storage.Data[alias]

	return shortURL.Source, ok
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Data: make(map[string]models.ShortURL),
	}
}
