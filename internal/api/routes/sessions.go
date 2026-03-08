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

// SessionRoutes handles all session lifecycle routes.
type SessionRoutes struct {
	sessionSvc *service.SessionService
	consentSvc *service.ConsentService
	screenSvc  *service.ContraindicationService
	config     *config.Configuration
	metrics    *metrics.Client
}

// NewSessionRoutes creates a new SessionRoutes instance.
func NewSessionRoutes(
	sessionSvc *service.SessionService,
	consentSvc *service.ConsentService,
	screenSvc *service.ContraindicationService,
	cfg *config.Configuration,
	m *metrics.Client,
) *SessionRoutes {
	return &SessionRoutes{
		sessionSvc: sessionSvc,
		consentSvc: consentSvc,
		screenSvc:  screenSvc,
		config:     cfg,
		metrics:    m,
	}
}

// RegisterRoutes registers all session lifecycle routes under the /sessions prefix.
func (sr *SessionRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/sessions", func(r chi.Router) {
		r.Use(middleware.RequireAuth(sr.config))
		r.Use(middleware.RequireRole(domain.RoleDoctor))

		// Session CRUD.
		r.Post("/", handlers.HandleCreateSession(sr.sessionSvc, sr.metrics))
		r.Get("/", handlers.HandleListSessions(sr.sessionSvc, sr.metrics))

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.HandleGetSession(sr.sessionSvc, sr.metrics))
			r.Put("/", handlers.HandleUpdateSession(sr.sessionSvc, sr.metrics))

			// State transition.
			r.Post("/transition", handlers.HandleTransitionSession(sr.sessionSvc, sr.metrics))

			// Consent.
			r.Post("/consent", handlers.HandleRecordConsent(sr.consentSvc, sr.metrics))
			r.Get("/consent", handlers.HandleGetConsent(sr.consentSvc, sr.metrics))
			r.Put("/consent", handlers.HandleUpdateConsent(sr.consentSvc, sr.metrics))

			// Screening.
			r.Post("/screening", handlers.HandleRecordScreening(sr.screenSvc, sr.metrics))
			r.Get("/screening", handlers.HandleGetScreening(sr.screenSvc, sr.metrics))
			r.Put("/screening", handlers.HandleUpdateScreening(sr.screenSvc, sr.metrics))

			// Modules.
			r.Post("/modules", handlers.HandleAddModule(sr.sessionSvc, sr.metrics))
			r.Get("/modules", handlers.HandleListModules(sr.sessionSvc, sr.metrics))
			r.Delete("/modules/{moduleId}", handlers.HandleRemoveModule(sr.sessionSvc, sr.metrics))
		})
	})
}
