package routes

import (
	"database/sql"

	"dermify-api/config"
	"dermify-api/internal/api/handlers"
	"dermify-api/internal/api/metrics"

	"github.com/go-chi/chi/v5"
)

// Manager handles all route registration
type Manager struct {
	authRoutes *AuthRoutes
	userRoutes *UserRoutes
	apiRoutes  *APIRoutes
	metrics    *metrics.Client
}

// NewManager creates a new route manager
func NewManager(db *sql.DB, cfg *config.Configuration, metrics *metrics.Client) *Manager {
	return &Manager{
		metrics:    metrics,
		authRoutes: NewAuthRoutes(db, cfg, metrics),
		userRoutes: NewUserRoutes(db, metrics),
		apiRoutes:  NewAPIRoutes(metrics),
	}
}

// RegisterAllRoutes registers all route modules
func (m *Manager) RegisterAllRoutes(router chi.Router) {
	// Register API v1 routes
	router.Route("/api/v1", func(r chi.Router) {
		m.apiRoutes.RegisterRoutes(r)
		m.authRoutes.RegisterRoutes(r)
		m.userRoutes.RegisterRoutes(r)
	})

	// Register metrics endpoint (outside API versioning)
	router.Get("/metrics", handlers.HandleMetrics(m.metrics.Registry))
}
