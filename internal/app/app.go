/*
Package app provides the core application setup and runtime for the URL shortener service.
*/
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
	"github.com/gururuby/shortener/internal/infra/server"
)

// Router defines the interface for HTTP request routing.
type Router interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// DB defines the interface for interaction with database layer
type DB interface {
	// Shutdown allows to gracefully shutdown database
	Shutdown(context.Context) error
}

// ShortURLStorage defines the interface for short URL persistence operations.
type ShortURLStorage interface {
	FindShortURL(ctx context.Context, alias string) (*entity.ShortURL, error)
	SaveShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (*entity.ShortURL, error)
	IsDBReady(ctx context.Context) error
}

// UserStorage defines the interface for user persistence operations.
type UserStorage interface {
	FindUser(ctx context.Context, userID int) (*userEntity.User, error)
	FindURLs(ctx context.Context, userID int) ([]*entity.ShortURL, error)
	SaveUser(ctx context.Context) (*userEntity.User, error)
	MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error
}

// App represents the main application container with all dependencies.
type App struct {
	ShortURLSStorage ShortURLStorage
	UserStorage      UserStorage
	Config           *config.Config
	Router           Router
	DB               DB
}

// New creates a new App instance with the given configuration.
func New(cfg *config.Config) *App {
	return &App{Config: cfg}
}

// Setup initializes all application dependencies in the correct order.
func (a *App) Setup() *App {
	ctx := context.Background()
	logger.Setup(a.Config.App.Env, a.Config.Log.Level)

	db, err := database.Setup(ctx, a.Config)
	if err != nil {
		log.Fatalf("cannot setup database: %s", err)
	}

	shortURLStg := shortURLStorage.Setup(db, a.Config)
	userStg := userStorage.Setup(db)
	r := router.Setup()
	auth := jwt.New(a.Config.Auth.SecretKey, a.Config.Auth.TokenTTL)

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
	a.DB = db

	return a
}

// Run starts the application server.
func (a *App) Run() {
	a.printWelcomeMessage()
	server.New(a.Router, a.Config, a.DB).Run()
}

func (a *App) printWelcomeMessage() {
	welcomeMsg := fmt.Sprintf("Starting %s server on %s",
		a.Config.AppInfo(),
		a.Config.Server.Address)
	logger.Log.Info(welcomeMsg)
}
