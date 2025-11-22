package character

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreationService handles character creation logic
type CreationService struct {
	// Dependencies would go here (e.g., EventStore, NPCRepository)
}

// NewCreationService creates a new CreationService
func NewCreationService() *CreationService {
	return &CreationService{}
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
		return nil, nil, fmt.Errorf("invalid species: %s", req.Species)
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
		return nil, nil, fmt.Errorf("invalid point buy allocation: %w", err)
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
