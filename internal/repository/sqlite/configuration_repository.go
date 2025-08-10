package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Titonu/configuration-management-service/internal/domain/entity"
	"github.com/Titonu/configuration-management-service/internal/domain/repository"
	"github.com/Titonu/configuration-management-service/pkg/errors"
	"time"

	// Import sqlite3 driver for database/sql
	_ "github.com/mattn/go-sqlite3"
)

// ConfigurationRepository implements the repository interface using SQLite
type ConfigurationRepository struct {
	db *sql.DB
}

// NewConfigurationRepository creates a new SQLite repository
func NewConfigurationRepository(dbPath string) (repository.ConfigurationRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Initialize database schema
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return &ConfigurationRepository{
		db: db,
	}, nil
}

// Initialize database schema
func initSchema(db *sql.DB) error {
	// Create configurations table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS configurations (
			name TEXT PRIMARY KEY,
			version INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			rollback_from INTEGER,
			rollback_to INTEGER
		)
	`)
	if err != nil {
		return err
	}

	// Create versions table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS versions (
			name TEXT NOT NULL,
			version INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL,
			is_rollback BOOLEAN NOT NULL DEFAULT 0,
			PRIMARY KEY (name, version)
		)
	`)
	if err != nil {
		return err
	}

	// Create version_data table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS version_data (
			name TEXT NOT NULL,
			version INTEGER NOT NULL,
			data TEXT NOT NULL,
			PRIMARY KEY (name, version)
		)
	`)
	if err != nil {
		return err
	}

	// Create schemas table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schemas (
			name TEXT PRIMARY KEY,
			schema TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// CreateConfiguration creates a new configuration
func (r *ConfigurationRepository) CreateConfiguration(config *entity.Configuration) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert into configurations table
	_, err = tx.Exec(
		"INSERT INTO configurations (name, version, created_at, updated_at) VALUES (?, ?, ?, ?)",
		config.Name, config.Version, config.CreatedAt, config.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Insert into versions table
	_, err = tx.Exec(
		"INSERT INTO versions (name, version, created_at, is_rollback) VALUES (?, ?, ?, ?)",
		config.Name, config.Version, config.CreatedAt, false,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateConfiguration updates an existing configuration
func (r *ConfigurationRepository) UpdateConfiguration(config *entity.Configuration) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update configurations table
	_, err = tx.Exec(
		"UPDATE configurations SET version = ?, updated_at = ?, rollback_from = ?, rollback_to = ? WHERE name = ?",
		config.Version, config.UpdatedAt, config.RollbackFrom, config.RollbackTo, config.Name,
	)
	if err != nil {
		return err
	}

	// Insert into versions table
	_, err = tx.Exec(
		"INSERT INTO versions (name, version, created_at, is_rollback) VALUES (?, ?, ?, ?)",
		config.Name, config.Version, config.UpdatedAt, config.RollbackFrom > 0,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetConfiguration retrieves a configuration by name
func (r *ConfigurationRepository) GetConfiguration(name string) (*entity.Configuration, error) {
	var config entity.Configuration
	var rollbackFrom, rollbackTo sql.NullInt64

	// Query configurations table
	err := r.db.QueryRow(
		"SELECT name, version, created_at, updated_at, rollback_from, rollback_to FROM configurations WHERE name = ?",
		name,
	).Scan(
		&config.Name,
		&config.Version,
		&config.CreatedAt,
		&config.UpdatedAt,
		&rollbackFrom,
		&rollbackTo,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Configuration", name)
		}
		return nil, err
	}

	// Handle NULL values for rollback fields
	if rollbackFrom.Valid {
		config.RollbackFrom = int(rollbackFrom.Int64)
	}
	if rollbackTo.Valid {
		config.RollbackTo = int(rollbackTo.Int64)
	}

	// Get data from version_data table
	var dataStr string
	err = r.db.QueryRow(
		"SELECT data FROM version_data WHERE name = ? AND version = ?",
		name, config.Version,
	).Scan(&dataStr)
	if err != nil {
		return nil, err
	}

	config.Data = json.RawMessage(dataStr)
	return &config, nil
}

// GetConfigurationVersion retrieves a specific version of a configuration
func (r *ConfigurationRepository) GetConfigurationVersion(name string, version int) (*entity.Configuration, error) {
	var config entity.Configuration

	// Check if version exists
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM versions WHERE name = ? AND version = ?)",
		name, version,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFoundError("Configuration version", fmt.Sprintf("%s:%d", name, version))
	}

	// Get version info
	var createdAt time.Time
	var isRollback bool
	err = r.db.QueryRow(
		"SELECT created_at, is_rollback FROM versions WHERE name = ? AND version = ?",
		name, version,
	).Scan(&createdAt, &isRollback)
	if err != nil {
		return nil, err
	}

	// Get data from version_data table
	var dataStr string
	err = r.db.QueryRow(
		"SELECT data FROM version_data WHERE name = ? AND version = ?",
		name, version,
	).Scan(&dataStr)
	if err != nil {
		return nil, err
	}

	// Get original creation time
	var originalCreatedAt time.Time
	err = r.db.QueryRow(
		"SELECT created_at FROM configurations WHERE name = ?",
		name,
	).Scan(&originalCreatedAt)
	if err != nil {
		return nil, err
	}

	config = entity.Configuration{
		Name:      name,
		Version:   version,
		Data:      json.RawMessage(dataStr),
		CreatedAt: originalCreatedAt,
		UpdatedAt: createdAt,
	}

	return &config, nil
}

// ListConfigurationVersions lists all versions of a configuration
func (r *ConfigurationRepository) ListConfigurationVersions(name string) (*entity.VersionList, error) {
	// Check if configuration exists
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM configurations WHERE name = ?)",
		name,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFoundError("Configuration", name)
	}

	// Query versions
	rows, err := r.db.Query(
		"SELECT version, created_at, is_rollback FROM versions WHERE name = ? ORDER BY version",
		name,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := []entity.VersionInfo{}
	for rows.Next() {
		var version entity.VersionInfo
		err := rows.Scan(&version.Version, &version.CreatedAt, &version.IsRollback)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return &entity.VersionList{
		Name:     name,
		Versions: versions,
	}, nil
}

// RegisterSchema registers a JSON schema for a configuration
func (r *ConfigurationRepository) RegisterSchema(configName string, schema json.RawMessage) error {
	// Check if schema already exists
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM schemas WHERE name = ?)",
		configName,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Update existing schema
		_, err = r.db.Exec(
			"UPDATE schemas SET schema = ? WHERE name = ?",
			string(schema), configName,
		)
	} else {
		// Insert new schema
		_, err = r.db.Exec(
			"INSERT INTO schemas (name, schema) VALUES (?, ?)",
			configName, string(schema),
		)
	}

	return err
}

// GetSchema retrieves the JSON schema for a configuration
func (r *ConfigurationRepository) GetSchema(configName string) (json.RawMessage, error) {
	var schemaStr string
	err := r.db.QueryRow(
		"SELECT schema FROM schemas WHERE name = ?",
		configName,
	).Scan(&schemaStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Schema", configName)
		}
		return nil, err
	}

	return json.RawMessage(schemaStr), nil
}

// StoreVersionData stores the raw data for a specific version
func (r *ConfigurationRepository) StoreVersionData(configName string, version int, data json.RawMessage) error {
	_, err := r.db.Exec(
		"INSERT OR REPLACE INTO version_data (name, version, data) VALUES (?, ?, ?)",
		configName, version, string(data),
	)
	return err
}

// GetVersionData retrieves the raw data for a specific version
func (r *ConfigurationRepository) GetVersionData(configName string, version int) (json.RawMessage, error) {
	var dataStr string
	err := r.db.QueryRow(
		"SELECT data FROM version_data WHERE name = ? AND version = ?",
		configName, version,
	).Scan(&dataStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Version data", fmt.Sprintf("%s:%d", configName, version))
		}
		return nil, err
	}

	return json.RawMessage(dataStr), nil
}

// Close closes the database connection
func (r *ConfigurationRepository) Close() error {
	return r.db.Close()
}
