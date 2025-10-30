package httputil

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/middlewares"
)

type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// JSON writes a JSON response with the given status code and value.
func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode json response", "error", err)
	}
}

// Error logs the detailed error internally and returns a sanitized error to the client.
// The internal error details are logged but not exposed in the HTTP response.
func Error(ctx context.Context, w http.ResponseWriter, status int, message string, internalErr error) {
	requestID := middlewares.GetRequestID(ctx)
	serviceName := middlewares.GetServiceName(ctx)

	if internalErr != nil {
		slog.ErrorContext(ctx, "request error",
			"error", internalErr.Error(),
			"status", status,
			"message", message,
			"request_id", requestID,
			"service", serviceName,
		)
	} else {
		slog.WarnContext(ctx, "request error",
			"status", status,
			"message", message,
			"request_id", requestID,
			"service", serviceName,
		)
	}

	response := errorResponse{
		Error:     message,
		RequestID: requestID,
	}

	JSON(w, status, response)
}

// InternalError returns a generic 500 error without exposing internal details.
func InternalError(ctx context.Context, w http.ResponseWriter, internalErr error) {
	Error(ctx, w, http.StatusInternalServerError, "internal server error", internalErr)
}
