package app

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gururuby/shortener/config"
	"github.com/gururuby/shortener/internal/domain/dao"
	"github.com/gururuby/shortener/internal/domain/entity"
	"github.com/gururuby/shortener/internal/domain/usecase"
	httpHandler "github.com/gururuby/shortener/internal/handler/http"
	apiHandler "github.com/gururuby/shortener/internal/handler/http/api"
	"github.com/gururuby/shortener/internal/infra/compress"
	fileDB "github.com/gururuby/shortener/internal/infra/db/file"
	memoryDB "github.com/gururuby/shortener/internal/infra/db/memory"
	nullDB "github.com/gururuby/shortener/internal/infra/db/null"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/internal/infra/utils/generator"
	"log"
	"net/http"
)

type DAO interface {
	FindByAlias(alias string) (*entity.ShortURL, error)
	Save(sourceURL string) (*entity.ShortURL, error)
}

type DB interface {
	Find(string) (*entity.ShortURL, error)
	Save(*entity.ShortURL) (*entity.ShortURL, error)
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

	cfg, setupErr = config.New()
	if setupErr != nil {
		log.Fatalf("cannot setup config: %s", setupErr)
	}

	if setupErr = logger.Initialize(cfg.App.Env); setupErr != nil {
		log.Fatalf("cannot setup logger: %s", setupErr)
	}

	gen := generator.New(cfg.App.AliasLength)

	db, setupErr = setupDB(cfg)
	if setupErr != nil {
		log.Fatalf("cannot setup database: %s", setupErr)
	}

	storage = dao.New(gen, cfg, db)

	router := chi.NewRouter()
	router.Use(logger.HandlerMiddleware)
	router.Use(compress.HandlerMiddleware)

	uc := usecase.NewUseCase(storage, cfg.App.BaseURL)

	httpHandler.Register(router, uc)
	apiHandler.Register(router, uc)

	return &App{Storage: storage, Config: cfg, Router: router}
}

func setupDB(cfg *config.Config) (DB, error) {
	var db DB
	var err error

	switch cfg.DB.Type {
	case "memory":
		db = memoryDB.New()
	case "file":
		if db, err = fileDB.New(cfg.FileStorage.Path); err != nil {
			log.Fatalf("cannot setup file DB: %s", err)
		}
	default:
		db = nullDB.New()
	}
	return db, err
}

func (a *App) Run() {
	logger.Log.Info(fmt.Sprintf("Starting %s server on %s", a.Config.AppInfo(), a.Config.Server.Address))
	log.Fatal(http.ListenAndServe(a.Config.Server.Address, a.Router))
}
