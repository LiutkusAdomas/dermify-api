package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

type createOrgRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type updateOrgRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
}

// HandleCreateOrganization creates a new organization and adds the creator as admin.
func HandleCreateOrganization(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		var req createOrgRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if req.Name == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "name is required")
			return
		}

		slug := req.Slug
		if slug == "" {
			slug = service.GenerateSlug(req.Name)
		}
		if !service.ValidSlug(slug) {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.OrgInvalidSlug, "slug must contain only lowercase letters, numbers, and hyphens")
			return
		}

		org, err := orgSvc.Create(r.Context(), req.Name, slug, req.Description, claims.UserID)
		if err != nil {
			handleOrgError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(org) //nolint:errcheck // response write
	}
}

// HandleListUserOrganizations lists organizations the authenticated user belongs to.
func HandleListUserOrganizations(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		orgs, err := orgSvc.ListByUser(r.Context(), claims.UserID)
		if err != nil {
			slog.Error("failed to list organizations", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.OrgLookupFailed, "failed to query organizations")
			return
		}

		if orgs == nil {
			orgs = []*domain.OrganizationWithRole{}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orgs) //nolint:errcheck // response write
	}
}

// HandleGetOrganization returns a single organization by ID.
func HandleGetOrganization(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		org, err := orgSvc.GetByID(r.Context(), membership.OrgID)
		if err != nil {
			handleOrgError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(org) //nolint:errcheck // response write
	}
}

// HandleUpdateOrganization updates organization details (admin only via middleware).
func HandleUpdateOrganization(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotAdmin, "admin access required")
			return
		}

		var req updateOrgRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		org, err := orgSvc.Update(r.Context(), membership.OrgID, req.Name, req.Description, req.LogoURL)
		if err != nil {
			handleOrgError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(org) //nolint:errcheck // response write
	}
}

// HandleListMembers lists all members of an organization.
func HandleListMembers(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		members, err := orgSvc.ListMembers(r.Context(), membership.OrgID)
		if err != nil {
			slog.Error("failed to list members", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.OrgLookupFailed, "failed to query members")
			return
		}

		if members == nil {
			members = []*domain.OrgMember{}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(members) //nolint:errcheck // response write
	}
}

type updateMemberRoleRequest struct {
	Role string `json:"role"`
}

// HandleUpdateMemberRole changes a member's role within an organization (admin only).
func HandleUpdateMemberRole(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotAdmin, "admin access required")
			return
		}

		userID, err := parseNamedIDParam(r, "userId")
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid user id")
			return
		}

		var req updateMemberRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if !domain.ValidOrgRole(req.Role) {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.RoleInvalidRole, "role must be admin, member, or viewer")
			return
		}

		if err := orgSvc.UpdateMemberRole(r.Context(), membership.OrgID, userID, req.Role); err != nil {
			handleOrgError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "role updated successfully"}) //nolint:errcheck // response write
	}
}

// HandleRemoveMember removes a member from an organization (admin only).
func HandleRemoveMember(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotAdmin, "admin access required")
			return
		}

		userID, err := parseNamedIDParam(r, "userId")
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid user id")
			return
		}

		if err := orgSvc.RemoveMember(r.Context(), membership.OrgID, userID); err != nil {
			handleOrgError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "member removed successfully"}) //nolint:errcheck // response write
	}
}

type inviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// HandleInviteUser creates an invitation and optionally sends an email (admin only).
func HandleInviteUser(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		membership := middleware.GetOrgMembership(r.Context())
		if claims == nil || membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotAdmin, "admin access required")
			return
		}

		var req inviteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if req.Email == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "email is required")
			return
		}

		if req.Role == "" {
			req.Role = domain.OrgRoleMember
		}
		if !domain.ValidOrgRole(req.Role) {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.RoleInvalidRole, "role must be admin, member, or viewer")
			return
		}

		inv, err := orgSvc.InviteUser(r.Context(), membership.OrgID, claims.UserID, req.Email, req.Role)
		if err != nil {
			handleOrgError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(inv) //nolint:errcheck // response write
	}
}

