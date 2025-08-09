//go:generate mockgen -destination=./mocks/mock.go -package=mocks . UserUseCase,ShortURLUseCase

/*
Package handler implements HTTP request handlers for URL shortening operations.

It provides:
- URL shortening and redirection endpoints
- User authentication and session management
- Request validation and error handling
- Support for both single and batch URL operations
*/
package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/shorturl/errors"
)

const (
	authCookieName        = "Authorization"  // Name of the authentication cookie
	createShortURLTimeout = time.Second * 30 // Timeout for URL creation operations
	shortensPath          = "/"              // Path for URL shortening endpoint
	shortenPath           = "/{alias}"       // Path pattern for URL redirection
)

// Router defines the interface for HTTP request routing.
type Router interface {
	// Post registers a handler for POST requests
	Post(path string, h http.HandlerFunc)
	// Get registers a handler for GET requests
	Get(path string, h http.HandlerFunc)
}

// ShortURLUseCase defines the interface for URL shortening business logic.
type ShortURLUseCase interface {
	// CreateShortURL generates a shortened URL for the given original URL
	CreateShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (string, error)
	// FindShortURL retrieves the original URL for a given short alias
	FindShortURL(ctx context.Context, alias string) (string, error)
	// BatchShortURLs processes multiple URLs in a single operation
	BatchShortURLs(ctx context.Context, urls []entity.BatchShortURLInput) []entity.BatchShortURLOutput
}

// UserUseCase defines the interface for user management operations.
type UserUseCase interface {
	// Authenticate verifies a user's credentials
	Authenticate(ctx context.Context, token string) (*userEntity.User, error)
	// Register creates a new user account
	Register(ctx context.Context) (*userEntity.User, error)
}

// handler implements the HTTP request handlers for URL operations.
type handler struct {
	userUC UserUseCase     // User management service
	urlUC  ShortURLUseCase // URL shortening service
	router Router          // HTTP router
}

// Register initializes and registers all URL shortening handlers.
// Parameters:
// - router: The HTTP router implementation
// - urlUC: URL shortening service
// - userUC: User management service
func Register(router Router, urlUC ShortURLUseCase, userUC UserUseCase) {
	h := handler{router: router, urlUC: urlUC, userUC: userUC}
	h.router.Get(shortenPath, h.FindShortURL())
	h.router.Post(shortensPath, h.CreateShortURL())
}

// CreateShortURL handles POST requests to create shortened URLs.
// Returns an HTTP handler function that:
// - Validates the request
// - Authenticates/registers the user
// - Creates the short URL
// - Returns appropriate responses:
//   - 201 Created for successful creation
//   - 409 Conflict if URL already exists
//   - 400/422 for invalid requests
func (h *handler) CreateShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err        error
			user       *userEntity.User
			reqBody    []byte
			shortURL   string
			statusCode = http.StatusCreated
		)

		ctx, cancel := context.WithTimeout(r.Context(), createShortURLTimeout)
		defer cancel()

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		reqBody, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sourceURL := string(reqBody)

		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}(r.Body)

		user, err = h.authUser(ctx, r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		shortURL, err = h.urlUC.CreateShortURL(r.Context(), user, sourceURL)

		if err != nil {
			if errors.Is(err, ucErrors.ErrShortURLAlreadyExist) {
				statusCode = http.StatusConflict
			} else {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(statusCode)

		_, err = io.WriteString(w, shortURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// FindShortURL handles GET requests to redirect to original URLs.
// Returns an HTTP handler function that:
// - Validates the request
// - Looks up the original URL
// - Returns appropriate responses:
//   - 307 Temporary Redirect for successful lookups
//   - 410 Gone for deleted URLs
//   - 422 for other errors
func (h *handler) FindShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}
		result, err := h.urlUC.FindShortURL(r.Context(), r.URL.Path)

		if err != nil {
			if errors.Is(err, ucErrors.ErrShortURLDeleted) {
				http.Error(w, err.Error(), http.StatusGone)
				return
			}
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Location", result)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// authUser handles user authentication via cookie or registration.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - r: HTTP request
// - w: HTTP response writer
// Returns:
// - *userEntity.User: Authenticated user
// - error: Authentication failure
func (h *handler) authUser(ctx context.Context, r *http.Request, w http.ResponseWriter) (*userEntity.User, error) {
	var (
		authCookie *http.Cookie
		user       *userEntity.User
		err        error
	)

	authCookie, err = r.Cookie(authCookieName)
	// If auth cookie was not passed
	if err != nil && errors.Is(err, http.ErrNoCookie) {
		// Register new User
		if user, err = h.userUC.Register(ctx); err != nil {
			return nil, err
		}

	} else { // If auth cookie exist, try to authenticate User
		if user, err = h.userUC.Authenticate(ctx, authCookie.Value); err != nil {
			// If auth cookie is invalid or user not found try to register new user
			if user, err = h.userUC.Register(ctx); err != nil {
				return nil, err
			}
		}
	}
	// Setup auth cookie
	http.SetCookie(w, &http.Cookie{Name: authCookieName, Value: user.AuthToken})

	return user, nil
}
