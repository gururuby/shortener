//go:generate mockgen -destination=./mocks/mock.go -package=mocks . PGDBPool

package db

import (
	"context"
	"embed"
	"errors"
	"github.com/gururuby/shortener/internal/config"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/pkg/retry"
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
	findShortURLQuery            = `SELECT original_url, uuid FROM urls WHERE urls.alias = $1`
	findUserQuery                = `SELECT id FROM users WHERE users.id = $1`
	findUserURLsQuery            = `SELECT alias, original_url FROM urls WHERE urls.user_id = $1`
	findShortURLBySourceURLQuery = `SELECT alias FROM urls WHERE urls.original_url = $1`
	saveShortURLQuery            = `INSERT INTO urls (alias, original_url) VALUES ($1, $2)`
	saveShortURLQueryWithUser    = `INSERT INTO urls (alias, original_url, user_id) VALUES ($1, $2, $3)`
	saveUserQuery                = `INSERT INTO users DEFAULT VALUES RETURNING id`
)

type PGDBPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Ping(ctx context.Context) error
}

type PGDB struct {
	pool PGDBPool
}

func New(ctx context.Context, cfg *config.Config) (*PGDB, error) {
	var (
		err  error
		pool *pgxpool.Pool
	)

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

	return &PGDB{
		pool: pool,
	}, nil
}

func newDBPool(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	var (
		pool   *pgxpool.Pool
		err    error
		cancel context.CancelFunc
	)

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

func (db *PGDB) FindUser(ctx context.Context, id int) (*userEntity.User, error) {
	user := userEntity.User{ID: id}
	err := db.pool.QueryRow(ctx, findUserQuery, id).Scan(&user.ID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dbErrors.ErrDBRecordNotFound
		} else {
			logger.Log.Error(err.Error())
			return nil, dbErrors.ErrDBQuery
		}
	}

	return &user, nil
}

func (db *PGDB) FindUserURLs(ctx context.Context, userID int) ([]*shortURLEntity.ShortURL, error) {
	var (
		alias       string
		originalURL string
		urls        []*shortURLEntity.ShortURL
	)

	rows, err := db.pool.Query(ctx, findUserURLsQuery, userID)
	if err != nil {
		logger.Log.Error(err.Error())
		return nil, dbErrors.ErrDBQuery
	}

	_, err = pgx.ForEachRow(rows, []any{&alias, &originalURL}, func() error {
		urls = append(urls, &shortURLEntity.ShortURL{Alias: alias, SourceURL: originalURL})
		return nil
	})

	if err != nil {
		logger.Log.Error(err.Error())
		return nil, dbErrors.ErrDBQuery
	}

	return urls, nil
}

func (db *PGDB) SaveUser(ctx context.Context) (*userEntity.User, error) {
	user := userEntity.User{}
	err := db.pool.QueryRow(ctx, saveUserQuery).Scan(&user.ID)
	if err != nil {
		logger.Log.Error(err.Error())
		return nil, dbErrors.ErrDBQuery
	}

	return &user, err
}

func (db *PGDB) FindShortURL(ctx context.Context, alias string) (*shortURLEntity.ShortURL, error) {
	shortURL := shortURLEntity.ShortURL{Alias: alias}
	err := db.pool.QueryRow(ctx, findShortURLQuery, alias).Scan(&shortURL.SourceURL, &shortURL.UUID)

	if err != nil {
		logger.Log.Error(err.Error())
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return &shortURL, nil
}

func (db *PGDB) SaveShortURL(ctx context.Context, shortURL *shortURLEntity.ShortURL) (*shortURLEntity.ShortURL, error) {
	var (
		err              error
		pgErr            *pgconn.PgError
		existingShortURL *shortURLEntity.ShortURL
	)

	if existingShortURL, err = db.findShortURLBySourceURL(ctx, shortURL.SourceURL); err == nil {
		return existingShortURL, dbErrors.ErrDBIsNotUnique
	}

	if errors.Is(err, dbErrors.ErrDBRecordNotFound) {
		if shortURL.UserID == 0 {
			if _, err = db.pool.Exec(ctx, saveShortURLQuery, shortURL.Alias, shortURL.SourceURL); err == nil {
				return shortURL, nil
			}
		} else {
			if _, err = db.pool.Exec(ctx, saveShortURLQueryWithUser, shortURL.Alias, shortURL.SourceURL, shortURL.UserID); err == nil {
				return shortURL, nil
			}
		}

		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return shortURL, dbErrors.ErrDBIsNotUnique
			}
			logger.Log.Error(err.Error())
			return nil, dbErrors.ErrDBQuery
		}
	}

	return nil, err
}

func (db *PGDB) findShortURLBySourceURL(ctx context.Context, sourceURL string) (*shortURLEntity.ShortURL, error) {
	shortURL := shortURLEntity.ShortURL{SourceURL: sourceURL}
	err := db.pool.QueryRow(ctx, findShortURLBySourceURLQuery, sourceURL).Scan(&shortURL.Alias)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dbErrors.ErrDBRecordNotFound
		} else {
			logger.Log.Error(err.Error())
			return nil, dbErrors.ErrDBQuery
		}

	}

	return &shortURL, nil
}

func (db *PGDB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}
