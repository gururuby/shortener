package usecase

import (
	"context"
	"testing"

	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/errors"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/user/errors"
	"github.com/gururuby/shortener/internal/domain/usecase/user/mocks"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	jwtErrors "github.com/gururuby/shortener/internal/infra/jwt/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Authenticate_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockUserStorage(ctrl)
	auth := mocks.NewMockAuthenticator(ctrl)
	ctx := context.Background()

	type storageRes struct {
		user *userEntity.User
	}

	tests := []struct {
		storageRes storageRes
		res        *userEntity.User
		name       string
		token      string
		ID         int
	}{
		{
			name:       "when read ID from token and record exist in db",
			ID:         1,
			storageRes: storageRes{user: &userEntity.User{ID: 1}},
			res:        &userEntity.User{ID: 1},
		},
	}
	for _, tt := range tests {
		auth.EXPECT().ReadUserID(tt.token).Return(tt.ID, nil)
		storage.EXPECT().FindUser(ctx, tt.ID).Return(tt.storageRes.user, nil).AnyTimes()
		uc := NewUserUseCase(auth, storage, "http://localhost:8080")

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.Authenticate(ctx, tt.token)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func Test_Authenticate_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockUserStorage(ctrl)
	auth := mocks.NewMockAuthenticator(ctrl)
	ctx := context.Background()

	type (
		authRes struct {
			err    error
			userID int
		}
		storageRes struct {
			user *userEntity.User
			err  error
		}
	)

	tests := []struct {
		authRes    authRes
		storageRes storageRes
		res        error
		name       string
		token      string
	}{
		{
			name:       "when authenticator return some error",
			token:      "token",
			authRes:    authRes{userID: 0, err: jwtErrors.ErrJWTUnexpectedSigningMethod},
			storageRes: storageRes{nil, nil},
			res:        ucErrors.ErrUserCannotAuthenticate,
		},
		{
			name:       "when authenticator is ok but record is not found",
			token:      "token",
			authRes:    authRes{userID: 1, err: nil},
			storageRes: storageRes{nil, storageErrors.ErrStorageRecordNotFound},
			res:        ucErrors.ErrUserNotFound,
		},
	}
	for _, tt := range tests {
		auth.EXPECT().ReadUserID(tt.token).Return(tt.authRes.userID, tt.authRes.err).AnyTimes()
		storage.EXPECT().FindUser(ctx, tt.authRes).Return(tt.storageRes.user, tt.storageRes.err).AnyTimes()
		uc := NewUserUseCase(auth, storage, "http://localhost:8080")

		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Authenticate(ctx, tt.token)
			require.Error(t, err, tt.res)
		})
	}
}

func Test_Register_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockUserStorage(ctrl)
	auth := mocks.NewMockAuthenticator(ctrl)
	ctx := context.Background()

	type (
		authRes struct {
			err   error
			token string
		}
		storageRes struct {
			user *userEntity.User
		}
	)

	tests := []struct {
		authRes    authRes
		storageRes storageRes
		res        *userEntity.User
		name       string
		token      string
	}{
		{
			name:       "when successfully create user and sign user ID",
			authRes:    authRes{token: "token", err: nil},
			storageRes: storageRes{user: &userEntity.User{ID: 1}},
			res:        &userEntity.User{AuthToken: "token", ID: 1},
		},
	}
	for _, tt := range tests {
		storage.EXPECT().SaveUser(ctx).Return(tt.storageRes.user, nil).Times(1)
		auth.EXPECT().SignUserID(tt.storageRes.user.ID).Return(tt.authRes.token, nil).Times(1)
		uc := NewUserUseCase(auth, storage, "http://localhost:8080")

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.Register(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func Test_Register_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockUserStorage(ctrl)
	auth := mocks.NewMockAuthenticator(ctrl)
	ctx := context.Background()

	type (
		authRes struct {
			err   error
			token string
		}
		storageRes struct {
			user *userEntity.User
			err  error
		}
	)

	tests := []struct {
		authRes    authRes
		storageRes storageRes
		res        error
		name       string
		token      string
	}{
		{
			name:       "when cannot create user in db",
			authRes:    authRes{token: "", err: nil},
			storageRes: storageRes{user: nil, err: storageErrors.ErrStorageIsNotReadyDB},
			res:        ucErrors.ErrUserCannotRegister,
		},
		{
			name:       "when authenticator cannot sign user ID",
			authRes:    authRes{token: "", err: jwtErrors.ErrJWTCannotSignData},
			storageRes: storageRes{user: &userEntity.User{ID: 1}, err: nil},
			res:        ucErrors.ErrUserCannotRegister,
		},
	}
	for _, tt := range tests {
		storage.EXPECT().SaveUser(ctx).Return(tt.storageRes.user, tt.storageRes.err).Times(1)
		if tt.storageRes.err == nil {
			auth.EXPECT().SignUserID(tt.storageRes.user.ID).Return(tt.authRes.token, tt.authRes.err).Times(1)
		}

		uc := NewUserUseCase(auth, storage, "http://localhost:8080")

		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Register(ctx)
			require.Error(t, err, tt.res)
		})
	}
}

