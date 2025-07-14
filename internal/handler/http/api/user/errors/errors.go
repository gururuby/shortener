// Package handler contains HTTP request handlers for URL shortening operations.
// It defines API-specific errors related to request validation and processing.
package handler

import "errors"

// Errors list
var (
	// ErrHandlerNoAliasesForDelete indicates a request to delete short URLs was made
	// without providing any aliases to delete.
	//
	// Typical cases:
	// - Empty array in JSON payload: `[]`
	// - Malformed input where aliases couldn't be parsed
	//
	ErrHandlerNoAliasesForDelete = errors.New("no aliases passed to delete short urls")
)
