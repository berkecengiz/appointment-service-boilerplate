package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDatabasePinger is a mock implementation of DatabasePinger interface
type MockDatabasePinger struct {
	mock.Mock
}

func (m *MockDatabasePinger) PingContext(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestLiveness_AlwaysReturnsOK(t *testing.T) {
	mockDB := new(MockDatabasePinger)
	handler := NewHealthHandler(mockDB)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler.Liveness(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])

	// No database call should be made for liveness
	mockDB.AssertNotCalled(t, "PingContext")
}

func TestReadiness_DatabaseHealthy(t *testing.T) {
	mockDB := new(MockDatabasePinger)
	handler := NewHealthHandler(mockDB)

	mockDB.On("PingContext", mock.Anything).Return(nil)

	req := httptest.NewRequest("GET", "/ready", nil)
	rr := httptest.NewRecorder()

	handler.Readiness(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ready", response["status"])

	mockDB.AssertExpectations(t)
}

func TestReadiness_DatabaseUnhealthy(t *testing.T) {
	mockDB := new(MockDatabasePinger)
	handler := NewHealthHandler(mockDB)

	mockDB.On("PingContext", mock.Anything).Return(errors.New("connection refused"))

	req := httptest.NewRequest("GET", "/ready", nil)
	rr := httptest.NewRecorder()

	handler.Readiness(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "database unavailable", response["error"])

	mockDB.AssertExpectations(t)
}

func TestReadiness_DatabaseTimeout(t *testing.T) {
	mockDB := new(MockDatabasePinger)
	handler := NewHealthHandler(mockDB)

	mockDB.On("PingContext", mock.Anything).Return(context.DeadlineExceeded)

	req := httptest.NewRequest("GET", "/ready", nil)
	rr := httptest.NewRecorder()

	handler.Readiness(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "database unavailable", response["error"])

	mockDB.AssertExpectations(t)
}

func TestReadiness_ReturnsJSON(t *testing.T) {
	mockDB := new(MockDatabasePinger)
	handler := NewHealthHandler(mockDB)

	mockDB.On("PingContext", mock.Anything).Return(nil)

	req := httptest.NewRequest("GET", "/ready", nil)
	rr := httptest.NewRecorder()

	handler.Readiness(rr, req)

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	assert.Contains(t, contentType, "application/json")

	mockDB.AssertExpectations(t)
}
