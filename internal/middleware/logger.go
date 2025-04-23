package middleware

import (
	"github.com/gururuby/shortener/internal/infra/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

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

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
