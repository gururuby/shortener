/*
Package db provides a null implementation of the database interface.

The NullDB is a no-op database implementation that:
- Implements all required database methods
- Returns empty results for all queries
- Never returns errors
- Useful for testing or when database functionality isn't needed
*/
package db

import (
	"context"

	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
)

// NullDB is a no-op database implementation that satisfies the database interface
// but doesn't actually store any data. All operations succeed but have no effect.
type NullDB struct{}

// New creates a new NullDB instance.
// Returns:
// - *NullDB: An initialized null database instance
func New() *NullDB {
	return &NullDB{}
}

// FindUser is a no-op implementation that always returns nil.
// Parameters:
// - ctx: Context (ignored)
// - id: User ID (ignored)
// Returns:
// - *userEntity.User: Always nil
// - error: Always nil
func (db *NullDB) FindUser(_ context.Context, _ int) (*userEntity.User, error) {
	return nil, nil
}

// FindUserURLs is a no-op implementation that always returns nil.
// Parameters:
// - ctx: Context (ignored)
// - userID: User ID (ignored)
// Returns:
// - []*shortURLEntity.ShortURL: Always nil
// - error: Always nil
func (db *NullDB) FindUserURLs(_ context.Context, _ int) ([]*shortURLEntity.ShortURL, error) {
	return nil, nil
}

// SaveUser is a no-op implementation that always returns nil.
// Parameters:
// - ctx: Context (ignored)
// Returns:
// - *userEntity.User: Always nil
// - error: Always nil
func (db *NullDB) SaveUser(_ context.Context) (*userEntity.User, error) {
	return nil, nil
}

// FindShortURL is a no-op implementation that always returns nil.
// Parameters:
// - ctx: Context (ignored)
// - alias: Short URL alias (ignored)
// Returns:
// - *shortURLEntity.ShortURL: Always nil
// - error: Always nil
func (db *NullDB) FindShortURL(_ context.Context, _ string) (*shortURLEntity.ShortURL, error) {
	return nil, nil
}

// findShortURLBySourceURL is a no-op implementation that always returns nil.
// Parameters:
// - ctx: Context (ignored)
// - sourceURL: Original URL (ignored)
// Returns:
// - *shortURLEntity.ShortURL: Always nil
// - error: Always nil
func (db *NullDB) findShortURLBySourceURL(_ context.Context, _ string) (*shortURLEntity.ShortURL, error) {
	return nil, nil
}

// SaveShortURL is a no-op implementation that returns the input unchanged.
// Parameters:
// - ctx: Context (ignored)
// - shortURL: URL to "save"
// Returns:
// - *shortURLEntity.ShortURL: Returns the input shortURL
// - error: Always nil
func (db *NullDB) SaveShortURL(_ context.Context, shortURL *shortURLEntity.ShortURL) (*shortURLEntity.ShortURL, error) {
	return shortURL, nil
}

// MarkURLAsDeleted is a no-op implementation that always succeeds.
// Parameters:
// - ctx: Context (ignored)
// - userID: User ID (ignored)
// - aliases: URLs to delete (ignored)
// Returns:
// - error: Always nil
func (db *NullDB) MarkURLAsDeleted(_ context.Context, _ int, _ []string) error {
	return nil
}

// Ping is a no-op implementation that always succeeds.
// Parameters:
// - ctx: Context (ignored)
// Returns:
// - error: Always nil
func (db *NullDB) Ping(_ context.Context) error {
	return nil
}

// Shutdown is a no-op implementation that always succeeds.
// Parameters:
// - ctx: Context (ignored)
// Returns:
// - error: Always nil
func (db *NullDB) Shutdown(_ context.Context) error { return nil }
