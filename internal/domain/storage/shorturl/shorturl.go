//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DB

package storage

import (
	"context"
	"errors"
	"github.com/gururuby/shortener/config"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/shorturl/errors"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	fileDB "github.com/gururuby/shortener/internal/infra/db/file"
	memoryDB "github.com/gururuby/shortener/internal/infra/db/memory"
	nullDB "github.com/gururuby/shortener/internal/infra/db/null"
	postgresqlDB "github.com/gururuby/shortener/internal/infra/db/postgresql"
	"github.com/gururuby/shortener/internal/infra/utils/generator"
	"log"
)

type DB interface {
	Find(string) (*entity.ShortURL, error)
	Save(*entity.ShortURL) (*entity.ShortURL, error)
	Ping() error
	Truncate()
}

type Generator interface {
	UUID() string
	Alias() string
}

type Storage struct {
	gen Generator
	db  DB
}

func Setup(ctx context.Context, cfg *config.Config) (*Storage, error) {
	var (
		db  DB
		err error
		gen = generator.New(cfg.App.AliasLength)
	)

	db, err = setupDB(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return New(gen, db), nil
}

func New(gen Generator, db DB) *Storage {
	return &Storage{gen: gen, db: db}
}

func (storage *Storage) FindByAlias(alias string) (*entity.ShortURL, error) {
	return storage.db.Find(alias)
}

func (storage *Storage) Save(sourceURL string) (*entity.ShortURL, error) {
	shortURL := entity.NewShortURL(storage.gen, sourceURL)
	res, err := storage.db.Save(shortURL)
	if err != nil {
		if errors.Is(err, dbErrors.ErrDBIsNotUnique) {
			return res, storageErrors.ErrStorageRecordIsNotUnique
		}
	}
	return res, err
}

func (storage *Storage) IsDBReady() error {
	return storage.db.Ping()
}

func (storage *Storage) Clear() {
	storage.db.Truncate()
}

func setupDB(ctx context.Context, cfg *config.Config) (db DB, err error) {
	switch cfg.Database.Type {
	case "memory":
		db = memoryDB.New()
	case "file":
		if db, err = fileDB.New(cfg.FileStorage.Path); err != nil {
			log.Fatalf("cannot setup file DB: %s", err)
		}
	case "postgresql":
		if db, err = postgresqlDB.New(ctx, cfg); err != nil {
			log.Fatalf("cannot setup postgresql DB: %s", err)
		}
	default:
		db = nullDB.New()
	}
	return
}
