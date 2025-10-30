package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_AllowsRequestsWithinLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute) // 5 requests per minute

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make 5 requests - all should succeed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("request %d: expected status %v, got %v", i+1, http.StatusOK, status)
		}
	}
}

func TestRateLimiter_BlocksRequestsOverLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute) // 3 requests per minute

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make 3 requests - should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("request %d: expected status %v, got %v", i+1, http.StatusOK, status)
		}
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusTooManyRequests {
		t.Errorf("expected rate limit status %v, got %v", http.StatusTooManyRequests, status)
	}
}

func TestRateLimiter_RefillsTokensAfterWindow(t *testing.T) {
	rl := NewRateLimiter(2, 100*time.Millisecond) // 2 requests per 100ms

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use first 2 tokens
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("initial request %d failed: got %v", i+1, status)
		}
	}

	// 3rd request should be blocked
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusTooManyRequests {
		t.Errorf("expected rate limit, got %v", status)
	}

	// Wait for window to reset
	time.Sleep(150 * time.Millisecond)

	// Should be able to make requests again
	req = httptest.NewRequest("GET", "/test", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("after refill: expected %v, got %v", http.StatusOK, status)
	}
}

func TestRateLimiter_SeparateBucketsPerKey(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	// Create handler that extracts service name from header for testing
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Service1 makes 2 requests - should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "service1"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("service1 request %d failed: got %v", i+1, status)
		}
	}

	// Service1's 3rd request should be blocked
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "service1"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusTooManyRequests {
		t.Errorf("service1 should be rate limited, got %v", status)
	}

	// Service2 should still have tokens available
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "service2"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("service2 should succeed, got %v", status)
	}
}
