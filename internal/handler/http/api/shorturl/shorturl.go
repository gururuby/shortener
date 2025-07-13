//go:generate mockgen -destination=./mocks/mock.go -package=mocks . ShortURLUseCase,UserUseCase

/*
Package handler implements HTTP request handlers for the URL shortener API.

It provides:
- REST endpoints for URL shortening operations
- Authentication and user management
- Request/response handling
- Error handling and status code management
*/
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/shorturl/errors"
	apiErrors "github.com/gururuby/shortener/internal/handler/http/api/shorturl/errors"
	"net/http"
	"time"
)

const (
	authCookieName        = "Authorization"  // Name of the authentication cookie
	createShortURLTimeout = time.Second * 30 // Timeout for short URL creation
	createShortURLPath    = "/api/shorten"   // Path for single URL shortening

	batchShortURLsTimeout = time.Second * 60     // Timeout for batch URL processing
	batchShortURLsPath    = "/api/shorten/batch" // Path for batch URL shortening
)

// Router defines the interface for HTTP request routing.
type Router interface {
	// Post registers a handler for POST requests at the specified path
	Post(path string, h http.HandlerFunc)
}

// ShortURLUseCase defines the interface for short URL business logic.
type ShortURLUseCase interface {
	// CreateShortURL generates a shortened URL for the given source URL
	CreateShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (string, error)

	// FindShortURL retrieves the original URL for a given short alias
	FindShortURL(ctx context.Context, alias string) (string, error)

	// BatchShortURLs processes multiple URLs in a single operation
	BatchShortURLs(ctx context.Context, urls []shortURLEntity.BatchShortURLInput) []shortURLEntity.BatchShortURLOutput
}

// UserUseCase defines the interface for user management operations.
type UserUseCase interface {
	// Authenticate verifies a user's credentials and returns user info
	Authenticate(ctx context.Context, token string) (*userEntity.User, error)

	// Register creates a new user account
	Register(ctx context.Context) (*userEntity.User, error)
}

// handler implements the HTTP request handlers for the API.
type handler struct {
	userUC UserUseCase     // User management service
	urlUC  ShortURLUseCase // URL shortening service
	router Router          // Request router
}

// errorResponse represents an API error response.
type errorResponse struct {
	StatusCode int    // HTTP status code
	Error      string // Error message
}

type (
	// createShortURLDTO defines the request/response structure for single URL shortening
	createShortURLDTO struct {
		request struct {
			URL string // Original URL to shorten
		}
		response struct {
			Result string // Generated short URL
		}
	}

	// batchShortURLsDTO defines the request/response structure for batch URL shortening
	batchShortURLsDTO struct {
		inputURLs  []shortURLEntity.BatchShortURLInput  // Input URLs to process
		outputURLs []shortURLEntity.BatchShortURLOutput // Resulting short URLs
	}
)

// Register sets up the API routes and their corresponding handlers.
// Parameters:
// - router: The HTTP router implementation
// - userUC: User management service
// - urlUC: URL shortening service
func Register(router Router, userUC UserUseCase, urlUC ShortURLUseCase) {
	h := handler{router: router, userUC: userUC, urlUC: urlUC}
	h.router.Post(batchShortURLsPath, h.BatchShortURLs())
	h.router.Post(createShortURLPath, h.CreateShortURL())
}

// CreateShortURL handles requests to create a single short URL.
// Returns an HTTP handler function that:
// - Validates the request
// - Authenticates/registers the user
// - Creates the short URL
// - Returns appropriate responses
func (h *handler) CreateShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err        error
			user       *userEntity.User
			statusCode = http.StatusCreated
			shortURL   string
			response   []byte
			dto        createShortURLDTO
			errRes     errorResponse
		)

		ctx, cancel := context.WithTimeout(r.Context(), createShortURLTimeout)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			errRes.Error = fmt.Sprintf("HTTP method %s is not allowed", r.Method)
			errRes.StatusCode = http.StatusMethodNotAllowed
			returnErrResponse(errRes, w)
			return
		}

		if err = json.NewDecoder(r.Body).Decode(&dto.request); err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusBadRequest
			returnErrResponse(errRes, w)
			return
		}

		user, err = h.authUser(ctx, r, w)
		if err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusUnprocessableEntity
			returnErrResponse(errRes, w)
			return
		}

		shortURL, err = h.urlUC.CreateShortURL(ctx, user, dto.request.URL)

		if err != nil {
			if errors.Is(err, ucErrors.ErrShortURLAlreadyExist) {
				statusCode = http.StatusConflict
			} else {
				errRes.Error = err.Error()
				errRes.StatusCode = http.StatusUnprocessableEntity
				returnErrResponse(errRes, w)
				return
			}

		}

		dto.response.Result = shortURL
		response, err = json.Marshal(dto.response)

		if err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusInternalServerError
			returnErrResponse(errRes, w)
			return
		}

		w.WriteHeader(statusCode)

		if _, err = w.Write(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// BatchShortURLs handles requests to create multiple short URLs in a batch.
// Returns an HTTP handler function that:
// - Validates the request
// - Processes URLs in batch
// - Returns appropriate responses
func (h *handler) BatchShortURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err      error
			response []byte
			dto      batchShortURLsDTO
			errRes   errorResponse
		)

		ctx, cancel := context.WithTimeout(r.Context(), batchShortURLsTimeout)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			errRes.Error = fmt.Sprintf("HTTP method %s is not allowed", r.Method)
			errRes.StatusCode = http.StatusMethodNotAllowed
			returnErrResponse(errRes, w)
			return
		}

		if err = json.NewDecoder(r.Body).Decode(&dto.inputURLs); err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusBadRequest
			returnErrResponse(errRes, w)
			return
		}

		if len(dto.inputURLs) == 0 {
			errRes.Error = apiErrors.ErrAPIEmptyBatch.Error()
			errRes.StatusCode = http.StatusBadRequest
			returnErrResponse(errRes, w)
			return
		}

		dto.outputURLs = h.urlUC.BatchShortURLs(ctx, dto.inputURLs)
		response, err = json.Marshal(dto.outputURLs)

		if err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusInternalServerError
			returnErrResponse(errRes, w)
			return
		}

		w.WriteHeader(http.StatusCreated)

		if _, err = w.Write(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// authUser handles user authentication via cookie or registration.
// Parameters:
// - ctx: Context for cancellation/timeout
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

// returnErrResponse writes an error response in JSON format.
// Parameters:
// - errResp: Error response details
// - w: HTTP response writer
func returnErrResponse(errResp errorResponse, w http.ResponseWriter) {
	w.WriteHeader(errResp.StatusCode)
	response, err := json.Marshal(errResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if _, err = w.Write(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
