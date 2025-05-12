package db

import (
	"context"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
)

type NullDB struct{}

func New() *NullDB {
	return &NullDB{}
}

func (db *NullDB) FindShortURL(_ context.Context, _ string) (*entity.ShortURL, error) {
	return nil, nil
}

func (db *NullDB) findBySourceURL(_ context.Context, _ string) (*entity.ShortURL, error) {
	return nil, nil
}

func (db *NullDB) SaveShortURL(_ context.Context, shortURL *entity.ShortURL) (*entity.ShortURL, error) {
	return shortURL, nil
}

func (db *NullDB) Ping(_ context.Context) error {
	return nil
}
