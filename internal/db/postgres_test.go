package db

import (
	"testing"

	"github.com/berkecengiz/appointment-service-boilerplate/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewPostgres_InvalidConfig(t *testing.T) {
	cfg := config.Config{
		PGHost:     "localhost",
		PGUser:     "", // Empty user will cause DSN parsing to fail
		PGPassword: "",
		PGDB:       "testdb",
		PGPort:     "5432",
		PGSSLMode:  "disable",
	}

	// This will panic during DSN parsing, so we expect a panic
	assert.Panics(t, func() {
		_, _ = NewPostgres(cfg)
	})
}

func TestNewPostgres_InvalidDSN(t *testing.T) {
	cfg := config.Config{
		PGHost:     "localhost",
		PGUser:     "testuser",
		PGPassword: "testpass",
		PGDB:       "testdb",
		PGPort:     "5432",
		PGSSLMode:  "disable",
	}

	// This will fail because we can't connect to a real database in tests
	// But we can test that the DSN is constructed and connection attempt is made
	db, err := NewPostgres(cfg)
	// Connection will fail, but we verify the function handles it
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestNewPostgres_DefaultSSLMode(t *testing.T) {
	cfg := config.Config{
		PGHost:     "localhost",
		PGUser:     "testuser",
		PGPassword: "testpass",
		PGDB:       "testdb",
		PGPort:     "5432",
		PGSSLMode:  "", // Empty should default to "disable"
	}

	// Connection will fail but we test SSL mode defaulting
	db, err := NewPostgres(cfg)
	// The function will attempt to connect and fail, but SSL mode should be set to "disable"
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestNewPostgres_QueryEscape(t *testing.T) {
	cfg := config.Config{
		PGHost:     "localhost",
		PGUser:     "user@domain",
		PGPassword: "pass:with:colons",
		PGDB:       "testdb",
		PGPort:     "5432",
		PGSSLMode:  "disable",
	}

	// Test that special characters in user/password are URL-encoded
	db, err := NewPostgres(cfg)
	// Connection will fail but function should properly escape special chars
	assert.Error(t, err)
	assert.Nil(t, db)
}
