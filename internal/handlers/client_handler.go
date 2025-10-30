package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/httputil"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/services"
	"github.com/go-chi/chi/v5"
)

// ClientHandler handles HTTP requests for client resources.
type ClientHandler struct {
	svc ClientService
}

// NewClientHandler constructs a handler backed by the provided client service.
func NewClientHandler(svc ClientService) *ClientHandler {
	return &ClientHandler{svc: svc}
}

// List godoc
// @Summary List clients
// @Description Returns all clients.
// @Tags clients
// @Accept json
// @Produce json
// @Success 200 {array} models.Client
// @Failure 500 {object} httputil.ErrorResponse
// @Security ApiKeyAuth
// @Router /clients [get]
func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	clients, err := h.svc.ListClients(r.Context())
	if err != nil {
		httputil.InternalError(r.Context(), w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, clients)
}

// GetByID godoc
// @Summary Get client by ID
// @Description Retrieves a single client by identifier.
// @Tags clients
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {object} models.Client
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Security ApiKeyAuth
// @Router /clients/{id} [get]
func (h *ClientHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	client, err := h.svc.GetClientByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(r.Context(), w, err)
		return
	}
	if client == nil {
		httputil.Error(r.Context(), w, http.StatusNotFound, "client not found", nil)
		return
	}
	httputil.JSON(w, http.StatusOK, client)
}

// Create godoc
// @Summary Create client
// @Description Creates a new client.
// @Tags clients
// @Accept json
// @Produce json
// @Param client body models.CreateClientRequest true "Client payload"
// @Success 201 {object} models.Client
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Security ApiKeyAuth
// @Router /clients [post]
func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.Name == "" || req.Email == "" || req.Phone == "" || req.Address == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "missing required fields", nil)
		return
	}

	client, err := h.svc.CreateClient(r.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrClientEmailExists) {
			httputil.Error(r.Context(), w, http.StatusConflict, "client email already exists", err)
			return
		}
		httputil.InternalError(r.Context(), w, err)
		return
	}
	httputil.JSON(w, http.StatusCreated, client)
}
