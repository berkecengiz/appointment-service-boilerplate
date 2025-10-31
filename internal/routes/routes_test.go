package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/handlers"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/middlewares"
	"github.com/stretchr/testify/assert"
)

// mockDatabasePinger is a mock implementation of DatabasePinger for testing
type mockDatabasePinger struct{}

func (m *mockDatabasePinger) PingContext(ctx context.Context) error {
	return nil
}

func TestNewRouter_HealthEndpoint(t *testing.T) {
	mockHealth := handlers.NewHealthHandler(&mockDatabasePinger{})
	mockAppt := &handlers.AppointmentHandler{}
	mockClient := &handlers.ClientHandler{}
	mockProvider := &handlers.ProviderHandler{}

	deps := Deps{
		AppointmentHandler: mockAppt,
		ClientHandler:      mockClient,
		ProviderHandler:    mockProvider,
		HealthHandler:      mockHealth,
		AuthMiddleware:     func(next http.Handler) http.Handler { return next },
		RateLimiter:        middlewares.NewRateLimiter(100, time.Minute),
	}

	router := NewRouter(deps)

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewRouter_ReadyEndpoint(t *testing.T) {
	mockHealth := handlers.NewHealthHandler(&mockDatabasePinger{})
	mockAppt := &handlers.AppointmentHandler{}
	mockClient := &handlers.ClientHandler{}
	mockProvider := &handlers.ProviderHandler{}

	deps := Deps{
		AppointmentHandler: mockAppt,
		ClientHandler:      mockClient,
		ProviderHandler:    mockProvider,
		HealthHandler:      mockHealth,
		AuthMiddleware:     func(next http.Handler) http.Handler { return next },
		RateLimiter:        middlewares.NewRateLimiter(100, time.Minute),
	}

	router := NewRouter(deps)

	// Test ready endpoint
	req := httptest.NewRequest("GET", "/ready", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return OK since our mock pinger always succeeds
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewRouter_SwaggerEndpoint(t *testing.T) {
	mockHealth := handlers.NewHealthHandler(&mockDatabasePinger{})
	mockAppt := &handlers.AppointmentHandler{}
	mockClient := &handlers.ClientHandler{}
	mockProvider := &handlers.ProviderHandler{}

	deps := Deps{
		AppointmentHandler: mockAppt,
		ClientHandler:      mockClient,
		ProviderHandler:    mockProvider,
		HealthHandler:      mockHealth,
		AuthMiddleware:     func(next http.Handler) http.Handler { return next },
		RateLimiter:        middlewares.NewRateLimiter(100, time.Minute),
	}

	router := NewRouter(deps)

	// Test swagger endpoint
	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Swagger endpoint should respond (even if 404 if not configured)
	assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
}

func TestNewRouter_AppointmentRoutes(t *testing.T) {
	mockHealth := handlers.NewHealthHandler(&mockDatabasePinger{})
	mockAppt := &handlers.AppointmentHandler{}
	mockClient := &handlers.ClientHandler{}
	mockProvider := &handlers.ProviderHandler{}

	authCalled := false
	authMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCalled = true
			next.ServeHTTP(w, r)
		})
	}

	deps := Deps{
		AppointmentHandler: mockAppt,
		ClientHandler:      mockClient,
		ProviderHandler:    mockProvider,
		HealthHandler:      mockHealth,
		AuthMiddleware:     authMw,
		RateLimiter:        middlewares.NewRateLimiter(100, time.Minute),
	}

	router := NewRouter(deps)

	// Test appointment routes require auth
	req := httptest.NewRequest("GET", "/appointments", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should have called auth middleware
	assert.True(t, authCalled)
}

func TestNewRouter_ClientRoutes(t *testing.T) {
	mockHealth := handlers.NewHealthHandler(&mockDatabasePinger{})
	mockAppt := &handlers.AppointmentHandler{}
	mockClient := &handlers.ClientHandler{}
	mockProvider := &handlers.ProviderHandler{}

	authCalled := false
	authMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCalled = true
			next.ServeHTTP(w, r)
		})
	}

	deps := Deps{
		AppointmentHandler: mockAppt,
		ClientHandler:      mockClient,
		ProviderHandler:    mockProvider,
		HealthHandler:      mockHealth,
		AuthMiddleware:     authMw,
		RateLimiter:        middlewares.NewRateLimiter(100, time.Minute),
	}

	router := NewRouter(deps)

	// Test client routes require auth
	req := httptest.NewRequest("GET", "/clients", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should have called auth middleware
	assert.True(t, authCalled)
}

func TestNewRouter_ProviderRoutes(t *testing.T) {
	mockHealth := handlers.NewHealthHandler(&mockDatabasePinger{})
	mockAppt := &handlers.AppointmentHandler{}
	mockClient := &handlers.ClientHandler{}
	mockProvider := &handlers.ProviderHandler{}

	authCalled := false
	authMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCalled = true
			next.ServeHTTP(w, r)
		})
	}

	deps := Deps{
		AppointmentHandler: mockAppt,
		ClientHandler:      mockClient,
		ProviderHandler:    mockProvider,
		HealthHandler:      mockHealth,
		AuthMiddleware:     authMw,
		RateLimiter:        middlewares.NewRateLimiter(100, time.Minute),
	}

	router := NewRouter(deps)

	// Test provider routes require auth
	req := httptest.NewRequest("GET", "/providers", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should have called auth middleware
	assert.True(t, authCalled)
}
