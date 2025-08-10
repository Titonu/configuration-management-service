package usecase

import (
	"encoding/json"
	"github.com/Titonu/configuration-management-service/internal/domain/entity"
	"github.com/Titonu/configuration-management-service/internal/domain/repository"
	"github.com/Titonu/configuration-management-service/internal/domain/usecase"
	"github.com/Titonu/configuration-management-service/pkg/errors"
	"github.com/Titonu/configuration-management-service/pkg/validator"
)

// ConfigurationUseCase implements the configuration service interface
type ConfigurationUseCase struct {
	repo      repository.ConfigurationRepository
	validator validator.Validator
}

// SetValidator sets the validator for testing purposes
func (uc *ConfigurationUseCase) SetValidator(v validator.Validator) {
	uc.validator = v
}

// NewConfigurationUseCase creates a new configuration use case
func NewConfigurationUseCase(repo repository.ConfigurationRepository) usecase.ConfigurationUsecase {
	return &ConfigurationUseCase{
		repo:      repo,
		validator: validator.NewJSONSchemaValidator(),
	}
}

// CreateConfiguration creates a new configuration
func (uc *ConfigurationUseCase) CreateConfiguration(name string, data json.RawMessage) (*entity.Configuration, error) {
	// Check if configuration already exists
	existingConfig, err := uc.repo.GetConfiguration(name)
	if err == nil && existingConfig != nil {
		return nil, errors.NewAlreadyExistsError("Configuration", name)
	}

	// Check if schema exists and validate against it
	schema, err := uc.repo.GetSchema(name)
	if err == nil && schema != nil {
		if err := uc.validator.ValidateJSON(schema, data); err != nil {
			return nil, err
		}
	}

	// Create new configuration
	config := entity.NewConfiguration(name, data)

	// Store in repository
	if err := uc.repo.CreateConfiguration(config); err != nil {
		return nil, errors.NewInternalError("Failed to create configuration", err.Error())
	}

	// Store version data
	if err := uc.repo.StoreVersionData(name, config.Version, data); err != nil {
		return nil, errors.NewInternalError("Failed to store version data", err.Error())
	}

	return config, nil
}

// UpdateConfiguration updates an existing configuration
func (uc *ConfigurationUseCase) UpdateConfiguration(name string, data json.RawMessage) (*entity.Configuration, error) {
	// Check if configuration exists
	existingConfig, err := uc.repo.GetConfiguration(name)
	if err != nil || existingConfig == nil {
		return nil, errors.NewNotFoundError("Configuration", name)
	}

	// Check if schema exists and validate against it
	schema, err := uc.repo.GetSchema(name)
	if err == nil && schema != nil {
		if err := uc.validator.ValidateJSON(schema, data); err != nil {
			return nil, err
		}
	}

	// Create new version
	newConfig := existingConfig.UpdateVersion(data)

	// Store in repository
	if err := uc.repo.UpdateConfiguration(newConfig); err != nil {
		return nil, errors.NewInternalError("Failed to update configuration", err.Error())
	}

	// Store version data
	if err := uc.repo.StoreVersionData(name, newConfig.Version, data); err != nil {
		return nil, errors.NewInternalError("Failed to store version data", err.Error())
	}

	return newConfig, nil
}

// GetConfiguration retrieves a configuration by name
func (uc *ConfigurationUseCase) GetConfiguration(name string) (*entity.Configuration, error) {
	config, err := uc.repo.GetConfiguration(name)
	if err != nil {
		return nil, errors.NewNotFoundError("Configuration", name)
	}

	return config, nil
}

// GetConfigurationVersion retrieves a specific version of a configuration
func (uc *ConfigurationUseCase) GetConfigurationVersion(name string, version int) (*entity.Configuration, error) {
	config, err := uc.repo.GetConfigurationVersion(name, version)
	if err != nil {
		return nil, errors.NewNotFoundError("Configuration version", name)
	}

	return config, nil
}

// ListConfigurationVersions lists all versions of a configuration
func (uc *ConfigurationUseCase) ListConfigurationVersions(name string) (*entity.VersionList, error) {
	// Check if configuration exists
	_, err := uc.repo.GetConfiguration(name)
	if err != nil {
		return nil, errors.NewNotFoundError("Configuration", name)
	}

	versions, err := uc.repo.ListConfigurationVersions(name)
	if err != nil {
		return nil, errors.NewInternalError("Failed to list configuration versions", err.Error())
	}

	return versions, nil
}

// RollbackConfiguration rolls back a configuration to a previous version
func (uc *ConfigurationUseCase) RollbackConfiguration(name string, targetVersion int) (*entity.Configuration, error) {
	// Check if configuration exists
	currentConfig, err := uc.repo.GetConfiguration(name)
	if err != nil || currentConfig == nil {
		return nil, errors.NewNotFoundError("Configuration", name)
	}

	// Check if target version exists
	targetData, err := uc.repo.GetVersionData(name, targetVersion)
	if err != nil || targetData == nil {
		return nil, errors.NewNotFoundError("Configuration version", name)
	}

	// Create new version from rollback
	newConfig := entity.NewVersionFromRollback(currentConfig, targetVersion, targetData)

	// Store in repository
	if err := uc.repo.UpdateConfiguration(newConfig); err != nil {
		return nil, errors.NewInternalError("Failed to rollback configuration", err.Error())
	}

	// Store version data
	if err := uc.repo.StoreVersionData(name, newConfig.Version, targetData); err != nil {
		return nil, errors.NewInternalError("Failed to store version data", err.Error())
	}

	return newConfig, nil
}

// RegisterSchema registers a JSON schema for a configuration
func (uc *ConfigurationUseCase) RegisterSchema(configName string, schema json.RawMessage) error {
	// Validate schema definition
	if err := uc.validator.ValidateSchemaDefinition(schema); err != nil {
		return err
	}

	// Store schema
	if err := uc.repo.RegisterSchema(configName, schema); err != nil {
		return errors.NewInternalError("Failed to register schema", err.Error())
	}

	return nil
}

// GetSchema retrieves the JSON schema for a configuration
func (uc *ConfigurationUseCase) GetSchema(configName string) (json.RawMessage, error) {
	schema, err := uc.repo.GetSchema(configName)
	if err != nil {
		return nil, errors.NewNotFoundError("Schema", configName)
	}

	return schema, nil
}

// ValidateConfigurationData validates configuration data against its schema
func (uc *ConfigurationUseCase) ValidateConfigurationData(configName string, data json.RawMessage) error {
	// Get schema
	schema, err := uc.repo.GetSchema(configName)
	if err != nil {
		return errors.NewNotFoundError("Schema", configName)
	}

	// Validate data against schema
	if err := uc.validator.ValidateJSON(schema, data); err != nil {
		return err
	}

	return nil
}
