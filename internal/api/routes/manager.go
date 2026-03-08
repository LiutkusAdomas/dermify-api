package routes

import (
	"database/sql"

	"dermify-api/config"
	"dermify-api/internal/api/handlers"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/repository/postgres"
	"dermify-api/internal/service"

	_ "dermify-api/docs"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// Manager handles all route registration.
type Manager struct {
	authRoutes     *AuthRoutes
	userRoutes     *UserRoutes
	apiRoutes      *APIRoutes
	roleRoutes     *RoleRoutes
	patientRoutes  *PatientRoutes
	registryRoutes *RegistryRoutes
	sessionRoutes  *SessionRoutes
	metrics        *metrics.Client
}

// NewManager creates a new route manager.
func NewManager(db *sql.DB, cfg *config.Configuration, m *metrics.Client) *Manager {
	roleRepo := postgres.NewPostgresRoleRepository(db)
	roleSvc := service.NewRoleService(roleRepo)

	patientRepo := postgres.NewPostgresPatientRepository(db)
	patientSvc := service.NewPatientService(patientRepo)

	registryRepo := postgres.NewPostgresRegistryRepository(db)
	registrySvc := service.NewRegistryService(registryRepo)

	sessionRepo := postgres.NewPostgresSessionRepository(db)
	consentRepo := postgres.NewPostgresConsentRepository(db)
	contraindicationRepo := postgres.NewPostgresContraindicationRepository(db)
	moduleRepo := postgres.NewPostgresModuleRepository(db)
	sessionSvc := service.NewSessionService(sessionRepo, consentRepo, moduleRepo)
	consentSvc := service.NewConsentService(consentRepo)
	contraindicationSvc := service.NewContraindicationService(contraindicationRepo)

	iplModuleRepo := postgres.NewPostgresIPLModuleRepository(db)
	ndyagModuleRepo := postgres.NewPostgresNdYAGModuleRepository(db)
	co2ModuleRepo := postgres.NewPostgresCO2ModuleRepository(db)
	rfModuleRepo := postgres.NewPostgresRFModuleRepository(db)
	energySvc := service.NewEnergyModuleService(sessionSvc, registrySvc, iplModuleRepo, ndyagModuleRepo, co2ModuleRepo, rfModuleRepo)

	fillerModuleRepo := postgres.NewPostgresFillerModuleRepository(db)
	botulinumModuleRepo := postgres.NewPostgresBotulinumModuleRepository(db)
	injectableSvc := service.NewInjectableModuleService(sessionSvc, registrySvc, fillerModuleRepo, botulinumModuleRepo)

	outcomeRepo := postgres.NewPostgresOutcomeRepository(db)
	outcomeSvc := service.NewOutcomeService(outcomeRepo, sessionSvc)

	signoffRepo := postgres.NewPostgresSignoffRepository(db)
	signoffSvc := service.NewSignoffService(sessionRepo, consentRepo, moduleRepo, outcomeRepo, signoffRepo)

	addendumRepo := postgres.NewPostgresAddendumRepository(db)
	addendumSvc := service.NewAddendumService(addendumRepo, sessionRepo)

	auditRepo := postgres.NewPostgresAuditRepository(db)
	auditSvc := service.NewAuditService(auditRepo)

	photoRepo := postgres.NewPostgresPhotoRepository(db)
	photoFileStore := service.NewLocalFileStore(cfg.Storage.BasePath)
	photoSvc := service.NewPhotoService(photoRepo, sessionRepo, moduleRepo, photoFileStore)

	userRepo := postgres.NewPostgresUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	authRepo := postgres.NewPostgresAuthRepository(db)
	authSvc := service.NewAuthService(authRepo, userRepo, roleSvc)

	return &Manager{
		metrics:        m,
		authRoutes:     NewAuthRoutes(authSvc, cfg, m),
		userRoutes:     NewUserRoutes(userSvc, cfg, m),
		apiRoutes:      NewAPIRoutes(m),
		roleRoutes:     NewRoleRoutes(roleSvc, cfg, m),
		patientRoutes:  NewPatientRoutes(patientSvc, cfg, m),
		registryRoutes: NewRegistryRoutes(registrySvc, cfg, m),
		sessionRoutes:  NewSessionRoutes(sessionSvc, consentSvc, contraindicationSvc, energySvc, injectableSvc, outcomeSvc, signoffSvc, addendumSvc, auditSvc, photoSvc, cfg.Storage.BasePath, cfg, m),
	}
}

// RegisterAllRoutes registers all route modules.
func (m *Manager) RegisterAllRoutes(router chi.Router) {
	// Register API v1 routes.
	router.Route("/api/v1", func(r chi.Router) {
		m.apiRoutes.RegisterRoutes(r)
		m.authRoutes.RegisterRoutes(r)
		m.userRoutes.RegisterRoutes(r)
		m.roleRoutes.RegisterRoutes(r)
		m.patientRoutes.RegisterRoutes(r)
		m.registryRoutes.RegisterRoutes(r)
		m.sessionRoutes.RegisterRoutes(r)
	})

	// Register metrics endpoint (outside API versioning).
	router.Get("/metrics", handlers.HandleMetrics(m.metrics.Registry))

	// Register Swagger UI.
	router.Get("/swagger/*", httpSwagger.WrapHandler)
}
