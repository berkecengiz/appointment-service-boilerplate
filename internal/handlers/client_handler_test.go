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
