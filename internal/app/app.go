package app

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gururuby/shortener/config"
	"github.com/gururuby/shortener/internal/domain/dao"
	"github.com/gururuby/shortener/internal/domain/entity"
	ucApp "github.com/gururuby/shortener/internal/domain/usecase/app"
	ucShortURL "github.com/gururuby/shortener/internal/domain/usecase/shorturl"
	handlerAPI "github.com/gururuby/shortener/internal/handler/http/api"
	handlerApp "github.com/gururuby/shortener/internal/handler/http/app"
	handlerShortURL "github.com/gururuby/shortener/internal/handler/http/shorturl"
	fileDB "github.com/gururuby/shortener/internal/infra/db/file"
	memoryDB "github.com/gururuby/shortener/internal/infra/db/memory"
	nullDB "github.com/gururuby/shortener/internal/infra/db/null"
	postgresqlDB "github.com/gururuby/shortener/internal/infra/db/postgresql"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/internal/infra/utils/generator"
	"github.com/gururuby/shortener/internal/middleware"
	"log"
	"net/http"
)

type DAO interface {
	FindByAlias(alias string) (*entity.ShortURL, error)
	Save(sourceURL string) (*entity.ShortURL, error)
	IsDBReady() error
	Clear()
}

type DB interface {
	Find(string) (*entity.ShortURL, error)
	Save(*entity.ShortURL) (*entity.ShortURL, error)
	Ping() error
	Truncate()
}

type Router interface {
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}

type App struct {
	Storage DAO
	Config  *config.Config
	Router  Router
}

func New(cfg *config.Config) *App {
	return &App{Config: cfg}
}

func (a *App) Setup() *App {
	var setupErr error
	var storage DAO
	var db DB

	ctx := context.Background()

	logger.Initialize(a.Config.App.Env, a.Config.Log.Level)

	gen := generator.New(a.Config.App.AliasLength)

	db, setupErr = setupDB(ctx, a.Config)
	if setupErr != nil {
		log.Fatalf("cannot setup database: %s", setupErr)
	}

	storage = dao.New(gen, db)

	router := chi.NewRouter()
	router.Use(middleware.Logging)
	router.Use(middleware.Compression)

	shortURLUseCase := ucShortURL.NewShortURLUseCase(storage, a.Config.App.BaseURL)
	appUseCase := ucApp.NewAppUseCase(storage)

	handlerShortURL.Register(router, shortURLUseCase)
	handlerApp.Register(router, appUseCase)
	handlerAPI.Register(router, shortURLUseCase)

	a.Storage = storage
	a.Router = router

	return a
}

func (a *App) Run() {
	logger.Log.Info(fmt.Sprintf("Starting %s server on %s", a.Config.AppInfo(), a.Config.Server.Address))
	log.Fatal(http.ListenAndServe(a.Config.Server.Address, a.Router))
}

func setupDB(ctx context.Context, cfg *config.Config) (db DB, err error) {
	switch cfg.Database.Type {
	case "memory":
		db = memoryDB.New()
	case "file":
		if db, err = fileDB.New(cfg.FileStorage.Path); err != nil {
			log.Fatalf("cannot setup file DB: %s", err)
		}
	case "postgresql":
		if db, err = postgresqlDB.New(ctx, cfg); err != nil {
			log.Fatalf("cannot setup postgresql DB: %s", err)
		}
	default:
		db = nullDB.New()
	}
	return
}
