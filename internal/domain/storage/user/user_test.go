package storage

import (
	"context"
	"github.com/gururuby/shortener/internal/domain/entity/user"
	"github.com/gururuby/shortener/internal/domain/storage/errors"
	storageMock "github.com/gururuby/shortener/internal/domain/storage/user/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_Storage_FindUser_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	ctx := context.Background()

	storage := UserStorage{db: db}

	type dbRecord struct {
		value *entity.User
		err   error
	}

	tests := []struct {
		name     string
		ID       int
		dbRecord dbRecord
	}{
		{
			name:     "when find user in db by ID",
			ID:       1,
			dbRecord: dbRecord{value: &entity.User{ID: 1}},
		},
	}
	for _, tt := range tests {
		db.EXPECT().FindUser(ctx, tt.ID).Return(tt.dbRecord.value, tt.dbRecord.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			res, err := storage.FindUser(ctx, tt.ID)
			require.NoError(t, err)
			require.Equal(t, tt.dbRecord.value, res)
		})
	}
}

func Test_Storage_FindUser_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	ctx := context.Background()

	storage := UserStorage{db: db}

	type result struct {
		value *entity.User
		err   error
	}

	tests := []struct {
		name   string
		ID     int
		result result
	}{
		{
			name:   "when cannot find user in db by ID",
			ID:     2,
			result: result{err: errors.ErrStorageRecordNotFound},
		},
	}
	for _, tt := range tests {
		db.EXPECT().FindUser(ctx, tt.ID).Return(tt.result.value, tt.result.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			_, err := storage.FindUser(ctx, tt.ID)
			require.Error(t, tt.result.err, err)
		})
	}
}

func Test_Storage_SaveUser_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	ctx := context.Background()
	user := &entity.User{ID: 1}

	storage := UserStorage{db: db}

	tests := []struct {
		name string
		res  *entity.User
	}{
		{
			name: "when save user in db",
			res:  user,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.EXPECT().SaveUser(ctx).Return(tt.res, nil)
			res, err := storage.SaveUser(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}
