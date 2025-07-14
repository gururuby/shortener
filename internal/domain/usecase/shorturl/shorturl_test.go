package usecase

import (
	"context"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/errors"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/shorturl/errors"
	"github.com/gururuby/shortener/internal/domain/usecase/shorturl/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_FindShortURL_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockShortURLStorage(ctrl)
	ctx := context.Background()

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
		storage.EXPECT().FindShortURL(ctx, "alias1").Return(tt.storageRes.shortURL, nil).AnyTimes()
		uc := NewShortURLUseCase(storage, "baseURL")

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.FindShortURL(ctx, tt.alias)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func Test_FindShortURL_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockShortURLStorage(ctrl)
	ctx := context.Background()

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
			storage.EXPECT().FindShortURL(ctx, tt.alias).Return(tt.storageRes.shortURL, tt.storageRes.err).AnyTimes()
			uc := NewShortURLUseCase(storage, "base")
			_, err := uc.FindShortURL(ctx, tt.alias)
			require.ErrorIs(t, tt.err, err)
		})
	}
}

func Benchmark_FindShortURL(b *testing.B) {
	ctrl := gomock.NewController(b)
	storage := mocks.NewMockShortURLStorage(ctrl)
	ctx := context.Background()

	storage.EXPECT().FindShortURL(ctx, "alias").Return(&entity.ShortURL{}, nil).AnyTimes()
	uc := NewShortURLUseCase(storage, "baseURL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.FindShortURL(ctx, "alias")
	}
}

func Test_CreateShortURL_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockShortURLStorage(ctrl)
	ctx := context.Background()

	type storageRes struct {
		shortURL *entity.ShortURL
	}

	tests := []struct {
		name       string
		sourceURL  string
		user       *userEntity.User
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
		storage.EXPECT().SaveShortURL(ctx, nil, tt.sourceURL).Return(tt.storageRes.shortURL, nil)
		uc := NewShortURLUseCase(storage, tt.baseURL)

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.CreateShortURL(ctx, nil, tt.sourceURL)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func Test_CreateShortURL_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockShortURLStorage(ctrl)
	ctx := context.Background()

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
		storage.EXPECT().SaveShortURL(ctx, nil, tt.sourceURL).Return(tt.storageRes.shortURL, tt.storageRes.err).AnyTimes()
		uc := NewShortURLUseCase(storage, tt.baseURL)

		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.CreateShortURL(ctx, nil, tt.sourceURL)
			require.ErrorIs(t, tt.err, err)
		})
	}
}

func Benchmark_CreateShortURL(b *testing.B) {
	ctrl := gomock.NewController(b)
	storage := mocks.NewMockShortURLStorage(ctrl)
	ctx := context.Background()

	storage.EXPECT().SaveShortURL(ctx, nil, "https://example.com").Return(&entity.ShortURL{}, nil).AnyTimes()
	uc := NewShortURLUseCase(storage, "baseURL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.CreateShortURL(ctx, nil, "https://example.com")
	}
}

func Test_BatchShortURLs_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockShortURLStorage(ctrl)
	ctx := context.Background()

	var urls []entity.BatchShortURLInput
	urls = append(urls,
		entity.BatchShortURLInput{CorrelationID: "1", OriginalURL: "https://ya.ru"},
		entity.BatchShortURLInput{CorrelationID: "2", OriginalURL: "https://ya.com"},
	)

	storage.EXPECT().SaveShortURL(ctx, nil, urls[0].OriginalURL).Return(&entity.ShortURL{Alias: "alias1"}, nil).Times(1)
	storage.EXPECT().SaveShortURL(ctx, nil, urls[1].OriginalURL).Return(&entity.ShortURL{Alias: "alias2"}, nil).Times(1)

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
			res := uc.BatchShortURLs(ctx, tt.urls)
			require.Equal(t, tt.result, res)
		})
	}
}

func Benchmark_BatchShortURLs(b *testing.B) {
	ctrl := gomock.NewController(b)
	storage := mocks.NewMockShortURLStorage(ctrl)
	ctx := context.Background()

	var urls []entity.BatchShortURLInput
	urls = append(urls,
		entity.BatchShortURLInput{CorrelationID: "1", OriginalURL: "https://ya.ru"},
		entity.BatchShortURLInput{CorrelationID: "2", OriginalURL: "https://ya.com"},
	)

	storage.EXPECT().SaveShortURL(ctx, nil, urls[0].OriginalURL).Return(&entity.ShortURL{Alias: "alias1"}, nil).AnyTimes()
	storage.EXPECT().SaveShortURL(ctx, nil, urls[1].OriginalURL).Return(&entity.ShortURL{Alias: "alias2"}, nil).AnyTimes()

	uc := NewShortURLUseCase(storage, "baseURL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uc.BatchShortURLs(ctx, urls)
	}
}
