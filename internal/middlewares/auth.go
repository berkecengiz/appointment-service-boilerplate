package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
)

type serviceNameKey struct{}

// NewAPIKeyMiddleware creates a middleware that validates API keys from the X-API-Key header.
// validKeys maps API key -> service name for identification and logging.
func NewAPIKeyMiddleware(validKeys map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				writeJSONError(w, r, http.StatusUnauthorized, "missing api key")
				return
			}

			serviceName, ok := validKeys[apiKey]
			if !ok {
				writeJSONError(w, r, http.StatusUnauthorized, "invalid api key")
				return
			}

			// Store service name in context for logging/auditing
			ctx := context.WithValue(r.Context(), serviceNameKey{}, serviceName)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetServiceName retrieves the service name from the request context.
// Returns empty string if not found.
func GetServiceName(ctx context.Context) string {
	if name, ok := ctx.Value(serviceNameKey{}).(string); ok {
		return name
	}
	return ""
}

// writeJSONError writes a JSON error response with the given status and message.
func writeJSONError(w http.ResponseWriter, r *http.Request, status int, message string) {
	requestID := GetRequestID(r.Context())

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	response := map[string]string{
		"error": message,
	}
	if requestID != "" {
		response["request_id"] = requestID
	}

	_ = json.NewEncoder(w).Encode(response)
}
