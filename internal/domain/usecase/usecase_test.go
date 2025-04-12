package usecase

import (
	"errors"
	"github.com/gururuby/shortener/internal/domain/dao/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestFindShortURL_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mock.NewMockDAO(ctrl)

	type daoRes struct {
		sourceURL string
	}

	tests := []struct {
		name   string
		alias  string
		daoRes daoRes
		res    string
	}{
		{
			name:   "when find sourceURL in db",
			alias:  "alias1",
			daoRes: daoRes{sourceURL: "https://ya.ru"},
			res:    "https://ya.ru",
		},
		{
			name:   "when alias passed with '/' prefix",
			alias:  "/alias1",
			daoRes: daoRes{sourceURL: "https://ya.ru"},
			res:    "https://ya.ru",
		},
	}
	for _, tt := range tests {
		dao.EXPECT().FindByAlias("alias1").Return(tt.daoRes.sourceURL, nil).AnyTimes()
		uc := UseCase{DAO: dao}

		t.Run(tt.name, func(t *testing.T) {
			res, _ := uc.FindShortURL(tt.alias)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestFindShortURL_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mock.NewMockDAO(ctrl)

	type daoRes struct {
		sourceURL string
		err       error
	}

	tests := []struct {
		name   string
		alias  string
		daoRes daoRes
		err    error
	}{
		{
			name:  "when passed empty alias",
			alias: "",
			err:   errors.New(EmptyAliasError),
		},
		{
			name:   "when source URL in db not found",
			alias:  "alias2",
			daoRes: daoRes{sourceURL: ""},
			err:    errors.New(SourceURLNotFoundError),
		},
		{
			name:   "when something went wrong with db",
			alias:  "alias3",
			daoRes: daoRes{sourceURL: "", err: errors.New("something went wrong")},
			err:    errors.New("something went wrong"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dao.EXPECT().FindByAlias(tt.alias).Return(tt.daoRes.sourceURL, tt.daoRes.err).AnyTimes()
			uc := UseCase{DAO: dao}
			_, err := uc.FindShortURL(tt.alias)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestCreateShortURL_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mock.NewMockDAO(ctrl)

	type daoRes struct {
		alias string
	}

	tests := []struct {
		name      string
		sourceURL string
		baseURL   string
		daoRes    daoRes
		res       string
	}{
		{
			name:      "when successfully stored short URL",
			sourceURL: "https://ya.ru",
			baseURL:   "http://localhost:8888",
			daoRes:    daoRes{alias: "alias"},
			res:       "http://localhost:8888/alias",
		},
	}
	for _, tt := range tests {
		dao.EXPECT().Save(tt.sourceURL).Return(tt.daoRes.alias, nil).AnyTimes()
		uc := UseCase{DAO: dao, baseURL: tt.baseURL}

		t.Run(tt.name, func(t *testing.T) {
			res, _ := uc.CreateShortURL(tt.sourceURL)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestCreateShortURL_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mock.NewMockDAO(ctrl)

	type daoRes struct {
		alias string
		err   error
	}

	tests := []struct {
		name      string
		sourceURL string
		baseURL   string
		daoRes    daoRes
		err       error
	}{
		{
			name:    "when passed empty base URL",
			baseURL: "",
			err:     errors.New(EmptyBaseURLError),
		},
		{
			name:    "when passed empty source URL",
			baseURL: "http://localhost:8888",
			err:     errors.New(EmptySourceURLError),
		},
		{
			name:      "when something went wrong with db",
			baseURL:   "http://localhost:8888",
			sourceURL: "https://ya.ru",
			daoRes:    daoRes{alias: "", err: errors.New("something went wrong")},
			err:       errors.New("something went wrong"),
		},
	}
	for _, tt := range tests {
		dao.EXPECT().Save(tt.sourceURL).Return(tt.daoRes.alias, tt.daoRes.err).AnyTimes()
		uc := UseCase{DAO: dao, baseURL: tt.baseURL}

		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.CreateShortURL(tt.sourceURL)
			require.Equal(t, tt.err, err)
		})
	}
}
