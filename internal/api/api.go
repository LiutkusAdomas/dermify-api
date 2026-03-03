package api

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"dermify-api/config"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	postgres "dermify-api/internal/pkg"
	"dermify-api/migrations"

	"github.com/go-chi/chi/v5"
)

const appName = "dermify-api"

type App struct {
	logger  *slog.Logger
	config  *config.Configuration
	metrics *metrics.Client
	db      *sql.DB
}

func New(configPath string) *App {
	if configPath == "" {
		configPath = "config.yaml"
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return &App{
		logger:  logger,
		config:  config.Configure(configPath),
		metrics: metrics.New(logger),
	}
}

// Start runs the API. It is triggered by the serve command.
//
//	@title						Dermify API
//	@version					1.0
//	@description				Backend REST API for the Dermify application
//	@host						localhost:8080
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
func (a *App) Start() {
	db, err := postgres.Open(a.config.Database.DSN(), a.logger)
	if err != nil {
		a.logger.Error("connecting to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	a.db = db
	defer a.db.Close()

	if err := postgres.MigrateFS(a.db, migrations.FS, "."); err != nil {
		a.logger.Error("running database migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}
	a.logger.Info("database migrations completed")

	r := chi.NewRouter()

	// Add CORS middleware first (before other middleware)
	r.Use(middleware.CORSWithConfig(a.config))

	// Add request ID and logging middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.NewLoggingMiddleware(a.logger))
	r.Use(a.NewMetricsMiddleware)

	a.createRoutes(r)

	port := fmt.Sprintf(":%d", a.config.Port)
	startupLog := fmt.Sprintf("starting API on port :%d", a.config.Port)
	a.logger.Info(startupLog)
	if err := http.ListenAndServe(port, r); err != nil {
		a.logger.Error("starting the API: ", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
