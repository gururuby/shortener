// Package errors defines domain-specific error conditions for the URL shortener service.
package errors

import "errors"

// Errors list
var (
	// ErrGeneratorEmptyAliasLength indicates an invalid configuration where
	// the requested alias length is zero or unset.
	//
	// This error occurs when:
	// - The alias generation configuration specifies length = 0
	// - No default length is provided
	// - Configuration loading fails to set this value
	//
	// Resolution steps:
	// 1. Verify your configuration file has 'alias_length' set
	// 2. Ensure environment variables override values properly
	// 3. Set reasonable default (e.g., 6-8 characters)
	//
	// Example valid configuration:
	//   alias_length: 7  # Must be positive integer
	ErrGeneratorEmptyAliasLength = errors.New("alias length is zero, please configure correct value")
)
