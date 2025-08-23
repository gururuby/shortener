// Package handler provides HTTP handlers for the application's API endpoints.
// It bridges the HTTP layer with the application's use cases.
package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gururuby/shortener/internal/domain/entity/stats"
	jsoniter "github.com/json-iterator/go"
)

var jsonIter = jsoniter.ConfigFastest

const (
	getStatsPath    = "/api/internal/stats" // Path for the stats endpoint
	getStatsTimeout = time.Second * 30      // Timeout for stats requests
)

// Router defines the interface for HTTP route registration.
type Router interface {
	Get(path string, h http.HandlerFunc)
}

// StatsUseCase defines the interface for statistics-related business logic operations.
type StatsUseCase interface {
	GetStats(ctx context.Context) (*entity.Stats, error)
}

// handler manages the HTTP handlers and their dependencies.
type handler struct {
	statsUC StatsUseCase // Use case for statistics operations
	router  Router       // Request router
}

// errorResponse represents a standardized error response for API errors.
type errorResponse struct {
	Error      string // Error message
	StatusCode int    // HTTP status code
}

// Register sets up all HTTP routes with their corresponding handlers.
func Register(router Router, statsUC StatsUseCase) {
	h := handler{router: router, statsUC: statsUC}
	h.router.Get(getStatsPath, h.GetStats())
}

// GetStats returns an HTTP handler function for the statistics endpoint.
// It handles GET requests to retrieve application statistics.
func (h *handler) GetStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err        error
			stats      *entity.Stats
			statusCode = http.StatusOK
			response   []byte
			errRes     errorResponse
		)

		ctx, cancel := context.WithTimeout(r.Context(), getStatsTimeout)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodGet {
			errRes.Error = fmt.Sprintf("HTTP method %s is not allowed", r.Method)
			errRes.StatusCode = http.StatusMethodNotAllowed
			returnErrResponse(errRes, w)
			return
		}

		stats, err = h.statsUC.GetStats(ctx)

		if err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusUnprocessableEntity
			returnErrResponse(errRes, w)
			return
		}

		response, err = jsonIter.Marshal(stats)

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

// returnErrResponse writes an error response in JSON format with the appropriate status code.
func returnErrResponse(errResp errorResponse, w http.ResponseWriter) {
	w.WriteHeader(errResp.StatusCode)
	response, err := jsonIter.Marshal(errResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if _, err = w.Write(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
