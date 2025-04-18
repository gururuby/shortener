package dao

import (
	"errors"
	"github.com/gururuby/shortener/config"
	"github.com/gururuby/shortener/internal/domain/dao/mock_dao"
	"github.com/gururuby/shortener/internal/domain/entity"
	"github.com/gururuby/shortener/internal/domain/entity/mock_entity"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestDAO_FindByAlias_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := mock_dao.NewMockDB(ctrl)
	gen := mock_entity.NewMockGenerator(ctrl)

	cfg, _ := config.New()
	dao := New(gen, cfg, db)

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
			res, _ := dao.FindByAlias(tt.alias)
			require.Equal(t, tt.dbRecord.value, res)
		})
	}
}

func TestDAO_FindByAlias_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := mock_dao.NewMockDB(ctrl)
	cfg, _ := config.New()
	gen := mock_entity.NewMockGenerator(ctrl)
	dao := New(gen, cfg, db)

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
			name:     "when cannot find record in db by alias",
			alias:    "unknown_alias",
			dbRecord: dbRecord{err: errors.New("not found URL")},
		},
	}
	for _, tt := range tests {
		db.EXPECT().Find(tt.alias).Return(tt.dbRecord.value, tt.dbRecord.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			_, err := dao.FindByAlias(tt.alias)
			require.Error(t, tt.dbRecord.err, err)
		})
	}
}

func TestDAO_Save_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := mock_dao.NewMockDB(ctrl)
	cfg, _ := config.New()

	gen := mock_entity.NewMockGenerator(ctrl)
	gen.EXPECT().UUID().Return("UUID")
	gen.EXPECT().Alias().Return("alias")

	dao := New(gen, cfg, db)

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
			res, _ := dao.Save(tt.sourceURL)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestDAO_Save_RetryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := mock_dao.NewMockDB(ctrl)
	cfg, _ := config.New()

	gen := mock_entity.NewMockGenerator(ctrl)
	gen.EXPECT().UUID().Return("UUID").AnyTimes()
	gen.EXPECT().Alias().Return("alias").AnyTimes()

	shortURL := entity.NewShortURL(gen, "https://ya.ru")

	dao := New(gen, cfg, db)

	type dbRecord struct {
		value *entity.ShortURL
		err   error
	}

	tests := []struct {
		name      string
		sourceURL string
		dbRecord  dbRecord
		err       error
		retryCnt  int
	}{
		{
			name:      "when try to save non unique value in db and retry to save",
			sourceURL: "https://ya.ru",
			dbRecord:  dbRecord{value: nil, err: errNonUnique},
			err:       errSave,
			retryCnt:  5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.EXPECT().Save(shortURL).Return(tt.dbRecord.value, tt.dbRecord.err).Times(tt.retryCnt)
			res, _ := dao.Save(tt.sourceURL)
			require.Error(t, tt.err, res)
		})
	}
}
