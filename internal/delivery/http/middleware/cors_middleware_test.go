package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.New()

	// Apply CORS middleware
	router.Use(CORSMiddleware())

	// Add a test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test normal GET request
	t.Run("Normal GET request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Check CORS headers
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, w.Header().Get("Access-Control-Expose-Headers"), "Authorization")
		assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
	})

	// Test OPTIONS preflight request
	t.Run("OPTIONS preflight request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check response - should be 204 No Content for OPTIONS
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Check CORS headers
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, w.Header().Get("Access-Control-Expose-Headers"), "Authorization")
		assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
	})
}
