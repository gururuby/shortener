package null

import (
	"context"
	"github.com/gururuby/shortener/internal/domain/entity"
)

type DB struct{}

func New() *DB {
	return &DB{}
}

func (db *DB) Find(_ context.Context, _ string) (*entity.ShortURL, error) {
	return nil, nil
}

func (db *DB) findBySourceURL(_ context.Context, _ string) (*entity.ShortURL, error) {
	return nil, nil
}

func (db *DB) Save(_ context.Context, shortURL *entity.ShortURL) (*entity.ShortURL, error) {
	return shortURL, nil
}

func (db *DB) Ping(_ context.Context) error {
	return nil
}
