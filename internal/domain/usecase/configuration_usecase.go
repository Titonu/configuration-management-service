package usecase

import (
	"encoding/json"
	"github.com/Titonu/configuration-management-service/internal/domain/entity"
)

// ConfigurationUsecase defines the interface for configuration business logic
type ConfigurationUsecase interface {
	// CreateConfiguration creates a new configuration
	CreateConfiguration(name string, data json.RawMessage) (*entity.Configuration, error)

	// UpdateConfiguration updates an existing configuration
	UpdateConfiguration(name string, data json.RawMessage) (*entity.Configuration, error)

	// GetConfiguration retrieves a configuration by name
	GetConfiguration(name string) (*entity.Configuration, error)

	// GetConfigurationVersion retrieves a specific version of a configuration
	GetConfigurationVersion(name string, version int) (*entity.Configuration, error)

	// ListConfigurationVersions lists all versions of a configuration
	ListConfigurationVersions(name string) (*entity.VersionList, error)

	// RollbackConfiguration rolls back a configuration to a previous version
	RollbackConfiguration(name string, targetVersion int) (*entity.Configuration, error)

	// RegisterSchema registers a JSON schema for a configuration
	RegisterSchema(configName string, schema json.RawMessage) error

	// GetSchema retrieves the JSON schema for a configuration
	GetSchema(configName string) (json.RawMessage, error)

	// ValidateConfigurationData validates configuration data against its schema
	ValidateConfigurationData(configName string, data json.RawMessage) error
}
