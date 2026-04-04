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
	// Rate limit public auth endpoints: 5 req/s per IP, burst of 10
	authLimiter := middleware.NewRateLimiter(5, 10)

	router.Route("/auth", func(r chi.Router) {
		r.With(authLimiter.Handler).Post("/register", handlers.HandleRegister(ar.authSvc, ar.metrics))
		r.With(authLimiter.Handler).Post("/login", handlers.HandleLogin(ar.authSvc, ar.config, ar.metrics))
		r.With(authLimiter.Handler).Post("/logout", handlers.HandleLogout(ar.authSvc, ar.metrics))
		r.With(authLimiter.Handler).Post("/refresh", handlers.HandleRefreshToken(ar.authSvc, ar.config, ar.metrics))

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
