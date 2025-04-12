package app

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gururuby/shortener/config"
	"github.com/gururuby/shortener/internal/domain/dao"
	"github.com/gururuby/shortener/internal/domain/usecase"
	httpHandler "github.com/gururuby/shortener/internal/handler/http"
	apiHandler "github.com/gururuby/shortener/internal/handler/http/api"
	memoryDB "github.com/gururuby/shortener/internal/infra/db/memory"
	nullDB "github.com/gururuby/shortener/internal/infra/db/null"
	"github.com/gururuby/shortener/internal/infra/logger"
	"log"
	"net/http"
)

type DAO interface {
	FindByAlias(alias string) (string, error)
	Save(sourceURL string) (string, error)
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
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("cannot setup config: %s", err)
	}

	if err = logger.Initialize(cfg.App.Env); err != nil {
		log.Fatalf("cannot setup logger: %s", err)
	}

	var storage DAO

	switch cfg.DB.Type {
	case "memory":
		storage = dao.New(memoryDB.New())
	default:
		storage = dao.New(nullDB.New())
	}

	router := chi.NewRouter()
	router.Use(logger.HandlerMiddleware)

	uc := usecase.NewUseCase(storage, cfg.App.BaseURL)

	httpHandler.Register(router, uc)
	apiHandler.Register(router, uc)

	return &App{Storage: storage, Config: cfg, Router: router}
}

func (a *App) Run() {
	logger.Log.Info(fmt.Sprintf("Starting %s server on %s", a.Config.AppInfo(), a.Config.Server.Address))
	log.Fatal(http.ListenAndServe(a.Config.Server.Address, a.Router))
}
