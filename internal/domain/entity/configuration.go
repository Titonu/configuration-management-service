package entity

import (
	"encoding/json"
	"time"
)

// Configuration represents a configuration entity with its metadata and data
type Configuration struct {
	Name      string          `json:"name"`
	Version   int             `json:"version"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at,omitempty"`

	// Fields for rollback operations
	RollbackFrom int `json:"rollback_from,omitempty"`
	RollbackTo   int `json:"rollback_to,omitempty"`
}

// VersionInfo represents version metadata for listing versions
type VersionInfo struct {
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	IsRollback bool      `json:"is_rollback,omitempty"`
}

// VersionList represents the response for listing versions
type VersionList struct {
	Name     string        `json:"name"`
	Versions []VersionInfo `json:"versions"`
}

// NewConfiguration creates a new Configuration with default values
func NewConfiguration(name string, data json.RawMessage) *Configuration {
	now := time.Now().UTC()
	return &Configuration{
		Name:      name,
		Version:   1,
		Data:      data,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewVersionFromRollback creates a new Configuration version from a rollback operation
func NewVersionFromRollback(config *Configuration, targetVersion int, targetData json.RawMessage) *Configuration {
	now := time.Now().UTC()
	return &Configuration{
		Name:         config.Name,
		Version:      config.Version + 1,
		Data:         targetData,
		CreatedAt:    config.CreatedAt, // Keep original creation time
		UpdatedAt:    now,
		RollbackFrom: config.Version,
		RollbackTo:   targetVersion,
	}
}

// UpdateVersion creates a new version of the configuration with updated data
func (c *Configuration) UpdateVersion(data json.RawMessage) *Configuration {
	now := time.Now().UTC()
	return &Configuration{
		Name:      c.Name,
		Version:   c.Version + 1,
		Data:      data,
		CreatedAt: c.CreatedAt, // Keep original creation time
		UpdatedAt: now,
	}
}
