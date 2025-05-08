package usecase

import (
	"github.com/gururuby/shortener/internal/domain/entity"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
	"github.com/gururuby/shortener/internal/domain/usecase/shorturl/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestFindShortURL_OK(t *testing.T) {
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
			name:   "when record exist in db",
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

func TestCreateShortURL_OK(t *testing.T) {
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
		dao.EXPECT().Save(tt.sourceURL).Return(tt.daoRes.shortURL, nil)
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
			err:     ucErrors.ErrShortURLInvalidBaseURL,
		},
		{
			name:    "when passed empty source URL",
			baseURL: "http://localhost:8888",
			err:     ucErrors.ErrShortURLInvalidSourceURL,
		},
		{
			name:      "when passed invalid source URL",
			sourceURL: "h://abcd",
			baseURL:   "http://localhost:8888",
			err:       ucErrors.ErrShortURLInvalidSourceURL,
		},
	}
	for _, tt := range tests {
		uc := NewShortURLUseCase(dao, tt.baseURL)

		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.CreateShortURL(tt.sourceURL)
			require.Error(t, tt.err, err)
		})
	}
}

func TestBatchShortURLs_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mocks.NewMockDAO(ctrl)

	var urls []entity.BatchShortURLInput
	urls = append(urls,
		entity.BatchShortURLInput{CorrelationID: "1", OriginalURL: "https://ya.ru"},
		entity.BatchShortURLInput{CorrelationID: "2", OriginalURL: "https://ya.com"},
	)

	dao.EXPECT().Save(urls[0].OriginalURL).Return(&entity.ShortURL{Alias: "alias1"}, nil).Times(1)
	dao.EXPECT().Save(urls[1].OriginalURL).Return(&entity.ShortURL{Alias: "alias2"}, nil).Times(1)

	tests := []struct {
		name    string
		baseURL string
		urls    []entity.BatchShortURLInput
		result  []entity.BatchShortURLOutput
	}{
		{
			name:    "when successfully batch proceed",
			baseURL: "http://localhost:8080",
			urls:    urls,
			result: []entity.BatchShortURLOutput{
				{CorrelationID: "1", ShortURL: "http://localhost:8080/alias1"},
				{CorrelationID: "2", ShortURL: "http://localhost:8080/alias2"},
			},
		},
	}
	for _, tt := range tests {
		uc := NewShortURLUseCase(dao, tt.baseURL)

		t.Run(tt.name, func(t *testing.T) {
			res := uc.BatchShortURLs(tt.urls)
			require.Equal(t, tt.result, res)
		})
	}
}
