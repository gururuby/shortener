//go:generate mockgen -destination=./mocks/mock.go -package=mocks . PGDBPool

/*
Package db implements a PostgreSQL database backend for the URL shortener service.

It provides:
- Persistent storage using PostgreSQL
- Database migrations using Goose
- Connection pooling for performance
- Comprehensive error handling
- Support for all required database operations
*/
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
	findShortURLQuery            = `SELECT original_url, uuid, is_deleted FROM urls WHERE urls.alias = $1`
	findUserQuery                = `SELECT id FROM users WHERE users.id = $1`
	findUserURLsQuery            = `SELECT alias, original_url FROM urls WHERE urls.user_id = $1`
	findShortURLBySourceURLQuery = `SELECT alias FROM urls WHERE urls.original_url = $1`
	saveShortURLQuery            = `INSERT INTO urls (alias, original_url) VALUES ($1, $2)`
	saveShortURLQueryWithUser    = `INSERT INTO urls (alias, original_url, user_id) VALUES ($1, $2, $3)`
	saveUserQuery                = `INSERT INTO users DEFAULT VALUES RETURNING id`
	markURLsAsDeletedQuery       = "UPDATE urls SET is_deleted = true WHERE user_id = $1 AND alias = ANY($2)"
)

// PGDBPool defines the interface for PostgreSQL database operations.
// This interface allows for mocking and testing database interactions.
type PGDBPool interface {
	// Exec executes a SQL command and returns the command tag
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	// Query executes a SQL query and returns the rows
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	// QueryRow executes a SQL query expected to return at most one row
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	// Ping checks if the database is available
	Ping(ctx context.Context) error
}

// PGDB implements the database interface using PostgreSQL as the backend.
type PGDB struct {
	pool PGDBPool // Connection pool for database operations
}

// New creates and initializes a new PGDB instance.
// It establishes a connection pool and runs database migrations.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - cfg: Database configuration
// Returns:
// - *PGDB: Initialized database instance
// - error: If connection or migration fails
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

// newDBPool creates a new PostgreSQL connection pool with retry logic.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - cfg: Database configuration
// Returns:
// - *pgxpool.Pool: Connection pool
// - error: If connection fails after retries
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

// FindUser retrieves a user by ID from the database.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - id: User ID to find
// Returns:
// - *userEntity.User: Found user
// - error: dbErrors.ErrDBRecordNotFound if user doesn't exist
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

// FindUserURLs retrieves all short URLs belonging to a user.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - userID: Owner's user ID
// Returns:
// - []*shortURLEntity.ShortURL: List of user's URLs
// - error: If query fails
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

// SaveUser creates a new user in the database.
// Parameters:
// - ctx: Context for cancellation/timeouts
// Returns:
// - *userEntity.User: Created user with ID
// - error: If insert fails
func (db *PGDB) SaveUser(ctx context.Context) (*userEntity.User, error) {
	user := userEntity.User{}
	err := db.pool.QueryRow(ctx, saveUserQuery).Scan(&user.ID)
	if err != nil {
		logger.Log.Error(err.Error())
		return nil, dbErrors.ErrDBQuery
	}

	return &user, nil
}

// FindShortURL retrieves a short URL by its alias.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - alias: Short URL identifier
// Returns:
// - *shortURLEntity.ShortURL: Found short URL
// - error: If URL doesn't exist or query fails
func (db *PGDB) FindShortURL(ctx context.Context, alias string) (*shortURLEntity.ShortURL, error) {
	shortURL := shortURLEntity.ShortURL{Alias: alias}
	err := db.pool.QueryRow(ctx, findShortURLQuery, alias).Scan(&shortURL.SourceURL, &shortURL.UUID, &shortURL.IsDeleted)

	if err != nil {
		logger.Log.Error(err.Error())
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return &shortURL, nil
}

// SaveShortURL stores a new short URL in the database.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - shortURL: URL to save
// Returns:
// - *shortURLEntity.ShortURL: Saved URL
// - error: If URL already exists or insert fails
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

// MarkURLAsDeleted marks the specified URLs as deleted for a user.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - userID: Owner's user ID
// - aliases: URLs to mark as deleted
// Returns:
// - error: If update fails
func (db *PGDB) MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error {
	_, err := db.pool.Exec(ctx, markURLsAsDeletedQuery, userID, aliases)
	return err
}

// findShortURLBySourceURL looks up a short URL by its original URL.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - sourceURL: Original long URL
// Returns:
// - *shortURLEntity.ShortURL: Found short URL
// - error: If URL doesn't exist or query fails
func (db *PGDB) findShortURLBySourceURL(ctx context.Context, sourceURL string) (*shortURLEntity.ShortURL, error) {
	shortURL := shortURLEntity.ShortURL{SourceURL: sourceURL}
	err := db.pool.QueryRow(ctx, findShortURLBySourceURLQuery, sourceURL).Scan(&shortURL.Alias)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dbErrors.ErrDBRecordNotFound
		}

		logger.Log.Error(err.Error())
		return nil, dbErrors.ErrDBQuery
	}

	return &shortURL, nil
}

// Ping checks if the database is available.
// Parameters:
// - ctx: Context for cancellation/timeouts
// Returns:
// - error: If database is unreachable
func (db *PGDB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}
