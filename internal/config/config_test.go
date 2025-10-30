package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Set required environment variables
	os.Setenv("API_KEYS", "service1:key1,service2:key2")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpass")
	os.Setenv("PG_DB", "testdb")
	defer cleanupEnv()

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.ServerPort)
	assert.Equal(t, "localhost", cfg.PGHost)
	assert.Equal(t, "testuser", cfg.PGUser)
	assert.Equal(t, "testpass", cfg.PGPassword)
	assert.Equal(t, "testdb", cfg.PGDB)
	assert.Len(t, cfg.APIKeys, 2)
	assert.Equal(t, "service1", cfg.APIKeys["key1"])
	assert.Equal(t, "service2", cfg.APIKeys["key2"])
}

func TestLoad_MissingAPIKeys(t *testing.T) {
	// Set all required vars except API_KEYS
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpass")
	os.Setenv("PG_DB", "testdb")
	defer cleanupEnv()

	_, err := Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API_KEYS")
}

func TestLoad_MissingMSSQLHost(t *testing.T) {
	os.Setenv("API_KEYS", "service1:key1")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpass")
	os.Setenv("PG_DB", "testdb")
	defer cleanupEnv()

	_, err := Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PG_HOST")
}

func TestLoad_MissingMSSQLUser(t *testing.T) {
	os.Setenv("API_KEYS", "service1:key1")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_PASSWORD", "testpass")
	os.Setenv("PG_DB", "testdb")
	defer cleanupEnv()

	_, err := Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PG_USER")
}

func TestLoad_MissingMSSQLPassword(t *testing.T) {
	os.Setenv("API_KEYS", "service1:key1")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_DB", "testdb")
	defer cleanupEnv()

	_, err := Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PG_PASSWORD")
}

func TestLoad_MissingMSSQLDB(t *testing.T) {
	os.Setenv("API_KEYS", "service1:key1")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpass")
	defer cleanupEnv()

	_, err := Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PG_DB")
}

func TestLoad_InvalidServerPort(t *testing.T) {
	os.Setenv("API_KEYS", "service1:key1")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpass")
	os.Setenv("PG_DB", "testdb")
	os.Setenv("SERVER_PORT", "invalid")
	defer cleanupEnv()

	_, err := Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SERVER_PORT")
}

func TestLoad_InvalidServerPortRange(t *testing.T) {
	os.Setenv("API_KEYS", "service1:key1")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpass")
	os.Setenv("PG_DB", "testdb")
	os.Setenv("SERVER_PORT", "70000")
	defer cleanupEnv()

	_, err := Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SERVER_PORT")
}

func TestLoad_InvalidMSSQLPort(t *testing.T) {
	os.Setenv("API_KEYS", "service1:key1")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpass")
	os.Setenv("PG_DB", "testdb")
	os.Setenv("PG_PORT", "invalid")
	defer cleanupEnv()

	_, err := Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PG_PORT")
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	os.Setenv("API_KEYS", "service1:key1")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpass")
	os.Setenv("PG_DB", "testdb")
	os.Setenv("LOG_LEVEL", "invalid")
	defer cleanupEnv()

	_, err := Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LOG_LEVEL")
}

func TestLoad_DefaultValues(t *testing.T) {
	os.Setenv("API_KEYS", "service1:key1")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpass")
	os.Setenv("PG_DB", "testdb")
	// Don't set optional values
	defer cleanupEnv()

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.ServerPort)   // default
	assert.Equal(t, "5432", cfg.PGPort)       // default
	assert.Equal(t, "disable", cfg.PGSSLMode) // default
	assert.Equal(t, "info", cfg.LogLevel)     // default
}

func TestLoad_CustomValues(t *testing.T) {
	os.Setenv("API_KEYS", "service1:key1")
	os.Setenv("PG_HOST", "dbserver")
	os.Setenv("PG_USER", "customuser")
	os.Setenv("PG_PASSWORD", "custompass")
	os.Setenv("PG_DB", "customdb")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("PG_PORT", "15432")
	os.Setenv("PG_SSLMODE", "require")
	os.Setenv("LOG_LEVEL", "debug")
	defer cleanupEnv()

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.ServerPort)
	assert.Equal(t, "15432", cfg.PGPort)
	assert.Equal(t, "require", cfg.PGSSLMode)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestParseAPIKeys_ValidFormat(t *testing.T) {
	raw := "service1:key1,service2:key2,service3:key3"

	result := parseAPIKeys(raw)

	assert.Len(t, result, 3)
	assert.Equal(t, "service1", result["key1"])
	assert.Equal(t, "service2", result["key2"])
	assert.Equal(t, "service3", result["key3"])
}

func TestParseAPIKeys_EmptyString(t *testing.T) {
	result := parseAPIKeys("")

	assert.Empty(t, result)
}

func TestParseAPIKeys_SinglePair(t *testing.T) {
	result := parseAPIKeys("myservice:mykey")

	assert.Len(t, result, 1)
	assert.Equal(t, "myservice", result["mykey"])
}

func TestParseAPIKeys_MultiplePairs(t *testing.T) {
	raw := "svc1:key1,svc2:key2"

	result := parseAPIKeys(raw)

	assert.Len(t, result, 2)
	assert.Equal(t, "svc1", result["key1"])
	assert.Equal(t, "svc2", result["key2"])
}

