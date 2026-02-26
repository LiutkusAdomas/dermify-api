package handlers

import (
	"dermify-api/internal/api/metrics"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// HandleListUsers handles listing all users
func HandleListUsers(metrics *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		users := []map[string]interface{}{
			{"id": 1, "username": "admin", "email": "admin@example.com"},
			{"id": 2, "username": "user1", "email": "user1@example.com"},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	}
}

// HandleCreateUser handles creating a new user
func HandleCreateUser(metrics *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := map[string]interface{}{
			"id":       3,
			"username": "newuser",
			"email":    "newuser@example.com",
			"message":  "User created successfully",
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// HandleGetUser handles getting a specific user
func HandleGetUser(metrics *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID := chi.URLParam(r, "id")

		response := map[string]interface{}{
			"id":       userID,
			"username": "user" + userID,
			"email":    "user" + userID + "@example.com",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// HandleUpdateUser handles updating a user
func HandleUpdateUser(metrics *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID := chi.URLParam(r, "id")

		response := map[string]interface{}{
			"id":       userID,
			"username": "updateduser" + userID,
			"email":    "updated" + userID + "@example.com",
			"message":  "User updated successfully",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// HandleDeleteUser handles deleting a user
func HandleDeleteUser(metrics *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID := chi.URLParam(r, "id")

		response := map[string]string{
			"message": "User " + userID + " deleted successfully",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
