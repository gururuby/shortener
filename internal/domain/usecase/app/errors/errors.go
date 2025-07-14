// Package usecase contains application business logic and acts as an intermediary
// between the presentation layer (e.g., HTTP handlers) and the data layer (e.g., database).
// It defines core application use cases and domain-specific errors.
package usecase

import "errors"

// Errors list
var (
	// ErrAppDBIsNotReady indicates that the application cannot proceed because
	// the database connection is not established or healthy.
	//
	// This error typically occurs during:
	// - Application startup
	// - Health checks
	// - When connection pools are exhausted
	//
	// Handling recommendations:
	// 1. At startup: Fail fast and crash the application
	// 2. During runtime: Return HTTP 503 (Service Unavailable) in web handlers
	// 3. In background jobs: Implement exponential backoff retry logic
	ErrAppDBIsNotReady = errors.New("db is not ready")
)
