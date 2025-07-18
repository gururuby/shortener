/*
Package db implements an in-memory database for the URL shortener service.

It provides:
- Fast in-memory storage for users and short URLs
- Basic CRUD operations without persistence
- Simple interface matching the database requirements
- Thread-unsafe operations (caller must handle synchronization)
*/
package db

import (
	"context"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
)

// MemoryDB represents an in-memory database implementation.
// It stores data in maps without persistence to disk.
type MemoryDB struct {
	shortURLs map[string]*shortURLEntity.ShortURL // Map of short URL aliases to entities
	users     map[int]*userEntity.User            // Map of user IDs to user entities
}

// New creates and initializes a new MemoryDB instance.
// Returns:
// - *MemoryDB: Empty initialized in-memory database
func New() *MemoryDB {
	return &MemoryDB{
		shortURLs: make(map[string]*shortURLEntity.ShortURL),
		users:     make(map[int]*userEntity.User),
	}
}

// FindUser retrieves a user by ID from memory.
// Parameters:
// - ctx: Context for cancellation/timeouts (unused)
// - id: User ID to find
// Returns:
// - *userEntity.User: Found user entity
// - error: dbErrors.ErrDBRecordNotFound if user doesn't exist
func (db *MemoryDB) FindUser(_ context.Context, id int) (*userEntity.User, error) {
	user, ok := db.users[id]
	if !ok {
		return nil, dbErrors.ErrDBRecordNotFound
	}
	return user, nil
}

// FindUserURLs retrieves all short URLs belonging to a user.
// Parameters:
// - ctx: Context for cancellation/timeouts (unused)
// - userID: Owner's user ID
// Returns:
// - []*shortURLEntity.ShortURL: List of user's URLs (empty slice if none)
// - error: Always nil
func (db *MemoryDB) FindUserURLs(_ context.Context, userID int) ([]*shortURLEntity.ShortURL, error) {
	var urls []*shortURLEntity.ShortURL

	for _, url := range db.shortURLs {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

// SaveUser creates and stores a new user in memory.
// Parameters:
// - ctx: Context for cancellation/timeouts (unused)
// Returns:
// - *userEntity.User: Created user with auto-incremented ID
// - error: Always nil
func (db *MemoryDB) SaveUser(_ context.Context) (*userEntity.User, error) {
	id := len(db.users) + 1
	user := &userEntity.User{ID: id}
	db.users[id] = user
	return user, nil
}

// FindShortURL retrieves a short URL by its alias.
// Parameters:
// - ctx: Context for cancellation/timeouts (unused)
// - alias: Short URL identifier
// Returns:
// - *shortURLEntity.ShortURL: Found short URL entity
// - error: dbErrors.ErrDBRecordNotFound if alias doesn't exist
func (db *MemoryDB) FindShortURL(_ context.Context, alias string) (*shortURLEntity.ShortURL, error) {
	shortURL, ok := db.shortURLs[alias]
	if !ok {
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return shortURL, nil
}

// MarkURLAsDeleted marks URLs as deleted (not implemented).
// Parameters:
// - ctx: Context for cancellation/timeouts (unused)
// - userID: Owner's user ID
// - aliases: URLs to mark as deleted
// Returns:
// - error: Always nil (not implemented)
func (db *MemoryDB) MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error {
	return nil
}

// findShortURLBySourceURL looks up a short URL by its original URL.
// Parameters:
// - ctx: Context for cancellation/timeouts (unused)
// - sourceURL: Original long URL
// Returns:
// - *shortURLEntity.ShortURL: Found short URL
// - error: dbErrors.ErrDBRecordNotFound if URL doesn't exist
func (db *MemoryDB) findShortURLBySourceURL(_ context.Context, sourceURL string) (*shortURLEntity.ShortURL, error) {
	var (
		shortURL  *shortURLEntity.ShortURL
		noRecords = true
	)

	for _, url := range db.shortURLs {
		if url.SourceURL == sourceURL {
			shortURL = url
			noRecords = false
			break
		}
	}

	if noRecords {
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return shortURL, nil
}

// SaveShortURL stores a new short URL in memory.
// Parameters:
// - ctx: Context for cancellation/timeouts (unused)
// - shortURL: URL entity to save
// Returns:
// - *shortURLEntity.ShortURL: Saved URL entity
// - error: dbErrors.ErrDBIsNotUnique if URL already exists
func (db *MemoryDB) SaveShortURL(ctx context.Context, shortURL *shortURLEntity.ShortURL) (*shortURLEntity.ShortURL, error) {
	existRecord, _ := db.findShortURLBySourceURL(ctx, shortURL.SourceURL)
	if existRecord != nil {
		return existRecord, dbErrors.ErrDBIsNotUnique
	}

	db.shortURLs[shortURL.Alias] = shortURL
	return shortURL, nil
}

// Ping checks if the database is available (always succeeds for in-memory).
// Parameters:
// - ctx: Context for cancellation/timeouts (unused)
// Returns:
// - error: Always nil
func (db *MemoryDB) Ping(_ context.Context) error {
	return nil
}
