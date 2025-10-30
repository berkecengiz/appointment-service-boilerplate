package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/models"
	"github.com/berkecengiz/appointment-service-boilerplate/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAppointmentService is a mock implementation of AppointmentService interface
type MockAppointmentService struct {
	mock.Mock
}

func (m *MockAppointmentService) ListAppointments(ctx context.Context, filter models.AppointmentFilter) ([]models.Appointment, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Appointment), args.Error(1)
}

func (m *MockAppointmentService) GetAppointmentByID(ctx context.Context, id string) (*models.Appointment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Appointment), args.Error(1)
}

func (m *MockAppointmentService) CreateAppointment(ctx context.Context, req models.CreateAppointmentRequest) (*models.Appointment, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Appointment), args.Error(1)
}

func (m *MockAppointmentService) CancelAppointment(ctx context.Context, id string, req models.CancelAppointmentRequest) (*models.Appointment, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Appointment), args.Error(1)
}

func TestCreate_InvalidJSON(t *testing.T) {
	handler := NewAppointmentHandler(services.NewAppointmentService(nil))

	invalidJSON := `{"client_id": "123", "provider_id": `
	req := httptest.NewRequest("POST", "/appointments", strings.NewReader(invalidJSON))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Check that response is JSON
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	// Check that error message doesn't expose internal details
	if errMsg, ok := response["error"].(string); ok {
		if strings.Contains(strings.ToLower(errMsg), "eof") || strings.Contains(strings.ToLower(errMsg), "unexpected end") {
			t.Errorf("error message exposes internal details: %s", errMsg)
		}
		// Verify it's the sanitized message
		if errMsg != "invalid request body" {
			t.Errorf("expected sanitized error message, got: %s", errMsg)
		}
	}

	// request_id will be empty in tests without middleware - that's okay
	// In production, middleware adds it
}

func TestCreate_MissingFields(t *testing.T) {
	handler := NewAppointmentHandler(services.NewAppointmentService(nil))

	req := models.CreateAppointmentRequest{
		ClientID: "123",
		// Missing other required fields
	}
	body, _ := json.Marshal(req)

	httpReq := httptest.NewRequest("POST", "/appointments", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, httpReq)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if errMsg, ok := response["error"].(string); !ok || errMsg != "missing required fields" {
		t.Errorf("expected 'missing required fields' error, got: %v", response["error"])
	}
}

func TestCreate_InvalidTimeRange(t *testing.T) {
	handler := NewAppointmentHandler(services.NewAppointmentService(nil))

	now := time.Now()
	req := models.CreateAppointmentRequest{
		ClientID:   "123",
		ProviderID: "456",
		Branch:     "Main",
		StartTime:  now,
		EndTime:    now.Add(-1 * time.Hour), // End before start
	}
	body, _ := json.Marshal(req)

	httpReq := httptest.NewRequest("POST", "/appointments", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, httpReq)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if errMsg, ok := response["error"].(string); !ok || errMsg != "invalid time range" {
		t.Errorf("expected 'invalid time range' error, got: %v", response["error"])
	}
}

func TestCreate_PastAppointment(t *testing.T) {
	handler := NewAppointmentHandler(services.NewAppointmentService(nil))

	twoDaysAgo := time.Now().Add(-48 * time.Hour)
	req := models.CreateAppointmentRequest{
		ClientID:   "123",
		ProviderID: "456",
		Branch:     "Main",
		StartTime:  twoDaysAgo,
		EndTime:    twoDaysAgo.Add(1 * time.Hour),
	}
	body, _ := json.Marshal(req)

	httpReq := httptest.NewRequest("POST", "/appointments", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, httpReq)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnprocessableEntity)
	}
}

// New tests for complete coverage

