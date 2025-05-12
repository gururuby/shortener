package storage

import (
	"context"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	entityMock "github.com/gururuby/shortener/internal/domain/entity/shorturl/mocks"
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/shorturl/errors"
	storageMock "github.com/gururuby/shortener/internal/domain/storage/shorturl/mocks"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_Storage_FindByAlias_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	gen := entityMock.NewMockGenerator(ctrl)
	ctx := context.Background()

	storage := Storage{gen: gen, db: db}

	type dbRecord struct {
		value *entity.ShortURL
		err   error
	}

	tests := []struct {
		name     string
		alias    string
		dbRecord dbRecord
	}{
		{
			name:     "when find short URL in db by alias",
			alias:    "alias",
			dbRecord: dbRecord{value: &entity.ShortURL{SourceURL: "https://ya.ru"}},
		},
	}
	for _, tt := range tests {
		db.EXPECT().FindShortURL(ctx, tt.alias).Return(tt.dbRecord.value, tt.dbRecord.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			res, err := storage.FindByAlias(ctx, tt.alias)
			require.NoError(t, err)
			require.Equal(t, tt.dbRecord.value, res)
		})
	}
}

func Test_Storage_FindByAlias_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	gen := entityMock.NewMockGenerator(ctrl)
	ctx := context.Background()

	storage := Storage{gen: gen, db: db}

	type result struct {
		value *entity.ShortURL
		err   error
	}

	tests := []struct {
		name   string
		alias  string
		result result
	}{
		{
			name:   "when cannot find record in db by alias",
			alias:  "unknown_alias",
			result: result{err: storageErrors.ErrStorageRecordNotFound},
		},
	}
	for _, tt := range tests {
		db.EXPECT().FindShortURL(ctx, tt.alias).Return(tt.result.value, tt.result.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			_, err := storage.FindByAlias(ctx, tt.alias)
			require.Error(t, tt.result.err, err)
		})
	}
}

func Test_Storage_Save_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	ctx := context.Background()

	gen := entityMock.NewMockGenerator(ctrl)
	gen.EXPECT().UUID().Return("UUID")
	gen.EXPECT().Alias().Return("alias")

	storage := Storage{gen: gen, db: db}

	tests := []struct {
		name      string
		sourceURL string
		res       *entity.ShortURL
	}{
		{
			name:      "when save short URL in db",
			sourceURL: "https://ya.ru",
			res: &entity.ShortURL{
				UUID:      "UUID",
				SourceURL: "https://ya.ru",
				Alias:     "alias",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.EXPECT().SaveShortURL(ctx, tt.res).Return(tt.res, nil)
			res, err := storage.Save(ctx, tt.sourceURL)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func Test_IsDBReady(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := storageMock.NewMockDB(ctrl)
	gen := entityMock.NewMockGenerator(ctrl)
	ctx := context.Background()

	storage := Storage{gen: gen, db: db}

	t.Run("when DB ping is ok", func(t *testing.T) {
		db.EXPECT().Ping(ctx).Return(nil)
		err := storage.IsDBReady(ctx)
		require.NoError(t, err)
	})

	t.Run("when DB ping is failed", func(t *testing.T) {
		db.EXPECT().Ping(ctx).Return(dbErrors.ErrDBIsNotHealthy)
		err := storage.IsDBReady(ctx)
		require.Error(t, storageErrors.ErrStorageIsNotReadyDB, err)
	})
}
