// Package entity defines the core domain models for the application.
// These models represent the fundamental business entities and their relationships.
package entity

// User represents an application user in the system.
// It contains the basic authentication information and identifier.
type User struct {
	AuthToken string
	ID        int
}
