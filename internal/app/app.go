package app

import (
	"context"
	"fmt"
	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	shortURLStorage "github.com/gururuby/shortener/internal/domain/storage/shorturl"
	userStorage "github.com/gururuby/shortener/internal/domain/storage/user"
	appUseCase "github.com/gururuby/shortener/internal/domain/usecase/app"
	shortURLUseCase "github.com/gururuby/shortener/internal/domain/usecase/shorturl"
	userUseCase "github.com/gururuby/shortener/internal/domain/usecase/user"
	apiShortURLHandler "github.com/gururuby/shortener/internal/handler/http/api/shorturl"
	apiUserHandler "github.com/gururuby/shortener/internal/handler/http/api/user"
	appHandler "github.com/gururuby/shortener/internal/handler/http/app"
	shortURLHandler "github.com/gururuby/shortener/internal/handler/http/shorturl"
	database "github.com/gururuby/shortener/internal/infra/db"
	"github.com/gururuby/shortener/internal/infra/jwt"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/internal/infra/router"
	"log"
	"net/http"
)

type ShortURLStorage interface {
	FindShortURL(ctx context.Context, alias string) (*entity.ShortURL, error)
	SaveShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (*entity.ShortURL, error)
	IsDBReady(ctx context.Context) error
}

type UserStorage interface {
	FindUser(ctx context.Context, userID int) (*userEntity.User, error)
	FindURLs(ctx context.Context, userID int) ([]*entity.ShortURL, error)
	SaveUser(ctx context.Context) (*userEntity.User, error)
}

type Router interface {
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}

type App struct {
	ShortURLSStorage ShortURLStorage
	UserStorage      UserStorage
	Config           *config.Config
	Router           Router
}

func New(cfg *config.Config) *App {
	return &App{Config: cfg}
}

func (a *App) Setup() *App {
	var (
		auth        *jwt.JWT
		setupErr    error
		shortURLStg ShortURLStorage
		userStg     UserStorage
		db          database.DB
	)

	ctx := context.Background()

	logger.Setup(a.Config.App.Env, a.Config.Log.Level)

	db, setupErr = database.Setup(ctx, a.Config)
	if setupErr != nil {
		log.Fatalf("cannot setup database: %s", setupErr)
	}

	shortURLStg = shortURLStorage.Setup(db, a.Config)
	userStg = userStorage.Setup(db)

	r := router.Setup()

	auth = jwt.New(a.Config.Auth.SecretKey, a.Config.Auth.TokenTTL)
	userUC := userUseCase.NewUserUseCase(auth, userStg, a.Config.App.BaseURL)

	urlUC := shortURLUseCase.NewShortURLUseCase(shortURLStg, a.Config.App.BaseURL)
	appUC := appUseCase.NewAppUseCase(shortURLStg)

	shortURLHandler.Register(r, urlUC, userUC)
	appHandler.Register(r, appUC)
	apiShortURLHandler.Register(r, userUC, urlUC)
	apiUserHandler.Register(r, userUC)

	a.ShortURLSStorage = shortURLStg
	a.UserStorage = userStg
	a.Router = r

	return a
}

func (a *App) Run() {
	welcomeMsg := fmt.Sprintf("Starting %s server on %s", a.Config.AppInfo(), a.Config.Server.Address)
	logger.Log.Info(welcomeMsg)
	log.Fatal(http.ListenAndServe(a.Config.Server.Address, a.Router))
}
