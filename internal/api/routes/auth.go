package routes

import (
	"database/sql"

	"dermify-api/config"
	"dermify-api/internal/api/handlers"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// AuthRoutes handles all authentication-related routes.
type AuthRoutes struct {
	db      *sql.DB
	roleSvc *service.RoleService
	config  *config.Configuration
	metrics *metrics.Client
}

// NewAuthRoutes creates a new AuthRoutes instance.
func NewAuthRoutes(db *sql.DB, roleSvc *service.RoleService, cfg *config.Configuration, m *metrics.Client) *AuthRoutes {
	return &AuthRoutes{
		db:      db,
		roleSvc: roleSvc,
		config:  cfg,
		metrics: m,
	}
}

// RegisterRoutes registers all auth routes under the /auth prefix.
func (ar *AuthRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", handlers.HandleRegister(ar.db, ar.roleSvc, ar.metrics))
		r.Post("/login", handlers.HandleLogin(ar.db, ar.config, ar.metrics))
		r.Post("/logout", handlers.HandleLogout(ar.db, ar.metrics))
		r.Post("/refresh", handlers.HandleRefreshToken(ar.db, ar.config, ar.metrics))

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(ar.config))
			r.Get("/me", handlers.HandleGetProfile(ar.db, ar.metrics))
		})
	})
}
