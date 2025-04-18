package memory

import (
	"errors"
	"github.com/gururuby/shortener/internal/domain/entity"
)

const (
	recordNotFoundError    = "record not found"
	recordIsNotUniqueError = "record is not unique"
)

type DB struct {
	shortURLs map[string]*entity.ShortURL
}

func New() *DB {
	return &DB{
		shortURLs: make(map[string]*entity.ShortURL),
	}
}

func (db *DB) Find(alias string) (*entity.ShortURL, error) {
	shortURL, ok := db.shortURLs[alias]
	if !ok {
		return nil, errors.New(recordNotFoundError)
	}

	return shortURL, nil
}

func (db *DB) Save(shortURL *entity.ShortURL) (*entity.ShortURL, error) {
	existing, _ := db.Find(shortURL.Alias)
	if existing != nil {
		return nil, errors.New(recordIsNotUniqueError)
	}

	db.shortURLs[shortURL.Alias] = shortURL

	return shortURL, nil
}
