package main

import (
	"context"
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
