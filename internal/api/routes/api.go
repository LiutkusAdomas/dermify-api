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

// RegisterRoutes registers general API routes
func (ar *APIRoutes) RegisterRoutes(router chi.Router) {
	// Health check and general endpoints
	router.Get("/health", handlers.HandleHealth())
	router.Get("/hello", handlers.HandleHello())
	router.Get("/foo", handlers.HandleFoo(ar.metrics))
}
