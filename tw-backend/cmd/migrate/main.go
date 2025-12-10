package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://admin:password123@localhost:5432/mud_core?sslmode=disable"
	}

	log.Printf("Connecting to database: %s", dbURL)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Enable PostGIS extension
	log.Println("Enabling PostGIS extension...")
	if _, err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis"); err != nil {
		log.Fatal("Failed to enable PostGIS:", err)
	}

	migrationsDir := "migrations/postgres"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatal(err)
	}

	var migrationFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}
	sort.Strings(migrationFiles)

	for _, file := range migrationFiles {
		log.Printf("Running migration: %s", file)
		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			log.Fatal(err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				log.Printf("Migration %s already applied (or object exists): %v", file, err)
			} else {
				log.Printf("Error running migration %s: %v", file, err)
				log.Fatal(err)
			}
		}
	}

	log.Println("Migrations completed successfully.")
}
