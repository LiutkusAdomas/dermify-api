package routes

import (
	"dermify-api/internal/api/handlers"
	"dermify-api/internal/api/metrics"

	"github.com/go-chi/chi/v5"
)

// APIRoutes handles general API routes
type APIRoutes struct {
	metrics *metrics.Client
}

// NewAPIRoutes creates a new APIRoutes instance
func NewAPIRoutes(metrics *metrics.Client) *APIRoutes {
	return &APIRoutes{
		metrics: metrics,
	}
}

// RegisterRoutes registers general API routes.
func (ar *APIRoutes) RegisterRoutes(router chi.Router) {
	router.Get("/health", handlers.HandleHealth())
}
