package usecase

import (
	"encoding/json"
	"github.com/Titonu/configuration-management-service/internal/domain/entity"
	"github.com/Titonu/configuration-management-service/internal/domain/repository"
	"github.com/Titonu/configuration-management-service/pkg/errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// NewTestConfigurationUseCase creates a new configuration use case for testing
// Returns the concrete type directly instead of the interface
func NewTestConfigurationUseCase(repo repository.ConfigurationRepository) *ConfigurationUseCase {
	return &ConfigurationUseCase{
		repo:      repo,
		validator: nil, // Will be set by test
	}
}

// MockConfigurationRepository is a mock implementation of repository.ConfigurationRepository
type MockConfigurationRepository struct {
	mock.Mock
}

func (m *MockConfigurationRepository) CreateConfiguration(config *entity.Configuration) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockConfigurationRepository) UpdateConfiguration(config *entity.Configuration) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockConfigurationRepository) GetConfiguration(name string) (*entity.Configuration, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Configuration), args.Error(1)
}

func (m *MockConfigurationRepository) GetConfigurationVersion(name string, version int) (*entity.Configuration, error) {
	args := m.Called(name, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Configuration), args.Error(1)
}

func (m *MockConfigurationRepository) ListConfigurationVersions(name string) (*entity.VersionList, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.VersionList), args.Error(1)
}

func (m *MockConfigurationRepository) RegisterSchema(name string, schema json.RawMessage) error {
	args := m.Called(name, schema)
	return args.Error(0)
}

func (m *MockConfigurationRepository) GetSchema(name string) (json.RawMessage, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(json.RawMessage), args.Error(1)
}

func (m *MockConfigurationRepository) StoreVersionData(configName string, version int, data json.RawMessage) error {
	args := m.Called(configName, version, data)
	return args.Error(0)
}

func (m *MockConfigurationRepository) GetVersionData(configName string, version int) (json.RawMessage, error) {
	args := m.Called(configName, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(json.RawMessage), args.Error(1)
}

// MockJSONSchemaValidator is a mock implementation of validator.JSONSchemaValidator
type MockJSONSchemaValidator struct {
	mock.Mock
}

func (m *MockJSONSchemaValidator) ValidateJSON(schema json.RawMessage, data json.RawMessage) error {
	args := m.Called(schema, data)
	return args.Error(0)
}

func (m *MockJSONSchemaValidator) ValidateSchemaDefinition(schema json.RawMessage) error {
	args := m.Called(schema)
	return args.Error(0)
}

func TestConfigurationUseCase_CreateConfiguration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		data := json.RawMessage(`{"key":"value"}`)

		// Check if configuration already exists
		mockRepo.On("GetConfiguration", name).Return(nil, errors.NewNotFoundError("Configuration", name))

		// No schema exists yet
		mockRepo.On("GetSchema", name).Return(nil, errors.NewNotFoundError("Schema", name))

		// Configuration creation should succeed
		mockRepo.On("CreateConfiguration", mock.AnythingOfType("*entity.Configuration")).Return(nil)
		mockRepo.On("StoreVersionData", name, 1, data).Return(nil)

		// Expected result
		expectedConfig := &entity.Configuration{
			Name:    name,
			Version: 1,
			Data:    data,
		}

		// Call the method
		result, err := useCase.CreateConfiguration(name, data)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, expectedConfig.Name, result.Name)
		assert.Equal(t, expectedConfig.Version, result.Version)
		assert.JSONEq(t, string(expectedConfig.Data), string(result.Data))
		mockRepo.AssertExpectations(t)
	})

	t.Run("WithSchemaValidation", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		mockValidator := new(MockJSONSchemaValidator)
		// Use concrete type directly
		useCase := NewTestConfigurationUseCase(mockRepo)
		useCase.SetValidator(mockValidator)

		// Test data
		name := "test-config"
		data := json.RawMessage(`{"key":"value"}`)
		schema := json.RawMessage(`{"type":"object","properties":{"key":{"type":"string"}}}`)

		// Check if configuration already exists
		mockRepo.On("GetConfiguration", name).Return(nil, errors.NewNotFoundError("Configuration", name))

		// Schema exists and is valid
		mockRepo.On("GetSchema", name).Return(schema, nil)
		mockValidator.On("ValidateJSON", schema, data).Return(nil)

		// Configuration creation should succeed
		mockRepo.On("CreateConfiguration", mock.AnythingOfType("*entity.Configuration")).Return(nil)
		mockRepo.On("StoreVersionData", name, 1, data).Return(nil)

		// Call the method
		result, err := useCase.CreateConfiguration(name, data)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, name, result.Name)
		assert.Equal(t, 1, result.Version)
		assert.JSONEq(t, string(data), string(result.Data))
		mockRepo.AssertExpectations(t)
		mockValidator.AssertExpectations(t)
	})

	t.Run("ValidationFailed", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		mockValidator := new(MockJSONSchemaValidator)
		// Use concrete type directly
		useCase := NewTestConfigurationUseCase(mockRepo)
		useCase.SetValidator(mockValidator)

		// Test data
		name := "test-config"
		data := json.RawMessage(`{"key":123}`) // Invalid: key should be string
		schema := json.RawMessage(`{"type":"object","properties":{"key":{"type":"string"}}}`)
		validationErr := errors.NewValidationFailedError("Invalid data", "key must be a string")

		// Check if configuration already exists
		mockRepo.On("GetConfiguration", name).Return(nil, errors.NewNotFoundError("Configuration", name))

		// Schema exists but validation fails
		mockRepo.On("GetSchema", name).Return(schema, nil)
		mockValidator.On("ValidateJSON", schema, data).Return(validationErr)

		// Call the method
		result, err := useCase.CreateConfiguration(name, data)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, validationErr, err)
		mockRepo.AssertExpectations(t)
		mockValidator.AssertExpectations(t)
	})
}

