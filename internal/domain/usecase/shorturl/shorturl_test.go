package usecase

import (
	"github.com/gururuby/shortener/internal/domain/entity"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
	"github.com/gururuby/shortener/internal/domain/usecase/shorturl/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestFindShortURL_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mocks.NewMockDAO(ctrl)

	type daoRes struct {
		shortURL *entity.ShortURL
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
			daoRes: daoRes{shortURL: &entity.ShortURL{SourceURL: "https://ya.ru"}},
			res:    "https://ya.ru",
		},
		{
			name:   "when alias passed with '/' prefix",
			alias:  "/alias1",
			daoRes: daoRes{shortURL: &entity.ShortURL{SourceURL: "https://ya.ru"}},
			res:    "https://ya.ru",
		},
	}
	for _, tt := range tests {
		dao.EXPECT().FindByAlias("alias1").Return(tt.daoRes.shortURL, nil).AnyTimes()
		uc := NewShortURLUseCase(dao, "baseURL")

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.FindShortURL(tt.alias)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestFindShortURL_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mocks.NewMockDAO(ctrl)

	type daoRes struct {
		shortURL *entity.ShortURL
		err      error
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
			err:   ucErrors.ErrShortURLEmptyAlias,
		},
		{
			name:   "when source URL in db not found",
			alias:  "alias2",
			daoRes: daoRes{shortURL: nil},
			err:    ucErrors.ErrShortURLSourceURLNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dao.EXPECT().FindByAlias(tt.alias).Return(tt.daoRes.shortURL, tt.daoRes.err).AnyTimes()
			uc := NewShortURLUseCase(dao, "base")
			_, err := uc.FindShortURL(tt.alias)
			require.Error(t, tt.err, err)
		})
	}
}

func TestCreateShortURL_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mocks.NewMockDAO(ctrl)

	type daoRes struct {
		shortURL *entity.ShortURL
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
			daoRes:    daoRes{shortURL: &entity.ShortURL{Alias: "alias"}},
			res:       "http://localhost:8888/alias",
		},
	}
	for _, tt := range tests {
		dao.EXPECT().Save(tt.sourceURL).Return(tt.daoRes.shortURL, nil).AnyTimes()
		uc := NewShortURLUseCase(dao, tt.baseURL)

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.CreateShortURL(tt.sourceURL)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestCreateShortURL_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mocks.NewMockDAO(ctrl)

	type daoRes struct {
		shortURL *entity.ShortURL
		err      error
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
			err:     ucErrors.ErrShortURLEmptyBaseURL,
		},
		{
			name:    "when passed empty source URL",
			baseURL: "http://localhost:8888",
			err:     ucErrors.ErrShortURLEmptySourceURL,
		},
	}
	for _, tt := range tests {
		dao.EXPECT().Save(tt.sourceURL).Return(tt.daoRes.shortURL, tt.daoRes.err).AnyTimes()
		uc := NewShortURLUseCase(dao, tt.baseURL)

		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.CreateShortURL(tt.sourceURL)
			require.Error(t, tt.err, err)
		})
	}
}
