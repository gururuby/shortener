package usecase

import (
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/shorturl/errors"
	"github.com/gururuby/shortener/internal/domain/usecase/app/mocks"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestPingDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)
	uc := NewAppUseCase(storage)

	t.Run("when all is ok with db", func(t *testing.T) {
		storage.EXPECT().IsDBReady().Return(nil)
		err := uc.PingDB()
		require.NoError(t, err)
	})

	t.Run("when something wrong with db", func(t *testing.T) {
		storage.EXPECT().IsDBReady().Return(storageErrors.ErrStorageIsNotReadyDB)
		err := uc.PingDB()
		require.ErrorIs(t, ucErrors.ErrAppDBIsNotReady, err)
	})
}
