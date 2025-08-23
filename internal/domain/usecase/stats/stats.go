// Package usecase contains application business logic and use cases.
// It implements the core functionality by coordinating between
// domain entities and storage interfaces.
package usecase

import (
	"context"

	statsEntity "github.com/gururuby/shortener/internal/domain/entity/stats"
)

//go:generate mockgen -destination=./mocks/mock.go -package=mocks . StatsStorage

// StatsStorage defines the interface for statistics storage operations.
// This interface abstracts the data retrieval layer from the business logic.
type StatsStorage interface {
	// GetStats retrieves application statistics from the storage layer.
	// Returns statistics data or an error if the operation fails.
	GetStats(ctx context.Context) (*statsEntity.Stats, error)
}

// StatsUseCase handles statistics-related business logic.
// It serves as an intermediary between the presentation layer
// and the data storage layer for statistics operations.
type StatsUseCase struct {
	storage StatsStorage
}

// NewStatsUseCase creates a new instance of StatsUseCase.
// It initializes the use case with the provided storage implementation.
func NewStatsUseCase(storage StatsStorage) *StatsUseCase {
	return &StatsUseCase{
		storage: storage,
	}
}

// GetStats retrieves application statistics by delegating to the storage layer.
// This method implements the business logic for fetching statistics data.
// Returns the statistics entity or an error if the operation fails.
func (s *StatsUseCase) GetStats(ctx context.Context) (*statsEntity.Stats, error) {
	return s.storage.GetStats(ctx)
}
