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
