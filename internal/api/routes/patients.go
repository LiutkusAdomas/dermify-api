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

// PatientRoutes handles all patient management routes.
type PatientRoutes struct {
	patientSvc *service.PatientService
	config     *config.Configuration
	metrics    *metrics.Client
}

// NewPatientRoutes creates a new PatientRoutes instance.
func NewPatientRoutes(patientSvc *service.PatientService, cfg *config.Configuration, m *metrics.Client) *PatientRoutes {
	return &PatientRoutes{
		patientSvc: patientSvc,
		config:     cfg,
		metrics:    m,
	}
}

// RegisterRoutes registers all patient routes under the /patients prefix.
func (pr *PatientRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/patients", func(r chi.Router) {
		r.Use(middleware.RequireAuth(pr.config))
		r.Use(middleware.RequireRole(domain.RoleDoctor, domain.RoleAdmin))

		r.Get("/", handlers.HandleListPatients(pr.patientSvc, pr.metrics))
		r.Post("/", handlers.HandleCreatePatient(pr.patientSvc, pr.metrics))
		r.Get("/{id}", handlers.HandleGetPatient(pr.patientSvc, pr.metrics))
		r.Put("/{id}", handlers.HandleUpdatePatient(pr.patientSvc, pr.metrics))
		r.Get("/{id}/sessions", handlers.HandleGetPatientSessions(pr.patientSvc, pr.metrics))
	})
}
