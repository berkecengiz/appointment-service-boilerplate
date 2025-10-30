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
