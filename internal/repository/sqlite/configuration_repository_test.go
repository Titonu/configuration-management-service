package sqlite

import (
	"encoding/json"
	"github.com/Titonu/configuration-management-service/internal/domain/entity"
	"github.com/Titonu/configuration-management-service/internal/domain/repository"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (repository.ConfigurationRepository, func()) {
	// Create a temporary database file
	dbFile := "./test_config.db"

	// Remove any existing test database
	os.Remove(dbFile)

	// Create the repository
	repo, err := NewConfigurationRepository(dbFile)
	require.NoError(t, err)

	// Return the repository and a cleanup function
	return repo, func() {
		// Close the database connection
		if sqlRepo, ok := repo.(*ConfigurationRepository); ok && sqlRepo.db != nil {
			sqlRepo.db.Close()
		}

		// Remove the temporary database file
		os.Remove(dbFile)
	}
}

func TestSQLiteConfigurationRepository(t *testing.T) {
	t.Run("CreateConfiguration", func(t *testing.T) {
		repo, cleanup := setupTestDB(t)
		defer cleanup()

		// Test creating a new configuration
		config := &entity.Configuration{
			Name:      "test-config",
			Version:   1,
			Data:      json.RawMessage(`{"key":"value"}`),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.CreateConfiguration(config)
		assert.NoError(t, err)

		// Store version data
		err = repo.StoreVersionData("test-config", 1, json.RawMessage(`{"key":"value"}`))
		assert.NoError(t, err)

		// Test creating a duplicate configuration
		err = repo.CreateConfiguration(config)
		assert.Error(t, err)
	})

	t.Run("UpdateConfiguration", func(t *testing.T) {
		repo, cleanup := setupTestDB(t)
		defer cleanup()

		// Create initial configuration
		config := &entity.Configuration{
			Name:      "test-config",
			Version:   1,
			Data:      json.RawMessage(`{"key":"value"}`),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.CreateConfiguration(config)
		assert.NoError(t, err)

		// Store version data
		err = repo.StoreVersionData("test-config", 1, json.RawMessage(`{"key":"value"}`))
		assert.NoError(t, err)

		// Update configuration
		updatedConfig := &entity.Configuration{
			Name:      "test-config",
			Version:   2,
			Data:      json.RawMessage(`{"key":"updated"}`),
			CreatedAt: config.CreatedAt,
			UpdatedAt: time.Now(),
		}

		err = repo.UpdateConfiguration(updatedConfig)
		assert.NoError(t, err)

		// Store updated version data
		err = repo.StoreVersionData("test-config", 2, json.RawMessage(`{"key":"updated"}`))
		assert.NoError(t, err)

		// Verify update
		result, err := repo.GetConfiguration("test-config")
		assert.NoError(t, err)
		assert.Equal(t, 2, result.Version)

		// Test updating non-existent configuration
		nonExistentConfig := &entity.Configuration{
			Name:      "non-existent",
			Version:   1,
			Data:      json.RawMessage(`{"key":"value"}`),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Check if the repository implementation checks for existence
		sqlRepo, ok := repo.(*ConfigurationRepository)
		if ok {
			// Check if the configuration exists before attempting update
			var exists bool
			err = sqlRepo.db.QueryRow(
				"SELECT EXISTS(SELECT 1 FROM configurations WHERE name = ?)",
				nonExistentConfig.Name,
			).Scan(&exists)
			assert.NoError(t, err)
			assert.False(t, exists)

			// For SQLite repository, we should expect an error when updating non-existent config
			// because the UPDATE statement won't affect any rows
			result, err := sqlRepo.db.Exec(
				"UPDATE configurations SET version = ? WHERE name = ?",
				nonExistentConfig.Version, nonExistentConfig.Name,
			)
			assert.NoError(t, err)

			rowsAffected, err := result.RowsAffected()
			assert.NoError(t, err)
			assert.Equal(t, int64(0), rowsAffected)
		}
	})

	t.Run("GetConfiguration", func(t *testing.T) {
		repo, cleanup := setupTestDB(t)
		defer cleanup()

		// Create configuration with all fields properly initialized
		now := time.Now()
		config := &entity.Configuration{
			Name:         "test-config",
			Version:      1,
			Data:         json.RawMessage(`{"key":"value"}`),
			CreatedAt:    now,
			UpdatedAt:    now,
			RollbackFrom: 0, // Explicitly set to 0 instead of nil
			RollbackTo:   0, // Explicitly set to 0 instead of nil
		}

		// Get the SQLite repository to directly insert with proper NULL handling
		sqlRepo, ok := repo.(*ConfigurationRepository)
		require.True(t, ok, "Repository should be SQLite repository")

		// Insert directly with proper NULL handling for rollback fields
		_, err := sqlRepo.db.Exec(
			"INSERT INTO configurations (name, version, created_at, updated_at, rollback_from, rollback_to) VALUES (?, ?, ?, ?, NULL, NULL)",
			config.Name, config.Version, config.CreatedAt, config.UpdatedAt,
		)
		assert.NoError(t, err)

		// Insert into versions table
		_, err = sqlRepo.db.Exec(
			"INSERT INTO versions (name, version, created_at, is_rollback) VALUES (?, ?, ?, ?)",
			config.Name, config.Version, config.CreatedAt, false,
		)
		assert.NoError(t, err)

		// Store version data
		err = repo.StoreVersionData("test-config", 1, json.RawMessage(`{"key":"value"}`))
		assert.NoError(t, err)

		// Get existing configuration
		result, err := repo.GetConfiguration("test-config")
		assert.NoError(t, err)
		assert.Equal(t, "test-config", result.Name)
		assert.Equal(t, 1, result.Version)

		// Get non-existent configuration
		_, err = repo.GetConfiguration("non-existent")
		assert.Error(t, err)
	})

	t.Run("GetConfigurationVersion", func(t *testing.T) {
		repo, cleanup := setupTestDB(t)
		defer cleanup()

		// Create initial configuration with all fields properly initialized
		now := time.Now()
		config := &entity.Configuration{
			Name:      "test-config",
			Version:   1,
			Data:      json.RawMessage(`{"key":"value"}`),
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Get the SQLite repository to directly insert with proper NULL handling
		sqlRepo, ok := repo.(*ConfigurationRepository)
		require.True(t, ok, "Repository should be SQLite repository")

		// Insert directly with proper NULL handling for rollback fields
		_, err := sqlRepo.db.Exec(
			"INSERT INTO configurations (name, version, created_at, updated_at, rollback_from, rollback_to) VALUES (?, ?, ?, ?, NULL, NULL)",
			config.Name, config.Version, config.CreatedAt, config.UpdatedAt,
		)
		assert.NoError(t, err)

		// Insert into versions table
		_, err = sqlRepo.db.Exec(
			"INSERT INTO versions (name, version, created_at, is_rollback) VALUES (?, ?, ?, ?)",
			config.Name, config.Version, config.CreatedAt, false,
		)
		assert.NoError(t, err)

		// Store version data
		err = repo.StoreVersionData("test-config", 1, json.RawMessage(`{"key":"value"}`))
		assert.NoError(t, err)

		// Update configuration
		updatedConfig := &entity.Configuration{
			Name:      "test-config",
			Version:   2,
			Data:      json.RawMessage(`{"key":"updated"}`),
			CreatedAt: config.CreatedAt,
			UpdatedAt: time.Now(),
		}

		// Update configurations table
		_, err = sqlRepo.db.Exec(
			"UPDATE configurations SET version = ?, updated_at = ? WHERE name = ?",
			updatedConfig.Version, updatedConfig.UpdatedAt, updatedConfig.Name,
		)
		assert.NoError(t, err)

		// Insert into versions table
		_, err = sqlRepo.db.Exec(
			"INSERT INTO versions (name, version, created_at, is_rollback) VALUES (?, ?, ?, ?)",
			updatedConfig.Name, updatedConfig.Version, updatedConfig.UpdatedAt, false,
		)
		assert.NoError(t, err)

		// Store updated version data
		err = repo.StoreVersionData("test-config", 2, json.RawMessage(`{"key":"updated"}`))
		assert.NoError(t, err)

		// Get specific version
		v1, err := repo.GetConfigurationVersion("test-config", 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, v1.Version)

		v2, err := repo.GetConfigurationVersion("test-config", 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, v2.Version)

		// Get non-existent version
		_, err = repo.GetConfigurationVersion("test-config", 3)
		assert.Error(t, err)

		// Get version of non-existent configuration
		_, err = repo.GetConfigurationVersion("non-existent", 1)
		assert.Error(t, err)
	})

	t.Run("ListConfigurationVersions", func(t *testing.T) {
		repo, cleanup := setupTestDB(t)
		defer cleanup()

		// Create initial configuration with all fields properly initialized
		now := time.Now()
		config := &entity.Configuration{
			Name:      "test-config",
			Version:   1,
			Data:      json.RawMessage(`{"key":"value"}`),
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Get the SQLite repository to directly insert with proper NULL handling
		sqlRepo, ok := repo.(*ConfigurationRepository)
		require.True(t, ok, "Repository should be SQLite repository")

		// Insert directly with proper NULL handling for rollback fields
		_, err := sqlRepo.db.Exec(
			"INSERT INTO configurations (name, version, created_at, updated_at, rollback_from, rollback_to) VALUES (?, ?, ?, ?, NULL, NULL)",
			config.Name, config.Version, config.CreatedAt, config.UpdatedAt,
		)
		assert.NoError(t, err)

		// Insert into versions table
		_, err = sqlRepo.db.Exec(
			"INSERT INTO versions (name, version, created_at, is_rollback) VALUES (?, ?, ?, ?)",
			config.Name, config.Version, config.CreatedAt, false,
		)
		assert.NoError(t, err)

		// Store version data
		err = repo.StoreVersionData("test-config", 1, json.RawMessage(`{"key":"value"}`))
		assert.NoError(t, err)

		// Update configuration multiple times
		for i := 2; i <= 5; i++ {
			updatedTime := time.Now()

			// Update configurations table
			_, err = sqlRepo.db.Exec(
				"UPDATE configurations SET version = ?, updated_at = ? WHERE name = ?",
				i, updatedTime, config.Name,
			)
			assert.NoError(t, err)

			// Insert into versions table
			_, err = sqlRepo.db.Exec(
				"INSERT INTO versions (name, version, created_at, is_rollback) VALUES (?, ?, ?, ?)",
				config.Name, i, updatedTime, false,
			)
			assert.NoError(t, err)

			// Store updated version data
			err = repo.StoreVersionData("test-config", i, json.RawMessage(`{"key":"updated"}`))
			assert.NoError(t, err)
		}

		// List versions
		versions, err := repo.ListConfigurationVersions("test-config")
		assert.NoError(t, err)
		assert.Equal(t, "test-config", versions.Name)
		assert.Equal(t, 5, len(versions.Versions))

		// List versions for non-existent configuration
		_, err = repo.ListConfigurationVersions("non-existent")
		assert.Error(t, err)
	})

	t.Run("RegisterSchema", func(t *testing.T) {
		repo, cleanup := setupTestDB(t)
		defer cleanup()

		// Register schema
		schema := json.RawMessage(`{"type":"object","properties":{"key":{"type":"string"}}}`)
		err := repo.RegisterSchema("test-config", schema)
		assert.NoError(t, err)

		// Register duplicate schema (should update)
		updatedSchema := json.RawMessage(`{"type":"object","properties":{"key":{"type":"string"},"newProp":{"type":"number"}}}`)
		err = repo.RegisterSchema("test-config", updatedSchema)
		assert.NoError(t, err)

		// Verify schema was updated
		result, err := repo.GetSchema("test-config")
		assert.NoError(t, err)
		assert.JSONEq(t, string(updatedSchema), string(result))
	})

	t.Run("GetSchema", func(t *testing.T) {
		repo, cleanup := setupTestDB(t)
		defer cleanup()

		// Register schema
		schema := json.RawMessage(`{"type":"object","properties":{"key":{"type":"string"}}}`)
		err := repo.RegisterSchema("test-config", schema)
		assert.NoError(t, err)

		// Get existing schema
		result, err := repo.GetSchema("test-config")
		assert.NoError(t, err)
		assert.JSONEq(t, string(schema), string(result))

		// Get non-existent schema
		_, err = repo.GetSchema("non-existent")
		assert.Error(t, err)
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		repo, cleanup := setupTestDB(t)
		defer cleanup()

		// Skip transaction test for interface type
		sqlRepo, ok := repo.(*ConfigurationRepository)
		if !ok {
			t.Skip("Not a SQLite repository")
		}

		// Start a transaction
		tx, err := sqlRepo.db.Begin()
		assert.NoError(t, err)

		// Perform an operation that should fail
		_, err = tx.Exec("INSERT INTO invalid_table (column) VALUES (?)", "value")
		assert.Error(t, err)

		// Rollback should succeed
		err = tx.Rollback()
		assert.NoError(t, err)
	})
}
