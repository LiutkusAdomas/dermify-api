package middleware

import (
	"context"
	"net/http"
	"strings"

	"dermify-api/config"
	"dermify-api/internal/api/apierrors"
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
				apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthMissingHeader, "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthInvalidHeaderFormat, "invalid authorization header format")
				return
			}

			claims, err := auth.ValidateAccessToken(parts[1], cfg.Auth.JWTSecret)
			if err != nil {
				apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthInvalidToken, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userClaimsKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware that checks whether the authenticated user
// has one of the specified roles. Must be used after RequireAuth.
func RequireRole(allowed ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			claims := GetUserClaims(r.Context())
			if claims == nil {
				apierrors.WriteError(w, http.StatusUnauthorized,
					apierrors.AuthNotAuthenticated, "not authenticated")
				return
			}

			for _, role := range allowed {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			apierrors.WriteError(w, http.StatusForbidden,
				apierrors.AuthInsufficientRole, "insufficient role permissions")
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
