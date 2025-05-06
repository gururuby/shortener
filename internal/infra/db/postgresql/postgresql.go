//go:generate mockgen -destination=./mocks/mock.go -package=mocks . Client

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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

const (
	findQuery = "SELECT original_url, uuid FROM urls WHERE urls.alias = $1 LIMIT 1"
	saveQuery = "INSERT INTO urls (alias, original_url) VALUES ($1, $2)"
)

type DBPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Ping(ctx context.Context) error
}

type DB struct {
	ctx  context.Context
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
		ctx:  ctx,
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

func (db *DB) Find(alias string) (*entity.ShortURL, error) {
	shortURL := entity.ShortURL{Alias: alias}
	err := db.pool.QueryRow(db.ctx, findQuery, alias).Scan(&shortURL.SourceURL, &shortURL.UUID)

	if err != nil {
		logger.Log.Error(err.Error())
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return &shortURL, nil
}

func (db *DB) Save(shortURL *entity.ShortURL) (*entity.ShortURL, error) {
	var err error
	var result pgconn.CommandTag

	result, err = db.pool.Exec(db.ctx, saveQuery, shortURL.Alias, shortURL.SourceURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			logger.Log.Error(pgErr.Error())
			if pgErr.Code == "23505" {
				return nil, dbErrors.ErrDBIsNotUnique
			}
			return nil, dbErrors.ErrDBQuery
		}
	}

	logger.Log.Info(result.String())

	return shortURL, nil
}

func (db *DB) Ping() error {
	return db.pool.Ping(context.Background())
}
