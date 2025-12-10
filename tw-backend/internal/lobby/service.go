package lobby

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"tw-backend/internal/auth"
)

type Service struct {
	authRepo auth.Repository
}

func NewService(authRepo auth.Repository) *Service {
	return &Service{
		authRepo: authRepo,
	}
}

// EnsureLobbyCharacter ensures the user has a character in the lobby
// Returns existing lobby character or creates a new one
func (s *Service) EnsureLobbyCharacter(ctx context.Context, userID uuid.UUID) (*auth.Character, error) {
	// Validate input
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID: cannot be nil")
	}

	// Check if character exists in lobby
	char, err := s.authRepo.GetCharacterByUserAndWorld(ctx, userID, LobbyWorldID)
	if err == nil && char != nil {
		log.Printf("[LOBBY] User %s already has lobby character %s", userID, char.CharacterID)
		return char, nil
	}

	// Log that we're creating a new lobby character
	log.Printf("[LOBBY] Creating new lobby character for user %s", userID)

	// Create new lobby character
	newChar, err := s.createLobbyCharacter(ctx, userID)
	if err != nil {
		log.Printf("[LOBBY] ERROR: Failed to create lobby character for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to create lobby character: %w", err)
	}

	log.Printf("[LOBBY] Successfully created lobby character %s for user %s", newChar.CharacterID, userID)
	return newChar, nil
}

// createLobbyCharacter creates a new character in the lobby world
// Copies appearance from user's most recent character if available
func (s *Service) createLobbyCharacter(ctx context.Context, userID uuid.UUID) (*auth.Character, error) {
	// Get user's existing characters to copy appearance from
	userChars, err := s.authRepo.GetUserCharacters(ctx, userID)

	// Default values for new lobby character
	name := "Ghost"
	role := "ghost"
	appearance := `{"form":"translucent","color":"pale"}`

	// Copy appearance from most recent character if available
	if err == nil && len(userChars) > 0 {
		lastChar := userChars[0] // Sorted by LastPlayed DESC or CreatedAt DESC
		if lastChar.Name != "" {
			name = lastChar.Name
			log.Printf("[LOBBY] Copying name '%s' from character %s", name, lastChar.CharacterID)
		}
		if lastChar.Appearance != "" {
			appearance = lastChar.Appearance
			log.Printf("[LOBBY] Copying appearance from character %s", lastChar.CharacterID)
		}
		// Maintain player role if they have characters
		role = "player"
	} else if err != nil {
		// Log warning but continue with default values
		log.Printf("[LOBBY] WARNING: Could not retrieve user characters: %v. Using default appearance.", err)
	}

	// Create the lobby character - position at center of lobby (5,5) for freedom of movement
	newChar := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      userID,
		WorldID:     LobbyWorldID,
		Name:        name,
		Role:        role,
		Appearance:  appearance,
		PositionX:   5.0, // Center of 0-10 lobby bounds
		PositionY:   5.0, // Center of 0-10 lobby bounds
		CreatedAt:   time.Now(),
	}

	// Persist to database
	if err := s.authRepo.CreateCharacter(ctx, newChar); err != nil {
		return nil, fmt.Errorf("failed to persist lobby character: %w", err)
	}

	return newChar, nil
}
