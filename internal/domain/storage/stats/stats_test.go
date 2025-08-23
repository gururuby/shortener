package storage

import (
	"context"
	"testing"

	statsEntity "github.com/gururuby/shortener/internal/domain/entity/stats"
	storageMock "github.com/gururuby/shortener/internal/domain/storage/stats/mocks"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Storage_GetStats_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockStatsDB(ctrl)
	ctx := context.Background()

	storage := StatsStorage{db: db}

	type dbResult struct {
		value *statsEntity.Stats
		err   error
	}

	tests := []struct {
		dbResult dbResult
		name     string
	}{
		{
			name:     "when stats returns data",
			dbResult: dbResult{value: &statsEntity.Stats{URLsCount: 10, UsersCount: 20}},
		},
	}
	for _, tt := range tests {
		db.EXPECT().GetResourcesCounts(ctx).Return(tt.dbResult.value, tt.dbResult.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			res, err := storage.GetStats(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.dbResult.value, res)
		})
	}
}

func Test_Storage_GetStats_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockStatsDB(ctrl)
	ctx := context.Background()

	storage := StatsStorage{db: db}

	type dbResult struct {
		value *statsEntity.Stats
		err   error
	}

	tests := []struct {
		dbResult dbResult
		name     string
	}{
		{
			name:     "when stats returns some db error",
			dbResult: dbResult{value: nil, err: dbErrors.ErrDBIsNotHealthy},
		},
	}
	for _, tt := range tests {
		db.EXPECT().GetResourcesCounts(ctx).Return(tt.dbResult.value, tt.dbResult.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			_, err := storage.GetStats(ctx)
			require.Error(t, err)
			require.Equal(t, tt.dbResult.err, err)
		})
	}
}
