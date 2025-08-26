package usecase

import (
	"context"
	"testing"

	entity "github.com/gururuby/shortener/internal/domain/entity/stats"
	"github.com/gururuby/shortener/internal/domain/usecase/stats/mocks"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_GetStats_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStatsStorage(ctrl)
	ctx := context.Background()

	type storageRes struct {
		stats *entity.Stats
		err   error
	}

	tests := []struct {
		name       string
		storageRes storageRes
	}{
		{
			name:       "when received stats from storage",
			storageRes: storageRes{stats: &entity.Stats{UsersCount: 123, URLsCount: 124}},
		},
	}
	for _, tt := range tests {
		storage.EXPECT().GetStats(ctx).Return(tt.storageRes.stats, tt.storageRes.err).Times(1)
		uc := NewStatsUseCase(storage)

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.GetStats(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.storageRes.stats, res)
		})
	}
}

func Test_GetStats_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStatsStorage(ctrl)
	ctx := context.Background()

	type storageRes struct {
		stats *entity.Stats
		err   error
	}

	tests := []struct {
		name       string
		storageRes storageRes
	}{
		{
			name:       "when received errors from storage",
			storageRes: storageRes{stats: nil, err: dbErrors.ErrDBIsNotHealthy},
		},
	}
	for _, tt := range tests {
		storage.EXPECT().GetStats(ctx).Return(tt.storageRes.stats, tt.storageRes.err).Times(1)
		uc := NewStatsUseCase(storage)

		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.GetStats(ctx)
			require.Error(t, err)
			require.Equal(t, tt.storageRes.err, err)
		})
	}
}
