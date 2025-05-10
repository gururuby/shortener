package usecase

import (
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/shorturl/errors"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
	"github.com/gururuby/shortener/internal/domain/usecase/shorturl/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestFindShortURL_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	type storageRes struct {
		shortURL *entity.ShortURL
	}

	tests := []struct {
		name       string
		alias      string
		storageRes storageRes
		res        string
	}{
		{
			name:       "when record exist in db",
			alias:      "alias1",
			storageRes: storageRes{shortURL: &entity.ShortURL{SourceURL: "https://ya.ru"}},
			res:        "https://ya.ru",
		},
		{
			name:       "when alias passed with '/' prefix",
			alias:      "/alias1",
			storageRes: storageRes{shortURL: &entity.ShortURL{SourceURL: "https://ya.ru"}},
			res:        "https://ya.ru",
		},
	}
	for _, tt := range tests {
		storage.EXPECT().FindByAlias("alias1").Return(tt.storageRes.shortURL, nil).AnyTimes()
		uc := NewShortURLUseCase(storage, "baseURL")

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.FindShortURL(tt.alias)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestFindShortURL_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	type storageRes struct {
		shortURL *entity.ShortURL
		err      error
	}

	tests := []struct {
		name       string
		alias      string
		storageRes storageRes
		err        error
	}{
		{
			name:  "when passed empty alias",
			alias: "",
			err:   ucErrors.ErrShortURLEmptyAlias,
		},
		{
			name:       "when source URL in db not found",
			alias:      "alias2",
			storageRes: storageRes{shortURL: nil},
			err:        ucErrors.ErrShortURLSourceURLNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.EXPECT().FindByAlias(tt.alias).Return(tt.storageRes.shortURL, tt.storageRes.err).AnyTimes()
			uc := NewShortURLUseCase(storage, "base")
			_, err := uc.FindShortURL(tt.alias)
			require.ErrorIs(t, tt.err, err)
		})
	}
}

func TestCreateShortURL_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	type storageRes struct {
		shortURL *entity.ShortURL
	}

	tests := []struct {
		name       string
		sourceURL  string
		baseURL    string
		storageRes storageRes
		res        string
	}{
		{
			name:       "when successfully stored short URL",
			sourceURL:  "https://ya.ru",
			baseURL:    "http://localhost:8888",
			storageRes: storageRes{shortURL: &entity.ShortURL{Alias: "alias"}},
			res:        "http://localhost:8888/alias",
		},
	}
	for _, tt := range tests {
		storage.EXPECT().Save(tt.sourceURL).Return(tt.storageRes.shortURL, nil)
		uc := NewShortURLUseCase(storage, tt.baseURL)

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.CreateShortURL(tt.sourceURL)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func TestCreateShortURL_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	type storageRes struct {
		shortURL *entity.ShortURL
		err      error
	}

	tests := []struct {
		name       string
		sourceURL  string
		baseURL    string
		storageRes storageRes
		err        error
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
		{
			name:      "when passed existing source URL",
			sourceURL: "http://ya.ru",
			baseURL:   "http://localhost:8888",
			storageRes: storageRes{
				shortURL: &entity.ShortURL{Alias: "alias"},
				err:      storageErrors.ErrStorageRecordIsNotUnique,
			},
			err: ucErrors.ErrShortURLAlreadyExist,
		},
	}
	for _, tt := range tests {
		storage.EXPECT().Save(tt.sourceURL).Return(tt.storageRes.shortURL, tt.storageRes.err).AnyTimes()
		uc := NewShortURLUseCase(storage, tt.baseURL)

		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.CreateShortURL(tt.sourceURL)
			require.ErrorIs(t, tt.err, err)
		})
	}
}

func TestBatchShortURLs_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	var urls []entity.BatchShortURLInput
	urls = append(urls,
		entity.BatchShortURLInput{CorrelationID: "1", OriginalURL: "https://ya.ru"},
		entity.BatchShortURLInput{CorrelationID: "2", OriginalURL: "https://ya.com"},
	)

	storage.EXPECT().Save(urls[0].OriginalURL).Return(&entity.ShortURL{Alias: "alias1"}, nil).Times(1)
	storage.EXPECT().Save(urls[1].OriginalURL).Return(&entity.ShortURL{Alias: "alias2"}, nil).Times(1)

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
		uc := NewShortURLUseCase(storage, tt.baseURL)

		t.Run(tt.name, func(t *testing.T) {
			res := uc.BatchShortURLs(tt.urls)
			require.Equal(t, tt.result, res)
		})
	}
}
