/*
Package db provides a database abstraction layer and factory for the URL shortener service.

It offers:
- A unified database interface for different storage backends
- Factory method for creating configured database instances
- Support for multiple database implementations:
  - In-memory (memory)
  - File-based (file)
  - PostgreSQL (postgresql)
  - Null/no-op (default)
*/
package db

import (
	"context"
	"log"

	"github.com/gururuby/shortener/internal/config"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	statsEntity "github.com/gururuby/shortener/internal/domain/entity/stats"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	fileDB "github.com/gururuby/shortener/internal/infra/db/file"
	memoryDB "github.com/gururuby/shortener/internal/infra/db/memory"
	nullDB "github.com/gururuby/shortener/internal/infra/db/null"
	postgresqlDB "github.com/gururuby/shortener/internal/infra/db/postgresql"
)

// DB defines the interface for all database operations in the application.
// Implementations must provide these methods for different storage backends.
type DB interface {
	// FindShortURL retrieves a short URL by its alias
	FindShortURL(ctx context.Context, alias string) (*shortURLEntity.ShortURL, error)

	GetResourcesCounts(ctx context.Context) (*statsEntity.Stats, error)
	// SaveShortURL stores a new short URL
	SaveShortURL(ctx context.Context, shortURL *shortURLEntity.ShortURL) (*shortURLEntity.ShortURL, error)

	// FindUser retrieves a user by ID
	FindUser(ctx context.Context, id int) (*userEntity.User, error)

	// FindUserURLs retrieves all short URLs belonging to a user
	FindUserURLs(ctx context.Context, id int) ([]*shortURLEntity.ShortURL, error)

	// MarkURLAsDeleted marks the specified URLs as deleted for a user
	MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error

	// SaveUser creates and stores a new user
	SaveUser(ctx context.Context) (*userEntity.User, error)

	// Ping checks if the database is available
	Ping(ctx context.Context) error

	// Shutdown allows to gracefully shutdown databases
	Shutdown(context.Context) error
}

// Setup initializes and returns the appropriate database implementation
// based on the configuration. This is the factory method for database instances.
//
// Parameters:
// - ctx: Context for cancellation/timeouts during setup
// - cfg: Application configuration containing database settings
//
// Returns:
// - DB: Initialized database instance
// - error: Any error that occurred during setup
//
// Supported database types:
// - "memory": In-memory database (memoryDB)
// - "file": File-based database (fileDB)
// - "postgresql": PostgreSQL database (postgresqlDB)
// - default: Null/no-op database (nullDB)
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
