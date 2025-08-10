package http

import (
	"github.com/Titonu/configuration-management-service/internal/delivery/http/handler"
	"github.com/Titonu/configuration-management-service/internal/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the routes for the API
func SetupRoutes(
	router *gin.Engine,
	configHandler *handler.ConfigurationHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	// API version group
	api := router.Group("/api/v1")

	// Apply authentication middleware
	api.Use(authMiddleware.Authenticate())

	// Configuration routes
	config := api.Group("/configurations")
	{
		// Create a new configuration
		config.POST("", configHandler.CreateConfiguration)

		// Get a configuration
		config.GET("/:name", configHandler.GetConfiguration)

		// Update a configuration
		config.PUT("/:name", configHandler.UpdateConfiguration)

		// List configuration versions
		config.GET("/:name/versions", configHandler.ListConfigurationVersions)

		// Get a specific version of a configuration
		config.GET("/:name/versions/:version", configHandler.GetConfigurationVersion)

		// Rollback a configuration to a previous version
		config.POST("/:name/rollback", configHandler.RollbackConfiguration)
	}

	// Schema routes
	schema := api.Group("/schemas")
	{
		// Register a schema for a configuration
		schema.POST("/:name", configHandler.RegisterSchema)

		// Get a schema for a configuration
		schema.GET("/:name", configHandler.GetSchema)
	}

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}
