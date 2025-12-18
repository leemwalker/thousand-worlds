package character

import (
	"context"
	"fmt"
	"time"

	"log"

	"tw-backend/internal/auth"
	"tw-backend/internal/game/constants"

	"github.com/google/uuid"

	apperrors "tw-backend/internal/errors"
)

// CreationService handles character creation logic
type CreationService struct {
	authRepo auth.Repository
}

// NewCreationService creates a new CreationService
func NewCreationService(authRepo auth.Repository) *CreationService {
	return &CreationService{
		authRepo: authRepo,
	}
}

// GenerationRequest contains data needed to generate a new character
type GenerationRequest struct {
	PlayerID        uuid.UUID      `json:"player_id"`
	Name            string         `json:"name"`
	Species         string         `json:"species"`
	VarianceSeed    int64          `json:"variance_seed"`
	PointBuyChoices map[string]int `json:"point_buy_choices"`
}

// InhabitationRequest contains data needed to inhabit an NPC
type InhabitationRequest struct {
	PlayerID uuid.UUID `json:"player_id"`
	NPCID    uuid.UUID `json:"npc_id"`
}

// GenerateCharacter creates a new character from scratch
func (s *CreationService) GenerateCharacter(req GenerationRequest) (*Character, *CharacterCreatedViaGenerationEvent, error) {
	// 1. Validate Species
	template := GetSpeciesTemplate(req.Species)
	if template.Name == "" {
		return nil, nil, apperrors.NewInvalidInput("invalid species: %s", req.Species)
	}

	// 2. Apply Variance
	_, varianceData := ApplyVariance(template.BaseAttrs, req.VarianceSeed)

	// Convert VarianceData back to Attributes for storage/event (simplified mapping)
	varianceAttrs := Attributes{
		Might: varianceData.Might, Agility: varianceData.Agility, Endurance: varianceData.Endurance,
		Reflexes: varianceData.Reflexes, Vitality: varianceData.Vitality,
		Intellect: varianceData.Intellect, Cunning: varianceData.Cunning, Willpower: varianceData.Willpower,
		Presence: varianceData.Presence, Intuition: varianceData.Intuition,
		Sight: varianceData.Sight, Hearing: varianceData.Hearing, Smell: varianceData.Smell,
		Taste: varianceData.Taste, Touch: varianceData.Touch,
	}

	// 3. Validate Point Buy
	if err := ValidatePointBuy(template.BaseAttrs, varianceAttrs, req.PointBuyChoices); err != nil {
		return nil, nil, apperrors.NewInvalidInput("invalid point buy allocation: %v", err)
	}

	// 4. Calculate Final Attributes
	finalAttrs := ApplyPointBuy(template.BaseAttrs, varianceAttrs, req.PointBuyChoices)

	// 5. Calculate Secondary Attributes
	secAttrs := CalculateSecondaryAttributes(finalAttrs)

	// 6. Create Character
	charID := uuid.New()
	character := &Character{
		ID:        charID,
		PlayerID:  req.PlayerID,
		Name:      req.Name,
		Species:   req.Species,
		BaseAttrs: finalAttrs,
		SecAttrs:  secAttrs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 7. Create Event
	event := &CharacterCreatedViaGenerationEvent{
		CharacterID:     charID,
		PlayerID:        req.PlayerID,
		Name:            req.Name,
		Species:         req.Species,
		BaseAttributes:  template.BaseAttrs,
		Variance:        varianceAttrs,
		PointBuyChoices: req.PointBuyChoices,
		FinalAttributes: finalAttrs,
		Timestamp:       character.CreatedAt,
	}

	return character, event, nil
}

// InhabitNPC takes over an existing NPC
func (s *CreationService) InhabitNPC(req InhabitationRequest) (*Character, *CharacterCreatedViaInhabitanceEvent, error) {
	// TODO: In Phase 3, we would fetch the NPC from the NPCRepository
	// For now, we mock the NPC data retrieval or assume it exists

	// Mock NPC Data retrieval
	// npc, err := s.npcRepo.Get(req.NPCID)

	charID := uuid.New()

	// Mock Baseline Snapshot (would come from NPC history)
	baseline := BehavioralBaseline{
		Aggression:   0.5,
		Generosity:   0.5,
		Honesty:      0.5,
		Sociability:  0.5,
		Recklessness: 0.5,
		Loyalty:      0.5,
	}

	// Create Event
	event := &CharacterCreatedViaInhabitanceEvent{
		CharacterID:      charID,
		PlayerID:         req.PlayerID,
		NPCID:            req.NPCID,
		BaselineSnapshot: baseline,
		Timestamp:        time.Now(),
	}

	// Create Character (Placeholder - in reality we'd convert NPC to Character)
	character := &Character{
		ID:        charID,
		PlayerID:  req.PlayerID,
		Name:      "Inhabited NPC", // Would come from NPC
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return character, event, nil
}

// EnsureCharacter ensures the user has a character in the specified world
// Returns existing character or creates a new one
func (s *CreationService) EnsureCharacter(ctx context.Context, userID uuid.UUID, worldID uuid.UUID) (*auth.Character, error) {
	// Validate input
	if userID == uuid.Nil {
		return nil, apperrors.NewInvalidInput("invalid user ID: cannot be nil")
	}

	// Check if character exists in this world
	char, err := s.authRepo.GetCharacterByUserAndWorld(ctx, userID, worldID)
	if err == nil && char != nil {
		// Character exists
		return char, nil
	}

	// Log that we're creating a new character
	log.Printf("[CHARACTER] Creating new character for user %s in world %s", userID, worldID)

	// Create new character
	newChar, err := s.createWorldCharacter(ctx, userID, worldID)
	if err != nil {
		log.Printf("[CHARACTER] ERROR: Failed to create character for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	log.Printf("[CHARACTER] Successfully created character %s for user %s", newChar.CharacterID, userID)
	return newChar, nil
}

// createWorldCharacter creates a new character in the specified world
// Copies appearance from user's most recent character if available
func (s *CreationService) createWorldCharacter(ctx context.Context, userID uuid.UUID, worldID uuid.UUID) (*auth.Character, error) {
	// Get user's existing characters to copy appearance from
	userChars, err := s.authRepo.GetUserCharacters(ctx, userID)

	// Default values
	name := "Traveler"
	role := "traveler"
	appearance := `{"form":"humanoid","style":"default"}`

	// Specific defaults for Lobby
	// Specific defaults for Lobby
	posX := 5.0
	posY := 5.0

	if constants.IsLobby(worldID) {
		name = "Ghost"
		role = "ghost"
		appearance = `{"form":"translucent","color":"pale"}`
		posX = 5.0
		posY = 2.0
	}

	// Copy appearance from most recent character if available
	if err == nil && len(userChars) > 0 {
		lastChar := userChars[0] // Sorted by LastPlayed DESC or CreatedAt DESC
		if lastChar.Name != "" {
			name = lastChar.Name
		}
		if lastChar.Appearance != "" {
			appearance = lastChar.Appearance
		}
		// Maintain player role if they have characters (unless it's a specific mechanic to downgrade)
		// For Lobby, we might want to keep them as "player" even if they look like a ghost?
		// or if they are entering a new world, they are a "traveler".
		if !constants.IsLobby(worldID) {
			role = "player"
		} else {
			// In lobby, returning players are "player" role usually?
			role = "player"
		}
	} else if err != nil {
		// Log warning but continue with default values
		log.Printf("[CHARACTER] WARNING: Could not retrieve user characters: %v. Using default appearance.", err)
	}

	// Create the character
	newChar := &auth.Character{
		CharacterID: uuid.New(),
		UserID:      userID,
		WorldID:     worldID,
		Name:        name,
		Role:        role,
		Appearance:  appearance,
		PositionX:   posX,
		PositionY:   posY,
		CreatedAt:   time.Now(),
	}

	// Persist to database
	if err := s.authRepo.CreateCharacter(ctx, newChar); err != nil {
		return nil, fmt.Errorf("failed to persist character: %w", err)
	}

	return newChar, nil
}
