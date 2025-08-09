//go:generate mockgen -destination=./mocks/mock.go -package=mocks . UserStorage,Authenticator

/*
Package usecase implements the business logic for user management operations.

It provides:
- User authentication and registration
- User URL management
- JWT token handling
- Error handling specific to user operations
*/
package usecase

import (
	"context"
	"errors"

	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/user/errors"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"github.com/gururuby/shortener/internal/infra/logger"
)

// UserStorage defines the interface for user persistence operations.
type UserStorage interface {
	// FindUser retrieves a user by ID.
	// Returns:
	// - *userEntity.User: The found user
	// - error: If user is not found or database operation fails
	FindUser(ctx context.Context, userID int) (*userEntity.User, error)

	// FindURLs retrieves all short URLs belonging to a user.
	// Returns:
	// - []*shortURLEntity.ShortURL: List of user's short URLs
	// - error: If database operation fails
	FindURLs(ctx context.Context, userID int) ([]*shortURLEntity.ShortURL, error)

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

// Authenticator defines the interface for user authentication operations.
type Authenticator interface {
	// SignUserID generates a JWT token for the given user ID.
	// Returns:
	// - string: The generated token
	// - error: If token generation fails
	SignUserID(userID int) (string, error)

	// ReadUserID extracts the user ID from a JWT token.
	// Returns:
	// - int: The user ID from the token
	// - error: If token is invalid or expired
	ReadUserID(tokenString string) (int, error)
}

// UserUseCase implements the business logic for user management.
type UserUseCase struct {
	auth    Authenticator // JWT authentication service
	storage UserStorage   // User persistence layer
	baseURL string        // Base URL for shortened links
}

// UserShortURL represents a shortened URL with its original URL.
type UserShortURL struct {
	ShortURL    string `json:"short_url"`    // The shortened URL
	OriginalURL string `json:"original_url"` // The original long URL
}

// NewUserUseCase creates a new instance of UserUseCase.
// Parameters:
// - auth: JWT authentication service
// - storage: User persistence layer
// - baseURL: Base URL for shortened links
// Returns:
// - *UserUseCase: Initialized user use case
func NewUserUseCase(auth Authenticator, storage UserStorage, baseURL string) *UserUseCase {
	return &UserUseCase{
		auth:    auth,
		storage: storage,
		baseURL: baseURL,
	}
}

// Authenticate verifies a user's JWT token and retrieves their information.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - token: JWT token to authenticate
// Returns:
// - *userEntity.User: Authenticated user with token
// - error: Specific authentication errors
func (u *UserUseCase) Authenticate(ctx context.Context, token string) (*userEntity.User, error) {
	var (
		userID int
		user   *userEntity.User
		err    error
	)

	if userID, err = u.auth.ReadUserID(token); err != nil {
		return nil, ucErrors.ErrUserCannotAuthenticate
	}

	if user, err = u.storage.FindUser(ctx, userID); err != nil {
		return nil, ucErrors.ErrUserNotFound
	}

	user.AuthToken = token
	return user, nil
}

// Register creates a new user account and generates an authentication token.
// Parameters:
// - ctx: Context for cancellation and timeouts
// Returns:
// - *userEntity.User: Newly created user with auth token
// - error: Specific registration errors
func (u *UserUseCase) Register(ctx context.Context) (*userEntity.User, error) {
	var (
		user  *userEntity.User
		token string
		err   error
	)

	if user, err = u.storage.SaveUser(ctx); err != nil {
		return nil, ucErrors.ErrUserCannotRegister
	}

	if token, err = u.auth.SignUserID(user.ID); err != nil {
		return nil, ucErrors.ErrUserCannotRegister
	}

	user.AuthToken = token

	return user, nil
}

// SaveUser persists a new user record.
// Parameters:
// - ctx: Context for cancellation and timeouts
// Returns:
// - *userEntity.User: Saved user entity
// - error: If save operation fails
func (u *UserUseCase) SaveUser(ctx context.Context) (*userEntity.User, error) {
	user, err := u.storage.SaveUser(ctx)
	if err != nil {
		return nil, ucErrors.ErrUserCannotSave
	}
	return user, nil
}

// FindUser retrieves a user by their ID.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - id: User ID to look up
// Returns:
// - *userEntity.User: Found user entity
// - error: Specific user lookup errors
func (u *UserUseCase) FindUser(ctx context.Context, id int) (*userEntity.User, error) {
	user, err := u.storage.FindUser(ctx, id)
	if err != nil {
		if errors.Is(err, dbErrors.ErrDBRecordNotFound) {
			return nil, ucErrors.ErrUserNotFound
		}
		return nil, ucErrors.ErrUserStorageNotWorking
	}
	return user, nil
}

// GetURLs retrieves all shortened URLs belonging to a user.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - user: The user whose URLs to retrieve
// Returns:
// - []*UserShortURL: List of user's URLs with full shortened URLs
// - error: If retrieval operation fails
func (u *UserUseCase) GetURLs(ctx context.Context, user *userEntity.User) ([]*UserShortURL, error) {
	var (
		shortURLs []*shortURLEntity.ShortURL
		userURLs  []*UserShortURL
		err       error
	)

	if shortURLs, err = u.storage.FindURLs(ctx, user.ID); err != nil {
		return nil, ucErrors.ErrUserStorageNotWorking
	}

	for _, shortURL := range shortURLs {
		userURLs = append(userURLs, &UserShortURL{
			ShortURL:    u.baseURL + "/" + shortURL.Alias,
			OriginalURL: shortURL.SourceURL,
		})
	}

	return userURLs, nil
}

// DeleteURLs marks the specified URLs as deleted for a user.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - user: The user owning the URLs
// - aliases: List of URL aliases to delete
// Note: Errors are logged but not returned to allow batch operations to continue
func (u *UserUseCase) DeleteURLs(ctx context.Context, user *userEntity.User, aliases []string) {
	err := u.storage.MarkURLAsDeleted(ctx, user.ID, aliases)
	if err != nil {
		logger.Log.Error(err.Error())
	}
}
