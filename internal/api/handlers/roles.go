package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/service"
)

type assignRoleRequest struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

type assignRoleResponse struct {
	UserID  int64  `json:"user_id"`
	Role    string `json:"role"`
	Message string `json:"message"`
}

// HandleAssignRole assigns a role to a user. Only accessible by Admin users.
//
//	@Summary		Assign role to user
//	@Description	Assigns a role (admin or doctor) to a user. Admin-only endpoint.
//	@Tags			roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		assignRoleRequest	true	"Role assignment details"
//	@Success		200		{object}	assignRoleResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/roles/assign [post]
func HandleAssignRole(svc *service.RoleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req assignRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if req.UserID == 0 || req.Role == "" {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationRequiredFields, "user_id and role are required")
			return
		}

		err := svc.AssignRole(r.Context(), req.UserID, req.Role)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidRole):
				apierrors.WriteError(w, http.StatusBadRequest,
					apierrors.RoleInvalidRole, "invalid role value")
			case errors.Is(err, service.ErrUserNotFound):
				apierrors.WriteError(w, http.StatusNotFound,
					apierrors.RoleUserNotFound, "user not found")
			default:
				apierrors.WriteError(w, http.StatusInternalServerError,
					apierrors.RoleAssignmentFailed, "failed to assign role")
			}
			return
		}

		m.IncrementRoleAssignmentCount()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(assignRoleResponse{ //nolint:errcheck // response write
			UserID:  req.UserID,
			Role:    req.Role,
			Message: "role assigned successfully",
		})
	}
}
