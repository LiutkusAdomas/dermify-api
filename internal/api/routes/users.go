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

// UserRoutes handles all user-related routes.
type UserRoutes struct {
	userSvc *service.UserService
	config  *config.Configuration
	metrics *metrics.Client
}

// NewUserRoutes creates a new UserRoutes instance.
func NewUserRoutes(userSvc *service.UserService, cfg *config.Configuration, m *metrics.Client) *UserRoutes {
	return &UserRoutes{
		userSvc: userSvc,
		config:  cfg,
		metrics: m,
	}
}

// RegisterRoutes registers all user routes under the /users prefix.
func (ur *UserRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/users", func(r chi.Router) {
		r.Use(middleware.RequireAuth(ur.config))
		r.Use(middleware.RequireRole(domain.RoleAdmin))

		r.Get("/", handlers.HandleListUsers(ur.userSvc, ur.metrics))
		r.Post("/", handlers.HandleCreateUser(ur.userSvc, ur.metrics))
		r.Get("/{id}", handlers.HandleGetUser(ur.userSvc, ur.metrics))
		r.Put("/{id}", handlers.HandleUpdateUser(ur.userSvc, ur.metrics))
		r.Delete("/{id}", handlers.HandleDeleteUser(ur.userSvc, ur.metrics))
	})
}
