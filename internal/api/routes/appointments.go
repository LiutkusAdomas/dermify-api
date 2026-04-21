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

// AppointmentRoutes handles appointment and availability routes.
type AppointmentRoutes struct {
	appointmentSvc *service.AppointmentService
	scheduleSvc    *service.ScheduleService
	orgSvc         *service.OrganizationService
	config         *config.Configuration
	metrics        *metrics.Client
}

// NewAppointmentRoutes creates a new AppointmentRoutes instance.
func NewAppointmentRoutes(appointmentSvc *service.AppointmentService, scheduleSvc *service.ScheduleService, orgSvc *service.OrganizationService, cfg *config.Configuration, m *metrics.Client) *AppointmentRoutes {
	return &AppointmentRoutes{
		appointmentSvc: appointmentSvc,
		scheduleSvc:    scheduleSvc,
		orgSvc:         orgSvc,
		config:         cfg,
		metrics:        m,
	}
}

// RegisterOrgRoutes registers appointment routes under /orgs/{orgId}.
func (ar *AppointmentRoutes) RegisterOrgRoutes(router chi.Router) {
	router.Route("/orgs/{orgId}/appointments", func(r chi.Router) {
		r.Use(middleware.RequireAuth(ar.config))
		r.Use(middleware.RequireOrgRole(ar.orgSvc, domain.OrgRoleAdmin, domain.OrgRoleDoctor, domain.OrgRoleReceptionist))

		r.Post("/", handlers.HandleCreateAppointment(ar.appointmentSvc, ar.metrics))
		r.Get("/", handlers.HandleListAppointments(ar.appointmentSvc, ar.metrics))

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.HandleGetAppointment(ar.appointmentSvc, ar.metrics))
			r.Put("/", handlers.HandleUpdateAppointment(ar.appointmentSvc, ar.metrics))
			r.Patch("/status", handlers.HandleUpdateAppointmentStatus(ar.appointmentSvc, ar.metrics))
			r.Post("/start-session", handlers.HandleStartSessionFromAppointment(ar.appointmentSvc, ar.metrics))
		})
	})

	router.Route("/orgs/{orgId}/availability", func(r chi.Router) {
		r.Use(middleware.RequireAuth(ar.config))
		r.Use(middleware.RequireOrgMember(ar.orgSvc))
		r.Get("/", handlers.HandleGetAvailability(ar.scheduleSvc, ar.appointmentSvc, ar.orgSvc, ar.metrics))
	})
}
