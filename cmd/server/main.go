package main

import (
	"github.com/Titonu/configuration-management-service/internal/delivery/http"
	"github.com/Titonu/configuration-management-service/internal/delivery/http/handler"
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

	// Set up routes
	http.SetupRoutes(router, configHandler)

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
