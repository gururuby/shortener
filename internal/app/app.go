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
}

type DB interface {
	Find(string) (*entity.ShortURL, error)
	Save(*entity.ShortURL) (*entity.ShortURL, error)
	Ping() error
}

type Router interface {
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}

type App struct {
	Storage DAO
	Config  *config.Config
	Router  Router
}

func Setup() *App {
	var setupErr error
	var cfg *config.Config
	var storage DAO
	var db DB

	ctx := context.Background()

	cfg, setupErr = config.New()
	if setupErr != nil {
		log.Fatalf("cannot setup config: %s", setupErr)
	}

	logger.Initialize(cfg.App.Env, cfg.Log.Level)

	gen := generator.New(cfg.App.AliasLength)

	db, setupErr = setupDB(ctx, cfg)
	if setupErr != nil {
		log.Fatalf("cannot setup database: %s", setupErr)
	}

	storage = dao.New(gen, cfg, db)

	router := chi.NewRouter()
	router.Use(middleware.Logging)
	router.Use(middleware.Compression)

	shortURLUseCase := ucShortURL.NewShortURLUseCase(storage, cfg.App.BaseURL)
	appUseCase := ucApp.NewAppUseCase(storage)

	handlerShortURL.Register(router, shortURLUseCase)
	handlerApp.Register(router, appUseCase)
	handlerAPI.Register(router, shortURLUseCase)

	return &App{Storage: storage, Config: cfg, Router: router}
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

func (a *App) Run() {
	logger.Log.Info(fmt.Sprintf("Starting %s server on %s", a.Config.AppInfo(), a.Config.Server.Address))
	log.Fatal(http.ListenAndServe(a.Config.Server.Address, a.Router))
}