func TestConfigurationUseCase_UpdateConfiguration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		data := json.RawMessage(`{"key":"updated"}`)
		existingConfig := &entity.Configuration{
			Name:    name,
			Version: 1,
			Data:    json.RawMessage(`{"key":"value"}`),
		}

		// Configuration exists
		mockRepo.On("GetConfiguration", name).Return(existingConfig, nil)

		// No schema exists
		mockRepo.On("GetSchema", name).Return(nil, errors.NewNotFoundError("Schema", name))

		// Update should succeed
		mockRepo.On("UpdateConfiguration", mock.AnythingOfType("*entity.Configuration")).Return(nil)
		mockRepo.On("StoreVersionData", name, 2, data).Return(nil)

		// Call the method
		result, err := useCase.UpdateConfiguration(name, data)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, name, result.Name)
		assert.Equal(t, 2, result.Version) // Version incremented
		assert.JSONEq(t, string(data), string(result.Data))
		mockRepo.AssertExpectations(t)
	})

	t.Run("ConfigurationNotFound", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		data := json.RawMessage(`{"key":"updated"}`)
		notFoundErr := errors.NewNotFoundError("Configuration", name)

		// Configuration doesn't exist
		mockRepo.On("GetConfiguration", name).Return(nil, notFoundErr)

		// Call the method
		result, err := useCase.UpdateConfiguration(name, data)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertNotCalled(t, "UpdateConfiguration")
		mockRepo.AssertNotCalled(t, "StoreVersionData")
		mockRepo.AssertExpectations(t)
	})

	t.Run("WithSchemaValidationSuccess", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		mockValidator := new(MockJSONSchemaValidator)
		uc := NewTestConfigurationUseCase(mockRepo)
		uc.SetValidator(mockValidator)

		// Test data
		name := "test-config"
		data := json.RawMessage(`{"key":"updated"}`)
		schema := json.RawMessage(`{"type":"object"}`)
		existingConfig := &entity.Configuration{
			Name:    name,
			Version: 1,
			Data:    json.RawMessage(`{"key":"original"}`),
		}

		// Configuration exists
		mockRepo.On("GetConfiguration", name).Return(existingConfig, nil)

		// Schema exists and validation passes
		mockRepo.On("GetSchema", name).Return(schema, nil)
		mockValidator.On("ValidateJSON", schema, data).Return(nil)

		// Update should succeed
		mockRepo.On("UpdateConfiguration", mock.AnythingOfType("*entity.Configuration")).Return(nil)
		mockRepo.On("StoreVersionData", name, 2, data).Return(nil)

		// Call the method
		result, err := uc.UpdateConfiguration(name, data)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, name, result.Name)
		assert.Equal(t, 2, result.Version) // Version incremented
		assert.Equal(t, data, result.Data)
		mockRepo.AssertExpectations(t)
		mockValidator.AssertExpectations(t)
	})

	t.Run("WithSchemaValidationFailure", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		mockValidator := new(MockJSONSchemaValidator)
		uc := NewTestConfigurationUseCase(mockRepo)
		uc.SetValidator(mockValidator)

		// Test data
		name := "test-config"
		data := json.RawMessage(`{"key":"updated"}`)
		schema := json.RawMessage(`{"type":"object","required":["required_field"]}`)
		existingConfig := &entity.Configuration{
			Name:    name,
			Version: 1,
			Data:    json.RawMessage(`{"key":"original"}`),
		}
		validationErr := errors.NewValidationFailedError("Validation failed", "required_field is required")

		// Configuration exists
		mockRepo.On("GetConfiguration", name).Return(existingConfig, nil)

		// Schema exists but validation fails
		mockRepo.On("GetSchema", name).Return(schema, nil)
		mockValidator.On("ValidateJSON", schema, data).Return(validationErr)

		// Call the method
		result, err := uc.UpdateConfiguration(name, data)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, validationErr, err)
		mockRepo.AssertNotCalled(t, "UpdateConfiguration")
		mockRepo.AssertNotCalled(t, "StoreVersionData")
		mockRepo.AssertExpectations(t)
		mockValidator.AssertExpectations(t)
	})
}

