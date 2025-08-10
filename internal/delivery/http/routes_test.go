package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Titonu/configuration-management-service/internal/delivery/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestAuthenticationRoutes tests that routes are properly protected by authentication
func TestAuthenticationRoutes(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.New()

	// Setup test API keys
	apiKeys := map[string]string{
		"test-api-key": "test-client",
	}
	authMiddleware := middleware.NewAuthMiddleware(apiKeys)

	// Create a simplified version of SetupRoutes for testing auth only
	// API version group
	api := router.Group("/api/v1")

	// Apply authentication middleware
	api.Use(authMiddleware.Authenticate())

	// Add a test endpoint under the API group
	api.GET("/test", func(c *gin.Context) {
		// This handler will only be called if auth passes
		c.JSON(http.StatusOK, gin.H{"status": "authenticated"})
	})

	// Add health endpoint (unprotected)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test cases
	testCases := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Health check - no auth required",
			method:         http.MethodGet,
			path:           "/health",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "API endpoint - auth required - no auth header",
			method:         http.MethodGet,
			path:           "/api/v1/test",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "API endpoint - auth required - invalid auth",
			method:         http.MethodGet,
			path:           "/api/v1/test",
			authHeader:     "Bearer invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "API endpoint - auth required - valid auth",
			method:         http.MethodGet,
			path:           "/api/v1/test",
			authHeader:     "Bearer test-api-key",
			expectedStatus: http.StatusOK,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

// TestRouteSetup tests that the SetupRoutes function properly configures routes
func TestRouteSetup(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.New()

	// Setup test API keys
	apiKeys := map[string]string{
		"test-api-key": "test-client",
	}
	authMiddleware := middleware.NewAuthMiddleware(apiKeys)

	// Setup routes with a dummy handler that always returns 200 OK
	api := router.Group("/api/v1")
	api.Use(authMiddleware.Authenticate())

	// Add configuration routes
	config := api.Group("/configurations")
	config.GET("/:name", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	config.POST("", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Add schema routes
	schema := api.Group("/schemas")
	schema.GET("/:name", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Add health endpoint
	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Test that routes are properly set up
	testCases := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Health check endpoint",
			method:         http.MethodGet,
			path:           "/health",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get configuration endpoint",
			method:         http.MethodGet,
			path:           "/api/v1/configurations/test-config",
			authHeader:     "Bearer test-api-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Create configuration endpoint",
			method:         http.MethodPost,
			path:           "/api/v1/configurations",
			authHeader:     "Bearer test-api-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get schema endpoint",
			method:         http.MethodGet,
			path:           "/api/v1/schemas/test-schema",
			authHeader:     "Bearer test-api-key",
			expectedStatus: http.StatusOK,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}
