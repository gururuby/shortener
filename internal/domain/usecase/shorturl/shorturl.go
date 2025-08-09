//go:generate mockgen -destination=./mocks/mock.go -package=mocks . ShortURLStorage

/*
Package usecase implements the business logic for URL shortening operations.

It provides:
- Short URL creation and lookup functionality
- Batch URL processing
- Input validation
- Error handling specific to URL operations
*/
package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/errors"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/shorturl/errors"
	"github.com/gururuby/shortener/pkg/validator"
)

// ShortURLStorage defines the interface for short URL persistence operations.
type ShortURLStorage interface {
	// FindShortURL retrieves a short URL by its alias.
	// Returns:
	// - *entity.ShortURL: The found short URL entity
	// - error: Any error that occurred during lookup
	FindShortURL(ctx context.Context, alias string) (*entity.ShortURL, error)

	// SaveShortURL creates and persists a new short URL.
	// Returns:
	// - *entity.ShortURL: The created short URL entity
	// - error: Any error that occurred during creation
	SaveShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (*entity.ShortURL, error)
}

// ShortURLUseCase implements the business logic for URL shortening operations.
type ShortURLUseCase struct {
	storage ShortURLStorage
	baseURL string
}

// NewShortURLUseCase creates a new instance of ShortURLUseCase.
// Parameters:
// - storage: Implementation of ShortURLStorage
// - baseURL: The base URL to use for shortened links
// Returns:
// - *ShortURLUseCase: Initialized use case instance
func NewShortURLUseCase(storage ShortURLStorage, baseURL string) *ShortURLUseCase {
	return &ShortURLUseCase{
		storage: storage,
		baseURL: baseURL,
	}
}

// CreateShortURL creates a new shortened URL from the source URL.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - user: The user creating the short URL (can be nil for anonymous)
// - sourceURL: The original URL to shorten
// Returns:
// - string: The full shortened URL (baseURL + alias)
// - error: Specific error for invalid URLs, duplicates, or storage failures
func (u *ShortURLUseCase) CreateShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (string, error) {
	if validator.IsInvalidURL(u.baseURL) {
		return "", ucErrors.ErrShortURLInvalidBaseURL
	}

	if validator.IsInvalidURL(sourceURL) {
		return "", ucErrors.ErrShortURLInvalidSourceURL
	}

	result, err := u.storage.SaveShortURL(ctx, user, sourceURL)

	if err != nil {
		if errors.Is(err, storageErrors.ErrStorageRecordIsNotUnique) {
			return u.baseURL + "/" + result.Alias, ucErrors.ErrShortURLAlreadyExist
		}
		return "", err
	}

	return u.baseURL + "/" + result.Alias, nil
}

// FindShortURL retrieves the original URL for a given alias.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - alias: The short URL identifier to look up
// Returns:
// - string: The original source URL
// - error: Specific error for missing, deleted, or invalid aliases
func (u *ShortURLUseCase) FindShortURL(ctx context.Context, alias string) (string, error) {
	alias = strings.TrimPrefix(alias, "/")

	if alias == "" {
		return "", ucErrors.ErrShortURLEmptyAlias
	}

	res, err := u.storage.FindShortURL(ctx, alias)
	if err != nil {
		return "", err
	}

	if res == nil {
		return "", ucErrors.ErrShortURLSourceURLNotFound
	}

	if res.IsDeleted {
		return "", ucErrors.ErrShortURLDeleted
	}

	return res.SourceURL, nil
}

// BatchShortURLs processes multiple URLs in a single operation.
// Parameters:
// - ctx: Context for cancellation and timeouts
// - urls: List of URLs to shorten with correlation IDs
// Returns:
// - []entity.BatchShortURLOutput: List of shortened URLs with correlation IDs
func (u *ShortURLUseCase) BatchShortURLs(ctx context.Context, urls []entity.BatchShortURLInput) []entity.BatchShortURLOutput {
	var res []entity.BatchShortURLOutput

	for _, url := range urls {
		shortURL, err := u.CreateShortURL(ctx, nil, url.OriginalURL)
		if err != nil {
			continue
		}
		res = append(res, entity.BatchShortURLOutput{
			CorrelationID: url.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return res
}
