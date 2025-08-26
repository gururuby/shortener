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
	statsStorage "github.com/gururuby/shortener/internal/domain/storage/stats"
	userStorage "github.com/gururuby/shortener/internal/domain/storage/user"
	appUseCase "github.com/gururuby/shortener/internal/domain/usecase/app"
	shortURLUseCase "github.com/gururuby/shortener/internal/domain/usecase/shorturl"
	statsUseCase "github.com/gururuby/shortener/internal/domain/usecase/stats"
	userUseCase "github.com/gururuby/shortener/internal/domain/usecase/user"
	appGRPC "github.com/gururuby/shortener/internal/grpc/app"
	grpcRegistry "github.com/gururuby/shortener/internal/grpc/registry"
	shorturlGRPC "github.com/gururuby/shortener/internal/grpc/shorturl"
	statsGRPC "github.com/gururuby/shortener/internal/grpc/stats"
	userGRPC "github.com/gururuby/shortener/internal/grpc/user"
	apiStatsHandler "github.com/gururuby/shortener/internal/handler/http/api/internal_stats"
	apiShortURLHandler "github.com/gururuby/shortener/internal/handler/http/api/shorturl"
	apiUserHandler "github.com/gururuby/shortener/internal/handler/http/api/user"
	appHandler "github.com/gururuby/shortener/internal/handler/http/app"
	shortURLHandler "github.com/gururuby/shortener/internal/handler/http/shorturl"
	"github.com/gururuby/shortener/internal/infra/server"
	grpcServer "github.com/gururuby/shortener/internal/infra/server/grpc"

	database "github.com/gururuby/shortener/internal/infra/db"
	"github.com/gururuby/shortener/internal/infra/jwt"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/internal/infra/router"
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
	serviceRegistry  *grpcRegistry.ServiceRegistry
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
	statsStg := statsStorage.Setup(db)

	r := router.Setup(a.Config.Server)
	auth := jwt.New(a.Config.Auth.SecretKey, a.Config.Auth.TokenTTL)

	userUC := userUseCase.NewUserUseCase(auth, userStg, a.Config.App.BaseURL)
	urlUC := shortURLUseCase.NewShortURLUseCase(shortURLStg, a.Config.App.BaseURL)
	statsUC := statsUseCase.NewStatsUseCase(statsStg)
	appUC := appUseCase.NewAppUseCase(shortURLStg)

	// Register HTTP handlers
	shortURLHandler.Register(r, urlUC, userUC)
	appHandler.Register(r, appUC)
	apiShortURLHandler.Register(r, userUC, urlUC)
	apiUserHandler.Register(r, userUC)
	apiStatsHandler.Register(r, statsUC)

	// Register GRPC servers
	appServer := appGRPC.NewServer(appUC)
	statsServer := statsGRPC.NewServer(statsUC)
	userServer := userGRPC.NewServer(userUC)
	shortURLServer := shorturlGRPC.NewServer(urlUC, userUC)

	serviceRegistry := grpcRegistry.NewServiceRegistry(appServer, statsServer, userServer, shortURLServer)

	a.ShortURLSStorage = shortURLStg
	a.UserStorage = userStg
	a.Router = r
	a.DB = db
	a.serviceRegistry = serviceRegistry

	return a
}

// Run starts the application server.
func (a *App) Run() {
	a.printWelcomeMessage()
	if a.Config.Server.GRPC.Enabled {
		grpcSrv, err := grpcServer.New(a.serviceRegistry, a.Config, a.DB)
		if err != nil {
			log.Fatalf("cannot start gRPC server: %s", err)
		}
		go grpcSrv.Run()
	}
	server.New(a.Router, a.Config, a.DB).Run()
}

func (a *App) printWelcomeMessage() {
	welcomeMsg := fmt.Sprintf("Starting %s server on %s",
		a.Config.AppInfo(),
		a.Config.Server.Address)
	logger.Log.Info(welcomeMsg)
}
