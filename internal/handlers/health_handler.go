package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/httputil"
)

// HealthHandler handles health check endpoints for the service.
type HealthHandler struct {
	db DatabasePinger
}

// NewHealthHandler creates a new health handler with the given database connection.
func NewHealthHandler(db DatabasePinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Liveness godoc
// @Summary Check service liveness
// @Description Returns 200 OK when the service process is running.
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
// Liveness handles GET /health and returns the service liveness status.
// Always returns 200 OK if the service is running.
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness godoc
// @Summary Check service readiness
// @Description Returns 200 OK when dependencies (database) respond in time.
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} httputil.ErrorResponse
// @Router /ready [get]
// Readiness handles GET /ready and returns the service readiness status.
// Returns 200 OK if database is accessible, 503 Service Unavailable otherwise.
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		httputil.Error(r.Context(), w, http.StatusServiceUnavailable, "database unavailable", err)
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
