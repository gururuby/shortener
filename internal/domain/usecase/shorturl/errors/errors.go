// Package usecase implements the core business logic for URL shortening operations.
// It defines domain-specific errors that may occur during URL shortening operations.
package usecase

import "errors"

// Errors list
var (
	// ErrShortURLAlreadyExist indicates an attempt to create a short URL with an alias
	// that already exists in the system.
	//
	// Resolution: Either:
	// - Choose a different alias
	// - Implement update logic (if appropriate)
	ErrShortURLAlreadyExist = errors.New("short URL already exist")

	// ErrShortURLInvalidBaseURL indicates the base URL used for generating short URLs
	// is malformed or invalid.
	//
	// Typical cases:
	// - Missing scheme (http/https)
	// - Invalid domain format
	// - Contains illegal characters
	ErrShortURLInvalidBaseURL = errors.New("invalid base URL, please specify valid URL")

	// ErrShortURLInvalidSourceURL indicates the provided long URL is invalid.
	//
	// Common validations failed:
	// - URL parsing failure
	// - Missing scheme
	// - Empty URL
	// - Exceeds maximum length
	ErrShortURLInvalidSourceURL = errors.New("invalid source URL, please specify valid URL")

	// ErrShortURLEmptyAlias indicates a request was made with an empty short URL identifier.
	//
	// Prevention:
	// - Validate input before processing
	// - Provide default generation if empty
	ErrShortURLEmptyAlias = errors.New("empty alias, please specify alias")

	// ErrShortURLSourceURLNotFound indicates the requested short URL doesn't exist
	// in the system (404 equivalent).
	//
	// Note: Distinct from deleted URLs (ErrShortURLDeleted)
	ErrShortURLSourceURLNotFound = errors.New("source URL not found")

	// ErrShortURLDeleted indicates the requested short URL was previously created
	// but has been soft-deleted.
	//
	// Business considerations:
	// - May want to track deletion timestamps
	// - Could allow recreation after cleanup period
	ErrShortURLDeleted = errors.New("short URL was deleted")
)
