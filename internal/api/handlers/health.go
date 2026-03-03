package handlers

import (
	"encoding/json"
	"net/http"
)

// HandleHealth handles health check endpoint.
//
//	@Summary		Health check
//	@Description	Returns the health status of the API
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/health [get]
func HandleHealth() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := HealthResponse{
			Status:    "healthy",
			Timestamp: "2024-01-01T00:00:00Z",
			Version:   "1.0.0",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response) //nolint:errcheck // response write
	}
}
