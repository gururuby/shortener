//go:generate mockgen -destination=./mocks/mock.go -package=mocks . UserStorage,Authenticator

package usecase

import (
	"context"
	"errors"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/user/errors"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
)

type UserStorage interface {
	FindUser(ctx context.Context, userID int) (*userEntity.User, error)
	FindURLs(ctx context.Context, userID int) ([]*shortURLEntity.ShortURL, error)
	SaveUser(ctx context.Context) (*userEntity.User, error)
}

type Authenticator interface {
	SignUserID(userID int) (string, error)
	ReadUserID(tokenString string) (int, error)
}

type UserUseCase struct {
	auth    Authenticator
	storage UserStorage
	baseURL string
}

type UserShortURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewUserUseCase(auth Authenticator, storage UserStorage, baseURL string) *UserUseCase {
	return &UserUseCase{
		auth:    auth,
		storage: storage,
		baseURL: baseURL,
	}
}

func (u *UserUseCase) Authenticate(ctx context.Context, token string) (*userEntity.User, error) {
	var (
		userID int
		user   *userEntity.User
		err    error
	)

	if userID, err = u.auth.ReadUserID(token); err != nil {
		return nil, ucErrors.ErrUserCannotAuthenticate
	}

	if user, err = u.storage.FindUser(ctx, userID); err != nil {
		return nil, ucErrors.ErrUserNotFound
	}

	user.AuthToken = token
	return user, nil
}

func (u *UserUseCase) Register(ctx context.Context) (*userEntity.User, error) {
	var (
		user  *userEntity.User
		token string
		err   error
	)

	if user, err = u.storage.SaveUser(ctx); err != nil {
		return nil, ucErrors.ErrUserCannotRegister
	}

	if token, err = u.auth.SignUserID(user.ID); err != nil {
		return nil, ucErrors.ErrUserCannotRegister
	}

	user.AuthToken = token

	return user, nil
}

func (u *UserUseCase) SaveUser(ctx context.Context) (*userEntity.User, error) {
	user, err := u.storage.SaveUser(ctx)
	if err != nil {
		return nil, ucErrors.ErrUserCannotSave
	}
	return user, nil
}

func (u *UserUseCase) FindUser(ctx context.Context, id int) (*userEntity.User, error) {
	user, err := u.storage.FindUser(ctx, id)
	if err != nil {
		if errors.Is(err, dbErrors.ErrDBRecordNotFound) {
			return nil, ucErrors.ErrUserNotFound
		}
		return nil, ucErrors.ErrUserStorageNotWorking
	}
	return user, nil
}

func (u *UserUseCase) GetURLs(ctx context.Context, user *userEntity.User) ([]*UserShortURL, error) {
	var (
		shortURLs []*shortURLEntity.ShortURL
		userURLs  []*UserShortURL
		err       error
	)

	if shortURLs, err = u.storage.FindURLs(ctx, user.ID); err != nil {
		return nil, ucErrors.ErrUserStorageNotWorking
	}

	for _, shortURL := range shortURLs {
		userURLs = append(userURLs, &UserShortURL{
			ShortURL:    u.baseURL + "/" + shortURL.Alias,
			OriginalURL: shortURL.SourceURL,
		})
	}

	return userURLs, nil
}
