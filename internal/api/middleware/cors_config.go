package middleware

import (
	"dermify-api/config"
	"net/http"
	"strings"
)

// CORSWithConfig creates a CORS middleware using configuration
func CORSWithConfig(cfg *config.Configuration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if isOriginAllowed(origin, cfg.CORS.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(cfg.CORS.AllowedOrigins) == 0 {
				// If no origins configured, allow all (development mode)
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			// Set allowed methods
			if len(cfg.CORS.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.CORS.AllowedMethods, ", "))
			} else {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			}

			// Set allowed headers
			if len(cfg.CORS.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.CORS.AllowedHeaders, ", "))
			} else {
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With")
			}

			w.Header().Set("Access-Control-Expose-Headers", "Link")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if the origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}
	return false
}
