package usecase

import (
	"context"
	daoErrors "github.com/gururuby/shortener/internal/domain/dao/errors"
	"github.com/gururuby/shortener/internal/domain/usecase/app/mocks"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestPingDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mocks.NewMockDAO(ctrl)
	uc := NewAppUseCase(dao)
	ctx := context.Background()

	t.Run("when all is ok with db", func(t *testing.T) {
		dao.EXPECT().IsDBReady(ctx).Return(nil)
		err := uc.PingDB(ctx)
		require.NoError(t, err)
	})

	t.Run("when something wrong with db", func(t *testing.T) {
		dao.EXPECT().IsDBReady(ctx).Return(daoErrors.ErrDAOIsNotReadyDB)
		err := uc.PingDB(ctx)
		require.ErrorIs(t, ucErrors.ErrAppDBIsNotReady, err)
	})
}
