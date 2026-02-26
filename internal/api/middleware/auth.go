package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"dermify-api/config"
	"dermify-api/internal/api/auth"
)

// userClaimsKey is the context key for authenticated user claims.
type userClaimsKey struct{}

// RequireAuth returns middleware that validates JWT access tokens from the
// Authorization header and stores the claims in the request context.
func RequireAuth(cfg *config.Configuration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			header := r.Header.Get("Authorization")
			if header == "" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "missing authorization header"})
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid authorization header format"})
				return
			}

			claims, err := auth.ValidateAccessToken(parts[1], cfg.Auth.JWTSecret)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired token"})
				return
			}

			ctx := context.WithValue(r.Context(), userClaimsKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserClaims extracts the authenticated user claims from the request context.
// Returns nil if no claims are present.
func GetUserClaims(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(userClaimsKey{}).(*auth.Claims); ok {
		return claims
	}
	return nil
}
