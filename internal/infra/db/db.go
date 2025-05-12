package db

import (
	"context"
	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	fileDB "github.com/gururuby/shortener/internal/infra/db/file"
	memoryDB "github.com/gururuby/shortener/internal/infra/db/memory"
	nullDB "github.com/gururuby/shortener/internal/infra/db/null"
	postgresqlDB "github.com/gururuby/shortener/internal/infra/db/postgresql"
	"log"
)

type DB interface {
	FindShortURL(ctx context.Context, alias string) (*entity.ShortURL, error)
	SaveShortURL(ctx context.Context, shortURL *entity.ShortURL) (*entity.ShortURL, error)
	Ping(ctx context.Context) error
}

func Setup(ctx context.Context, cfg *config.Config) (db DB, err error) {
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
