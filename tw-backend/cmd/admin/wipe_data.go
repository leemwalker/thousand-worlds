package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection string - assume local default or env
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://admin:test_password_123456@127.0.0.1:5432/mud_core?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Delete all users
	// This will cascade to characters, sessions, etc.
	fmt.Println("Deleting all users...")
	_, err = db.ExecContext(ctx, "DELETE FROM users")
	if err != nil {
		log.Printf("Error deleting users: %v", err)
	} else {
		fmt.Println("Users deleted.")
	}

	// 2. Delete all worlds except Lobby
	// Lobby ID is 00000000-0000-0000-0000-000000000000
	fmt.Println("Deleting all non-Lobby worlds...")
	_, err = db.ExecContext(ctx, "DELETE FROM worlds WHERE id != '00000000-0000-0000-0000-000000000000'")
	if err != nil {
		log.Printf("Error deleting worlds: %v", err)
	} else {
		fmt.Println("Worlds deleted.")
	}

	fmt.Println("Wipe complete.")
}
