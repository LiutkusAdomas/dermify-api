package api

import (
	"dermify-api/internal/api/routes"

	"github.com/go-chi/chi/v5"
)

func (a *App) createRoutes(router *chi.Mux) {
	routeManager := routes.NewManager(a.db, a.config, a.metrics)
	routeManager.RegisterAllRoutes(router)
}
