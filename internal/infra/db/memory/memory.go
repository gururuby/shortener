package memory

import (
	"github.com/gururuby/shortener/internal/domain/entity"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
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
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return shortURL, nil
}

func (db *DB) Save(shortURL *entity.ShortURL) (*entity.ShortURL, error) {
	existing, _ := db.Find(shortURL.Alias)
	if existing != nil {
		return nil, dbErrors.ErrDBIsNotUnique
	}

	db.shortURLs[shortURL.Alias] = shortURL

	return shortURL, nil
}

func (db *DB) Ping() error {
	return nil
}