func Test_FindUser_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockUserStorage(ctrl)
	auth := mocks.NewMockAuthenticator(ctrl)
	ctx := context.Background()

	type storageRes struct {
		user *userEntity.User
	}

	tests := []struct {
		storageRes storageRes
		res        *userEntity.User
		name       string
		ID         int
	}{
		{
			name:       "when record exist in db",
			ID:         1,
			storageRes: storageRes{user: &userEntity.User{ID: 1}},
			res:        &userEntity.User{ID: 1},
		},
	}
	for _, tt := range tests {
		storage.EXPECT().FindUser(ctx, tt.ID).Return(tt.storageRes.user, nil).AnyTimes()
		uc := NewUserUseCase(auth, storage, "http://localhost:8080")

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.FindUser(ctx, tt.ID)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func Test_FindUser_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockUserStorage(ctrl)
	auth := mocks.NewMockAuthenticator(ctrl)
	ctx := context.Background()

	type storageRes struct {
		user *userEntity.User
		err  error
	}

	tests := []struct {
		storageRes storageRes
		err        error
		name       string
		ID         int
	}{
		{
			name:       "when passed unknown ID",
			ID:         0,
			storageRes: storageRes{user: nil, err: dbErrors.ErrDBRecordNotFound},
			err:        ucErrors.ErrUserNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.EXPECT().FindUser(ctx, tt.ID).Return(tt.storageRes.user, tt.storageRes.err).AnyTimes()
			uc := NewUserUseCase(auth, storage, "http://localhost:8080")
			_, err := uc.FindUser(ctx, tt.ID)
			require.ErrorIs(t, tt.err, err)
		})
	}
}

func Test_SaveUser_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockUserStorage(ctrl)
	auth := mocks.NewMockAuthenticator(ctrl)
	ctx := context.Background()

	type storageRes struct {
		user *userEntity.User
	}

	tests := []struct {
		storageRes storageRes
		res        *userEntity.User
		name       string
	}{
		{
			name:       "when successfully stored user",
			storageRes: storageRes{user: &userEntity.User{ID: 1}},
			res:        &userEntity.User{ID: 1},
		},
	}
	for _, tt := range tests {
		storage.EXPECT().SaveUser(ctx).Return(tt.storageRes.user, nil)
		uc := NewUserUseCase(auth, storage, "http://localhost:8080")

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.SaveUser(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.res, res)
		})
	}
}

func Test_SaveUser_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockUserStorage(ctrl)
	auth := mocks.NewMockAuthenticator(ctrl)
	ctx := context.Background()

	type storageRes struct {
		user *userEntity.User
		err  error
	}

	tests := []struct {
		storageRes storageRes
		err        error
		name       string
	}{
		{
			name: "when something wrong with DB",
			storageRes: storageRes{
				user: nil,
				err:  storageErrors.ErrStorageIsNotReadyDB,
			},
			err: ucErrors.ErrUserCannotSave,
		},
	}
	for _, tt := range tests {
		storage.EXPECT().SaveUser(ctx).Return(tt.storageRes.user, tt.storageRes.err).AnyTimes()
		uc := NewUserUseCase(auth, storage, "http://localhost:8080")

		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.SaveUser(ctx)
			require.ErrorIs(t, tt.err, err)
		})
	}
}

func Test_GetURLs_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockUserStorage(ctrl)
	auth := mocks.NewMockAuthenticator(ctrl)
	ctx := context.Background()

	urls := make([]*shortURLEntity.ShortURL, 0)
	urls = append(urls, &shortURLEntity.ShortURL{Alias: "alias", SourceURL: "https://ya.ru"})

	userURLs := make([]*UserShortURL, 0)
	userURLs = append(userURLs, &UserShortURL{
		OriginalURL: "https://ya.ru",
		ShortURL:    "http://localhost:8080/alias",
	})

	type storageRes struct {
		err  error
		urls []*shortURLEntity.ShortURL
	}

	tests := []struct {
		name       string
		storageRes storageRes
		res        []*UserShortURL
	}{
		{
			name:       "when success authenticate user and get their urls",
			storageRes: storageRes{urls: urls, err: nil},
			res:        userURLs,
		},
	}
	for _, tt := range tests {
		storage.EXPECT().FindURLs(ctx, 1).Return(tt.storageRes.urls, tt.storageRes.err).Times(1)
		uc := NewUserUseCase(auth, storage, "http://localhost:8080")

		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.GetURLs(ctx, &userEntity.User{ID: 1})
			require.NoError(t, err)
			require.ElementsMatch(t, tt.res, res)
		})
	}
}

func Test_GetURLs_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockUserStorage(ctrl)
	auth := mocks.NewMockAuthenticator(ctrl)
	ctx := context.Background()

	type (
		storageRes struct {
			err  error
			urls []*shortURLEntity.ShortURL
		}
	)

	tests := []struct {
		res        error
		name       string
		token      string
		storageRes storageRes
	}{
		{
			name:       "when something went wrong with storage",
			storageRes: storageRes{urls: nil, err: storageErrors.ErrStorageIsNotReadyDB},
			res:        ucErrors.ErrUserStorageNotWorking,
		},
	}
	for _, tt := range tests {
		storage.EXPECT().FindURLs(ctx, 1).Return(tt.storageRes.urls, tt.storageRes.err).AnyTimes()
		uc := NewUserUseCase(auth, storage, "http://localhost:8080")

		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.GetURLs(ctx, &userEntity.User{ID: 1})
			require.Error(t, err, tt.res)
		})
	}
}
