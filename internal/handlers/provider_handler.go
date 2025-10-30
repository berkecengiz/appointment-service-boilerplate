package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/httputil"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
	"github.com/go-chi/chi/v5"
)

// ProviderHandler manages HTTP requests for provider resources.
type ProviderHandler struct {
	svc ProviderService
}

// NewProviderHandler returns a handler for the provider service.
func NewProviderHandler(svc ProviderService) *ProviderHandler {
	return &ProviderHandler{svc: svc}
}

// List godoc
// @Summary List providers
// @Description Returns all providers.
// @Tags providers
// @Accept json
// @Produce json
// @Success 200 {array} models.Provider
// @Failure 500 {object} httputil.ErrorResponse
// @Security ApiKeyAuth
// @Router /providers [get]
func (h *ProviderHandler) List(w http.ResponseWriter, r *http.Request) {
	providers, err := h.svc.ListProviders(r.Context())
	if err != nil {
		httputil.InternalError(r.Context(), w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, providers)
}

// GetByID godoc
// @Summary Get provider by ID
// @Description Retrieves a single provider by identifier.
// @Tags providers
// @Accept json
// @Produce json
// @Param id path string true "Provider ID"
// @Success 200 {object} models.Provider
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Security ApiKeyAuth
// @Router /providers/{id} [get]
func (h *ProviderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	provider, err := h.svc.GetProviderByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(r.Context(), w, err)
		return
	}
	if provider == nil {
		httputil.Error(r.Context(), w, http.StatusNotFound, "provider not found", nil)
		return
	}
	httputil.JSON(w, http.StatusOK, provider)
}

// Create godoc
// @Summary Create provider
// @Description Creates a new provider.
// @Tags providers
// @Accept json
// @Produce json
// @Param provider body models.CreateProviderRequest true "Provider payload"
// @Success 201 {object} models.Provider
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Security ApiKeyAuth
// @Router /providers [post]
func (h *ProviderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.Name == "" || req.Email == "" || req.Phone == "" || req.Address == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "missing required fields", nil)
		return
	}

	provider, err := h.svc.CreateProvider(r.Context(), req)
	if err != nil {
		httputil.InternalError(r.Context(), w, err)
		return
	}
	httputil.JSON(w, http.StatusCreated, provider)
}
