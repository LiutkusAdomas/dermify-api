package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/auth"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

type createUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateUserRequest struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Bio      *string `json:"bio"`
}

// HandleListUsers returns all users.
//
//	@Summary		List users
//	@Description	Returns a list of all users (admin only)
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		UserResponse
//	@Failure		500	{object}	apierrors.ErrorResponse
//	@Router			/users [get]
func HandleListUsers(svc *service.UserService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		users, err := svc.List(r.Context())
		if err != nil {
			slog.Error("failed to list users", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.UserOperationFailed, "failed to list users")
			return
		}

		resp := make([]UserResponse, len(users))
		for i, u := range users {
			resp[i] = toUserResponse(u)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp) //nolint:errcheck // response write
	}
}

// HandleCreateUser creates a new user (admin only).
//
//	@Summary		Create user
//	@Description	Creates a new user (admin only)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		createUserRequest	true	"User details"
//	@Success		201		{object}	UserResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/users [post]
func HandleCreateUser(svc *service.UserService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req createUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if req.Username == "" || req.Email == "" || req.Password == "" {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationRequiredFields, "username, email, and password are required")
			return
		}

		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.InternalPasswordProcessing, "failed to process password")
			return
		}

		user := &domain.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: hash,
		}

		if err := svc.Create(r.Context(), user); err != nil {
			handleUserError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toUserResponse(user)) //nolint:errcheck // response write
	}
}

// HandleGetUser returns a single user by ID (admin only).
//
//	@Summary		Get user
//	@Description	Returns a specific user by ID (admin only)
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	UserResponse
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/users/{id} [get]
func HandleGetUser(svc *service.UserService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.UserInvalidData, "invalid user ID")
			return
		}

		user, err := svc.GetByID(r.Context(), id)
		if err != nil {
			handleUserError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toUserResponse(user)) //nolint:errcheck // response write
	}
}

// HandleUpdateUser updates a user by ID (admin only).
//
//	@Summary		Update user
//	@Description	Updates an existing user by ID (admin only)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int					true	"User ID"
//	@Param			request	body		updateUserRequest	true	"Updated user details"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Router			/users/{id} [put]
func HandleUpdateUser(svc *service.UserService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.UserInvalidData, "invalid user ID")
			return
		}

		var req updateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		user := &domain.User{
			ID:       id,
			Username: req.Username,
			Email:    req.Email,
			Bio:      req.Bio,
		}

		if err := svc.Update(r.Context(), user); err != nil {
			handleUserError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toUserResponse(user)) //nolint:errcheck // response write
	}
}

// HandleDeleteUser deletes a user by ID (admin only).
//
//	@Summary		Delete user
//	@Description	Deletes a user by ID (admin only)
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	MessageResponse
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/users/{id} [delete]
func HandleDeleteUser(svc *service.UserService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.UserInvalidData, "invalid user ID")
			return
		}

		if err := svc.Delete(r.Context(), id); err != nil {
			handleUserError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{ //nolint:errcheck // response write
			Message: "user deleted successfully",
		})
	}
}

// toUserResponse converts a domain user to an API response.
func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Bio:       u.Bio,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// handleUserError maps service user errors to HTTP responses.
func handleUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.UserNotFound, "user not found")
	case errors.Is(err, service.ErrUserAlreadyExists):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.UserAlreadyExists, "username or email already exists")
	case errors.Is(err, service.ErrInvalidUserData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.UserInvalidData, "invalid user data")
	default:
		slog.Error("user operation failed", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.UserOperationFailed, "user operation failed")
	}
}
