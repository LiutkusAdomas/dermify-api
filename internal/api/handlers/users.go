package handlers

import (
	"dermify-api/internal/api/metrics"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// HandleListUsers handles listing all users.
//
//	@Summary		List users
//	@Description	Returns a list of all users
//	@Tags			users
//	@Produce		json
//	@Success		200	{array}		UserResponse
//	@Router			/users [get]
func HandleListUsers(metrics *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		users := []UserResponse{
			{ID: 1, Username: "admin", Email: "admin@example.com"},
			{ID: 2, Username: "user1", Email: "user1@example.com"},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users) //nolint:errcheck // response write
	}
}

// HandleCreateUser handles creating a new user.
//
//	@Summary		Create user
//	@Description	Creates a new user
//	@Tags			users
//	@Produce		json
//	@Success		201	{object}	CreateUserResponse
//	@Router			/users [post]
func HandleCreateUser(metrics *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := CreateUserResponse{
			ID:       3,
			Username: "newuser",
			Email:    "newuser@example.com",
			Message:  "User created successfully",
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response) //nolint:errcheck // response write
	}
}

// HandleGetUser handles getting a specific user.
//
//	@Summary		Get user
//	@Description	Returns a specific user by ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	UserResponse
//	@Router			/users/{id} [get]
func HandleGetUser(metrics *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID := chi.URLParam(r, "id")

		response := UserResponse{
			ID:       0,
			Username: "user" + userID,
			Email:    "user" + userID + "@example.com",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response) //nolint:errcheck // response write
	}
}

// HandleUpdateUser handles updating a user.
//
//	@Summary		Update user
//	@Description	Updates an existing user by ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	UpdateUserResponse
//	@Router			/users/{id} [put]
func HandleUpdateUser(metrics *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID := chi.URLParam(r, "id")

		response := UpdateUserResponse{
			ID:       userID,
			Username: "updateduser" + userID,
			Email:    "updated" + userID + "@example.com",
			Message:  "User updated successfully",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response) //nolint:errcheck // response write
	}
}

// HandleDeleteUser handles deleting a user.
//
//	@Summary		Delete user
//	@Description	Deletes a user by ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	MessageResponse
//	@Router			/users/{id} [delete]
func HandleDeleteUser(metrics *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID := chi.URLParam(r, "id")

		response := MessageResponse{
			Message: "User " + userID + " deleted successfully",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response) //nolint:errcheck // response write
	}
}
