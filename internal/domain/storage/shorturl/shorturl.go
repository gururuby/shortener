//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DB

package storage

import (
	"context"
	"errors"
	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/shorturl/errors"
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

type Storage struct {
	gen Generator
	db  DB
}

func Setup(db DB, cfg *config.Config) *Storage {
	return &Storage{gen: generator.New(cfg.App.AliasLength), db: db}
}

func (storage *Storage) FindByAlias(ctx context.Context, alias string) (*entity.ShortURL, error) {
	return storage.db.FindShortURL(ctx, alias)
}

func (storage *Storage) Save(ctx context.Context, sourceURL string) (*entity.ShortURL, error) {
	shortURL := entity.NewShortURL(storage.gen, sourceURL)
	res, err := storage.db.SaveShortURL(ctx, shortURL)
	if err != nil {
		if errors.Is(err, dbErrors.ErrDBIsNotUnique) {
			return res, storageErrors.ErrStorageRecordIsNotUnique
		}
	}
	return res, err
}

func (storage *Storage) IsDBReady(ctx context.Context) error {
	return storage.db.Ping(ctx)
}
