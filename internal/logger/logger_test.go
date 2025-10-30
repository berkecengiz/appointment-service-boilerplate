package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	testCases := []struct {
		name     string
		level    string
		expected bool
	}{
		{"Debug", "debug", true},
		{"Info", "info", true},
		{"Warn", "warn", true},
		{"Warning", "warning", true},
		{"Error", "error", true},
		{"Invalid_DefaultsToInfo", "invalid", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Init(tc.level)
			assert.NotNil(t, defaultLogger)
		})
	}
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	ctx := context.Background()
	Info(ctx, "test info", "key", "value")

	output := buf.String()
	assert.Contains(t, output, "test info")
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	ctx := context.Background()
	Error(ctx, "test error", "error", "something went wrong")

	output := buf.String()
	assert.Contains(t, output, "test error")
	assert.Contains(t, output, "something went wrong")
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	ctx := context.Background()
	Warn(ctx, "test warn", "reason", "suspicious activity")

	output := buf.String()
	assert.Contains(t, output, "test warn")
	assert.Contains(t, output, "suspicious activity")
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	ctx := context.Background()
	Debug(ctx, "test debug", "detail", "extra info")

	output := buf.String()
	assert.Contains(t, output, "test debug")
	assert.Contains(t, output, "extra info")
}

func TestWith(t *testing.T) {
	Init("info")

	logger := With("service", "test-service", "version", "1.0")

	assert.NotNil(t, logger)
}

func TestFromContext(t *testing.T) {
	Init("info")

	ctx := context.Background()
	logger := FromContext(ctx)

	assert.NotNil(t, logger)
	assert.Equal(t, defaultLogger, logger)
}
