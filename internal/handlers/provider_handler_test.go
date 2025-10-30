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

type stubProviderService struct {
	listFn   func(context.Context) ([]models.Provider, error)
	getFn    func(context.Context, string) (*models.Provider, error)
	createFn func(context.Context, models.CreateProviderRequest) (*models.Provider, error)
}

func (s *stubProviderService) ListProviders(ctx context.Context) ([]models.Provider, error) {
	if s.listFn != nil {
		return s.listFn(ctx)
	}
	return nil, nil
}

func (s *stubProviderService) GetProviderByID(ctx context.Context, id string) (*models.Provider, error) {
	if s.getFn != nil {
		return s.getFn(ctx, id)
	}
	return nil, nil
}

func (s *stubProviderService) CreateProvider(ctx context.Context, req models.CreateProviderRequest) (*models.Provider, error) {
	if s.createFn != nil {
		return s.createFn(ctx, req)
	}
	return nil, nil
}

func TestProviderHandler_CreateSuccess(t *testing.T) {
	svc := &stubProviderService{
		createFn: func(ctx context.Context, req models.CreateProviderRequest) (*models.Provider, error) {
			return &models.Provider{ID: "1", Name: req.Name, Email: req.Email}, nil
		},
	}
	handler := NewProviderHandler(svc)

	payload := models.CreateProviderRequest{
		Name:    "Dr. Smith",
		Email:   "smith@example.com",
		Phone:   "123",
		Address: "Street",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/providers", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestProviderHandler_CreateConflict(t *testing.T) {
	svc := &stubProviderService{
		createFn: func(ctx context.Context, req models.CreateProviderRequest) (*models.Provider, error) {
			return nil, services.ErrProviderEmailExists
		},
	}
	handler := NewProviderHandler(svc)

	payload := models.CreateProviderRequest{
		Name:    "Dr. Smith",
		Email:   "smith@example.com",
		Phone:   "123",
		Address: "Street",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/providers", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestProviderHandler_CreateServerError(t *testing.T) {
	svc := &stubProviderService{
		createFn: func(ctx context.Context, req models.CreateProviderRequest) (*models.Provider, error) {
			return nil, errors.New("db down")
		},
	}
	handler := NewProviderHandler(svc)

	payload := models.CreateProviderRequest{
		Name:    "Dr. Smith",
		Email:   "smith@example.com",
		Phone:   "123",
		Address: "Street",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/providers", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestProviderHandler_ListSuccess(t *testing.T) {
	svc := &stubProviderService{
		listFn: func(ctx context.Context) ([]models.Provider, error) {
			return []models.Provider{
				{ID: "1", Name: "Alpha", Email: "alpha@example.com"},
				{ID: "2", Name: "Beta", Email: "beta@example.com"},
			}, nil
		},
	}
	handler := NewProviderHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/providers", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []models.Provider
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestProviderHandler_ListError(t *testing.T) {
	svc := &stubProviderService{
		listFn: func(ctx context.Context) ([]models.Provider, error) {
			return nil, errors.New("query failed")
		},
	}
	handler := NewProviderHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/providers", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestProviderHandler_GetByIDSuccess(t *testing.T) {
	svc := &stubProviderService{
		getFn: func(ctx context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Name: "Alpha", Email: "alpha@example.com"}, nil
		},
	}
	handler := NewProviderHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/providers/1", nil)
	rr := httptest.NewRecorder()

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Provider
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "1", resp.ID)
}

func TestProviderHandler_GetByIDNotFound(t *testing.T) {
	svc := &stubProviderService{
		getFn: func(ctx context.Context, id string) (*models.Provider, error) {
			return nil, nil
		},
	}
	handler := NewProviderHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/providers/1", nil)
	rr := httptest.NewRecorder()

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestProviderHandler_GetByIDError(t *testing.T) {
	svc := &stubProviderService{
		getFn: func(ctx context.Context, id string) (*models.Provider, error) {
			return nil, errors.New("query failed")
		},
	}
	handler := NewProviderHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/providers/1", nil)
	rr := httptest.NewRecorder()

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