func TestConfigurationUseCase_GetConfiguration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		config := &entity.Configuration{
			Name:    name,
			Version: 1,
			Data:    json.RawMessage(`{"key":"value"}`),
		}

		// Configuration exists
		mockRepo.On("GetConfiguration", name).Return(config, nil)

		// Call the method
		result, err := useCase.GetConfiguration(name)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, config, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "non-existent"
		notFoundErr := errors.NewNotFoundError("Configuration", name)

		// Configuration doesn't exist
		mockRepo.On("GetConfiguration", name).Return(nil, notFoundErr)

		// Call the method
		result, err := useCase.GetConfiguration(name)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notFoundErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestConfigurationUseCase_GetConfigurationVersion(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		version := 1
		config := &entity.Configuration{
			Name:    name,
			Version: version,
			Data:    json.RawMessage(`{"key":"value"}`),
		}

		// Version exists
		mockRepo.On("GetConfigurationVersion", name, version).Return(config, nil)

		// Call the method
		result, err := useCase.GetConfigurationVersion(name, version)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, config, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestConfigurationUseCase_ListConfigurationVersions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		// Parse time strings to time.Time
		time1, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		time2, _ := time.Parse(time.RFC3339, "2023-01-02T00:00:00Z")

		versions := &entity.VersionList{
			Name: name,
			Versions: []entity.VersionInfo{
				{Version: 1, CreatedAt: time1},
				{Version: 2, CreatedAt: time2},
			},
		}

		// Check if configuration exists
		mockRepo.On("GetConfiguration", name).Return(&entity.Configuration{Name: name}, nil)

		// Versions exist
		mockRepo.On("ListConfigurationVersions", name).Return(versions, nil)

		// Call the method
		result, err := useCase.ListConfigurationVersions(name)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, versions, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestConfigurationUseCase_RollbackConfiguration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		targetVersion := 1
		currentVersion := 2
		currentConfig := &entity.Configuration{
			Name:    name,
			Version: currentVersion,
			Data:    json.RawMessage(`{"key":"updated"}`),
		}
		targetData := json.RawMessage(`{"key":"original"}`)

		// Current configuration exists
		mockRepo.On("GetConfiguration", name).Return(currentConfig, nil)

		// Target version exists
		mockRepo.On("GetVersionData", name, targetVersion).Return(targetData, nil)

		// Rollback should succeed
		mockRepo.On("UpdateConfiguration", mock.AnythingOfType("*entity.Configuration")).Return(nil)
		mockRepo.On("StoreVersionData", name, currentVersion+1, targetData).Return(nil)

		// Call the method
		result, err := useCase.RollbackConfiguration(name, targetVersion)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, name, result.Name)
		assert.Equal(t, currentVersion+1, result.Version) // Version incremented
		assert.Equal(t, currentVersion, result.RollbackFrom)
		assert.Equal(t, targetVersion, result.RollbackTo)
		assert.JSONEq(t, string(targetData), string(result.Data))
		mockRepo.AssertExpectations(t)
	})

	t.Run("ConfigurationNotFound", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		targetVersion := 1
		notFoundErr := errors.NewNotFoundError("Configuration", name)

		// Configuration doesn't exist
		mockRepo.On("GetConfiguration", name).Return(nil, notFoundErr)

		// Call the method
		result, err := useCase.RollbackConfiguration(name, targetVersion)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertNotCalled(t, "GetVersionData")
		mockRepo.AssertNotCalled(t, "UpdateConfiguration")
		mockRepo.AssertNotCalled(t, "StoreVersionData")
		mockRepo.AssertExpectations(t)
	})

	t.Run("TargetVersionNotFound", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		targetVersion := 1
		currentVersion := 2
		currentConfig := &entity.Configuration{
			Name:    name,
			Version: currentVersion,
			Data:    json.RawMessage(`{"key":"updated"}`),
		}
		notFoundErr := errors.NewNotFoundError("Version", "1")

		// Current configuration exists
		mockRepo.On("GetConfiguration", name).Return(currentConfig, nil)

		// Target version doesn't exist
		mockRepo.On("GetVersionData", name, targetVersion).Return(nil, notFoundErr)

		// Call the method
		result, err := useCase.RollbackConfiguration(name, targetVersion)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertNotCalled(t, "UpdateConfiguration")
		mockRepo.AssertNotCalled(t, "StoreVersionData")
		mockRepo.AssertExpectations(t)
	})

	t.Run("RollbackToSameVersion", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		currentVersion := 2
		currentConfig := &entity.Configuration{
			Name:    name,
			Version: currentVersion,
			Data:    json.RawMessage(`{"key":"updated"}`),
		}

		// Current configuration exists
		mockRepo.On("GetConfiguration", name).Return(currentConfig, nil)

		// Mock GetVersionData since the implementation calls it regardless of version check
		mockRepo.On("GetVersionData", name, currentVersion).Return(currentConfig.Data, nil)

		// Mock UpdateConfiguration since the implementation calls it
		mockRepo.On("UpdateConfiguration", mock.AnythingOfType("*entity.Configuration")).Return(nil)

		// Mock StoreVersionData since the implementation calls it
		mockRepo.On("StoreVersionData", name, currentVersion+1, currentConfig.Data).Return(nil)

		// Call the method with same version
		result, err := useCase.RollbackConfiguration(name, currentVersion)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.Name)
		assert.Equal(t, currentVersion+1, result.Version) // Version incremented
		assert.Equal(t, currentVersion, result.RollbackFrom)
		assert.Equal(t, currentVersion, result.RollbackTo)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RollbackToFutureVersion", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		currentVersion := 2
		futureVersion := 3
		currentConfig := &entity.Configuration{
			Name:    name,
			Version: currentVersion,
			Data:    json.RawMessage(`{"key":"updated"}`),
		}

		// Current configuration exists
		mockRepo.On("GetConfiguration", name).Return(currentConfig, nil)

		// Mock GetVersionData since the implementation calls it regardless of version check
		// Return a not found error for future version
		notFoundErr := errors.NewNotFoundError("Configuration version", name)
		mockRepo.On("GetVersionData", name, futureVersion).Return(nil, notFoundErr)

		// Call the method with future version
		result, err := useCase.RollbackConfiguration(name, futureVersion)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertNotCalled(t, "UpdateConfiguration")
		mockRepo.AssertNotCalled(t, "StoreVersionData")
		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdateFailed", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		useCase := NewConfigurationUseCase(mockRepo)

		// Test data
		name := "test-config"
		targetVersion := 1
		currentVersion := 2
		currentConfig := &entity.Configuration{
			Name:    name,
			Version: currentVersion,
			Data:    json.RawMessage(`{"key":"updated"}`),
		}
		targetData := json.RawMessage(`{"key":"original"}`)
		updateErr := errors.NewInternalError("Database error", nil)

		// Current configuration exists
		mockRepo.On("GetConfiguration", name).Return(currentConfig, nil)

		// Target version exists
		mockRepo.On("GetVersionData", name, targetVersion).Return(targetData, nil)

		// Update fails
		mockRepo.On("UpdateConfiguration", mock.AnythingOfType("*entity.Configuration")).Return(updateErr)

		// Call the method
		result, err := useCase.RollbackConfiguration(name, targetVersion)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		// Check that it's an internal error with the expected message
		assert.Contains(t, err.Error(), "Failed to rollback configuration")
		mockRepo.AssertNotCalled(t, "StoreVersionData")
		mockRepo.AssertExpectations(t)
	})
}

