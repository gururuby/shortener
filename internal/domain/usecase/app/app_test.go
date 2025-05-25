package usecase

import (
	"context"
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/errors"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/app/errors"
	"github.com/gururuby/shortener/internal/domain/usecase/app/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_PingDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)
	ctx := context.Background()
	uc := NewAppUseCase(storage)

	t.Run("when all is ok with db", func(t *testing.T) {
		storage.EXPECT().IsDBReady(ctx).Return(nil)
		err := uc.PingDB(ctx)
		require.NoError(t, err)
	})

	t.Run("when something wrong with db", func(t *testing.T) {
		storage.EXPECT().IsDBReady(ctx).Return(storageErrors.ErrStorageIsNotReadyDB)
		err := uc.PingDB(ctx)
		require.ErrorIs(t, ucErrors.ErrAppDBIsNotReady, err)
	})
}
