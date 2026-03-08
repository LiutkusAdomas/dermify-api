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

	return &Manager{
		metrics:        m,
		authRoutes:     NewAuthRoutes(db, roleSvc, cfg, m),
		userRoutes:     NewUserRoutes(db, m),
		apiRoutes:      NewAPIRoutes(m),
		roleRoutes:     NewRoleRoutes(roleSvc, cfg, m),
		patientRoutes:  NewPatientRoutes(patientSvc, cfg, m),
		registryRoutes: NewRegistryRoutes(registrySvc, cfg, m),
		sessionRoutes:  NewSessionRoutes(sessionSvc, consentSvc, contraindicationSvc, cfg, m),
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