func TestConfigurationUseCase_RegisterSchema(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		mockValidator := new(MockJSONSchemaValidator)
		// Use concrete type directly
		useCase := NewTestConfigurationUseCase(mockRepo)
		useCase.SetValidator(mockValidator)

		// Test data
		name := "test-config"
		schema := json.RawMessage(`{"type":"object","properties":{"key":{"type":"string"}}}`)

		// Validate schema
		mockValidator.On("ValidateSchemaDefinition", schema).Return(nil)

		// Register schema
		mockRepo.On("RegisterSchema", name, schema).Return(nil)

		// Call the method
		err := useCase.RegisterSchema(name, schema)

		// Assertions
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockValidator.AssertExpectations(t)
	})

	t.Run("InvalidSchema", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		mockValidator := new(MockJSONSchemaValidator)
		// Use concrete type directly
		useCase := NewTestConfigurationUseCase(mockRepo)
		useCase.SetValidator(mockValidator)

		// Test data
		name := "test-config"
		invalidSchema := json.RawMessage(`{"type":"invalid"}`)
		validationErr := errors.NewValidationFailedError("Invalid schema", "unknown type: invalid")

		// Validate schema fails
		mockValidator.On("ValidateSchemaDefinition", invalidSchema).Return(validationErr)

		// Call the method
		err := useCase.RegisterSchema(name, invalidSchema)

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, validationErr, err)
		mockRepo.AssertNotCalled(t, "RegisterSchema")
		mockValidator.AssertExpectations(t)
	})
}

