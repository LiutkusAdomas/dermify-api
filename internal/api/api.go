package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	if err := a.config.Validate(); err != nil {
		a.logger.Error("invalid configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

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

	// Panic recovery first — catches panics from all downstream handlers
	r.Use(middleware.Recoverer(a.logger))

	// Add CORS middleware (before other middleware)
	r.Use(middleware.CORSWithConfig(a.config))

	// Request timeout — cancels context after 30s
	r.Use(middleware.Timeout(30 * time.Second))

	// Add request ID and logging middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.NewLoggingMiddleware(a.logger))
	r.Use(a.NewMetricsMiddleware)

	a.createRoutes(r)

	port := fmt.Sprintf(":%d", a.config.Port)

	server := &http.Server{
		Addr:              port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// Start server in background
	go func() {
		a.logger.Info("starting API", slog.Int("port", a.config.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	a.logger.Info("shutdown signal received", slog.String("signal", sig.String()))

	// Graceful shutdown with 15s deadline
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		a.logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	a.logger.Info("server stopped gracefully")
}
