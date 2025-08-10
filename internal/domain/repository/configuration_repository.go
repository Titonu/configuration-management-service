package repository

import (
	"encoding/json"
	"github.com/Titonu/configuration-management-service/internal/domain/entity"
)

// ConfigurationRepository defines the interface for configuration storage operations
type ConfigurationRepository interface {
	// CreateConfiguration creates a new configuration
	CreateConfiguration(config *entity.Configuration) error

	// UpdateConfiguration updates an existing configuration
	UpdateConfiguration(config *entity.Configuration) error

	// GetConfiguration retrieves a configuration by name
	GetConfiguration(name string) (*entity.Configuration, error)

	// GetConfigurationVersion retrieves a specific version of a configuration
	GetConfigurationVersion(name string, version int) (*entity.Configuration, error)

	// ListConfigurationVersions lists all versions of a configuration
	ListConfigurationVersions(name string) (*entity.VersionList, error)

	// RegisterSchema registers a JSON schema for a configuration
	RegisterSchema(configName string, schema json.RawMessage) error

	// GetSchema retrieves the JSON schema for a configuration
	GetSchema(configName string) (json.RawMessage, error)

	// StoreVersionData stores the raw data for a specific version
	StoreVersionData(configName string, version int, data json.RawMessage) error

	// GetVersionData retrieves the raw data for a specific version
	GetVersionData(configName string, version int) (json.RawMessage, error)
}
