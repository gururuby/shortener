package db

import (
	"errors"
	"github.com/gururuby/shortener/internal/domain/entity"
)

const (
	sourceURLNotFoundError = "source URL not found"
)

type DB struct {
	Data map[string]entity.ShortURL
}

func New() *DB {
	return &DB{
		Data: make(map[string]entity.ShortURL),
	}
}

func (db *DB) Find(alias string) (string, error) {
	res, ok := db.Data[alias]
	if !ok {
		return "", errors.New(sourceURLNotFoundError)
	}

	return res.SourceURL, nil

}

func (db *DB) Save(sourceURL string) (string, error) {
	shortURL := entity.NewShortURL(sourceURL)
	db.Data[shortURL.Alias] = shortURL

	return shortURL.Alias, nil
}