// HandleListPendingInvitations lists the current user's pending invitations.
func HandleListPendingInvitations(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		invitations, err := orgSvc.ListUserInvitations(r.Context(), claims.Email)
		if err != nil {
			slog.Error("failed to list invitations", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.OrgLookupFailed, "failed to query invitations")
			return
		}

		if invitations == nil {
			invitations = []*domain.OrgInvitation{}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(invitations) //nolint:errcheck // response write
	}
}

// HandleAcceptInvitation accepts a pending invitation by token.
func HandleAcceptInvitation(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		token := getURLParam(r, "token")
		if token == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "token is required")
			return
		}

		if err := orgSvc.AcceptInvitation(r.Context(), token, claims.UserID, claims.Email); err != nil {
			handleOrgError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "invitation accepted"}) //nolint:errcheck // response write
	}
}

// HandleDeclineInvitation declines a pending invitation by token.
func HandleDeclineInvitation(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		token := getURLParam(r, "token")
		if token == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "token is required")
			return
		}

		if err := orgSvc.DeclineInvitation(r.Context(), token, claims.Email); err != nil {
			handleOrgError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "invitation declined"}) //nolint:errcheck // response write
	}
}

// HandleListOrgInvitations lists all invitations for an organization (admin only).
func HandleListOrgInvitations(orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotAdmin, "admin access required")
			return
		}

		invitations, err := orgSvc.ListOrgInvitations(r.Context(), membership.OrgID)
		if err != nil {
			slog.Error("failed to list org invitations", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.OrgLookupFailed, "failed to query invitations")
			return
		}

		if invitations == nil {
			invitations = []*domain.OrgInvitation{}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(invitations) //nolint:errcheck // response write
	}
}

// handleOrgError maps service org errors to HTTP responses.
func handleOrgError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrOrgNotFound):
		apierrors.WriteError(w, http.StatusNotFound, apierrors.OrgNotFound, "organization not found")
	case errors.Is(err, service.ErrOrgSlugExists):
		apierrors.WriteError(w, http.StatusConflict, apierrors.OrgSlugExists, "organization slug already exists")
	case errors.Is(err, service.ErrOrgNotMember):
		apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member of this organization")
	case errors.Is(err, service.ErrOrgNotAdmin):
		apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotAdmin, "admin access required")
	case errors.Is(err, service.ErrOrgMemberNotFound):
		apierrors.WriteError(w, http.StatusNotFound, apierrors.OrgMemberNotFound, "membership not found")
	case errors.Is(err, service.ErrOrgLastAdmin):
		apierrors.WriteError(w, http.StatusBadRequest, apierrors.OrgLastAdmin, "cannot remove the last admin")
	case errors.Is(err, service.ErrOrgNoFieldsToUpdate):
		apierrors.WriteError(w, http.StatusBadRequest, apierrors.OrgNoFieldsToUpdate, "no fields to update")
	case errors.Is(err, service.ErrOrgAlreadyMember):
		apierrors.WriteError(w, http.StatusConflict, apierrors.OrgAlreadyMember, "user is already a member of this organization")
	case errors.Is(err, service.ErrInvitationExists):
		apierrors.WriteError(w, http.StatusConflict, apierrors.InvitationExists, "a pending invitation already exists for this email")
	case errors.Is(err, service.ErrInvitationNotFound):
		apierrors.WriteError(w, http.StatusNotFound, apierrors.InvitationNotFound, "invitation not found or expired")
	case errors.Is(err, service.ErrInvitationWrongEmail):
		apierrors.WriteError(w, http.StatusForbidden, apierrors.InvitationWrongEmail, "this invitation is for a different email address")
	default:
		slog.Error("organization operation failed", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError, apierrors.OrgLookupFailed, "internal error")
	}
}
