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

// RequireOrgRole validates that the authenticated user has one of the specified
// roles within the organization identified by the {orgId} URL parameter.
func RequireOrgRole(orgSvc *service.OrganizationService, allowed ...string) func(http.Handler) http.Handler {
	allowedSet := make(map[string]bool, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = true
	}
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

			if !allowedSet[role] {
				apierrors.WriteError(w, http.StatusForbidden, apierrors.AuthInsufficientRole, "insufficient role permissions")
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

// RequireOrgScheduleAccess allows:
// - admin/receptionist to manage any doctor's schedule
// - doctor to manage only their own schedule
func RequireOrgScheduleAccess(orgSvc *service.OrganizationService) func(http.Handler) http.Handler {
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
			doctorID, err := strconv.ParseInt(chi.URLParam(r, "doctorId"), 10, 64)
			if err != nil {
				apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid doctor id")
				return
			}

			role, err := orgSvc.GetMemberRole(r.Context(), orgID, claims.UserID)
			if err != nil {
				apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member of this organization")
				return
			}

			allowed := role == "admin" || role == "receptionist" || (role == "doctor" && claims.UserID == doctorID)
			if !allowed {
				apierrors.WriteError(w, http.StatusForbidden, apierrors.AuthInsufficientRole, "insufficient role permissions")
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
