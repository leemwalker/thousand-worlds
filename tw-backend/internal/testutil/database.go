package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// SetupTestDB creates a connection to the test database
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		// Default to local test database
		dbURL = "postgres://admin:test_password_123456@localhost:5432/mud_core?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Verify connection
	err = db.Ping()
	require.NoError(t, err, "Failed to ping test database")

	return db
}

// TruncateTables cleans all tables for a fresh test state
func TruncateTables(t *testing.T, db *sql.DB) {
	t.Helper()

	// Tables to fully truncate (order matters due to foreign keys)
	tables := []string{
		"characters", // Has FK to users and worlds
		"sessions",   // Has FK to users
		"users",      // Referenced by characters and sessions
		"world_interviews",
		"world_configurations",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// Table might not exist in all test scenarios, log but don't fail
			t.Logf("Warning: failed to truncate %s: %v", table, err)
		}
	}

	// For worlds table, delete all EXCEPT the lobby world (00000000-0000-0000-0000-000000000000)
	// This preserves the lobby world which is required for lobby character tests
	_, err := db.Exec("DELETE FROM worlds WHERE id != '00000000-0000-0000-0000-000000000000'")
	if err != nil {
		t.Logf("Warning: failed to clean worlds table: %v", err)
	}
}

// RunMigrations executes all migration files
func RunMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	// Enable PostGIS
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
	require.NoError(t, err, "Failed to enable PostGIS")

	migrationFiles := []string{
		"000001_create_worlds_table.up.sql",
		"000013_create_auth_tables.up.sql",
		"000014_create_interview_tables.up.sql",
		"000015_add_character_role_and_appearance.up.sql",
		"000016_add_character_description_occupation.up.sql",
		"000017_add_performance_indexes.up.sql",
		"000018_create_lobby_world.up.sql",
		"000019_add_username_to_users.up.sql",
		"000020_add_world_name_to_configurations.up.sql",
		"000021_add_last_world_id_to_users.up.sql",
	}

	// Try to find the migrations directory
	// We check multiple depths to handle tests in different package levels
	basePaths := []string{
		"../../migrations/postgres",       // for internal/package
		"../../../migrations/postgres",    // for internal/group/package
		"../../../../migrations/postgres", // for internal/group/subgroup/package
		"migrations/postgres",             // for root
	}

	var migrationDir string
	for _, path := range basePaths {
		if _, err := os.Stat(path); err == nil {
			migrationDir = path
			break
		}
	}

	if migrationDir == "" {
		t.Log("Warning: could not find migrations directory")
		return
	}

	for _, file := range migrationFiles {
		path := fmt.Sprintf("%s/%s", migrationDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Logf("Skipping migration %s: %v", file, err)
			continue
		}

		_, err = db.Exec(string(content))
		if err != nil {
			// Migration might already be applied
			t.Logf("Warning: migration %s: %v", file, err)
		}
	}
}

// CloseDB closes the database connection
func CloseDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if db != nil {
		err := db.Close()
		require.NoError(t, err, "Failed to close database")
	}
}
