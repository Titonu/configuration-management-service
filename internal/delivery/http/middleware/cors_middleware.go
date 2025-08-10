package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
)

// CORSMiddleware handles CORS for API requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow requests from any origin
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// Allow credentials
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// Allow all common headers including those used by OpenAPI UI
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Access-Control-Request-Headers, Access-Control-Request-Method")
		// Allow all common methods
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		// Allow headers to be exposed to the browser
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization")
		// Set max age for preflight requests
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			log.Printf("[CORS] Handling OPTIONS preflight request")
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
