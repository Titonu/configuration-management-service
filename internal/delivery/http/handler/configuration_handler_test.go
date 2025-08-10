package handler

import (
	"bytes"
	"encoding/json"
	"github.com/Titonu/configuration-management-service/internal/domain/entity"
	"github.com/Titonu/configuration-management-service/internal/domain/usecase"
	"github.com/Titonu/configuration-management-service/pkg/errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConfigurationService is a mock implementation of service.ConfigurationUsecase
type MockConfigurationService struct {
	mock.Mock
}

func (m *MockConfigurationService) CreateConfiguration(name string, data json.RawMessage) (*entity.Configuration, error) {
	args := m.Called(name, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Configuration), args.Error(1)
}

func (m *MockConfigurationService) UpdateConfiguration(name string, data json.RawMessage) (*entity.Configuration, error) {
	args := m.Called(name, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Configuration), args.Error(1)
}

func (m *MockConfigurationService) GetConfiguration(name string) (*entity.Configuration, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Configuration), args.Error(1)
}

func (m *MockConfigurationService) GetConfigurationVersion(name string, version int) (*entity.Configuration, error) {
	args := m.Called(name, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Configuration), args.Error(1)
}

func (m *MockConfigurationService) ListConfigurationVersions(name string) (*entity.VersionList, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.VersionList), args.Error(1)
}

func (m *MockConfigurationService) RollbackConfiguration(name string, targetVersion int) (*entity.Configuration, error) {
	args := m.Called(name, targetVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Configuration), args.Error(1)
}

func (m *MockConfigurationService) RegisterSchema(name string, schema json.RawMessage) error {
	args := m.Called(name, schema)
	return args.Error(0)
}

func (m *MockConfigurationService) GetSchema(name string) (json.RawMessage, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(json.RawMessage), args.Error(1)
}

func (m *MockConfigurationService) ValidateConfigurationData(configName string, data json.RawMessage) error {
	args := m.Called(configName, data)
	return args.Error(0)
}

func setupRouter(mockService usecase.ConfigurationUsecase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := NewConfigurationHandler(mockService)

	v1 := router.Group("/api/v1")
	{
		// Configuration endpoints
		v1.POST("/configurations", handler.CreateConfiguration)
		v1.PUT("/configurations/:name", handler.UpdateConfiguration)
		v1.GET("/configurations/:name", handler.GetConfiguration)
		v1.GET("/configurations/:name/versions", handler.ListConfigurationVersions)
		v1.GET("/configurations/:name/versions/:version", handler.GetConfigurationVersion)
		v1.POST("/configurations/:name/rollback", handler.RollbackConfiguration)

		// Schema endpoints
		v1.POST("/schemas/:name", handler.RegisterSchema)
		v1.GET("/schemas/:name", handler.GetSchema)
	}

	return router
}

