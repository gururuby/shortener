package dao

import (
	"errors"
	"github.com/gururuby/shortener/internal/infra/db/memory/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestShortURLDAO_FindByAlias_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := mock.NewMockshortURLDB(ctrl)
	dao := NewShortURLDAO(db)

	type dbRes struct {
		res string
		err error
	}

	tests := []struct {
		name  string
		alias string
		dbRes dbRes
		res   string
	}{
		{
			name:  "when find short URL in db by alias",
			alias: "alias",
			dbRes: dbRes{res: "https://ya.ru"},
			res:   "https://ya.ru",
		},
	}
	for _, tt := range tests {
		db.EXPECT().Find(tt.alias).Return(tt.dbRes.res, tt.dbRes.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			res, _ := dao.FindByAlias(tt.alias)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestShortURLDAO_FindByAlias_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := mock.NewMockshortURLDB(ctrl)
	dao := NewShortURLDAO(db)

	type dbRes struct {
		res string
		err error
	}

	tests := []struct {
		name  string
		alias string
		dbRes dbRes
		err   error
	}{
		{
			name:  "when cannot find short URL in db by alias",
			alias: "unknown_alias",
			dbRes: dbRes{res: "", err: errors.New("not found URL")},
			err:   errors.New("not found URL"),
		},
	}
	for _, tt := range tests {
		db.EXPECT().Find(tt.alias).Return(tt.dbRes.res, tt.dbRes.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			_, err := dao.FindByAlias(tt.alias)
			require.Error(t, tt.err, err)
		})
	}
}

func TestShortURLDAO_Save_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := mock.NewMockshortURLDB(ctrl)
	dao := NewShortURLDAO(db)

	type dbRes struct {
		res string
		err error
	}

	tests := []struct {
		name      string
		sourceURL string
		dbRes     dbRes
		res       string
	}{
		{
			name:      "when save short URL in db",
			sourceURL: "https://ya.ru",
			dbRes:     dbRes{res: "alias"},
			res:       "alias",
		},
	}
	for _, tt := range tests {
		db.EXPECT().Save(tt.sourceURL).Return(tt.dbRes.res, tt.dbRes.err).Times(1)

		t.Run(tt.name, func(t *testing.T) {
			res, _ := dao.Save(tt.sourceURL)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestShortURLDAO_Save_RetryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := mock.NewMockshortURLDB(ctrl)
	dao := NewShortURLDAO(db)

	type dbRes struct {
		res string
		err error
	}

	tests := []struct {
		name      string
		sourceURL string
		dbRes     dbRes
		err       error
		retryCnt  int
	}{
		{
			name:      "when try to save non unique value in db and retry to save",
			sourceURL: "https://ya.ru",
			dbRes:     dbRes{res: "", err: errNonUnique},
			err:       errSave,
			retryCnt:  5,
		},
	}
	for _, tt := range tests {
		db.EXPECT().Save(tt.sourceURL).Return(tt.dbRes.res, tt.dbRes.err).Times(tt.retryCnt)

		t.Run(tt.name, func(t *testing.T) {
			res, _ := dao.Save(tt.sourceURL)
			require.Error(t, tt.err, res)
		})
	}
}
