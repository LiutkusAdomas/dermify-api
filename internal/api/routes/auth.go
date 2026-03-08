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
	config  *config.Configuration
	metrics *metrics.Client
}

// NewAuthRoutes creates a new AuthRoutes instance.
func NewAuthRoutes(authSvc *service.AuthService, cfg *config.Configuration, m *metrics.Client) *AuthRoutes {
	return &AuthRoutes{
		authSvc: authSvc,
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
		})
	})
}
