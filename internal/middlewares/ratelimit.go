package middlewares

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple in-memory token bucket rate limiter per API key.
type RateLimiter struct {
	buckets map[string]*bucket
	mu      sync.RWMutex
	rate    int           // requests per window
	window  time.Duration // time window
}

type bucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a rate limiter with specified rate and window.
// Example: NewRateLimiter(100, time.Minute) = 100 requests per minute.
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
	}

	// Clean up old buckets periodically
	go rl.cleanup()

	return rl
}

// Middleware returns a middleware function that enforces rate limiting per API key.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract service name from context (set by auth middleware)
		serviceName := GetServiceName(r.Context())
		if serviceName == "" {
			// If no service name, use IP as fallback
			serviceName = r.RemoteAddr
		}

		if !rl.allow(serviceName) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// allow checks if a request is allowed for the given key
func (rl *RateLimiter) allow(key string) bool {
	rl.mu.RLock()
	b, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check after acquiring write lock
		b, exists = rl.buckets[key]
		if !exists {
			b = &bucket{
				tokens:     rl.rate,
				lastRefill: time.Now(),
			}
			rl.buckets[key] = b
		}
		rl.mu.Unlock()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	if elapsed >= rl.window {
		b.tokens = rl.rate
		b.lastRefill = now
	}

	// Check if tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// cleanup removes stale buckets periodically to prevent memory leak
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window * 2)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.buckets {
			b.mu.Lock()
			if now.Sub(b.lastRefill) > rl.window*3 {
				delete(rl.buckets, key)
			}
			b.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}
