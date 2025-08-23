//go:generate mockgen -destination=./mocks/mock.go -package=mocks . ShortURLDB

/*
Package storage provides data persistence implementations for the application.

It includes:
- Database interfaces and abstractions
- Storage implementations
- Error handling for storage operations
*/
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

// ShortURLDB defines the interface for short URL database operations.
type ShortURLDB interface {
	// FindShortURL retrieves a short URL by its alias.
	// Returns:
	// - *entity.ShortURL: The found short URL
	// - error: Any error that occurred during lookup
	FindShortURL(ctx context.Context, alias string) (*entity.ShortURL, error)

	// SaveShortURL persists a short URL record.
	// Returns:
	// - *entity.ShortURL: The saved short URL
	// - error: Any error that occurred during save
	SaveShortURL(ctx context.Context, shortURL *entity.ShortURL) (*entity.ShortURL, error)

	// Ping checks the database connection health.
	// Returns:
	// - error: Any connection error
	Ping(ctx context.Context) error
}

// Generator defines the interface for generating unique identifiers.
type Generator interface {
	// UUID generates a universally unique identifier.
	UUID() string

	// Alias generates a short, URL-friendly identifier.
	// Returns:
	// - string: The generated alias
	// - error: Any generation error
	Alias() (string, error)
}

// ShortURLStorage implements the storage layer for short URLs.
// It combines database operations with ID generation.
type ShortURLStorage struct {
	gen Generator  // ID generator
	db  ShortURLDB // Database interface
}

// Setup creates and initializes a new ShortURLStorage instance.
// Parameters:
// - db: Database implementation
// - cfg: Application configuration
// Returns:
// - *ShortURLStorage: Initialized storage instance
func Setup(db ShortURLDB, cfg *config.Config) *ShortURLStorage {
	return &ShortURLStorage{gen: generator.New(cfg.App.AliasLength), db: db}
}

// FindShortURL retrieves a short URL by its alias.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - alias: The short URL identifier to look up
// Returns:
// - *entity.ShortURL: The found short URL
// - error: Any error that occurred during lookup
func (s *ShortURLStorage) FindShortURL(ctx context.Context, alias string) (*entity.ShortURL, error) {
	return s.db.FindShortURL(ctx, alias)
}

// SaveShortURL creates and persists a new short URL.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - user: The user creating the short URL (can be nil for anonymous)
// - sourceURL: The original URL to shorten
// Returns:
// - *entity.ShortURL: The created short URL
// - error: Any error that occurred during creation or save
func (s *ShortURLStorage) SaveShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (*entity.ShortURL, error) {
	shortURL, err := entity.NewShortURL(s.gen, user, sourceURL)
	if err != nil {
		return nil, err
	}
	res, err := s.db.SaveShortURL(ctx, shortURL)
	if err != nil {
		if errors.Is(err, dbErrors.ErrDBIsNotUnique) {
			return res, storageErrors.ErrStorageRecordIsNotUnique
		}
	}
	return res, err
}

// IsDBReady checks if the database connection is healthy.
// Parameters:
// - ctx: Context for cancellation and timeouts
// Returns:
// - error: Any connection error
func (s *ShortURLStorage) IsDBReady(ctx context.Context) error {
	return s.db.Ping(ctx)
}
