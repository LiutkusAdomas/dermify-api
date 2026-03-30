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

// ServiceTypeRoutes handles service type routes nested under an org.
type ServiceTypeRoutes struct {
	serviceTypeSvc *service.ServiceTypeService
	orgSvc         *service.OrganizationService
	config         *config.Configuration
	metrics        *metrics.Client
}

// NewServiceTypeRoutes creates a new ServiceTypeRoutes instance.
func NewServiceTypeRoutes(serviceTypeSvc *service.ServiceTypeService, orgSvc *service.OrganizationService, cfg *config.Configuration, m *metrics.Client) *ServiceTypeRoutes {
	return &ServiceTypeRoutes{
		serviceTypeSvc: serviceTypeSvc,
		orgSvc:         orgSvc,
		config:         cfg,
		metrics:        m,
	}
}

// RegisterOrgRoutes registers service type routes under /orgs/{orgId}/service-types.
func (sr *ServiceTypeRoutes) RegisterOrgRoutes(router chi.Router) {
	router.Route("/orgs/{orgId}/service-types", func(r chi.Router) {
		r.Use(middleware.RequireAuth(sr.config))

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireOrgMember(sr.orgSvc))
			r.Get("/", handlers.HandleListServiceTypes(sr.serviceTypeSvc, sr.metrics))
			r.Get("/{id}", handlers.HandleGetServiceType(sr.serviceTypeSvc, sr.metrics))
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireOrgRole(sr.orgSvc, domain.OrgRoleAdmin))
			r.Post("/", handlers.HandleCreateServiceType(sr.serviceTypeSvc, sr.metrics))
			r.Put("/{id}", handlers.HandleUpdateServiceType(sr.serviceTypeSvc, sr.metrics))
			r.Delete("/{id}", handlers.HandleDeleteServiceType(sr.serviceTypeSvc, sr.metrics))
		})
	})
}
