/*
Package app provides the core application setup and runtime for the URL shortener service.

It handles:
- Application configuration
- Dependency initialization (database, logger, JWT)
- HTTP router setup
- Storage layer abstraction
- Use case registration
- HTTP server lifecycle with graceful shutdown

The package follows clean architecture principles with clear separation between:
- Domain entities and business logic
- Infrastructure concerns
- Application composition
*/
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	"go.uber.org/zap"
)

// ShortURLStorage defines the interface for short URL persistence operations.
// Implementations should provide thread-safe access to short URL storage.
type ShortURLStorage interface {
	// FindShortURL retrieves a short URL entity by its alias.
	// Returns entity.ShortURL if found or error if alias doesn't exist.
	FindShortURL(ctx context.Context, alias string) (*entity.ShortURL, error)

	// SaveShortURL persists a new short URL for the given user and original URL.
	// Returns the created short URL or error if persistence fails.
	SaveShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (*entity.ShortURL, error)

	// IsDBReady checks if the database connection is healthy.
	// Returns error if connection check fails.
	IsDBReady(ctx context.Context) error
}

// UserStorage defines the interface for user persistence operations.
// Implementations should provide thread-safe access to user data.
type UserStorage interface {
	// FindUser retrieves a user by ID.
	// Returns userEntity.User if found or error if user doesn't exist.
	FindUser(ctx context.Context, userID int) (*userEntity.User, error)

	// FindURLs retrieves all short URLs belonging to a user.
	// Returns slice of entity.ShortURL or error if query fails.
	FindURLs(ctx context.Context, userID int) ([]*entity.ShortURL, error)

	// SaveUser creates and persists a new user.
	// Returns the created user or error if persistence fails.
	SaveUser(ctx context.Context) (*userEntity.User, error)

	// MarkURLAsDeleted soft-deletes the specified URLs for a user.
	// Returns error if update operation fails.
	MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error
}

// Router defines the interface for HTTP request routing.
// Implementations should route requests to appropriate handlers.
type Router interface {
	// ServeHTTP handles HTTP requests and routes them to appropriate handlers.
	// Implements http.Handler interface.
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}

// App represents the main application container with all dependencies.
// It serves as the composition root for the application.
type App struct {
	ShortURLSStorage ShortURLStorage // Storage for short URL operations
	UserStorage      UserStorage     // Storage for user operations
	Config           *config.Config  // Application configuration
	Router           Router          // HTTP request router
}

// New creates a new App instance with the given configuration.
// The constructor returns an uninitialized App - Setup() must be called before use.
//
// Parameters:
// - cfg: Application configuration
//
// Returns:
// - *App: Uninitialized application instance
func New(cfg *config.Config) *App {
	return &App{Config: cfg}
}

// Setup initializes all application dependencies in the correct order:
// 1. Logger configuration
// 2. Database connections
// 3. JWT authentication
// 4. Storage implementations
// 5. Use case registration
// 6. HTTP handler routing
//
// The method follows the dependency injection pattern and returns the configured
// App instance for method chaining.
//
// Returns:
// - *App: Configured application instance ready for use
func (a *App) Setup() *App {
	var (
		auth        *jwt.JWT
		setupErr    error
		shortURLStg ShortURLStorage
		userStg     UserStorage
		db          database.DB
	)

	ctx := context.Background()

	// Initialize logger first since other components may need it
	logger.Setup(a.Config.App.Env, a.Config.Log.Level)

	// Setup database connection
	db, setupErr = database.Setup(ctx, a.Config)
	if setupErr != nil {
		log.Fatalf("cannot setup database: %s", setupErr)
	}

	// Initialize storage implementations
	shortURLStg = shortURLStorage.Setup(db, a.Config)
	userStg = userStorage.Setup(db)

	// Configure HTTP router
	r := router.Setup()

	// Setup JWT authentication
	auth = jwt.New(a.Config.Auth.SecretKey, a.Config.Auth.TokenTTL)

	// Initialize use cases
	userUC := userUseCase.NewUserUseCase(auth, userStg, a.Config.App.BaseURL)
	urlUC := shortURLUseCase.NewShortURLUseCase(shortURLStg, a.Config.App.BaseURL)
	appUC := appUseCase.NewAppUseCase(shortURLStg)

	// Register HTTP handlers
	shortURLHandler.Register(r, urlUC, userUC)
	appHandler.Register(r, appUC)
	apiShortURLHandler.Register(r, userUC, urlUC)
	apiUserHandler.Register(r, userUC)

	// Set initialized dependencies
	a.ShortURLSStorage = shortURLStg
	a.UserStorage = userStg
	a.Router = r

	return a
}

// Run starts the web server with graceful shutdown capabilities.
// The method handles:
// - HTTP/HTTPS server startup based on configuration
// - OS signal handling (SIGTERM, SIGINT, SIGQUIT)
// - Graceful shutdown with configurable timeout
// - Error logging and reporting
//
// This is a blocking call that runs until the server is shut down.
// The server uses timeouts from configuration for:
// - Read operations
// - Write operations
// - Idle connections
// - Graceful shutdown
//
// Example:
//
//	app := New(config).Setup()
//	app.Run() // Blocks here
func (a *App) Run() {
	welcomeMsg := fmt.Sprintf("Starting %s server on %s", a.Config.AppInfo(), a.Config.Server.Address)
	logger.Log.Info(welcomeMsg)

	// Create HTTP server with configured timeouts
	server := &http.Server{
		Addr:         a.Config.Server.Address,
		Handler:      a.Router,
		ReadTimeout:  a.Config.Server.ReadTimeout,
		WriteTimeout: a.Config.Server.WriteTimeout,
		IdleTimeout:  a.Config.Server.IdleTimeout,
	}

	// Channel to listen for server errors
	serverErr := make(chan error, 1)

	// Start server in goroutine
	go func() {
		if a.Config.Server.HTTPS.Enabled {
			// Validate TLS configuration
			if a.Config.Server.HTTPS.CertFile == "" || a.Config.Server.HTTPS.KeyFile == "" {
				logger.Log.Fatal("HTTPS is enabled but certificate or key file is not specified")
				return
			}

			logger.Log.Info("Starting HTTPS server")
			serverErr <- server.ListenAndServeTLS(
				a.Config.Server.HTTPS.CertFile,
				a.Config.Server.HTTPS.KeyFile,
			)
		} else {
			logger.Log.Info("Starting HTTP server")
			serverErr <- server.ListenAndServe()
		}
	}()

	// Channel to listen for OS signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt,
		syscall.SIGTERM, // Kubernetes/Docker termination signal
		syscall.SIGINT,  // Ctrl+C from terminal
		syscall.SIGQUIT, // Ctrl+\ from terminal
	)

	// Wait for either server error or OS signal
	select {
	case err := <-serverErr:
		logger.Log.Error("Server error", zap.Error(err))
		return

	case sig := <-interrupt:
		logger.Log.Info("Received signal. Shutting down gracefully...",
			zap.String("signal", sig.String()))

		// Create shutdown context with configured timeout
		ctx, cancel := context.WithTimeout(context.Background(), a.Config.App.ShutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			logger.Log.Error("Graceful shutdown failed:", zap.Error(err))
			// Force close if graceful shutdown fails
			if closeErr := server.Close(); closeErr != nil {
				logger.Log.Error("Forced shutdown failed:", zap.Error(closeErr))
			}
		}

		logger.Log.Info("Server shutdown complete")
	}
}
