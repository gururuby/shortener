package storage

import (
	"context"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	"github.com/gururuby/shortener/internal/domain/entity/user"
	"github.com/gururuby/shortener/internal/domain/storage/errors"
	storageMock "github.com/gururuby/shortener/internal/domain/storage/user/mocks"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
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
		dbRecord dbRecord
		name     string
		ID       int
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
		result result
		name   string
		ID     int
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
		res  *entity.User
		name string
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

func Test_Storage_SaveUser_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	ctx := context.Background()

	storage := UserStorage{db: db}

	tests := []struct {
		err  error
		name string
	}{
		{
			name: "when cannot save user in db",
			err:  dbErrors.ErrDBQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.EXPECT().SaveUser(ctx).Return(nil, tt.err)
			_, err := storage.SaveUser(ctx)
			require.Error(t, err)
		})
	}
}

func Test_Storage_FindURLs_OK(t *testing.T) {
	var urls []*shortURLEntity.ShortURL

	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	ctx := context.Background()
	storage := UserStorage{db: db}

	tests := []struct {
		name   string
		res    []*shortURLEntity.ShortURL
		userID int
	}{
		{
			name:   "when find user URLs in db",
			userID: 1,
			res:    urls,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.EXPECT().FindUserURLs(ctx, 1).Return(tt.res, nil)
			res, err := storage.FindURLs(ctx, tt.userID)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func Test_Storage_FindURLs_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	ctx := context.Background()

	storage := UserStorage{db: db}

	tests := []struct {
		err    error
		name   string
		userID int
	}{
		{
			name:   "when something went wrong with db query",
			userID: 1,
			err:    dbErrors.ErrDBQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.EXPECT().FindUserURLs(ctx, 1).Return(nil, tt.err)
			_, err := storage.FindURLs(ctx, tt.userID)
			require.Error(t, err)
		})
	}
}

func Test_Storage_MarkURLAsDeleted_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	ctx := context.Background()
	storage := UserStorage{db: db}

	tests := []struct {
		err     error
		name    string
		aliases []string
		userID  int
	}{
		{
			name:    "when successfully marks URLs as deleted",
			userID:  1,
			err:     nil,
			aliases: []string{"some_alias"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.EXPECT().MarkURLAsDeleted(ctx, 1, tt.aliases).Return(tt.err)
			err := storage.MarkURLAsDeleted(ctx, tt.userID, tt.aliases)
			require.NoError(t, err)
		})
	}
}

func Test_Storage_MarkURLAsDeleted_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	ctx := context.Background()
	storage := UserStorage{db: db}

	tests := []struct {
		err     error
		name    string
		aliases []string
		userID  int
	}{
		{
			name:    "when marks URLs as deleted was failed with error",
			userID:  1,
			err:     dbErrors.ErrDBQuery,
			aliases: []string{"some_alias"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.EXPECT().MarkURLAsDeleted(ctx, 1, tt.aliases).Return(tt.err)
			err := storage.MarkURLAsDeleted(ctx, tt.userID, tt.aliases)
			require.Error(t, err)
		})
	}
}
