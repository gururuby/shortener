/*
Package router provides HTTP routing functionality for the application.

It features:
- Chi router implementation with common middleware
- Standardized HTTP method routing
- Debug profiling endpoint
- Interface for router abstraction
*/
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/middleware"
)

// Router defines the interface for HTTP request routing.
// Implementations should provide methods for registering route handlers
// and serving HTTP requests.
type Router interface {
	// Post registers a handler for HTTP POST requests at the specified path
	Post(path string, h http.HandlerFunc)

	// Get registers a handler for HTTP GET requests at the specified path
	Get(path string, h http.HandlerFunc)

	// Delete registers a handler for HTTP DELETE requests at the specified path
	Delete(path string, h http.HandlerFunc)

	// ServeHTTP dispatches the request to the handler whose pattern matches
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}

// Setup creates and configures a new router instance with default middleware.
// The returned router includes:
// - Request logging middleware
// - Response compression middleware
// - Debug profiling endpoint at /debug
//
// Returns:
// - Router: Configured router instance ready for route registration
func Setup(srvConfig config.Server) Router {
	router := chi.NewRouter()
	router.Use(middleware.Logging)
	router.Use(middleware.Compression)
	router.Use(middleware.TrustedSubnetMiddleware(srvConfig.TrustedSubnet))

	return router
}
