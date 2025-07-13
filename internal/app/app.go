/*
Package app provides the core application setup and runtime for the URL shortener service.

It handles:
- Application configuration
- Dependency initialization (database, logger, JWT)
- HTTP router setup
- Storage layer abstraction
- Use case registration
- HTTP server lifecycle
*/
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

// ShortURLStorage defines the interface for short URL persistence operations.
type ShortURLStorage interface {
	// FindShortURL retrieves a short URL entity by its alias.
	FindShortURL(ctx context.Context, alias string) (*entity.ShortURL, error)

	// SaveShortURL persists a new short URL for the given user and original URL.
	SaveShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (*entity.ShortURL, error)

	// IsDBReady checks if the database connection is healthy.
	IsDBReady(ctx context.Context) error
}

// UserStorage defines the interface for user persistence operations.
type UserStorage interface {
	// FindUser retrieves a user by ID.
	FindUser(ctx context.Context, userID int) (*userEntity.User, error)

	// FindURLs retrieves all short URLs belonging to a user.
	FindURLs(ctx context.Context, userID int) ([]*entity.ShortURL, error)

	// SaveUser creates and persists a new user.
	SaveUser(ctx context.Context) (*userEntity.User, error)

	// MarkURLAsDeleted soft-deletes the specified URLs for a user.
	MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error
}

// Router defines the interface for HTTP request routing.
type Router interface {
	// ServeHTTP handles HTTP requests and routes them to appropriate handlers.
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}

// App represents the main application container with all dependencies.
type App struct {
	ShortURLSStorage ShortURLStorage // Storage for short URL operations
	UserStorage      UserStorage     // Storage for user operations
	Config           *config.Config  // Application configuration
	Router           Router          // HTTP request router
}

// New creates a new App instance with the given configuration.
// The returned App needs to be initialized using Setup() before use.
func New(cfg *config.Config) *App {
	return &App{Config: cfg}
}

// Setup initializes all application dependencies including:
// - Logger configuration
// - Database connections
// - JWT authentication
// - Storage implementations
// - Use case registrations
// - HTTP handler routing
//
// Returns the configured App instance for method chaining.
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

// Run starts the HTTP server and begins listening for requests.
// This is a blocking call that runs until the server is shut down.
// Server address is taken from the application configuration.
func (a *App) Run() {
	welcomeMsg := fmt.Sprintf("Starting %s server on %s", a.Config.AppInfo(), a.Config.Server.Address)
	logger.Log.Info(welcomeMsg)
	log.Fatal(http.ListenAndServe(a.Config.Server.Address, a.Router))
}
