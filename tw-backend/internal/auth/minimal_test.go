package auth_test

import (
	"context"
	"testing"
	"time"

	"tw-backend/internal/auth"
	"tw-backend/internal/testutil"

	"github.com/google/uuid"
)

// Test_MinimalCharacterCreate tests just character creation with minimal data
func Test_MinimalCharacterCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutil.SetupTestDB(t)
	defer testutil.CloseDB(t, db)
	testutil.TruncateTables(t, db)

	repo := auth.NewPostgresRepository(db)

	// Create user
	user := &auth.User{
		UserID:    uuid.New(),
		Email:     "minimal@example.com",
		CreatedAt: time.Now(),
	}
	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create world
	worldID := uuid.New()
	_, err = db.Exec(`
		INSERT INTO worlds (id, name, shape, created_at)
		VALUES ($1, $2, $3, $4)
	`, worldID, "Test World", "sphere", time.Now())
	if err != nil {
		t.Fatalf("Failed to create world: %v", err)
	}

	// Create character with valid JSON appearance
	char1 := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      user.UserID,
		WorldID:     worldID,
		Name:        "Character with JSON",
		Role:        "player",
		Appearance:  `{"hair":"brown"}`,
		CreatedAt:   time.Now(),
	}
	err = repo.CreateCharacter(context.Background(), char1)
	if err != nil {
		t.Fatalf("Failed to create character with JSON appearance: %v", err)
	}
	t.Logf("✓ Created character with JSON appearance")

	// Create character with empty appearance in DIFFERENT world (unique constraint)
	worldID2 := uuid.New()
	_, err = db.Exec(`
		INSERT INTO worlds (id, name, shape, created_at)
		VALUES ($1, $2, $3, $4)
	`, worldID2, "Test World 2", "sphere", time.Now())
	if err != nil {
		t.Fatalf("Failed to create world 2: %v", err)
	}

	char2 := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      user.UserID,
		WorldID:     worldID2, // Different world
		Name:        "Character without appearance",
		Role:        "watcher",
		Appearance:  "",
		CreatedAt:   time.Now(),
	}
	err = repo.CreateCharacter(context.Background(), char2)
	if err != nil {
		t.Fatalf("Failed to create character with empty appearance: %v", err)
	}
	t.Logf("✓ Created character with empty appearance")

	// Retrieve and verify
	retrieved1, err := repo.GetCharacter(context.Background(), char1.CharacterID)
	if err != nil {
		t.Fatalf("Failed to retrieve character 1: %v", err)
	}
	t.Logf("✓ Retrieved character 1: %s, appearance: %s", retrieved1.Name, retrieved1.Appearance)

	retrieved2, err := repo.GetCharacter(context.Background(), char2.CharacterID)
	if err != nil {
		t.Fatalf("Failed to retrieve character 2: %v", err)
	}
	t.Logf("✓ Retrieved character 2: %s, appearance: %s", retrieved2.Name, retrieved2.Appearance)
}
