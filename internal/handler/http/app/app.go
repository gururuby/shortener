//go:generate mockgen -destination=./mocks/mock.go -package=mocks . AppUseCase

/*
Package handler implements HTTP request handlers for application health checks.

It provides:
- Health check endpoints
- Database connectivity testing
- Basic request validation
*/
package handler

import (
	"context"
	"fmt"
	"net/http"
)

const (
	pingDBPath = "/ping" // Endpoint path for database health check
)

// Router defines the interface for HTTP request routing.
type Router interface {
	// Get registers a handler for GET requests at the specified path
	Get(path string, h http.HandlerFunc)
}

// AppUseCase defines the interface for application-level operations.
type AppUseCase interface {
	// PingDB checks the database connection status
	// Returns:
	// - error: If database is unreachable
	PingDB(ctx context.Context) error
}

// handler implements the HTTP request handlers for application operations.
type handler struct {
	uc     AppUseCase // Application use case implementation
	router Router     // HTTP router
}

// Register sets up the application health check routes.
// Parameters:
// - router: The HTTP router implementation
// - uc: Application use case implementation
func Register(router Router, uc AppUseCase) {
	h := handler{router: router, uc: uc}
	h.router.Get(pingDBPath, h.PingDB())
}

// PingDB handles requests to check database connectivity.
// Returns an HTTP handler function that:
// - Validates the request method
// - Checks database status
// - Returns appropriate status codes:
//   - 200 OK if database is reachable
//   - 422 Unprocessable Entity if database is unreachable
//   - 405 Method Not Allowed for invalid HTTP methods
func (h *handler) PingDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		if r.Method != http.MethodGet {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		err = h.uc.PingDB(r.Context())

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
