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
		dbURL = "postgres://admin:password123@localhost:5432/mud_core?sslmode=disable"
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

	tables := []string{
		"characters",
		"users",
		"interviews",
		"worlds",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// Table might not exist in all test scenarios, log but don't fail
			t.Logf("Warning: failed to truncate %s: %v", table, err)
		}
	}
}

// RunMigrations executes all migration files
func RunMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	// Enable PostGIS
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
	require.NoError(t, err, "Failed to enable PostGIS")

	migrationFiles := []string{
		"migrations/postgres/000013_create_auth_tables.up.sql",
		"migrations/postgres/000015_add_character_role_and_appearance.up.sql",
		"migrations/postgres/000016_add_character_description_occupation.up.sql",
	}

	for _, file := range migrationFiles {
		content, err := os.ReadFile(file)
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
