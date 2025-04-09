package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/gururuby/shortener/config"
	"github.com/gururuby/shortener/internal/domain/dao"
	"github.com/gururuby/shortener/internal/domain/usecase"
	httpHandler "github.com/gururuby/shortener/internal/handler/http"
	"github.com/gururuby/shortener/internal/infra/db/memory"
	"github.com/gururuby/shortener/internal/infra/db/null"
	"github.com/gururuby/shortener/internal/infra/logger"
	"log"
	"net/http"
)

const (
	shortURLCreatePath = "/"
	shortURLFindPath   = "/{alias}"
)

type shortURLDAO interface {
	FindByAlias(alias string) (string, error)
	Save(sourceURL string) (string, error)
}

type App struct {
	dao    shortURLDAO
	router chi.Router
	cfg    *config.Config
}

func NewApp() App {
	app := App{}
	app.setupConfig()
	app.setupLogger()
	app.setupDAO()
	app.setupRouter()
	app.setupHandler()
	return app
}

func (a *App) Run() {
	log.Fatal(http.ListenAndServe(a.cfg.Server.Address, a.router))
}

func (a *App) setupConfig() {
	cfg, err := config.New()

	if err != nil {
		log.Fatalf("cannot setup config: %s", err)
	}

	a.cfg = cfg
}

func (a *App) setupLogger() {
	err := logger.Initialize(a.cfg.App.Env)
	if err != nil {
		log.Fatalf("cannot setup logger: %s", err)
	}
}

func (a *App) setupDAO() {
	switch a.cfg.DB.Type {
	case "memory":
		a.dao = dao.NewShortURLDAO(memory.NewShortURLDB())
	default:
		a.dao = dao.NewShortURLDAO(null.NewShortURLDB())
	}
}

func (a *App) setupRouter() {
	router := chi.NewRouter()
	router.Use(logger.HandlerMiddleware)
	a.router = router
}

func (a *App) setupHandler() {
	uc := usecase.NewShortURLUseCase(a.dao, a.cfg.App.BaseURL)
	handler := httpHandler.NewShortURLHandler(uc)

	a.router.Post(shortURLCreatePath, handler.CreateShortURL())
	a.router.Get(shortURLFindPath, handler.FindShortURL())
}
