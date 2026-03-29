package routes

import (
	"dermify-api/config"
	"dermify-api/internal/api/handlers"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// AuthRoutes handles all authentication-related routes.
type AuthRoutes struct {
	authSvc *service.AuthService
	userSvc *service.UserService
	orgSvc  *service.OrganizationService
	config  *config.Configuration
	metrics *metrics.Client
}

// NewAuthRoutes creates a new AuthRoutes instance.
func NewAuthRoutes(authSvc *service.AuthService, userSvc *service.UserService, orgSvc *service.OrganizationService, cfg *config.Configuration, m *metrics.Client) *AuthRoutes {
	return &AuthRoutes{
		authSvc: authSvc,
		userSvc: userSvc,
		orgSvc:  orgSvc,
		config:  cfg,
		metrics: m,
	}
}

// RegisterRoutes registers all auth routes under the /auth prefix.
func (ar *AuthRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", handlers.HandleRegister(ar.authSvc, ar.metrics))
		r.Post("/login", handlers.HandleLogin(ar.authSvc, ar.config, ar.metrics))
		r.Post("/logout", handlers.HandleLogout(ar.authSvc, ar.metrics))
		r.Post("/refresh", handlers.HandleRefreshToken(ar.authSvc, ar.config, ar.metrics))

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(ar.config))
			r.Get("/me", handlers.HandleGetProfile(ar.authSvc, ar.metrics))
			r.Put("/preferences", handlers.HandleUpdatePreferences(ar.userSvc, ar.metrics))
			r.Get("/invitations", handlers.HandleListPendingInvitations(ar.orgSvc, ar.metrics))
			r.Post("/invitations/{token}/accept", handlers.HandleAcceptInvitation(ar.orgSvc, ar.metrics))
			r.Post("/invitations/{token}/decline", handlers.HandleDeclineInvitation(ar.orgSvc, ar.metrics))
		})
	})
}
