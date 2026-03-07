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

// RegistryRoutes handles all registry routes.
type RegistryRoutes struct {
	registrySvc *service.RegistryService
	config      *config.Configuration
	metrics     *metrics.Client
}

// NewRegistryRoutes creates a new RegistryRoutes instance.
func NewRegistryRoutes(registrySvc *service.RegistryService, cfg *config.Configuration, m *metrics.Client) *RegistryRoutes {
	return &RegistryRoutes{
		registrySvc: registrySvc,
		config:      cfg,
		metrics:     m,
	}
}

// RegisterRoutes registers all registry routes under the /registry prefix.
func (rr *RegistryRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/registry", func(r chi.Router) {
		r.Use(middleware.RequireAuth(rr.config))
		r.Use(middleware.RequireRole(domain.RoleDoctor, domain.RoleAdmin))

		r.Get("/devices", handlers.HandleListDevices(rr.registrySvc, rr.metrics))
		r.Get("/devices/{id}", handlers.HandleGetDevice(rr.registrySvc, rr.metrics))
		r.Get("/products", handlers.HandleListProducts(rr.registrySvc, rr.metrics))
		r.Get("/products/{id}", handlers.HandleGetProduct(rr.registrySvc, rr.metrics))
		r.Get("/indication-codes", handlers.HandleListIndicationCodes(rr.registrySvc, rr.metrics))
		r.Get("/clinical-endpoints", handlers.HandleListClinicalEndpoints(rr.registrySvc, rr.metrics))
	})
}
