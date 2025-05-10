package null

import (
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
)

type DB struct{}

func New() *DB {
	return &DB{}
}

func (db *DB) Find(_ string) (*entity.ShortURL, error) {
	return nil, nil
}

func (db *DB) findBySourceURL(_ string) (*entity.ShortURL, error) {
	return nil, nil
}

func (db *DB) Save(shortURL *entity.ShortURL) (*entity.ShortURL, error) {
	return shortURL, nil
}

func (db *DB) Ping() error {
	return nil
}

func (db *DB) Truncate() {
}
