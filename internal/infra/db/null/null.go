package db

import (
	"context"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
)

type NullDB struct{}

func New() *NullDB {
	return &NullDB{}
}

func (db *NullDB) FindUser(_ context.Context, _ int) (*userEntity.User, error) {
	return nil, nil
}

func (db *NullDB) FindUserURLs(_ context.Context, _ int) ([]*shortURLEntity.ShortURL, error) {
	return nil, nil
}

func (db *NullDB) SaveUser(_ context.Context) (*userEntity.User, error) {
	return nil, nil
}

func (db *NullDB) FindShortURL(_ context.Context, _ string) (*shortURLEntity.ShortURL, error) {
	return nil, nil
}
func (db *NullDB) findShortURLBySourceURL(_ context.Context, _ string) (*shortURLEntity.ShortURL, error) {
	return nil, nil
}
func (db *NullDB) SaveShortURL(_ context.Context, shortURL *shortURLEntity.ShortURL) (*shortURLEntity.ShortURL, error) {
	return shortURL, nil
}

func (db *NullDB) Ping(_ context.Context) error {
	return nil
}
