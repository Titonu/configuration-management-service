package middleware

import (
	"github.com/Titonu/configuration-management-service/pkg/errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles authentication for API requests
type AuthMiddleware struct {
	apiKeys map[string]string // map of API key to user/client ID
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(apiKeys map[string]string) *AuthMiddleware {
	return &AuthMiddleware{
		apiKeys: apiKeys,
	}
}

// Authenticate returns a middleware function that validates API keys
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")

		// Check if Authorization header exists
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.NewErrorResponse(
				"API key is required",
				errors.ErrorCodeUnauthorized,
				nil,
			))
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.NewErrorResponse(
				"Invalid authorization format",
				errors.ErrorCodeUnauthorized,
				nil,
			))
			return
		}

		// Get the API key
		apiKey := parts[1]

		// Validate API key
		clientID, valid := m.apiKeys[apiKey]
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.NewErrorResponse(
				"Invalid API key",
				errors.ErrorCodeUnauthorized,
				nil,
			))
			return
		}

		// Set client ID in context for later use
		c.Set("client_id", clientID)

		// Continue to the next middleware/handler
		c.Next()
	}
}
