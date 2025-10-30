package routes

import (
	"net/http"
	"time"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/handlers"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	maxRequestBodySize = 1 << 20 // 1MB
	requestTimeout     = 30 * time.Second
	compressionLevel   = 5
)

type Deps struct {
	AppointmentHandler *handlers.AppointmentHandler
	ClientHandler      *handlers.ClientHandler
	ProviderHandler    *handlers.ProviderHandler
	HealthHandler      *handlers.HealthHandler
	AuthMiddleware     func(http.Handler) http.Handler
	RateLimiter        *middlewares.RateLimiter
}

func NewRouter(d Deps) *chi.Mux {
	r := chi.NewRouter()

	// Global middlewares
	r.Use(middlewares.RequestID) // Add request ID to context
	r.Use(middleware.RealIP)     // Get real IP from X-Real-IP or X-Forwarded-For
	r.Use(middleware.Recoverer)  // Recover from panics
	r.Use(middleware.RequestSize(maxRequestBodySize))
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(middleware.Compress(compressionLevel))

	// Health endpoints (no auth required)
	r.Get("/health", d.HealthHandler.Liveness)
	r.Get("/ready", d.HealthHandler.Readiness)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Group(func(pr chi.Router) {
		pr.Use(d.AuthMiddleware)
		pr.Use(d.RateLimiter.Middleware)

		pr.Route("/appointments", func(ar chi.Router) {
			ar.Get("/", d.AppointmentHandler.List)
			ar.Post("/", d.AppointmentHandler.Create)
			ar.Get("/{id}", d.AppointmentHandler.GetByID)
		})

		pr.Route("/clients", func(cr chi.Router) {
			cr.Get("/", d.ClientHandler.List)
			cr.Post("/", d.ClientHandler.Create)
			cr.Get("/{id}", d.ClientHandler.GetByID)
		})

		pr.Route("/providers", func(prr chi.Router) {
			prr.Get("/", d.ProviderHandler.List)
			prr.Post("/", d.ProviderHandler.Create)
			prr.Get("/{id}", d.ProviderHandler.GetByID)
		})
	})

	return r
}
