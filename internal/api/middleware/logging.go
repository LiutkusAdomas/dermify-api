package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// NewLoggingMiddleware creates a middleware that logs HTTP requests using slog.
// It logs the request method, path, status code, and duration for each request.
func NewLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer to capture the status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Process the request
			next.ServeHTTP(ww, r)

			// Calculate duration
			duration := time.Since(start)
			durationMs := float64(duration.Nanoseconds()) / 1e6

			// Get request ID from context
			requestID := GetRequestID(r.Context())

			// Log the request
			logger.Info("request completed",
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Float64("duration_ms", durationMs),
			)
		})
	}
}
