package character

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCharacter(t *testing.T) {
	service := NewCreationService()
	playerID := uuid.New()

	req := GenerationRequest{
		PlayerID:     playerID,
		Name:         "Test Hero",
		Species:      SpeciesHuman,
		VarianceSeed: 1,
		PointBuyChoices: map[string]int{
			AttrMight: 10, // Valid +10
		},
	}

	char, event, err := service.GenerateCharacter(req)
	assert.NoError(t, err)
	assert.NotNil(t, char)
	assert.NotNil(t, event)

	assert.Equal(t, playerID, char.PlayerID)
	assert.Equal(t, "Test Hero", char.Name)
	assert.Equal(t, SpeciesHuman, char.Species)

	// Check attributes: Base(50) + Variance(Seed 1 for Might is -3) + PointBuy(10) = 57
	// Wait, let's check the variance logic again.
	// Seed 1: Might variance is rng.Intn(11) - 5.
	// I should probably not rely on exact random values in this high-level test,
	// but rather check that properties hold.

	assert.True(t, char.BaseAttrs.Might > 0)
	assert.True(t, char.SecAttrs.MaxHP > 0)

	// Verify Event
	assert.Equal(t, char.ID, event.CharacterID)
	assert.Equal(t, req.PointBuyChoices, event.PointBuyChoices)
}

func TestGenerateCharacter_InvalidSpecies(t *testing.T) {
	service := NewCreationService()
	req := GenerationRequest{
		Species: "Alien",
	}
	_, _, err := service.GenerateCharacter(req)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "invalid species")
	}
}

func TestGenerateCharacter_InvalidPointBuy(t *testing.T) {
	service := NewCreationService()
	req := GenerationRequest{
		Species: SpeciesHuman,
		PointBuyChoices: map[string]int{
			AttrMight: 100, // Invalid
		},
	}
	_, _, err := service.GenerateCharacter(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid point buy")
}

func TestInhabitNPC(t *testing.T) {
	service := NewCreationService()
	playerID := uuid.New()
	npcID := uuid.New()

	req := InhabitationRequest{
		PlayerID: playerID,
		NPCID:    npcID,
	}

	char, event, err := service.InhabitNPC(req)
	assert.NoError(t, err)
	assert.NotNil(t, char)
	assert.NotNil(t, event)

	assert.Equal(t, playerID, char.PlayerID)
	assert.Equal(t, npcID, event.NPCID)
	assert.NotNil(t, event.BaselineSnapshot)
}
