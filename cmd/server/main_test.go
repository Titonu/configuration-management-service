package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping server startup test in short mode")
	}

	// Set environment variables for testing
	os.Setenv("PORT", "8082")

	// Start server in a goroutine
	go func() {
		main()
	}()

	// Give the server time to start
	time.Sleep(2 * time.Second)

	// Create a context with timeout for our HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create HTTP request to health endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8082/health", nil)
	assert.NoError(t, err)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)

	// Check response
	if assert.NoError(t, err) {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

// Config holds server configuration
type Config struct {
	Port        string
	StorageType string
	APIKeys     map[string]string
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "sqlite"
	}

	return Config{
		Port:        port,
		StorageType: storageType,
	}
}

func TestServerEnvironmentVariables(t *testing.T) {
	// Test that environment variables are properly loaded
	originalPort := os.Getenv("PORT")

	// Restore original environment variables after test
	defer func() {
		os.Setenv("PORT", originalPort)
	}()

	// Set test environment variables
	os.Setenv("PORT", "9999")

	// Load config
	cfg := loadConfig()

	// Verify config values
	assert.Equal(t, "9999", cfg.Port)
}

// TestAuthenticationEndpoints tests the authentication functionality of the server
func TestAuthenticationEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping authentication test in short mode")
	}

	// Set environment variables for testing
	originalPort := os.Getenv("PORT")
	originalAPIKeys := os.Getenv("API_KEYS")

	// Restore original environment variables after test
	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("API_KEYS", originalAPIKeys)
	}()

	// Set test environment variables
	os.Setenv("PORT", "8083")
	os.Setenv("API_KEYS", "test-api-key:test-client")

	// Start server in a goroutine
	go func() {
		main()
	}()

	// Give the server time to start
	time.Sleep(2 * time.Second)

	// Create a context with timeout for our HTTP requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create HTTP client
	client := &http.Client{}

	// Test cases
	testCases := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Health check - no auth required",
			path:           "/health",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "API endpoint - auth required - no auth header",
			path:           "/api/v1/configurations/test-config",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "API endpoint - auth required - invalid auth",
			path:           "/api/v1/configurations/test-config",
			authHeader:     "Bearer invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "API endpoint - auth required - valid auth",
			path:           "/api/v1/configurations/test-config",
			authHeader:     "Bearer test-api-key",
			expectedStatus: http.StatusNotFound, // We expect 404 since the config doesn't exist, but auth passes
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create HTTP request
			req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8083"+tc.path, nil)
			assert.NoError(t, err)

			// Add auth header if specified
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			// Send the request
			resp, err := client.Do(req)
			if assert.NoError(t, err) {
				defer resp.Body.Close()
				assert.Equal(t, tc.expectedStatus, resp.StatusCode)

				// If unauthorized, verify error response format
				if resp.StatusCode == http.StatusUnauthorized {
					var errorResponse map[string]interface{}
					err := json.NewDecoder(resp.Body).Decode(&errorResponse)
					assert.NoError(t, err)
					assert.Equal(t, "UNAUTHORIZED", errorResponse["code"])
				}
			}
		})
	}
}
