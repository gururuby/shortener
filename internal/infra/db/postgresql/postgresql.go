//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DBPool

package postgresql

import (
	"context"
	"embed"
	"errors"
	"github.com/gururuby/shortener/config"
	"github.com/gururuby/shortener/internal/domain/entity"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/internal/infra/utils/retry"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

const (
	findQuery            = `SELECT original_url, uuid FROM urls WHERE urls.alias = $1`
	findBySourceURLQuery = `SELECT alias FROM urls WHERE urls.original_url = $1`
	saveQuery            = `INSERT INTO urls (alias, original_url) VALUES ($1, $2)`
)

type DBPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Ping(ctx context.Context) error
}

type DB struct {
	pool DBPool
}

func New(ctx context.Context, cfg *config.Config) (*DB, error) {
	var err error
	var pool *pgxpool.Pool

	goose.SetBaseFS(migrations)
	if err = goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	pool, err = newDBPool(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	dbFromPool := stdlib.OpenDBFromPool(pool)
	if err = goose.Up(dbFromPool, "migrations"); err != nil {
		return nil, err
	}

	if err = dbFromPool.Close(); err != nil {
		return nil, err
	}

	return &DB{
		pool: pool,
	}, nil
}

func newDBPool(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error
	var cancel context.CancelFunc

	err = utils.Retry(func() error {
		ctx, cancel = context.WithTimeout(ctx, cfg.ConnTryDelay)
		defer cancel()

		pool, err = pgxpool.New(ctx, cfg.DSN)

		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}

		return nil

	}, cfg.ConnTryTimes, cfg.ConnTryDelay)

	return pool, err
}

func (db *DB) Find(ctx context.Context, alias string) (*entity.ShortURL, error) {
	shortURL := entity.ShortURL{Alias: alias}
	err := db.pool.QueryRow(ctx, findQuery, alias).Scan(&shortURL.SourceURL, &shortURL.UUID)

	if err != nil {
		logger.Log.Error(err.Error())
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return &shortURL, nil
}

func (db *DB) Save(ctx context.Context, shortURL *entity.ShortURL) (*entity.ShortURL, error) {
	var (
		err              error
		pgErr            *pgconn.PgError
		existingShortURL *entity.ShortURL
	)

	if existingShortURL, err = db.findBySourceURL(ctx, shortURL.SourceURL); err == nil {
		return existingShortURL, dbErrors.ErrDBIsNotUnique
	}

	if errors.Is(err, dbErrors.ErrDBRecordNotFound) {
		if _, err = db.pool.Exec(ctx, saveQuery, shortURL.Alias, shortURL.SourceURL); err == nil {
			return shortURL, nil
		}

		logger.Log.Error(err.Error())
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return shortURL, dbErrors.ErrDBIsNotUnique
			}
			return nil, dbErrors.ErrDBQuery
		}
	}

	return nil, err
}

func (db *DB) findBySourceURL(ctx context.Context, sourceURL string) (*entity.ShortURL, error) {
	shortURL := entity.ShortURL{SourceURL: sourceURL}
	err := db.pool.QueryRow(ctx, findBySourceURLQuery, sourceURL).Scan(&shortURL.Alias)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dbErrors.ErrDBRecordNotFound
		} else {
			return nil, dbErrors.ErrDBQuery
		}

	}

	return &shortURL, nil
}

func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}
