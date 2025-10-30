package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/httputil"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
	"github.com/go-chi/chi/v5"
)

// AppointmentHandler handles HTTP requests for appointment operations.
type AppointmentHandler struct {
	svc AppointmentService
}

// NewAppointmentHandler creates a new appointment handler with the given service.
func NewAppointmentHandler(svc AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{svc: svc}
}

// List handles GET /appointments and returns filtered appointments.
// Query parameters: date (YYYY-MM-DD), customer_id, provider_id, branch.
func (h *AppointmentHandler) List(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date != "" {
		if _, err := time.Parse("2006-01-02", date); err != nil {
			httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid date format (expected YYYY-MM-DD)", nil)
			return
		}
	}

	filter := models.AppointmentFilter{
		Date:       date,
		CustomerID: r.URL.Query().Get("customer_id"),
		ProviderID: r.URL.Query().Get("provider_id"),
		Branch:     r.URL.Query().Get("branch"),
	}
	list, err := h.svc.ListAppointments(r.Context(), filter)
	if err != nil {
		httputil.InternalError(r.Context(), w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, list)
}

// GetByID handles GET /appointments/{id} and returns a specific appointment.
func (h *AppointmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	a, err := h.svc.GetAppointmentByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(r.Context(), w, err)
		return
	}
	if a == nil {
		httputil.Error(r.Context(), w, http.StatusNotFound, "appointment not found", nil)
		return
	}
	httputil.JSON(w, http.StatusOK, a)
}

// Create handles POST /appointments and creates a new appointment.
// Request body should contain customer_id, provider_id, branch, start_time, end_time, and optional notes.
func (h *AppointmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.CustomerID == "" || req.ProviderID == "" || req.Branch == "" || req.StartTime.IsZero() || req.EndTime.IsZero() {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "missing required fields", nil)
		return
	}
	if !req.EndTime.After(req.StartTime) || req.StartTime.Before(time.Now().Add(-24*time.Hour)) {
		httputil.Error(r.Context(), w, http.StatusUnprocessableEntity, "invalid time range", nil)
		return
	}

	a, err := h.svc.CreateAppointment(r.Context(), req)
	if err != nil {
		httputil.InternalError(r.Context(), w, err)
		return
	}
	httputil.JSON(w, http.StatusCreated, a)
}
