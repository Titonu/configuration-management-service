package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Titonu/configuration-management-service/internal/delivery/http/handler"
	"github.com/Titonu/configuration-management-service/internal/domain/repository"
	"github.com/Titonu/configuration-management-service/internal/domain/usecase"
	"github.com/Titonu/configuration-management-service/internal/repository/sqlite"
	implUsecase "github.com/Titonu/configuration-management-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ConfigurationAPITestSuite is a test suite for the Configuration API
type ConfigurationAPITestSuite struct {
	suite.Suite
	router        *gin.Engine
	dbPath        string
	configRepo    repository.ConfigurationRepository
	configUseCase usecase.ConfigurationUsecase
}

// SetupSuite sets up the test suite
func (suite *ConfigurationAPITestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a temporary database for testing
	suite.dbPath = "../../data/test_config.db"

	// Ensure directory exists
	dir := suite.dbPath[:strings.LastIndex(suite.dbPath, "/")]
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		suite.T().Fatalf("Failed to create test database directory: %v", err)
	}

	// Remove existing test database if it exists
	os.Remove(suite.dbPath)

	// Initialize SQLite repository
	configRepo, err := sqlite.NewConfigurationRepository(suite.dbPath)
	if err != nil {
		suite.T().Fatalf("Failed to initialize SQLite repository: %v", err)
	}
	suite.configRepo = configRepo

	// Initialize usecase - using the implementation from internal/usecase
	suite.configUseCase = implUsecase.NewConfigurationUseCase(suite.configRepo)

	// Initialize handlers
	configHandler := handler.NewConfigurationHandler(suite.configUseCase)

	// Initialize router
	suite.router = gin.New()
	suite.router.Use(gin.Recovery())

	// Set up routes
	// API version group
	api := suite.router.Group("/api/v1")

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
	suite.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}

// SetupTest resets the database state before each test
func (suite *ConfigurationAPITestSuite) SetupTest() {
	// Clean up and recreate the test database
	os.Remove(suite.dbPath)

	// Re-initialize SQLite repository
	configRepo, err := sqlite.NewConfigurationRepository(suite.dbPath)
	if err != nil {
		suite.T().Fatalf("Failed to initialize SQLite repository: %v", err)
	}
	suite.configRepo = configRepo

	// Re-initialize usecase
	suite.configUseCase = implUsecase.NewConfigurationUseCase(suite.configRepo)

	// Re-initialize handlers and update router
	configHandler := handler.NewConfigurationHandler(suite.configUseCase)

	// Reset routes
	suite.router = gin.New()
	suite.router.Use(gin.Recovery())

	// Set up routes
	api := suite.router.Group("/api/v1")

	// Configuration routes
	config := api.Group("/configurations")
	{
		config.POST("", configHandler.CreateConfiguration)
		config.GET("/:name", configHandler.GetConfiguration)
		config.PUT("/:name", configHandler.UpdateConfiguration)
		config.GET("/:name/versions", configHandler.ListConfigurationVersions)
		config.GET("/:name/versions/:version", configHandler.GetConfigurationVersion)
		config.POST("/:name/rollback", configHandler.RollbackConfiguration)
	}

	// Schema routes
	schema := api.Group("/schemas")
	{
		schema.POST("/:name", configHandler.RegisterSchema)
		schema.GET("/:name", configHandler.GetSchema)
	}

	// Health check endpoint
	suite.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}

// TearDownSuite tears down the test suite
func (suite *ConfigurationAPITestSuite) TearDownSuite() {
	// Clean up the test database
	os.Remove(suite.dbPath)
}

