package memory

import (
	"errors"
	"github.com/gururuby/shortener/internal/domain/entity"
)

const (
	sourceURLNotFoundError = "source URL not found"
)

type ShortURLDB struct {
	Data map[string]entity.ShortURL
}

func NewShortURLDB() *ShortURLDB {
	return &ShortURLDB{
		Data: make(map[string]entity.ShortURL),
	}
}

func (db *ShortURLDB) Find(alias string) (string, error) {
	res, ok := db.Data[alias]
	if !ok {
		return "", errors.New(sourceURLNotFoundError)
	}

	return res.SourceURL, nil

}

func (db *ShortURLDB) Save(sourceURL string) (string, error) {
	shortURL := entity.NewShortURL(sourceURL)
	db.Data[shortURL.Alias] = shortURL

	return shortURL.Alias, nil
}
