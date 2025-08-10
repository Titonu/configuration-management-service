package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	// Set up test API keys
	apiKeys := map[string]string{
		"valid-key":   "test-client",
		"another-key": "another-client",
	}

	// Create middleware
	authMiddleware := NewAuthMiddleware(apiKeys)

	// Set up Gin router for testing
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(authMiddleware.Authenticate())

	// Add a test handler
	router.GET("/test", func(c *gin.Context) {
		// Get client ID from context
		clientID, exists := c.Get("client_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "client ID not found in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"client_id": clientID})
	})

	t.Run("ValidToken", func(t *testing.T) {
		// Create request with valid token
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-key")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "test-client")
	})

	t.Run("AnotherValidToken", func(t *testing.T) {
		// Create request with another valid token
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer another-key")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "another-client")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Create request with invalid token
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-key")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("MissingToken", func(t *testing.T) {
		// Create request without token
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("MalformedToken", func(t *testing.T) {
		// Create request with malformed token
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "NotBearer token")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("EmptyToken", func(t *testing.T) {
		// Create request with empty token
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