func TestCreate_Success(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	now := time.Now().Add(24 * time.Hour)
	req := models.CreateAppointmentRequest{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
	}

	expectedAppt := &models.Appointment{
		ID:         "appt123",
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
		Status:     "scheduled",
	}

	mockSvc.On("CreateAppointment", mock.Anything, mock.MatchedBy(func(r models.CreateAppointmentRequest) bool {
		return r.ClientID == "client123" && r.ProviderID == "provider456" && r.Branch == "Main"
	})).Return(expectedAppt, nil)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/appointments", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, httpReq)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response models.Appointment
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "appt123", response.ID)
	assert.Equal(t, "client123", response.ClientID)
	mockSvc.AssertExpectations(t)
}

func TestCreate_ProviderConflict(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	now := time.Now().Add(24 * time.Hour)
	req := models.CreateAppointmentRequest{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
	}

	mockSvc.On("CreateAppointment", mock.Anything, mock.MatchedBy(func(r models.CreateAppointmentRequest) bool {
		return r.ProviderID == "provider456"
	})).Return(nil, services.ErrAppointmentConflict)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/appointments", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, httpReq)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "provider unavailable for requested time", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestCreate_ClientConflict(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	now := time.Now().Add(24 * time.Hour)
	req := models.CreateAppointmentRequest{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
	}

	mockSvc.On("CreateAppointment", mock.Anything, mock.MatchedBy(func(r models.CreateAppointmentRequest) bool {
		return r.ClientID == "client123"
	})).Return(nil, services.ErrClientAlreadyBooked)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/appointments", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, httpReq)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "client already has an appointment on this date", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestCreate_ServiceError(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	now := time.Now().Add(24 * time.Hour)
	req := models.CreateAppointmentRequest{
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
	}

	mockSvc.On("CreateAppointment", mock.Anything, mock.MatchedBy(func(r models.CreateAppointmentRequest) bool {
		return r.ClientID == "client123"
	})).Return(nil, errors.New("database error"))

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/appointments", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, httpReq)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestList_Success(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	now := time.Now()
	appointments := []models.Appointment{
		{
			ID:         "appt1",
			ClientID:   "client123",
			ProviderID: "provider456",
			Branch:     "Main",
			StartTime:  now,
			EndTime:    now.Add(time.Hour),
			Status:     "scheduled",
		},
		{
			ID:         "appt2",
			ClientID:   "client456",
			ProviderID: "provider789",
			Branch:     "East",
			StartTime:  now.Add(2 * time.Hour),
			EndTime:    now.Add(3 * time.Hour),
			Status:     "scheduled",
		},
	}

	filter := models.AppointmentFilter{
		ClientID: "client123",
	}

	mockSvc.On("ListAppointments", mock.Anything, filter).Return(appointments, nil)

	httpReq := httptest.NewRequest("GET", "/appointments?client_id=client123", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []models.Appointment
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "appt1", response[0].ID)
	mockSvc.AssertExpectations(t)
}

