package routes

import (
	"dermify-api/config"
	"dermify-api/internal/api/handlers"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// ScheduleRoutes handles doctor schedule routes.
type ScheduleRoutes struct {
	scheduleSvc *service.ScheduleService
	orgSvc      *service.OrganizationService
	config      *config.Configuration
	metrics     *metrics.Client
}

// NewScheduleRoutes creates a new ScheduleRoutes instance.
func NewScheduleRoutes(scheduleSvc *service.ScheduleService, orgSvc *service.OrganizationService, cfg *config.Configuration, m *metrics.Client) *ScheduleRoutes {
	return &ScheduleRoutes{
		scheduleSvc: scheduleSvc,
		orgSvc:      orgSvc,
		config:      cfg,
		metrics:     m,
	}
}

// RegisterOrgRoutes registers schedule routes under /orgs/{orgId}.
func (sr *ScheduleRoutes) RegisterOrgRoutes(router chi.Router) {
	router.Route("/orgs/{orgId}/doctors/{doctorId}/schedule", func(r chi.Router) {
		r.Use(middleware.RequireAuth(sr.config))
		r.Use(middleware.RequireOrgScheduleAccess(sr.orgSvc))

		r.Get("/", handlers.HandleGetWorkingHours(sr.scheduleSvc, sr.metrics))
		r.Put("/", handlers.HandleSetWorkingHours(sr.scheduleSvc, sr.metrics))
		r.Get("/overrides", handlers.HandleListOverrides(sr.scheduleSvc, sr.metrics))
		r.Post("/overrides", handlers.HandleCreateOverride(sr.scheduleSvc, sr.metrics))
		r.Delete("/overrides/{id}", handlers.HandleDeleteOverride(sr.scheduleSvc, sr.metrics))
	})
}
