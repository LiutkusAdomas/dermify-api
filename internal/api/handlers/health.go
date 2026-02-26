package handlers

import (
	"encoding/json"
	"net/http"
)

// HandleHealth handles health check endpoint
func HandleHealth() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
			"version":   "1.0.0",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
