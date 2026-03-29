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
	signoffSvc    *service.SignoffService
	addendumSvc   *service.AddendumService
	auditSvc      *service.AuditService
	photoSvc      *service.PhotoService
	storagePath   string
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
	signoffSvc *service.SignoffService,
	addendumSvc *service.AddendumService,
	auditSvc *service.AuditService,
	photoSvc *service.PhotoService,
	storagePath string,
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
		signoffSvc:    signoffSvc,
		addendumSvc:   addendumSvc,
		auditSvc:      auditSvc,
		photoSvc:      photoSvc,
		storagePath:   storagePath,
		config:        cfg,
		metrics:       m,
	}
}

// RegisterRoutes registers all session lifecycle routes under the /sessions prefix.
func (sr *SessionRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/sessions", func(r chi.Router) {
		r.Use(middleware.RequireAuth(sr.config))
		r.Use(middleware.RequireRole(domain.RoleDoctor, domain.RoleAdmin))

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

			// Sign-off routes.
			r.Get("/signoff/readiness", handlers.HandleGetSignOffReadiness(sr.signoffSvc, sr.metrics))
			r.Post("/signoff", handlers.HandleSignOffSession(sr.signoffSvc, sr.metrics))
			r.Post("/lock", handlers.HandleLockSession(sr.signoffSvc, sr.metrics))

			// Addendum routes.
			r.Post("/addendums", handlers.HandleCreateAddendum(sr.addendumSvc, sr.metrics))
			r.Get("/addendums", handlers.HandleListAddendums(sr.addendumSvc, sr.metrics))
			r.Get("/addendums/{addendumId}", handlers.HandleGetAddendum(sr.addendumSvc, sr.metrics))

			// Audit trail route.
			r.Get("/audit", handlers.HandleGetAuditTrail(sr.auditSvc, sr.metrics))

			// Photo routes.
			r.Route("/photos", func(r chi.Router) {
				r.Post("/before", handlers.HandleUploadBeforePhoto(sr.photoSvc, sr.metrics))
				r.Get("/", handlers.HandleListSessionPhotos(sr.photoSvc, sr.metrics))
				r.Route("/{photoId}", func(r chi.Router) {
					r.Get("/", handlers.HandleGetPhoto(sr.photoSvc, sr.metrics))
					r.Get("/file", handlers.HandleServePhotoFile(sr.photoSvc, sr.storagePath))
					r.Delete("/", handlers.HandleDeletePhoto(sr.photoSvc, sr.metrics))
				})
				// Label photo upload requires moduleId.
				r.Post("/label/{moduleId}", handlers.HandleUploadLabelPhoto(sr.photoSvc, sr.metrics))
			})
		})
	})
}
