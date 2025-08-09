/*
Package middleware provides HTTP middleware components for request logging.

It features:
- Request/response logging with detailed metrics
- Structured logging using zap logger
- Capture of response status codes and sizes
- Duration measurement for performance monitoring
*/
package middleware

import (
	"net/http"
	"time"

	"github.com/gururuby/shortener/internal/infra/logger"
	"go.uber.org/zap"
)

// Logging is middleware that logs HTTP requests and responses.
// It captures:
// - HTTP method
// - Request path
// - Response status code
// - Response duration
// - Response size
//
// Logs are emitted in structured format using the application logger.
func Logging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		resp := &responseData{}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   resp,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Log.Info("shortener",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", resp.status),
			zap.Duration("duration", duration),
			zap.Int("size", resp.size),
		)
	}
	return http.HandlerFunc(logFn)
}

// responseData holds captured response metrics for logging.
type responseData struct {
	status int // HTTP status code
	size   int // Response body size in bytes
}

// loggingResponseWriter wraps http.ResponseWriter to capture response data.
type loggingResponseWriter struct {
	http.ResponseWriter               // Embedded original ResponseWriter
	responseData        *responseData // Pointer to shared response metrics
}

// Write captures the response size while writing to the underlying ResponseWriter.
// Implements the io.Writer interface.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader captures the status code while writing headers.
// Overrides the http.ResponseWriter interface method.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
