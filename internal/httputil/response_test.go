package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON_ValidStruct(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"message": "success"}

	JSON(rr, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "success", response["message"])
}

func TestJSON_NilValue(t *testing.T) {
	rr := httptest.NewRecorder()

	JSON(rr, http.StatusNoContent, nil)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
	assert.Empty(t, rr.Body.String())
}

func TestJSON_ComplexStruct(t *testing.T) {
	rr := httptest.NewRecorder()

	type TestData struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	data := TestData{
		ID:    "123",
		Name:  "Test",
		Count: 42,
	}

	JSON(rr, http.StatusCreated, data)

	// Verify status code
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify content type
	assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))

	// Verify serialization
	var response TestData
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "123", response.ID)
	assert.Equal(t, "Test", response.Name)
	assert.Equal(t, 42, response.Count)
}

func TestJSON_UnserializableValue(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	slog.SetDefault(logger)

	rr := httptest.NewRecorder()

	// Channels cannot be JSON encoded
	unserializable := map[string]interface{}{
		"channel": make(chan int),
	}

	JSON(rr, http.StatusOK, unserializable)

	// Should still set headers and status code
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

	// Log should contain error message
	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "failed to encode json response")
}

func TestError_WithInternalError(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	slog.SetDefault(logger)

	// Create a request with request ID through middleware
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "req-123")

	handler := middlewares.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		internalErr := errors.New("database connection failed")
		Error(r.Context(), w, http.StatusInternalServerError, "internal server error", internalErr)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response errorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", response.Error)
	assert.Equal(t, "req-123", response.RequestID)

	// Check log contains error details
	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "database connection failed")
	assert.Contains(t, logOutput, "req-123")
}

func TestError_WithoutInternalError(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	rr := httptest.NewRecorder()

	Error(ctx, rr, http.StatusBadRequest, "missing required fields", nil)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response errorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "missing required fields", response.Error)

	// Should log at warning level, not error level
	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "missing required fields")
}

func TestError_IncludesRequestID(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-request-id")

	handler := middlewares.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Error(r.Context(), w, http.StatusNotFound, "resource not found", nil)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var response errorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "test-request-id", response.RequestID)
}

func TestError_IncludesServiceName(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	slog.SetDefault(logger)

	validKeys := map[string]string{
		"test-api-key": "test-service",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-api-key")

	authMiddleware := middlewares.NewAPIKeyMiddleware(validKeys)
	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Error(r.Context(), w, http.StatusUnauthorized, "unauthorized", nil)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Service name should be in logs
	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "test-service")
}

func TestError_SanitizesMessage(t *testing.T) {
	ctx := context.Background()
	rr := httptest.NewRecorder()

	// Internal error has sensitive details
	internalErr := errors.New("SQL injection attempt: DROP TABLE users")

	// But response should only show sanitized message
	Error(ctx, rr, http.StatusBadRequest, "invalid input", internalErr)

	// Verify JSON content type
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

	var response errorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	// Response should not contain internal error details
	assert.Equal(t, "invalid input", response.Error)
	assert.NotContains(t, response.Error, "DROP TABLE")
	assert.NotContains(t, response.Error, "SQL injection")
}

func TestInternalError_CallsError(t *testing.T) {
	ctx := context.Background()
	rr := httptest.NewRecorder()

	internalErr := errors.New("unexpected database error")

	InternalError(ctx, rr, internalErr)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response errorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", response.Error)
}

func TestInternalError_UsesCorrectStatusCode(t *testing.T) {
	ctx := context.Background()
	rr := httptest.NewRecorder()

	InternalError(ctx, rr, errors.New("any error"))

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestInternalError_UsesGenericMessage(t *testing.T) {
	ctx := context.Background()
	rr := httptest.NewRecorder()

	// Even if internal error is specific
	specificErr := errors.New("deadlock detected in row 42 of table users")

	InternalError(ctx, rr, specificErr)

	var response errorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	// Response should use generic message
	assert.Equal(t, "internal server error", response.Error)
	assert.NotContains(t, response.Error, "deadlock")
	assert.NotContains(t, response.Error, "users")
}

func TestError_DifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		message    string
	}{
		{"BadRequest", http.StatusBadRequest, "bad request"},
		{"Unauthorized", http.StatusUnauthorized, "unauthorized"},
		{"Forbidden", http.StatusForbidden, "forbidden"},
		{"NotFound", http.StatusNotFound, "not found"},
		{"InternalServerError", http.StatusInternalServerError, "internal error"},
		{"ServiceUnavailable", http.StatusServiceUnavailable, "service unavailable"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			rr := httptest.NewRecorder()

			Error(ctx, rr, tc.statusCode, tc.message, nil)

			assert.Equal(t, tc.statusCode, rr.Code)

			var response errorResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tc.message, response.Error)
		})
	}
}
