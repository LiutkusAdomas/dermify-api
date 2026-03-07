package routes

import (
	"dermify-api/config"
	"dermify-api/internal/api/handlers"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/domain"
	"dermify-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// RoleRoutes handles all role management routes.
type RoleRoutes struct {
	roleSvc *service.RoleService
	config  *config.Configuration
	metrics *metrics.Client
}

// NewRoleRoutes creates a new RoleRoutes instance.
func NewRoleRoutes(roleSvc *service.RoleService, cfg *config.Configuration, m *metrics.Client) *RoleRoutes {
	return &RoleRoutes{
		roleSvc: roleSvc,
		config:  cfg,
		metrics: m,
	}
}

// RegisterRoutes registers all role routes under the /roles prefix.
func (rr *RoleRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/roles", func(r chi.Router) {
		r.Use(middleware.RequireAuth(rr.config))
		r.Use(middleware.RequireRole(domain.RoleAdmin))
		r.Post("/assign", handlers.HandleAssignRole(rr.roleSvc, rr.metrics))
	})
}
