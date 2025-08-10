package main

import (
	"github.com/Titonu/configuration-management-service/internal/delivery/http"
	"github.com/Titonu/configuration-management-service/internal/delivery/http/handler"
	"github.com/Titonu/configuration-management-service/internal/delivery/http/middleware"
	"github.com/Titonu/configuration-management-service/internal/repository/sqlite"
	"github.com/Titonu/configuration-management-service/internal/usecase"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set up Gin mode
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)
	// Initialize router
	router := gin.Default()

	dbPath := os.Getenv("SQLITE_DB_PATH")
	if dbPath == "" {
		dbPath = "data/config.db"
	}
	// Ensure directory exists
	dir := dbPath[:strings.LastIndex(dbPath, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Initialize SQLite repository
	configRepo, err := sqlite.NewConfigurationRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite repository: %v", err)
	}
	log.Printf("Using SQLite storage at %s", dbPath)

	// Initialize usecase
	configUseCase := usecase.NewConfigurationUseCase(configRepo)

	// Initialize handlers
	configHandler := handler.NewConfigurationHandler(configUseCase)

	// Set up API keys (from environment or configuration)
	apiKeys := parseAPIKeys(os.Getenv("API_KEYS"))
	if len(apiKeys) == 0 {
		// Add a default API key for development
		apiKeys["dev-api-key"] = "development"
		log.Println("WARNING: Using default API key. Set API_KEYS environment variable for production.")
	}

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(apiKeys)

	// Set up routes
	http.SetupRoutes(router, configHandler, authMiddleware)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// parseAPIKeys parses API keys from environment variable
// Format: key1:client1,key2:client2
func parseAPIKeys(keysStr string) map[string]string {
	result := make(map[string]string)

	if keysStr == "" {
		return result
	}

	pairs := strings.Split(keysStr, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, ":")
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return result
}
