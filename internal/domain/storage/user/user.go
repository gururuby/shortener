//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DB

/*
Package storage provides data persistence implementations for user-related operations.

It includes:
- Database interfaces for user management
- Storage layer implementations
- Methods for user and URL operations
*/
package storage

import (
	"context"

	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
)

// UserDB defines the interface for user database operations.
type UserDB interface {
	// FindUser retrieves a user by their ID.
	// Returns:
	// - *userEntity.User: The found user
	// - error: If user is not found or database operation fails
	FindUser(ctx context.Context, id int) (*userEntity.User, error)

	// FindUserURLs retrieves all short URLs belonging to a user.
	// Returns:
	// - []*shortURLEntity.ShortURL: List of user's short URLs
	// - error: If database operation fails
	FindUserURLs(ctx context.Context, id int) ([]*shortURLEntity.ShortURL, error)

	// SaveUser creates and persists a new user.
	// Returns:
	// - *userEntity.User: The created user
	// - error: If database operation fails
	SaveUser(ctx context.Context) (*userEntity.User, error)

	// MarkURLAsDeleted soft-deletes the specified URLs for a user.
	// Returns:
	// - error: If database operation fails or URLs don't belong to user
	MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error
}

// UserStorage implements the storage layer for user operations.
// It acts as an intermediary between the domain and database layers.
type UserStorage struct {
	db UserDB // Database interface implementation
}

// Setup creates and initializes a new UserStorage instance.
// Parameters:
// - db: The database implementation to use
// Returns:
// - *UserStorage: Initialized storage instance
func Setup(db UserDB) *UserStorage {
	return &UserStorage{db: db}
}

// FindURLs retrieves all short URLs belonging to a user.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - id: User ID to look up
// Returns:
// - []*shortURLEntity.ShortURL: List of user's short URLs
// - error: If operation fails
func (s *UserStorage) FindURLs(ctx context.Context, id int) ([]*shortURLEntity.ShortURL, error) {
	return s.db.FindUserURLs(ctx, id)
}

// MarkURLAsDeleted marks the specified URLs as deleted for a user.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - userID: Owner of the URLs
// - aliases: List of URL aliases to mark as deleted
// Returns:
// - error: If operation fails or URLs don't belong to user
func (s *UserStorage) MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error {
	return s.db.MarkURLAsDeleted(ctx, userID, aliases)
}

// FindUser retrieves a user by their ID.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - id: User ID to look up
// Returns:
// - *userEntity.User: The found user
// - error: If user is not found or operation fails
func (s *UserStorage) FindUser(ctx context.Context, id int) (*userEntity.User, error) {
	return s.db.FindUser(ctx, id)
}

// SaveUser creates and persists a new user.
// Returns:
// - *userEntity.User: The created user
// - error: If operation fails
func (s *UserStorage) SaveUser(ctx context.Context) (*userEntity.User, error) {
	return s.db.SaveUser(ctx)
}
