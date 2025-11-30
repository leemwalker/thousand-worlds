package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents an authenticated user
type User struct {
	UserID       uuid.UUID  `json:"user_id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Never expose password hash
	CreatedAt    time.Time  `json:"created_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
}

// Character represents a player character
type Character struct {
	CharacterID uuid.UUID  `json:"character_id"`
	UserID      uuid.UUID  `json:"user_id"`
	WorldID     uuid.UUID  `json:"world_id"`
	Name        string     `json:"name"`
	Position    *Position  `json:"position,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	LastPlayed  *time.Time `json:"last_played,omitempty"`
}

// Position represents geographic position
type Position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Repository defines data access methods
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error

	// Character operations
	CreateCharacter(ctx context.Context, char *Character) error
	GetCharacter(ctx context.Context, characterID uuid.UUID) (*Character, error)
	GetUserCharacters(ctx context.Context, userID uuid.UUID) ([]*Character, error)
	GetCharacterByUserAndWorld(ctx context.Context, userID, worldID uuid.UUID) (*Character, error)
	UpdateCharacter(ctx context.Context, char *Character) error
}
