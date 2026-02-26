package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// requestIDKey is the context key for the request ID.
type requestIDKey struct{}

// RequestIDHeader is the header name for the request ID.
const RequestIDHeader = "X-Request-ID"

// RequestID is a middleware that generates a unique request ID for each request.
// It stores the ID in the request context and adds it to the response headers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (e.g., from upstream proxy)
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to response header
		w.Header().Set(RequestIDHeader, requestID)

		// Store request ID in context
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from the context.
// Returns an empty string if no request ID is found.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
