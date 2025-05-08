package dao

import (
	daoErrors "github.com/gururuby/shortener/internal/domain/dao/errors"
	daoMock "github.com/gururuby/shortener/internal/domain/dao/mocks"
	"github.com/gururuby/shortener/internal/domain/entity"
	entityMock "github.com/gururuby/shortener/internal/domain/entity/mocks"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestDAO_FindByAlias_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := daoMock.NewMockDB(ctrl)
	gen := entityMock.NewMockGenerator(ctrl)

	dao := New(gen, db)

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
		db.EXPECT().Find(tt.alias).Return(tt.dbRecord.value, tt.dbRecord.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			res, err := dao.FindByAlias(tt.alias)
			require.NoError(t, err)
			require.Equal(t, tt.dbRecord.value, res)
		})
	}
}

func TestDAO_FindByAlias_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := daoMock.NewMockDB(ctrl)
	gen := entityMock.NewMockGenerator(ctrl)
	dao := New(gen, db)

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
			result: result{err: daoErrors.ErrDAORecordNotFound},
		},
	}
	for _, tt := range tests {
		db.EXPECT().Find(tt.alias).Return(tt.result.value, tt.result.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			_, err := dao.FindByAlias(tt.alias)
			require.Error(t, tt.result.err, err)
		})
	}
}

func TestDAO_Save_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := daoMock.NewMockDB(ctrl)

	gen := entityMock.NewMockGenerator(ctrl)
	gen.EXPECT().UUID().Return("UUID")
	gen.EXPECT().Alias().Return("alias")

	dao := New(gen, db)

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
			db.EXPECT().Save(tt.res).Return(tt.res, nil)
			res, err := dao.Save(tt.sourceURL)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestIsDBReady(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := daoMock.NewMockDB(ctrl)

	gen := entityMock.NewMockGenerator(ctrl)
	dao := New(gen, db)

	t.Run("when DB ping is ok", func(t *testing.T) {
		db.EXPECT().Ping().Return(nil)
		err := dao.IsDBReady()
		require.NoError(t, err)
	})

	t.Run("when DB ping is failed", func(t *testing.T) {
		db.EXPECT().Ping().Return(dbErrors.ErrDBIsNotHealthy)
		err := dao.IsDBReady()
		require.Error(t, daoErrors.ErrDAOIsNotReadyDB, err)
	})
}
