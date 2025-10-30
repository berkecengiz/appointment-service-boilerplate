package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyMiddleware_ValidKey(t *testing.T) {
	validKeys := map[string]string{
		"test-key-123": "test-service",
	}

	middleware := NewAPIKeyMiddleware(validKeys)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "success"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestAPIKeyMiddleware_MissingKey(t *testing.T) {
	validKeys := map[string]string{
		"test-key-123": "test-service",
	}

	middleware := NewAPIKeyMiddleware(validKeys)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	// No API key header
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestAPIKeyMiddleware_InvalidKey(t *testing.T) {
	validKeys := map[string]string{
		"test-key-123": "test-service",
	}

	middleware := NewAPIKeyMiddleware(validKeys)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestAPIKeyMiddleware_EmptyKey(t *testing.T) {
	validKeys := map[string]string{
		"test-key-123": "test-service",
	}

	middleware := NewAPIKeyMiddleware(validKeys)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestGetServiceName(t *testing.T) {
	validKeys := map[string]string{
		"test-key-123": "test-service",
	}

	middleware := NewAPIKeyMiddleware(validKeys)

	var capturedServiceName string
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedServiceName = GetServiceName(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if capturedServiceName != "test-service" {
		t.Errorf("GetServiceName returned wrong service name: got %v want %v", capturedServiceName, "test-service")
	}
}

func TestAPIKeyMiddleware_WithRequestID(t *testing.T) {
	validKeys := map[string]string{
		"test-key-123": "test-service",
	}

	authMiddleware := NewAPIKeyMiddleware(validKeys)

	// Chain with RequestID middleware to test writeJSONError with request_id
	handler := RequestID(authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	// Request without API key should return error with request_id
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-req-id")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	// Check response includes request_id
	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to parse JSON response: %v", err)
	}

	if response["request_id"] != "test-req-id" {
		t.Errorf("expected request_id in response, got: %v", response)
	}
}