func TestList_WithFilters(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	filter := models.AppointmentFilter{
		Date:       "2025-10-20",
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
	}

	mockSvc.On("ListAppointments", mock.Anything, filter).Return([]models.Appointment{}, nil)

	httpReq := httptest.NewRequest("GET", "/appointments?date=2025-10-20&client_id=client123&provider_id=provider456&branch=Main", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestList_ServiceError(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	mockSvc.On("ListAppointments", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))

	httpReq := httptest.NewRequest("GET", "/appointments", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, httpReq)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestList_EmptyResult(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	mockSvc.On("ListAppointments", mock.Anything, mock.Anything).Return([]models.Appointment{}, nil)

	httpReq := httptest.NewRequest("GET", "/appointments", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []models.Appointment
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Empty(t, response)
	mockSvc.AssertExpectations(t)
}

func TestList_InvalidDateFormat(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	httpReq := httptest.NewRequest("GET", "/appointments?date=2024-13-01", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid date format (expected YYYY-MM-DD)", response["error"])

	mockSvc.AssertNotCalled(t, "ListAppointments", mock.Anything, mock.Anything)
}

func TestGetByID_Success(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	now := time.Now()
	appointment := &models.Appointment{
		ID:         "appt123",
		ClientID:   "client456",
		ProviderID: "provider789",
		Branch:     "Main",
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
		Status:     "scheduled",
	}

	mockSvc.On("GetAppointmentByID", mock.Anything, "appt123").Return(appointment, nil)

	httpReq := httptest.NewRequest("GET", "/appointments/appt123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "appt123")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.GetByID(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.Appointment
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "appt123", response.ID)
	assert.Equal(t, "client456", response.ClientID)
	mockSvc.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	mockSvc.On("GetAppointmentByID", mock.Anything, "nonexistent").Return(nil, nil)

	httpReq := httptest.NewRequest("GET", "/appointments/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.GetByID(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "appointment not found", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestGetByID_ServiceError(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	mockSvc.On("GetAppointmentByID", mock.Anything, "appt123").Return(nil, errors.New("database error"))

	httpReq := httptest.NewRequest("GET", "/appointments/appt123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "appt123")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.GetByID(rr, httpReq)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestCancel_Success(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	now := time.Now()
	expected := &models.Appointment{
		ID:         "appt123",
		ClientID:   "client123",
		ProviderID: "provider456",
		Branch:     "Main",
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
		Status:     "cancelled",
	}

	mockSvc.On("CancelAppointment", mock.Anything, "appt123", mock.MatchedBy(func(req models.CancelAppointmentRequest) bool {
		return req.ClientID == "client123"
	})).Return(expected, nil)

	body := `{"client_id":"client123"}`
	httpReq := httptest.NewRequest("POST", "/appointments/appt123/cancel", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "appt123")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Cancel(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.Appointment
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", response.Status)
	mockSvc.AssertExpectations(t)
}

func TestCancel_InvalidJSON(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	httpReq := httptest.NewRequest("POST", "/appointments/appt123/cancel", strings.NewReader("{"))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "appt123")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Cancel(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertNotCalled(t, "CancelAppointment", mock.Anything, mock.Anything, mock.Anything)
}

func TestCancel_MissingClientID(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	body := `{"client_id":""}`
	httpReq := httptest.NewRequest("POST", "/appointments/appt123/cancel", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "appt123")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Cancel(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertNotCalled(t, "CancelAppointment", mock.Anything, mock.Anything, mock.Anything)
}

func TestCancel_NotFound(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	mockSvc.On("CancelAppointment", mock.Anything, "appt123", mock.AnythingOfType("models.CancelAppointmentRequest")).
		Return(nil, services.ErrAppointmentNotFound)

	body := `{"client_id":"client123"}`
	httpReq := httptest.NewRequest("POST", "/appointments/appt123/cancel", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "appt123")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Cancel(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "appointment not found", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestCancel_Forbidden(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	mockSvc.On("CancelAppointment", mock.Anything, "appt123", mock.AnythingOfType("models.CancelAppointmentRequest")).
		Return(nil, services.ErrAppointmentForbidden)

	body := `{"client_id":"client123"}`
	httpReq := httptest.NewRequest("POST", "/appointments/appt123/cancel", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "appt123")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Cancel(rr, httpReq)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "appointment not owned by client", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestCancel_AlreadyCancelled(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	mockSvc.On("CancelAppointment", mock.Anything, "appt123", mock.AnythingOfType("models.CancelAppointmentRequest")).
		Return(nil, services.ErrAppointmentAlreadyCancelled)

	body := `{"client_id":"client123"}`
	httpReq := httptest.NewRequest("POST", "/appointments/appt123/cancel", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "appt123")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Cancel(rr, httpReq)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "appointment already cancelled", response["error"])
	mockSvc.AssertExpectations(t)
}

func TestCancel_ServiceError(t *testing.T) {
	mockSvc := new(MockAppointmentService)
	handler := NewAppointmentHandler(mockSvc)

	mockSvc.On("CancelAppointment", mock.Anything, "appt123", mock.AnythingOfType("models.CancelAppointmentRequest")).
		Return(nil, errors.New("database error"))

	body := `{"client_id":"client123"}`
	httpReq := httptest.NewRequest("POST", "/appointments/appt123/cancel", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "appt123")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Cancel(rr, httpReq)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", response["error"])
	mockSvc.AssertExpectations(t)
}
