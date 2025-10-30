package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	docs "github.com/berkecengiz/appointment-service-boilerplate/docs"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/config"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/db"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/handlers"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/logger"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/middlewares"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/routes"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/services"
)

const (
	rateLimitPerMinute = 100
	rateLimitWindow    = time.Minute

	httpReadTimeout  = 10 * time.Second
	httpWriteTimeout = 15 * time.Second
	httpIdleTimeout  = 60 * time.Second
	shutdownTimeout  = 10 * time.Second
)

// @title Appointment Service API
// @version 1.0
// @description API for managing appointments, availability, and health checks.
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	docs.SwaggerInfo.BasePath = "/"
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)
	slog.Info("starting hms-service", "port", cfg.ServerPort, "log_level", cfg.LogLevel)

	// Connect to database
	bunDB, err := db.NewPostgres(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer bunDB.Close()
	slog.Info("database connected successfully")

	// Run database migrations at startup to ensure schema is current.
	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migrateCancel()
	if err := db.RunMigrations(migrateCtx, bunDB); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrations applied")

	// Initialize services and handlers
	apptSvc := services.NewAppointmentService(bunDB)
	apptHandler := handlers.NewAppointmentHandler(apptSvc)
	healthHandler := handlers.NewHealthHandler(bunDB)

	// Initialize middlewares
	authMw := middlewares.NewAPIKeyMiddleware(cfg.APIKeys)
	rateLimiter := middlewares.NewRateLimiter(rateLimitPerMinute, rateLimitWindow)

	// Setup router
	r := routes.NewRouter(routes.Deps{
		AppointmentHandler: apptHandler,
		HealthHandler:      healthHandler,
		AuthMiddleware:     authMw,
		RateLimiter:        rateLimiter,
	})

	// Configure HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
		IdleTimeout:  httpIdleTimeout,
	}

	// Start server in goroutine
	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	shutdownOnSignal(srv)
}

func shutdownOnSignal(srv *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sig := <-stop

	slog.Info("shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		_ = srv.Close()
	} else {
		slog.Info("server shutdown complete")
	}
}
