package routes

import (
	"database/sql"

	"dermify-api/internal/api/handlers"
	"dermify-api/internal/api/metrics"

	"github.com/go-chi/chi/v5"
)

// UserRoutes handles all user-related routes
type UserRoutes struct {
	db      *sql.DB
	metrics *metrics.Client
}

// NewUserRoutes creates a new UserRoutes instance
func NewUserRoutes(db *sql.DB, metrics *metrics.Client) *UserRoutes {
	return &UserRoutes{
		db:      db,
		metrics: metrics,
	}
}

// RegisterRoutes registers all user routes under the /users prefix
func (ur *UserRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/users", func(r chi.Router) {
		r.Get("/", handlers.HandleListUsers(ur.metrics))
		r.Post("/", handlers.HandleCreateUser(ur.metrics))
		r.Get("/{id}", handlers.HandleGetUser(ur.metrics))
		r.Put("/{id}", handlers.HandleUpdateUser(ur.metrics))
		r.Delete("/{id}", handlers.HandleDeleteUser(ur.metrics))
	})
}
