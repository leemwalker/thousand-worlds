package testutil

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigration000017_Success verifies all indexes are created successfully
func TestMigration000017_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CloseDB(t, db)

	// Enable PostGIS (required for some tables)
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
	require.NoError(t, err)

	// Run prerequisite migrations (create tables first)
	runPrerequisiteMigrations(t, db)

	// Read and execute migration 000017
	content, err := os.ReadFile("../../migrations/postgres/000017_add_performance_indexes.up.sql")
	require.NoError(t, err, "Failed to read migration file")

	_, err = db.Exec(string(content))
	require.NoError(t, err, "Migration 000017 should execute without errors")

	// Verify indexes were created
	expectedIndexes := []string{
		"idx_characters_world_position",
		"idx_users_last_login",
		"idx_worlds_metadata_gin",
		"idx_sessions_expires_user",
		// NOTE: idx_sessions_active is removed in the fix
	}

	for _, indexName := range expectedIndexes {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes 
				WHERE indexname = $1
			)
		`, indexName).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Index %s should exist", indexName)
	}

	t.Log("✓ Migration 000017 executed successfully and created all required indexes")
}

// TestMigration000017_NoImmutableError verifies no IMMUTABLE function errors occur
func TestMigration000017_NoImmutableError(t *testing.T) {
	db := SetupTestDB(t)
	defer CloseDB(t, db)

	// Enable PostGIS
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
	require.NoError(t, err)

	// Run prerequisite migrations
	runPrerequisiteMigrations(t, db)

	// Read and execute migration 000017
	content, err := os.ReadFile("../../migrations/postgres/000017_add_performance_indexes.up.sql")
	require.NoError(t, err)

	_, err = db.Exec(string(content))

	// Assert no error (specifically no IMMUTABLE error)
	require.NoError(t, err, "Should not have IMMUTABLE function errors")

	// If there was an error, it would contain "IMMUTABLE" - verify it doesn't
	if err != nil {
		assert.NotContains(t, err.Error(), "IMMUTABLE", "Should not have IMMUTABLE function errors")
		assert.NotContains(t, err.Error(), "index predicate", "Should not have index predicate errors")
	}

	t.Log("✓ Migration 000017 has no IMMUTABLE function errors")
}

// TestMigration000017_Idempotency verifies migration can run multiple times
func TestMigration000017_Idempotency(t *testing.T) {
	db := SetupTestDB(t)
	defer CloseDB(t, db)

	// Enable PostGIS
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
	require.NoError(t, err)

	// Run prerequisite migrations
	runPrerequisiteMigrations(t, db)

	// Read migration content
	content, err := os.ReadFile("../../migrations/postgres/000017_add_performance_indexes.up.sql")
	require.NoError(t, err)

	// Run migration first time
	_, err = db.Exec(string(content))
	require.NoError(t, err, "First execution should succeed")

	// Run migration second time - should not error due to IF NOT EXISTS
	_, err = db.Exec(string(content))
	require.NoError(t, err, "Second execution should succeed (idempotent)")

	// Run migration third time for good measure
	_, err = db.Exec(string(content))
	require.NoError(t, err, "Third execution should succeed (idempotent)")

	t.Log("✓ Migration 000017 is idempotent and can run multiple times")
}

// TestMigration000017_DownMigration verifies rollback removes all indexes
func TestMigration000017_DownMigration(t *testing.T) {
	db := SetupTestDB(t)
	defer CloseDB(t, db)

	// Enable PostGIS
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
	require.NoError(t, err)

	// Run prerequisite migrations
	runPrerequisiteMigrations(t, db)

	// Run up migration
	upContent, err := os.ReadFile("../../migrations/postgres/000017_add_performance_indexes.up.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(upContent))
	require.NoError(t, err)

	// Verify indexes exist before rollback
	var indexCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pg_indexes 
		WHERE indexname IN (
			'idx_characters_world_position',
			'idx_users_last_login',
			'idx_worlds_metadata_gin',
			'idx_sessions_expires_user'
		)
	`).Scan(&indexCount)
	require.NoError(t, err)
	assert.Greater(t, indexCount, 0, "Indexes should exist before rollback")

	// Run down migration
	downContent, err := os.ReadFile("../../migrations/postgres/000017_add_performance_indexes.down.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(downContent))
	require.NoError(t, err, "Down migration should execute successfully")

	// Verify indexes are removed
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pg_indexes 
		WHERE indexname IN (
			'idx_characters_world_position',
			'idx_users_last_login',
			'idx_worlds_metadata_gin',
			'idx_sessions_expires_user'
		)
	`).Scan(&indexCount)
	require.NoError(t, err)
	assert.Equal(t, 0, indexCount, "All indexes should be removed after rollback")

	t.Log("✓ Down migration successfully removes all indexes")
}

// runPrerequisiteMigrations runs migrations needed before 000017
func runPrerequisiteMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	prerequisiteMigrations := []string{
		"../../migrations/postgres/000001_create_worlds_table.up.sql",
		"../../migrations/postgres/000013_create_auth_tables.up.sql",
		"../../migrations/postgres/000014_create_interview_tables.up.sql",
		"../../migrations/postgres/000015_add_character_role_and_appearance.up.sql",
		"../../migrations/postgres/000016_add_character_description_occupation.up.sql",
	}

	for _, migration := range prerequisiteMigrations {
		content, err := os.ReadFile(migration)
		if err != nil {
			t.Logf("Warning: Could not read %s: %v", migration, err)
			continue
		}

		_, err = db.Exec(string(content))
		if err != nil {
			// Might already exist from previous test - log but continue
			t.Logf("Warning: Migration %s: %v", migration, err)
		}
	}
}
