package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/httputil"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/services"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/utils"
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

// List godoc
// @Summary List appointments
// @Description Returns appointments filtered by optional query parameters.
// @Tags appointments
// @Accept json
// @Produce json
// @Param date query string false "Filter by date (YYYY-MM-DD)"
// @Param client_id query string false "Filter by client identifier"
// @Param provider_id query string false "Filter by provider identifier"
// @Param branch query string false "Filter by branch"
// @Success 200 {array} models.Appointment
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Security ApiKeyAuth
// @Router /appointments [get]
// List handles GET /appointments and returns filtered appointments.
// Query parameters: date (YYYY-MM-DD), client_id, provider_id, branch, status, start_date, end_date, limit, offset.
func (h *AppointmentHandler) List(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date != "" {
		if _, err := time.Parse("2006-01-02", date); err != nil {
			httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid date format (expected YYYY-MM-DD)", nil)
			return
		}
	}

	startDate := r.URL.Query().Get("start_date")
	if startDate != "" {
		if _, err := time.Parse("2006-01-02", startDate); err != nil {
			httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid start_date format (expected YYYY-MM-DD)", nil)
			return
		}
	}

	endDate := r.URL.Query().Get("end_date")
	if endDate != "" {
		if _, err := time.Parse("2006-01-02", endDate); err != nil {
			httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid end_date format (expected YYYY-MM-DD)", nil)
			return
		}
	}

	status := r.URL.Query().Get("status")
	if status != "" && !models.ValidateStatus(status) {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid status", nil)
		return
	}

	// Parse pagination parameters
	limit := 50 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		if limit, err = utils.ParseInt(limitStr); err != nil {
			httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid limit parameter", nil)
			return
		}
		if limit > 200 {
			limit = 200 // max limit
		}
		if limit < 1 {
			limit = 1
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		var err error
		if offset, err = utils.ParseInt(offsetStr); err != nil {
			httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid offset parameter", nil)
			return
		}
		if offset < 0 {
			offset = 0
		}
	}

	filter := models.AppointmentFilter{
		Date:       date,
		ClientID:   r.URL.Query().Get("client_id"),
		ProviderID: r.URL.Query().Get("provider_id"),
		Branch:     r.URL.Query().Get("branch"),
		Status:     status,
		StartDate:  startDate,
		EndDate:    endDate,
		Limit:      limit,
		Offset:     offset,
	}
	list, total, err := h.svc.ListAppointments(r.Context(), filter)
	if err != nil {
		httputil.InternalError(r.Context(), w, err)
		return
	}

	response := map[string]interface{}{
		"data":   list,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}
	httputil.JSON(w, http.StatusOK, response)
}

// GetByID godoc
// @Summary Get appointment by ID
// @Description Retrieves a single appointment by its identifier.
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Success 200 {object} models.Appointment
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Security ApiKeyAuth
// @Router /appointments/{id} [get]
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

// Create godoc
// @Summary Create appointment
// @Description Creates a new appointment with the provided payload.
// @Tags appointments
// @Accept json
// @Produce json
// @Param appointment body models.CreateAppointmentRequest true "Appointment payload"
// @Success 201 {object} models.Appointment
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Security ApiKeyAuth
// @Router /appointments [post]
// Create handles POST /appointments and creates a new appointment.
// Request body should contain client_id, provider_id, branch, start_time, end_time, and optional notes.
func (h *AppointmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.ClientID == "" || req.ProviderID == "" || req.Branch == "" || req.StartTime.IsZero() || req.EndTime.IsZero() {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "missing required fields", nil)
		return
	}
	if !req.EndTime.After(req.StartTime) || req.StartTime.Before(time.Now().Add(-24*time.Hour)) {
		httputil.Error(r.Context(), w, http.StatusUnprocessableEntity, "invalid time range", nil)
		return
	}

	a, err := h.svc.CreateAppointment(r.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrClientNotFound) {
			httputil.Error(r.Context(), w, http.StatusNotFound, "client not found", err)
			return
		}
		if errors.Is(err, services.ErrProviderNotFound) {
			httputil.Error(r.Context(), w, http.StatusNotFound, "provider not found", err)
			return
		}
		if errors.Is(err, services.ErrClientAlreadyBooked) {
			httputil.Error(r.Context(), w, http.StatusConflict, "client already has an appointment on this date", err)
			return
		}
		if errors.Is(err, services.ErrAppointmentConflict) {
			httputil.Error(r.Context(), w, http.StatusConflict, "provider unavailable for requested time", err)
			return
		}
		httputil.InternalError(r.Context(), w, err)
		return
	}
	httputil.JSON(w, http.StatusCreated, a)
}

// Update godoc
// @Summary Update appointment
// @Description Updates an existing appointment's time, notes, or status.
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Param update body models.UpdateAppointmentRequest true "Update payload"
// @Success 200 {object} models.Appointment
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Security ApiKeyAuth
// @Router /appointments/{id} [put]
// Update handles PUT /appointments/{id} and updates an appointment.
func (h *AppointmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UpdateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.ClientID == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "client_id is required", nil)
		return
	}

	// Get existing appointment to validate time ranges
	existingAppointment, err := h.svc.GetAppointmentByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(r.Context(), w, err)
		return
	}
	if existingAppointment == nil {
		httputil.Error(r.Context(), w, http.StatusNotFound, "appointment not found", nil)
		return
	}

	// Determine the final start and end times for validation
	finalStartTime := existingAppointment.StartTime
	finalEndTime := existingAppointment.EndTime

	if req.StartTime != nil {
		finalStartTime = *req.StartTime
	}
	if req.EndTime != nil {
		finalEndTime = *req.EndTime
	}

	// Validate time range
	if !finalEndTime.After(finalStartTime) {
		httputil.Error(r.Context(), w, http.StatusUnprocessableEntity, "end_time must be after start_time", nil)
		return
	}

	appointment, err := h.svc.UpdateAppointment(r.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrAppointmentNotFound):
			httputil.Error(r.Context(), w, http.StatusNotFound, "appointment not found", err)
		case errors.Is(err, services.ErrAppointmentForbidden):
			httputil.Error(r.Context(), w, http.StatusForbidden, "appointment not owned by client", err)
		case errors.Is(err, services.ErrClientAlreadyBooked):
			httputil.Error(r.Context(), w, http.StatusConflict, "client already has an appointment on this date", err)
		case errors.Is(err, services.ErrAppointmentConflict):
			httputil.Error(r.Context(), w, http.StatusConflict, "provider unavailable for requested time", err)
		case errors.Is(err, services.ErrInvalidStatus):
			httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid status value", err)
		case errors.Is(err, services.ErrInvalidStatusTransition):
			httputil.Error(r.Context(), w, http.StatusUnprocessableEntity, "invalid status transition", err)
		default:
			// Check for time range validation errors from service layer
			errMsg := err.Error()
			if strings.Contains(errMsg, "end time must be after start time") {
				httputil.Error(r.Context(), w, http.StatusUnprocessableEntity, "end_time must be after start_time", err)
				return
			}
			httputil.InternalError(r.Context(), w, err)
		}
		return
	}

	httputil.JSON(w, http.StatusOK, appointment)
}
