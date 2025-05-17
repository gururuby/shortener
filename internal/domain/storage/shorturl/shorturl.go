//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DB

package storage

import (
	"context"
	"errors"
	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/errors"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"github.com/gururuby/shortener/pkg/generator"
)

type DB interface {
	FindShortURL(ctx context.Context, alias string) (*entity.ShortURL, error)
	SaveShortURL(ctx context.Context, shortURL *entity.ShortURL) (*entity.ShortURL, error)
	Ping(ctx context.Context) error
}

type Generator interface {
	UUID() string
	Alias() string
}

type ShortURLStorage struct {
	gen Generator
	db  DB
}

func Setup(db DB, cfg *config.Config) *ShortURLStorage {
	return &ShortURLStorage{gen: generator.New(cfg.App.AliasLength), db: db}
}

func (s *ShortURLStorage) FindShortURL(ctx context.Context, alias string) (*entity.ShortURL, error) {
	return s.db.FindShortURL(ctx, alias)
}

func (s *ShortURLStorage) SaveShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (*entity.ShortURL, error) {
	shortURL := entity.NewShortURL(s.gen, user, sourceURL)
	res, err := s.db.SaveShortURL(ctx, shortURL)
	if err != nil {
		if errors.Is(err, dbErrors.ErrDBIsNotUnique) {
			return res, storageErrors.ErrStorageRecordIsNotUnique
		}
	}
	return res, err
}

func (s *ShortURLStorage) IsDBReady(ctx context.Context) error {
	return s.db.Ping(ctx)
}
