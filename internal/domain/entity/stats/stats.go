// Package entity defines the core domain models for the application.
// These models represent the fundamental business entities and their relationships.
package entity

// Stats represents application statistics containing various metrics.
// This structure is typically used to provide summary information about
// the system's current state and usage metrics.
//
// The struct fields are annotated with JSON tags for proper serialization
// when sending statistics via HTTP responses.
type Stats struct {
	URLsCount  int64 `json:"urls"`  // Total number of URLs in the system
	UsersCount int64 `json:"users"` // Total number of registered users
}