func TestParseAPIKeys_MalformedEntries(t *testing.T) {
	// Missing colon should be skipped
	raw := "service1:key1,invalidentry,service2:key2"

	result := parseAPIKeys(raw)

	assert.Len(t, result, 2)
	assert.Equal(t, "service1", result["key1"])
	assert.Equal(t, "service2", result["key2"])
}

func TestParseAPIKeys_WithWhitespace(t *testing.T) {
	raw := " service1 : key1 , service2 : key2 "

	result := parseAPIKeys(raw)

	assert.Len(t, result, 2)
	assert.Equal(t, "service1", result["key1"])
	assert.Equal(t, "service2", result["key2"])
}

func TestParseAPIKeys_EmptyServiceName(t *testing.T) {
	// Empty service name should be skipped
	raw := ":key1,service2:key2"

	result := parseAPIKeys(raw)

	assert.Len(t, result, 1)
	assert.Equal(t, "service2", result["key2"])
}

func TestParseAPIKeys_EmptyKey(t *testing.T) {
	// Empty key should be skipped
	raw := "service1:,service2:key2"

	result := parseAPIKeys(raw)

	assert.Len(t, result, 1)
	assert.Equal(t, "service2", result["key2"])
}

func TestParseAPIKeys_MultipleColons(t *testing.T) {
	// Should split on first colon only
	raw := "service1:key:with:colons,service2:key2"

	result := parseAPIKeys(raw)

	assert.Len(t, result, 2)
	assert.Equal(t, "service1", result["key:with:colons"])
	assert.Equal(t, "service2", result["key2"])
}

func TestGetEnv_ValueExists(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")

	assert.Equal(t, "test_value", result)
}

func TestGetEnv_ValueMissing(t *testing.T) {
	result := getEnv("NONEXISTENT_VAR", "default")

	assert.Equal(t, "default", result)
}

func TestGetEnv_EmptyString(t *testing.T) {
	os.Setenv("TEST_VAR", "")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")

	assert.Equal(t, "default", result)
}

func TestGetEnvBool_True(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	result := getEnvBool("TEST_BOOL", false)

	assert.True(t, result)
}

func TestGetEnvBool_False(t *testing.T) {
	os.Setenv("TEST_BOOL", "false")
	defer os.Unsetenv("TEST_BOOL")

	result := getEnvBool("TEST_BOOL", true)

	assert.False(t, result)
}

func TestGetEnvBool_Invalid(t *testing.T) {
	os.Setenv("TEST_BOOL", "invalid")
	defer os.Unsetenv("TEST_BOOL")

	result := getEnvBool("TEST_BOOL", true)

	assert.True(t, result) // Should return default
}

func TestGetEnvBool_Missing(t *testing.T) {
	result := getEnvBool("NONEXISTENT_BOOL", false)

	assert.False(t, result) // Should return default
}

func TestLoad_LogLevelVariations(t *testing.T) {
	testCases := []struct {
		name     string
		logLevel string
		valid    bool
	}{
		{"Debug", "debug", true},
		{"Info", "info", true},
		{"Warn", "warn", true},
		{"Warning", "warning", true},
		{"Error", "error", true},
		{"Invalid", "invalid", false},
		{"Empty", "", false}, // Will use default "info" but then validate
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("API_KEYS", "service1:key1")
			os.Setenv("PG_HOST", "localhost")
			os.Setenv("PG_USER", "testuser")
			os.Setenv("PG_PASSWORD", "testpass")
			os.Setenv("PG_DB", "testdb")
			if tc.logLevel != "" {
				os.Setenv("LOG_LEVEL", tc.logLevel)
			}
			defer cleanupEnv()

			cfg, err := Load()

			if tc.valid {
				require.NoError(t, err)
				if tc.logLevel != "" {
					assert.Equal(t, tc.logLevel, cfg.LogLevel)
				} else {
					assert.Equal(t, "info", cfg.LogLevel) // default
				}
			} else if tc.logLevel != "" {
				assert.Error(t, err)
			}
		})
	}
}

func TestSplitAndTrim_Basic(t *testing.T) {
	result := splitAndTrim("a,b,c", ",")

	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestSplitAndTrim_WithWhitespace(t *testing.T) {
	result := splitAndTrim(" a , b , c ", ",")

	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestSplitAndTrim_EmptyParts(t *testing.T) {
	result := splitAndTrim("a,,b", ",")

	assert.Equal(t, []string{"a", "b"}, result)
}

func TestSplitAndTrim_EmptyString(t *testing.T) {
	result := splitAndTrim("", ",")

	assert.Empty(t, result)
}

// Helper function to clean up environment variables
func cleanupEnv() {
	os.Unsetenv("API_KEYS")
	os.Unsetenv("PG_HOST")
	os.Unsetenv("PG_PORT")
	os.Unsetenv("PG_USER")
	os.Unsetenv("PG_PASSWORD")
	os.Unsetenv("PG_DB")
	os.Unsetenv("PG_SSLMODE")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("LOG_LEVEL")
}