func TestCreateConfiguration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		reqBody := map[string]interface{}{
			"name": "test-config",
			"data": map[string]interface{}{
				"key": "value",
			},
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Mock service response
		expectedConfig := &entity.Configuration{
			Name:    "test-config",
			Version: 1,
			Data:    json.RawMessage(`{"key":"value"}`),
		}

		mockService.On("CreateConfiguration", "test-config", mock.AnythingOfType("json.RawMessage")).Return(expectedConfig, nil)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configurations", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "test-config", response["name"])
		assert.Equal(t, float64(1), response["version"])

		mockService.AssertExpectations(t)
	})

	t.Run("BadRequest", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Invalid JSON
		reqJSON := []byte(`{"name": "test-config", "data": invalid}`)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configurations", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		reqBody := map[string]interface{}{
			"name": "test-config",
			"data": map[string]interface{}{
				"key": "value",
			},
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Mock service error
		mockService.On("CreateConfiguration", "test-config", mock.AnythingOfType("json.RawMessage")).
			Return(nil, errors.NewValidationFailedError("Invalid request", errors.NewValidationError("Request", "invalid request")))

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configurations", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestUpdateConfiguration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		reqBody := map[string]interface{}{
			"data": map[string]interface{}{
				"key": "updated",
			},
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Mock service response
		expectedConfig := &entity.Configuration{
			Name:    "test-config",
			Version: 2,
			Data:    json.RawMessage(`{"key":"updated"}`),
		}

		mockService.On("UpdateConfiguration", "test-config", mock.AnythingOfType("json.RawMessage")).Return(expectedConfig, nil)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/configurations/test-config", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "test-config", response["name"])
		assert.Equal(t, float64(2), response["version"])

		mockService.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		reqBody := map[string]interface{}{
			"data": map[string]interface{}{
				"key": "value",
			},
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Mock service error
		mockService.On("UpdateConfiguration", "non-existent", mock.AnythingOfType("json.RawMessage")).
			Return(nil, errors.NewNotFoundError("Configuration", "test-config"))

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/configurations/non-existent", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestGetConfiguration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Mock service response
		expectedConfig := &entity.Configuration{
			Name:    "test-config",
			Version: 1,
			Data:    json.RawMessage(`{"key":"value"}`),
		}

		mockService.On("GetConfiguration", "test-config").Return(expectedConfig, nil)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configurations/test-config", nil)

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "test-config", response["name"])
		assert.Equal(t, float64(1), response["version"])

		mockService.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Mock service error
		mockService.On("GetConfiguration", "non-existent").
			Return(nil, errors.NewNotFoundError("Configuration", "test-config"))

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configurations/non-existent", nil)

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestGetConfigurationVersion(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Mock service response
		expectedConfig := &entity.Configuration{
			Name:    "test-config",
			Version: 1,
			Data:    json.RawMessage(`{"key":"value"}`),
		}

		mockService.On("GetConfigurationVersion", "test-config", 1).Return(expectedConfig, nil)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configurations/test-config/versions/1", nil)

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "test-config", response["name"])
		assert.Equal(t, float64(1), response["version"])

		mockService.AssertExpectations(t)
	})

	t.Run("InvalidVersion", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Create request with invalid version
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configurations/test-config/versions/invalid", nil)

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Mock service error
		mockService.On("GetConfigurationVersion", "test-config", 999).
			Return(nil, errors.NewNotFoundError("Version", "1"))

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configurations/test-config/versions/999", nil)

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestListConfigurationVersions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Mock service response
		expectedVersions := &entity.VersionList{
			Name: "test-config",
			Versions: []entity.VersionInfo{
				{Version: 1, CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
				{Version: 2, CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
		}

		mockService.On("ListConfigurationVersions", "test-config").Return(expectedVersions, nil)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configurations/test-config/versions", nil)

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "test-config", response["name"])
		versions, ok := response["versions"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(versions))

		mockService.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Mock service error
		mockService.On("ListConfigurationVersions", "non-existent").
			Return(nil, errors.NewNotFoundError("Configuration", "test-config"))

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/configurations/non-existent/versions", nil)

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestRollbackConfiguration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		reqBody := map[string]interface{}{
			"target_version": 1,
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Mock service response
		expectedConfig := &entity.Configuration{
			Name:         "test-config",
			Version:      3,
			Data:         json.RawMessage(`{"key":"original"}`),
			RollbackFrom: 1,
		}

		mockService.On("RollbackConfiguration", "test-config", 1).Return(expectedConfig, nil)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configurations/test-config/rollback", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "test-config", response["name"])
		assert.Equal(t, float64(3), response["version"])
		assert.Equal(t, float64(1), response["rollback_from"])

		mockService.AssertExpectations(t)
	})

	t.Run("BadRequest", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Invalid JSON
		reqJSON := []byte(`{"target_version": "invalid"}`)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configurations/test-config/rollback", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		reqBody := map[string]interface{}{
			"target_version": 1,
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Mock service error
		mockService.On("RollbackConfiguration", "non-existent", 1).
			Return(nil, errors.NewNotFoundError("Configuration", "test-config"))

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configurations/non-existent/rollback", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestRegisterSchema(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		schema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"key": map[string]interface{}{
					"type": "string",
				},
			},
		}

		schemaJSON, _ := json.Marshal(schema)

		// Mock service response
		mockService.On("RegisterSchema", "test-config", mock.AnythingOfType("json.RawMessage")).Return(nil)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/schemas/test-config", bytes.NewBuffer(schemaJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("BadRequest", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Invalid JSON
		schemaJSON := []byte(`{"type": "object", "properties": invalid}`)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/schemas/test-config", bytes.NewBuffer(schemaJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidationError", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		schema := map[string]interface{}{
			"type": "invalid",
		}

		schemaJSON, _ := json.Marshal(schema)

		// Mock service error
		mockService.On("RegisterSchema", "test-config", mock.AnythingOfType("json.RawMessage")).
			Return(errors.NewInvalidRequestError("Invalid schema", errors.NewValidationError("schema", "invalid schema")))

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/schemas/test-config", bytes.NewBuffer(schemaJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestGetSchema(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Mock service response
		schema := json.RawMessage(`{"type":"object","properties":{"key":{"type":"string"}}}`)
		mockService.On("GetSchema", "test-config").Return(schema, nil)

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/schemas/test-config", nil)

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "object", response["type"])

		mockService.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockService := new(MockConfigurationService)
		router := setupRouter(mockService)

		// Mock service error
		mockService.On("GetSchema", "non-existent").
			Return(nil, errors.NewNotFoundError("Schema", "test-config"))

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/schemas/non-existent", nil)

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}
