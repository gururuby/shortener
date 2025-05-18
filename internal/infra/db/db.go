package db

import (
	"context"
	"github.com/gururuby/shortener/internal/config"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	fileDB "github.com/gururuby/shortener/internal/infra/db/file"
	memoryDB "github.com/gururuby/shortener/internal/infra/db/memory"
	nullDB "github.com/gururuby/shortener/internal/infra/db/null"
	postgresqlDB "github.com/gururuby/shortener/internal/infra/db/postgresql"
	"log"
)

type DB interface {
	FindShortURL(ctx context.Context, alias string) (*shortURLEntity.ShortURL, error)
	SaveShortURL(ctx context.Context, shortURL *shortURLEntity.ShortURL) (*shortURLEntity.ShortURL, error)
	FindUser(ctx context.Context, id int) (*userEntity.User, error)
	FindUserURLs(ctx context.Context, id int) ([]*shortURLEntity.ShortURL, error)
	MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error
	SaveUser(ctx context.Context) (*userEntity.User, error)
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
