package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestID_GeneratesIDWhenNotProvided(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		assert.NotEmpty(t, requestID)
		assert.Len(t, requestID, 32) // 16 bytes = 32 hex characters
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequestID_UsesExistingIDFromHeader(t *testing.T) {
	existingID := "my-custom-request-id"

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		assert.Equal(t, existingID, requestID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, existingID)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequestID_AddsIDToResponseHeader(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	responseID := rr.Header().Get(RequestIDHeader)
	assert.NotEmpty(t, responseID)
	assert.Len(t, responseID, 32)
}

func TestRequestID_ResponseHeaderMatchesContext(t *testing.T) {
	var contextID string

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	responseID := rr.Header().Get(RequestIDHeader)
	assert.Equal(t, responseID, contextID)
}

func TestRequestID_PreservesExistingIDInResponse(t *testing.T) {
	existingID := "preserved-request-id"

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, existingID)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	responseID := rr.Header().Get(RequestIDHeader)
	assert.Equal(t, existingID, responseID)
}

func TestRequestID_AddsToContext(t *testing.T) {
	var capturedID string

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.NotEmpty(t, capturedID)
}

func TestGetRequestID_RetrievesFromContext(t *testing.T) {
	testID := "test-request-id-123"

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retrievedID := GetRequestID(r.Context())
		assert.NotEmpty(t, retrievedID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, testID)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetRequestID_ReturnsEmptyForMissingContext(t *testing.T) {
	// Create a request without the middleware
	req := httptest.NewRequest("GET", "/test", nil)

	requestID := GetRequestID(req.Context())

	assert.Empty(t, requestID)
}

func TestGenerateRequestID_Returns32CharHex(t *testing.T) {
	id := generateRequestID()

	assert.Len(t, id, 32)

	// Check that it's valid hex
	for _, c := range id {
		valid := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
		assert.True(t, valid, "Character %c is not valid hex", c)
	}
}

func TestGenerateRequestID_UniqueIDs(t *testing.T) {
	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateRequestID()
		assert.False(t, ids[id], "Duplicate ID generated: %s", id)
		ids[id] = true
	}
}

func TestRequestID_WorksWithMultipleRequests(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		w.Header().Set("Echo-ID", requestID)
		w.WriteHeader(http.StatusOK)
	}))

	// Make multiple requests
	ids := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		id := rr.Header().Get("Echo-ID")
		assert.NotEmpty(t, id)
		ids = append(ids, id)
	}

	// All IDs should be unique
	uniqueIDs := make(map[string]bool)
	for _, id := range ids {
		assert.False(t, uniqueIDs[id], "Found duplicate request ID")
		uniqueIDs[id] = true
	}
}

func TestRequestID_HeaderConstant(t *testing.T) {
	// Verify the header constant is correct
	assert.Equal(t, "X-Request-ID", RequestIDHeader)
}

func TestRequestID_CaseInsensitiveHeader(t *testing.T) {
	// HTTP headers are case-insensitive, but Go's http.Request.Header.Get handles this
	existingID := "case-test-id"

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		assert.Equal(t, existingID, requestID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	// Set with different case
	req.Header.Set("x-request-id", existingID)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
