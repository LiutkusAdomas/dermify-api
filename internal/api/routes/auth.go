package routes

import (
	"database/sql"

	"dermify-api/config"
	"dermify-api/internal/api/handlers"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"

	"github.com/go-chi/chi/v5"
)

// AuthRoutes handles all authentication-related routes
type AuthRoutes struct {
	db      *sql.DB
	config  *config.Configuration
	metrics *metrics.Client
}

// NewAuthRoutes creates a new AuthRoutes instance
func NewAuthRoutes(db *sql.DB, cfg *config.Configuration, metrics *metrics.Client) *AuthRoutes {
	return &AuthRoutes{
		db:      db,
		config:  cfg,
		metrics: metrics,
	}
}

// RegisterRoutes registers all auth routes under the /auth prefix
func (ar *AuthRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", handlers.HandleRegister(ar.db, ar.metrics))
		r.Post("/login", handlers.HandleLogin(ar.db, ar.config, ar.metrics))
		r.Post("/logout", handlers.HandleLogout(ar.db, ar.metrics))
		r.Post("/refresh", handlers.HandleRefreshToken(ar.db, ar.config, ar.metrics))

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(ar.config))
			r.Get("/me", handlers.HandleGetProfile(ar.db, ar.metrics))
		})
	})
}
