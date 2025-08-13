//go:generate mockgen -destination=./mocks/mock.go -package=mocks . UserUseCase

/*
Package handler implements HTTP request handlers for user-related operations.

It provides:
- User URL management endpoints
- Authentication and session handling
- Request/response processing
- Error handling and status code management
*/
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	"github.com/gururuby/shortener/internal/domain/usecase/user"
	handlerErrors "github.com/gururuby/shortener/internal/handler/http/api/user/errors"
)

// Available constants
const (
	authCookieName    = "Authorization"  // Name of the authentication cookie
	getURLsTimeout    = time.Second * 30 // Timeout for GET URLs operation
	deleteURLsTimeout = time.Second * 30 // Timeout for DELETE URLs operation
	URLsPath          = "/api/user/urls" // Base path for user URL operations
)

// Router defines the interface for HTTP request routing.
type Router interface {
	// Get registers a handler for GET requests at the specified path
	Get(path string, h http.HandlerFunc)
	// Delete registers a handler for DELETE requests at the specified path
	Delete(path string, h http.HandlerFunc)
}

// UserUseCase defines the interface for user-related business logic.
type UserUseCase interface {
	// GetURLs retrieves all shortened URLs belonging to a user
	GetURLs(ctx context.Context, user *userEntity.User) ([]*usecase.UserShortURL, error)
	// DeleteURLs removes the specified URLs belonging to a user
	DeleteURLs(ctx context.Context, user *userEntity.User, aliases []string)
	// Authenticate verifies a user's credentials
	Authenticate(ctx context.Context, token string) (*userEntity.User, error)
	// Register creates a new user account
	Register(ctx context.Context) (*userEntity.User, error)
}

// handler implements the HTTP request handlers for user operations.
type handler struct {
	userUC UserUseCase // User business logic service
	router Router      // Request router
}

// errorResponse represents an API error response.
type errorResponse struct {
	Error      string
	StatusCode int
}

// Register sets up the user-related API routes and their handlers.
// Parameters:
// - router: The HTTP router implementation
// - userUC: User business logic service
func Register(router Router, userUC UserUseCase) {
	h := handler{router: router, userUC: userUC}
	h.router.Get(URLsPath, h.GetURLs())
	h.router.Delete(URLsPath, h.DeleteURLs())
}

// GetURLs handles GET requests to retrieve a user's shortened URLs.
// Returns an HTTP handler function that:
// - Authenticates the user
// - Retrieves their URLs
// - Returns appropriate responses
func (h *handler) GetURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err        error
			statusCode int
			response   []byte
			errRes     errorResponse
			user       *userEntity.User
			userURLs   []*usecase.UserShortURL
		)

		ctx, cancel := context.WithTimeout(r.Context(), getURLsTimeout)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodGet {
			errRes.Error = fmt.Sprintf("HTTP method %s is not allowed", r.Method)
			errRes.StatusCode = http.StatusMethodNotAllowed
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

		userURLs, err = h.userUC.GetURLs(ctx, user)
		if err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusInternalServerError
			returnErrResponse(errRes, w)
			return
		}

		if len(userURLs) == 0 {
			statusCode = http.StatusNoContent
			response = []byte("{}")
		} else {
			statusCode = http.StatusOK
			response, err = json.Marshal(userURLs)
			if err != nil {
				errRes.Error = err.Error()
				errRes.StatusCode = http.StatusInternalServerError
				returnErrResponse(errRes, w)
				return
			}
		}

		w.WriteHeader(statusCode)

		if _, err = w.Write(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	}
}

// DeleteURLs handles DELETE requests to remove user's shortened URLs.
// Returns an HTTP handler function that:
// - Authenticates the user
// - Validates the request
// - Deletes specified URLs
// - Returns appropriate responses
func (h *handler) DeleteURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err     error
			errRes  errorResponse
			user    *userEntity.User
			aliases []string
		)

		ctx, cancel := context.WithTimeout(r.Context(), deleteURLsTimeout)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodDelete {
			errRes.Error = fmt.Sprintf("HTTP method %s is not allowed", r.Method)
			errRes.StatusCode = http.StatusMethodNotAllowed
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

		if err = json.NewDecoder(r.Body).Decode(&aliases); err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusBadRequest
			returnErrResponse(errRes, w)
			return
		}

		if len(aliases) == 0 {
			errRes.Error = handlerErrors.ErrHandlerNoAliasesForDelete.Error()
			errRes.StatusCode = http.StatusBadRequest
			returnErrResponse(errRes, w)
			return
		}

		h.userUC.DeleteURLs(ctx, user, aliases)
		w.WriteHeader(http.StatusAccepted)
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
