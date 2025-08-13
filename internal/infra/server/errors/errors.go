// Package errors defines common error types used across the application.
// These errors provide consistent error handling and messaging.
package errors

import "errors"

// Error definitions for the application.
// These should be used instead of creating new error strings inline.
var (
	// ErrServerInvalidTLSConfig indicates that the TLS configuration is invalid.
	// This typically occurs when HTTPS is enabled but certificate files are missing or invalid.
	//
	// Example cases:
	// - Certificate file path is empty
	// - Key file path is empty
	// - Certificate files are not readable
	// - Certificate and key don't match
	ErrServerInvalidTLSConfig = errors.New("invalid TLS configuration")
)
