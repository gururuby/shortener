// Package storage provides interfaces and implementations for data storage operations.
// It defines abstractions for accessing and managing application data.
package storage

import (
	"context"

	statsEntity "github.com/gururuby/shortener/internal/domain/entity/stats"
)

//go:generate mockgen -destination=./mocks/mock.go -package=mocks . StatsDB

// StatsDB defines the interface for statistics database operations.
// Implementations of this interface are responsible for retrieving
// statistical data from the underlying data storage.
type StatsDB interface {
	// GetResourcesCounts retrieves the counts of various resources from the database.
	// Returns a Stats entity containing the counts or an error if the operation fails.
	GetResourcesCounts(ctx context.Context) (*statsEntity.Stats, error)
}

// StatsStorage provides a wrapper around StatsDB with additional business logic.
// It serves as an intermediate layer between the application and data storage.
type StatsStorage struct {
	db StatsDB
}

// Setup creates and returns a new instance of StatsStorage.
// It initializes the storage with the provided StatsDB implementation.
func Setup(db StatsDB) *StatsStorage {
	return &StatsStorage{db: db}
}

// GetStats retrieves application statistics from the underlying database.
// It delegates the actual data retrieval to the configured StatsDB implementation.
// Returns statistics data or an error if the operation fails.
func (s *StatsStorage) GetStats(ctx context.Context) (*statsEntity.Stats, error) {
	return s.db.GetResourcesCounts(ctx)
}