// TestRegisterSchema tests the schema registration endpoint
func (suite *ConfigurationAPITestSuite) TestRegisterSchema() {
	t := suite.T()

	// Schema for payment-config
	schema := []byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"max_limit": {
				"type": "integer"
			},
			"enabled": {
				"type": "boolean"
			}
		},
		"required": ["max_limit", "enabled"]
	}`)

	// Create a request to register the schema
	req := httptest.NewRequest("POST", "/api/v1/schemas/payment-config", bytes.NewBuffer(schema))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "payment-config", response["name"])
	assert.Equal(t, "schema registered successfully", response["status"])
}

// TestGetSchema tests the schema retrieval endpoint
func (suite *ConfigurationAPITestSuite) TestGetSchema() {
	t := suite.T()

	// First register a schema
	suite.TestRegisterSchema()

	// Create a request to get the schema
	req := httptest.NewRequest("GET", "/api/v1/schemas/payment-config", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var schema map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &schema)
	assert.NoError(t, err)

	// Verify schema structure
	assert.Equal(t, "object", schema["type"])
	properties, ok := schema["properties"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, properties, "max_limit")
	assert.Contains(t, properties, "enabled")
}

// TestCreateConfiguration tests the configuration creation endpoint
func (suite *ConfigurationAPITestSuite) TestCreateConfiguration() {
	t := suite.T()

	// First register a schema
	suite.TestRegisterSchema()

	// Configuration data
	configData := []byte(`{
		"name": "payment-config",
		"data": {
			"max_limit": 1000,
			"enabled": true
		}
	}`)

	// Create a request to create a configuration
	req := httptest.NewRequest(http.MethodPost, "/api/v1/configurations", bytes.NewBuffer(configData))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "payment-config", response["name"])
	assert.Equal(t, float64(1), response["version"])

	// Verify data
	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1000), data["max_limit"])
	assert.Equal(t, true, data["enabled"])
}

// TestGetConfiguration tests the configuration retrieval endpoint
func (suite *ConfigurationAPITestSuite) TestGetConfiguration() {
	t := suite.T()

	// First create a configuration
	suite.TestCreateConfiguration()

	// Create a request to get the configuration
	req := httptest.NewRequest("GET", "/api/v1/configurations/payment-config", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "payment-config", response["name"])
	assert.Equal(t, float64(1), response["version"])

	// Verify data
	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1000), data["max_limit"])
	assert.Equal(t, true, data["enabled"])
}

// TestUpdateConfiguration tests the configuration update endpoint
func (suite *ConfigurationAPITestSuite) TestUpdateConfiguration() {
	t := suite.T()

	// First create a configuration
	suite.TestCreateConfiguration()

	// Updated configuration data
	updateData := []byte(`{
		"data": {
			"max_limit": 2000,
			"enabled": false
		}
	}`)

	// Create a request to update the configuration
	req := httptest.NewRequest("PUT", "/api/v1/configurations/payment-config", bytes.NewBuffer(updateData))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "payment-config", response["name"])
	assert.Equal(t, float64(2), response["version"])

	// Verify updated data
	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(2000), data["max_limit"])
	assert.Equal(t, false, data["enabled"])
}

// TestListConfigurationVersions tests the configuration versions listing endpoint
func (suite *ConfigurationAPITestSuite) TestListConfigurationVersions() {
	t := suite.T()

	// First create and update a configuration
	suite.TestUpdateConfiguration()

	// Create a request to list configuration versions
	req := httptest.NewRequest("GET", "/api/v1/configurations/payment-config/versions", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "payment-config", response["name"])

	// Verify versions
	versions, ok := response["versions"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(versions))

	// Check version 1
	version1 := versions[0].(map[string]interface{})
	assert.Equal(t, float64(1), version1["version"])

	// Check version 2
	version2 := versions[1].(map[string]interface{})
	assert.Equal(t, float64(2), version2["version"])
}

// TestGetConfigurationVersion tests the specific configuration version retrieval endpoint
func (suite *ConfigurationAPITestSuite) TestGetConfigurationVersion() {
	t := suite.T()

	// First register schema and create a configuration
	suite.TestRegisterSchema()

	// Create configuration data
	configData := []byte(`{
		"name": "payment-config",
		"data": {
			"max_limit": 1000,
			"enabled": true
		}
	}`)

	// Create a request to create a configuration
	createReq := httptest.NewRequest("POST", "/api/v1/configurations", bytes.NewBuffer(configData))
	createReq.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	createW := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(createW, createReq)

	// Assert the response
	assert.Equal(t, http.StatusCreated, createW.Code)

	// Create a request to get version 1 of the configuration
	getReq := httptest.NewRequest("GET", "/api/v1/configurations/payment-config/versions/1", nil)

	// Create a response recorder
	getW := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(getW, getReq)

	// Assert the response
	assert.Equal(t, http.StatusOK, getW.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(getW.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "payment-config", response["name"])
	assert.Equal(t, float64(1), response["version"])

	// Verify data is from version 1
	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1000), data["max_limit"])
	assert.Equal(t, true, data["enabled"])
}

// TestRollbackConfiguration tests the configuration rollback endpoint
func (suite *ConfigurationAPITestSuite) TestRollbackConfiguration() {
	t := suite.T()

	// First create and update a configuration
	suite.TestUpdateConfiguration()

	// Rollback data
	rollbackData := []byte(`{
		"target_version": 1
	}`)

	// Create a request to rollback the configuration
	req := httptest.NewRequest(http.MethodPost, "/api/v1/configurations/payment-config/rollback", bytes.NewBuffer(rollbackData))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "payment-config", response["name"])
	assert.Equal(t, float64(3), response["version"])
	// Note: is_rollback field is not present in the response

	// Verify data is rolled back to version 1
	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1000), data["max_limit"])
	assert.Equal(t, true, data["enabled"])
}

// TestHealthCheck tests the health check endpoint
func (suite *ConfigurationAPITestSuite) TestHealthCheck() {
	t := suite.T()

	// Create a request to the health check endpoint
	req := httptest.NewRequest("GET", "/health", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

// TestInvalidSchemaRegistration tests schema registration with invalid schema
func (suite *ConfigurationAPITestSuite) TestInvalidSchemaRegistration() {
	t := suite.T()

	// Invalid schema
	invalidSchema := []byte(`{
		"type": "invalid"
	}`)

	// Create a request to register the schema
	req := httptest.NewRequest(http.MethodPost, "/api/v1/schemas/invalid-schema", bytes.NewBuffer(invalidSchema))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestInvalidConfigurationCreation tests configuration creation with invalid data
func (suite *ConfigurationAPITestSuite) TestInvalidConfigurationCreation() {
	t := suite.T()

	// First register a schema
	suite.TestRegisterSchema()

	// Invalid configuration data (missing required field)
	invalidConfigData := []byte(`{
		"name": "payment-config",
		"data": {
			"max_limit": 1000
		}
	}`)

	// Create a request to create a configuration
	req := httptest.NewRequest(http.MethodPost, "/api/v1/configurations", bytes.NewBuffer(invalidConfigData))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestConfigurationNotFound tests retrieving a non-existent configuration
func (suite *ConfigurationAPITestSuite) TestConfigurationNotFound() {
	t := suite.T()

	// Create a request to get a non-existent configuration
	req := httptest.NewRequest("GET", "/api/v1/configurations/non-existent", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	suite.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestConfigurationAPITestSuite runs the test suite
func TestConfigurationAPITestSuite(t *testing.T) {
	suite.Run(t, new(ConfigurationAPITestSuite))
}
