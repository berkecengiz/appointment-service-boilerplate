package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type stubClientService struct {
	listFn   func(context.Context) ([]models.Client, error)
	getFn    func(context.Context, string) (*models.Client, error)
	createFn func(context.Context, models.CreateClientRequest) (*models.Client, error)
}

func (s *stubClientService) ListClients(ctx context.Context) ([]models.Client, error) {
	if s.listFn != nil {
		return s.listFn(ctx)
	}
	return nil, nil
}

func (s *stubClientService) GetClientByID(ctx context.Context, id string) (*models.Client, error) {
	if s.getFn != nil {
		return s.getFn(ctx, id)
	}
	return nil, nil
}

func (s *stubClientService) CreateClient(ctx context.Context, req models.CreateClientRequest) (*models.Client, error) {
	if s.createFn != nil {
		return s.createFn(ctx, req)
	}
	return nil, nil
}

func TestClientHandler_CreateSuccess(t *testing.T) {
	svc := &stubClientService{
		createFn: func(ctx context.Context, req models.CreateClientRequest) (*models.Client, error) {
			return &models.Client{ID: "1", Name: req.Name, Email: req.Email}, nil
		},
	}
	handler := NewClientHandler(svc)

	payload := models.CreateClientRequest{
		Name:    "Jane",
		Email:   "jane@example.com",
		Phone:   "123",
		Address: "Street",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/clients", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestClientHandler_CreateConflict(t *testing.T) {
	svc := &stubClientService{
		createFn: func(ctx context.Context, req models.CreateClientRequest) (*models.Client, error) {
			return nil, services.ErrClientEmailExists
		},
	}
	handler := NewClientHandler(svc)

	payload := models.CreateClientRequest{
		Name:    "Jane",
		Email:   "jane@example.com",
		Phone:   "123",
		Address: "Street",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/clients", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestClientHandler_CreateServerError(t *testing.T) {
	svc := &stubClientService{
		createFn: func(ctx context.Context, req models.CreateClientRequest) (*models.Client, error) {
			return nil, errors.New("db down")
		},
	}
	handler := NewClientHandler(svc)

	payload := models.CreateClientRequest{
		Name:    "Jane",
		Email:   "jane@example.com",
		Phone:   "123",
		Address: "Street",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/clients", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestClientHandler_ListSuccess(t *testing.T) {
	svc := &stubClientService{
		listFn: func(ctx context.Context) ([]models.Client, error) {
			return []models.Client{
				{ID: "1", Name: "Alice", Email: "alice@example.com"},
				{ID: "2", Name: "Bob", Email: "bob@example.com"},
			}, nil
		},
	}
	handler := NewClientHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/clients", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []models.Client
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestClientHandler_ListError(t *testing.T) {
	svc := &stubClientService{
		listFn: func(ctx context.Context) ([]models.Client, error) {
			return nil, errors.New("query failed")
		},
	}
	handler := NewClientHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/clients", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestClientHandler_GetByIDSuccess(t *testing.T) {
	svc := &stubClientService{
		getFn: func(ctx context.Context, id string) (*models.Client, error) {
			return &models.Client{ID: id, Name: "Alice", Email: "alice@example.com"}, nil
		},
	}
	handler := NewClientHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	rr := httptest.NewRecorder()

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Client
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "1", resp.ID)
}

func TestClientHandler_GetByIDNotFound(t *testing.T) {
	svc := &stubClientService{
		getFn: func(ctx context.Context, id string) (*models.Client, error) {
			return nil, nil
		},
	}
	handler := NewClientHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	rr := httptest.NewRecorder()

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestClientHandler_GetByIDError(t *testing.T) {
	svc := &stubClientService{
		getFn: func(ctx context.Context, id string) (*models.Client, error) {
			return nil, errors.New("query failed")
		},
	}
	handler := NewClientHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/clients/1", nil)
	rr := httptest.NewRecorder()

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
