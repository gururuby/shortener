//go:generate mockgen -destination=./mocks/mock.go -package=mocks . Generator

/*
Package entity defines the core domain entities and value objects for the URL shortener service.

It includes:
- Interface for ID/alias generation
- Business entity definitions
- Factory methods for entity creation
- Data transfer objects for batch operations
*/
package entity

import userEntity "github.com/gururuby/shortener/internal/domain/entity/user"

// Generator defines the interface for generating unique identifiers and URL aliases.
// Implementations should ensure generated values are sufficiently unique.
type Generator interface {
	// UUID generates a universally unique identifier.
	UUID() string

	// Alias generates a short, URL-friendly identifier.
	// Returns:
	// - string: The generated alias
	// - error: Any generation error
	Alias() (string, error)
}

// ShortURL represents a shortened URL entity in the system.
// It tracks the relationship between original URLs and their shortened versions.
type ShortURL struct {
	UUID      string
	SourceURL string
	Alias     string
	UserID    int
	IsDeleted bool
}

// BatchShortURLInput represents the input structure for batch URL shortening operations.
// Used when creating multiple short URLs in a single request.
type BatchShortURLInput struct {
	CorrelationID string `json:"correlation_id"` // Client-provided ID for matching requests to responses
	OriginalURL   string `json:"original_url"`   // URL to be shortened
}

// BatchShortURLOutput represents the output structure for batch URL shortening operations.
// Contains the results of creating multiple short URLs.
type BatchShortURLOutput struct {
	CorrelationID string `json:"correlation_id"` // Echoes the client-provided correlation ID
	ShortURL      string `json:"short_url"`      // Generated shortened URL
}

// NewShortURL creates and initializes a new ShortURL entity.
//
// Parameters:
// - g: Generator implementation for creating IDs and aliases
// - user: User entity creating the short URL (can be nil for anonymous)
// - sourceURL: Original URL to be shortened
//
// Returns:
// - *ShortURL: The created short URL entity
// - error: Any error that occurred during generation
func NewShortURL(g Generator, user *userEntity.User, sourceURL string) (*ShortURL, error) {
	alias, err := g.Alias()
	if err != nil {
		return nil, err
	}
	shortURL := &ShortURL{
		UUID:      g.UUID(),
		Alias:     alias,
		SourceURL: sourceURL,
	}

	if user != nil {
		shortURL.UserID = user.ID
	}
	return shortURL, err
}