func TestConfigurationUseCase_ValidateConfigurationData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		mockValidator := new(MockJSONSchemaValidator)
		uc := NewTestConfigurationUseCase(mockRepo)
		uc.SetValidator(mockValidator)

		name := "test-config"
		schema := json.RawMessage(`{"type":"object"}`)
		data := json.RawMessage(`{"key":"value"}`)

		// Schema exists
		mockRepo.On("GetSchema", name).Return(schema, nil)

		// Validation succeeds
		mockValidator.On("ValidateJSON", schema, data).Return(nil)

		// Call the method
		err := uc.ValidateConfigurationData(name, data)

		// Assertions
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockValidator.AssertExpectations(t)
	})

	t.Run("SchemaNotFound", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		mockValidator := new(MockJSONSchemaValidator)
		uc := NewTestConfigurationUseCase(mockRepo)
		uc.SetValidator(mockValidator)

		name := "test-config"
		data := json.RawMessage(`{"key":"value"}`)
		notFoundErr := errors.NewNotFoundError("Schema", name)

		// Schema doesn't exist
		mockRepo.On("GetSchema", name).Return(nil, notFoundErr)

		// Call the method
		err := uc.ValidateConfigurationData(name, data)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("ValidationFailed", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		mockValidator := new(MockJSONSchemaValidator)
		uc := NewTestConfigurationUseCase(mockRepo)
		uc.SetValidator(mockValidator)

		name := "test-config"
		schema := json.RawMessage(`{"type":"object","required":["required_field"]}`)
		data := json.RawMessage(`{"key":"value"}`)
		validationErr := errors.NewValidationFailedError("Validation failed", "required_field is required")

		// Schema exists
		mockRepo.On("GetSchema", name).Return(schema, nil)

		// Validation fails
		mockValidator.On("ValidateJSON", schema, data).Return(validationErr)

		// Call the method
		err := uc.ValidateConfigurationData(name, data)

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, validationErr, err)
		mockRepo.AssertExpectations(t)
		mockValidator.AssertExpectations(t)
	})
}

func TestConfigurationUseCase_GetSchema(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		mockValidator := new(MockJSONSchemaValidator)
		uc := NewTestConfigurationUseCase(mockRepo)
		uc.SetValidator(mockValidator)

		name := "test-config"
		schema := json.RawMessage(`{"type":"object"}`)

		mockRepo.On("GetSchema", name).Return(schema, nil)

		result, err := uc.GetSchema(name)

		assert.NoError(t, err)
		assert.Equal(t, schema, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo := new(MockConfigurationRepository)
		mockValidator := new(MockJSONSchemaValidator)
		uc := NewTestConfigurationUseCase(mockRepo)
		uc.SetValidator(mockValidator)

		name := "test-config"
		notFoundErr := errors.NewNotFoundError("Schema", name)

		mockRepo.On("GetSchema", name).Return(nil, notFoundErr)

		result, err := uc.GetSchema(name)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Error(t, err)
		// Check if it's a NotFoundError without using type assertion
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}
