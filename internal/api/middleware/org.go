package middleware

import (
	"context"
	"net/http"
	"strconv"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/service"

	"github.com/go-chi/chi/v5"
)

type orgContextKey struct{}

// OrgMembership holds organization membership details from the context.
type OrgMembership struct {
	OrgID  int64
	UserID int64
	Role   string
}

// RequireOrgMember validates that the authenticated user is a member of the
// organization identified by the {orgId} URL parameter.
func RequireOrgMember(orgSvc *service.OrganizationService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			claims := GetUserClaims(r.Context())
			if claims == nil {
				apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthNotAuthenticated, "not authenticated")
				return
			}

			orgID, err := strconv.ParseInt(chi.URLParam(r, "orgId"), 10, 64)
			if err != nil {
				apierrors.WriteError(w, http.StatusBadRequest, apierrors.OrgNotFound, "invalid organization id")
				return
			}

			role, err := orgSvc.GetMemberRole(r.Context(), orgID, claims.UserID)
			if err != nil {
				apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member of this organization")
				return
			}

			membership := &OrgMembership{
				OrgID:  orgID,
				UserID: claims.UserID,
				Role:   role,
			}
			ctx := context.WithValue(r.Context(), orgContextKey{}, membership)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireOrgAdmin validates that the authenticated user is an admin of the
// organization identified by the {orgId} URL parameter.
func RequireOrgAdmin(orgSvc *service.OrganizationService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			claims := GetUserClaims(r.Context())
			if claims == nil {
				apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthNotAuthenticated, "not authenticated")
				return
			}

			orgID, err := strconv.ParseInt(chi.URLParam(r, "orgId"), 10, 64)
			if err != nil {
				apierrors.WriteError(w, http.StatusBadRequest, apierrors.OrgNotFound, "invalid organization id")
				return
			}

			role, err := orgSvc.GetMemberRole(r.Context(), orgID, claims.UserID)
			if err != nil || role != "admin" {
				apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotAdmin, "admin access required")
				return
			}

			membership := &OrgMembership{
				OrgID:  orgID,
				UserID: claims.UserID,
				Role:   role,
			}
			ctx := context.WithValue(r.Context(), orgContextKey{}, membership)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetOrgMembership extracts the organization membership from the request context.
func GetOrgMembership(ctx context.Context) *OrgMembership {
	if m, ok := ctx.Value(orgContextKey{}).(*OrgMembership); ok {
		return m
	}
	return nil
}
