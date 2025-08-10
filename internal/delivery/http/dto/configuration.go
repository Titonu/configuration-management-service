package dto

import (
	"encoding/json"
	"time"
)

// ConfigurationCreateRequest represents the request body for creating a new configuration
type ConfigurationCreateRequest struct {
	Name string          `json:"name" binding:"required"`
	Data json.RawMessage `json:"data" binding:"required"`
}

// ConfigurationUpdateRequest represents the request body for updating a configuration
type ConfigurationUpdateRequest struct {
	Data json.RawMessage `json:"data" binding:"required"`
}

// RollbackRequest represents the request body for rolling back a configuration
type RollbackRequest struct {
	TargetVersion int `json:"target_version" binding:"required"`
}

// VersionInfo represents version metadata for listing versions
type VersionInfo struct {
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	IsRollback bool      `json:"is_rollback,omitempty"`
}

// VersionListResponse represents the response for listing versions
type VersionListResponse struct {
	Name     string        `json:"name"`
	Versions []VersionInfo `json:"versions"`
}
