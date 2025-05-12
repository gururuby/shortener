package app

import (
	"context"
	"fmt"
	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	shortURLStorage "github.com/gururuby/shortener/internal/domain/storage/shorturl"
	appUseCase "github.com/gururuby/shortener/internal/domain/usecase/app"
	shortURLUseCase "github.com/gururuby/shortener/internal/domain/usecase/shorturl"
	apiHandler "github.com/gururuby/shortener/internal/handler/http/api"
	appHandler "github.com/gururuby/shortener/internal/handler/http/app"
	shortURLHandler "github.com/gururuby/shortener/internal/handler/http/shorturl"
	database "github.com/gururuby/shortener/internal/infra/db"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/internal/infra/router"
	"log"
	"net/http"
)

type Storage interface {
	FindByAlias(ctx context.Context, alias string) (*entity.ShortURL, error)
	Save(ctx context.Context, sourceURL string) (*entity.ShortURL, error)
	IsDBReady(ctx context.Context) error
}

type Router interface {
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}

type App struct {
	Storage Storage
	Config  *config.Config
	Router  Router
}

func New(cfg *config.Config) *App {
	return &App{Config: cfg}
}

func (a *App) Setup() *App {
	var setupErr error
	var shortURLStg Storage
	var db database.DB

	ctx := context.Background()

	logger.Setup(a.Config.App.Env, a.Config.Log.Level)

	db, setupErr = database.Setup(ctx, a.Config)
	if setupErr != nil {
		log.Fatalf("cannot setup database: %s", setupErr)
	}

	shortURLStg = shortURLStorage.Setup(db, a.Config)

	r := router.Setup()

	shortURLUC := shortURLUseCase.NewShortURLUseCase(shortURLStg, a.Config.App.BaseURL)
	appUC := appUseCase.NewAppUseCase(shortURLStg)

	shortURLHandler.Register(r, shortURLUC)
	appHandler.Register(r, appUC)
	apiHandler.Register(r, shortURLUC)

	a.Storage = shortURLStg
	a.Router = r

	return a
}

func (a *App) Run() {
	welcomeMsg := fmt.Sprintf("Starting %s server on %s", a.Config.AppInfo(), a.Config.Server.Address)
	logger.Log.Info(welcomeMsg)
	log.Fatal(http.ListenAndServe(a.Config.Server.Address, a.Router))
}
