package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration values loaded from environment variables.
type Config struct {
	ServerPort string

	PGHost     string
	PGPort     string
	PGUser     string
	PGPassword string
	PGDB       string
	PGSSLMode  string

	APIKeys map[string]string // map[apiKey]serviceName

	LogLevel string
}

// Load loads configuration from environment variables. It first attempts to load a local .env file
// if present, then reads from the process environment.
func Load() (Config, error) {
	_ = godotenv.Load() // best-effort; ignore if missing

	cfg := Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),

		PGHost:     getEnv("PG_HOST", ""),
		PGPort:     getEnv("PG_PORT", "5432"),
		PGUser:     getEnv("PG_USER", ""),
		PGPassword: getEnv("PG_PASSWORD", ""),
		PGDB:       getEnv("PG_DB", ""),
		PGSSLMode:  getEnv("PG_SSLMODE", "disable"),

		APIKeys:  parseAPIKeys(getEnv("API_KEYS", "")),
		LogLevel: strings.ToLower(getEnv("LOG_LEVEL", "info")),
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func parseAPIKeys(raw string) map[string]string {
	result := make(map[string]string)
	if raw == "" {
		return result
	}

	pairs := splitAndTrim(raw, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			serviceName := strings.TrimSpace(parts[0])
			apiKey := strings.TrimSpace(parts[1])
			if serviceName != "" && apiKey != "" {
				result[apiKey] = serviceName
			}
		}
	}
	return result
}

func validate(c Config) error {
	// Validate API keys
	if len(c.APIKeys) == 0 {
		return fmt.Errorf("API_KEYS must be set with at least one key in format: service1:key1,service2:key2")
	}

	// Validate database configuration
	if c.PGHost == "" {
		return fmt.Errorf("PG_HOST must be set")
	}
	if c.PGUser == "" {
		return fmt.Errorf("PG_USER must be set")
	}
	if c.PGPassword == "" {
		return fmt.Errorf("PG_PASSWORD must be set")
	}
	if c.PGDB == "" {
		return fmt.Errorf("PG_DB must be set")
	}

	// Validate port number
	port, err := strconv.Atoi(c.ServerPort)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("SERVER_PORT must be a valid port number (1-65535)")
	}

	pgPort, err := strconv.Atoi(c.PGPort)
	if err != nil || pgPort < 1 || pgPort > 65535 {
		return fmt.Errorf("PG_PORT must be a valid port number (1-65535)")
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "warning": true, "error": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error (got: %s)", c.LogLevel)
	}

	return nil
}
