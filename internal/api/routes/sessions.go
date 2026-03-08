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
	sessionSvc    *service.SessionService
	consentSvc    *service.ConsentService
	screenSvc     *service.ContraindicationService
	energySvc     *service.EnergyModuleService
	injectableSvc *service.InjectableModuleService
	outcomeSvc    *service.OutcomeService
	config        *config.Configuration
	metrics       *metrics.Client
}

// NewSessionRoutes creates a new SessionRoutes instance.
func NewSessionRoutes(
	sessionSvc *service.SessionService,
	consentSvc *service.ConsentService,
	screenSvc *service.ContraindicationService,
	energySvc *service.EnergyModuleService,
	injectableSvc *service.InjectableModuleService,
	outcomeSvc *service.OutcomeService,
	cfg *config.Configuration,
	m *metrics.Client,
) *SessionRoutes {
	return &SessionRoutes{
		sessionSvc:    sessionSvc,
		consentSvc:    consentSvc,
		screenSvc:     screenSvc,
		energySvc:     energySvc,
		injectableSvc: injectableSvc,
		outcomeSvc:    outcomeSvc,
		config:        cfg,
		metrics:       m,
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

			// Energy module type-specific routes.
			r.Route("/modules/ipl", func(r chi.Router) {
				r.Post("/", handlers.HandleCreateIPLModule(sr.energySvc, sr.metrics))
				r.Get("/{moduleId}", handlers.HandleGetIPLModule(sr.energySvc, sr.metrics))
				r.Put("/{moduleId}", handlers.HandleUpdateIPLModule(sr.energySvc, sr.metrics))
			})
			r.Route("/modules/ndyag", func(r chi.Router) {
				r.Post("/", handlers.HandleCreateNdYAGModule(sr.energySvc, sr.metrics))
				r.Get("/{moduleId}", handlers.HandleGetNdYAGModule(sr.energySvc, sr.metrics))
				r.Put("/{moduleId}", handlers.HandleUpdateNdYAGModule(sr.energySvc, sr.metrics))
			})
			r.Route("/modules/co2", func(r chi.Router) {
				r.Post("/", handlers.HandleCreateCO2Module(sr.energySvc, sr.metrics))
				r.Get("/{moduleId}", handlers.HandleGetCO2Module(sr.energySvc, sr.metrics))
				r.Put("/{moduleId}", handlers.HandleUpdateCO2Module(sr.energySvc, sr.metrics))
			})
			r.Route("/modules/rf", func(r chi.Router) {
				r.Post("/", handlers.HandleCreateRFModule(sr.energySvc, sr.metrics))
				r.Get("/{moduleId}", handlers.HandleGetRFModule(sr.energySvc, sr.metrics))
				r.Put("/{moduleId}", handlers.HandleUpdateRFModule(sr.energySvc, sr.metrics))
			})

			// Injectable module type-specific routes.
			r.Route("/modules/filler", func(r chi.Router) {
				r.Post("/", handlers.HandleCreateFillerModule(sr.injectableSvc, sr.metrics))
				r.Get("/{moduleId}", handlers.HandleGetFillerModule(sr.injectableSvc, sr.metrics))
				r.Put("/{moduleId}", handlers.HandleUpdateFillerModule(sr.injectableSvc, sr.metrics))
			})
			r.Route("/modules/botulinum", func(r chi.Router) {
				r.Post("/", handlers.HandleCreateBotulinumModule(sr.injectableSvc, sr.metrics))
				r.Get("/{moduleId}", handlers.HandleGetBotulinumModule(sr.injectableSvc, sr.metrics))
				r.Put("/{moduleId}", handlers.HandleUpdateBotulinumModule(sr.injectableSvc, sr.metrics))
			})

			// Outcome routes (session-level singleton, like consent).
			r.Post("/outcome", handlers.HandleRecordOutcome(sr.outcomeSvc, sr.metrics))
			r.Get("/outcome", handlers.HandleGetOutcome(sr.outcomeSvc, sr.metrics))
			r.Put("/outcome", handlers.HandleUpdateOutcome(sr.outcomeSvc, sr.metrics))
		})
	})
}
