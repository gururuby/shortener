package app

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gururuby/shortener/config"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	"github.com/gururuby/shortener/internal/domain/storage/shorturl"
	appUseCase "github.com/gururuby/shortener/internal/domain/usecase/app"
	shortURLUseCase "github.com/gururuby/shortener/internal/domain/usecase/shorturl"
	apiHandler "github.com/gururuby/shortener/internal/handler/http/api"
	appHandler "github.com/gururuby/shortener/internal/handler/http/app"
	shortURLHandler "github.com/gururuby/shortener/internal/handler/http/shorturl"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/internal/middleware"
	"log"
	"net/http"
)

type Storage interface {
	FindByAlias(alias string) (*entity.ShortURL, error)
	Save(sourceURL string) (*entity.ShortURL, error)
	IsDBReady() error
	Clear()
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
	var stg Storage

	ctx := context.Background()

	logger.Initialize(a.Config.App.Env, a.Config.Log.Level)

	stg, setupErr = storage.Setup(ctx, a.Config)

	if setupErr != nil {
		log.Fatalf("cannot setup storage: %s", setupErr)
	}

	router := chi.NewRouter()
	router.Use(middleware.Logging)
	router.Use(middleware.Compression)

	shortURLUC := shortURLUseCase.NewShortURLUseCase(stg, a.Config.App.BaseURL)
	appUC := appUseCase.NewAppUseCase(stg)

	shortURLHandler.Register(router, shortURLUC)
	appHandler.Register(router, appUC)
	apiHandler.Register(router, shortURLUC)

	a.Storage = stg
	a.Router = router

	return a
}

func (a *App) Run() {
	welcomeMsg := fmt.Sprintf("Starting %s server on %s", a.Config.AppInfo(), a.Config.Server.Address)
	logger.Log.Info(welcomeMsg)
	log.Fatal(http.ListenAndServe(a.Config.Server.Address, a.Router))
}
