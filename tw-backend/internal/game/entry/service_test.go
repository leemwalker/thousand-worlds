package entry

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tw-backend/internal/game/constants"
	"tw-backend/internal/world/interview"
)

// TestGetEntryOptions_Success tests getting entry options for a valid world
func TestGetEntryOptions_Success(t *testing.T) {
	repo := interview.NewMockRepository()
	service := NewService(repo)

	worldID := uuid.New()
	ctx := context.Background()

	// Create world configuration
	config := &interview.WorldConfiguration{
		WorldID:         &worldID,
		SentientSpecies: []string{"Human", "Elf", "Dwarf"},
	}
	repo.SaveConfiguration(ctx, config)

	// Get entry options
	options, err := service.GetEntryOptions(ctx, worldID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, options)
	assert.True(t, options.CanEnterAsWatcher, "Should allow watcher mode")
	assert.True(t, options.CanCreateCustom, "Should allow custom creation")
	assert.Len(t, options.AvailableNPCs, 5, "Should generate 5 NPCs")
}

// TestGetEntryOptions_LobbyWorld tests rejection of lobby world
func TestGetEntryOptions_LobbyWorld(t *testing.T) {
	repo := interview.NewMockRepository()
	service := NewService(repo)

	ctx := context.Background()

	// Try to get entry options for lobby
	options, err := service.GetEntryOptions(ctx, constants.LobbyWorldID)

	// Assert error
	require.Error(t, err)
	assert.Nil(t, options)
	assert.Contains(t, err.Error(), "lobby")
}

// TestGetEntryOptions_NonExistentWorld tests handling of non-existent world
func TestGetEntryOptions_NonExistentWorld(t *testing.T) {
	repo := interview.NewMockRepository()
	service := NewService(repo)

	worldID := uuid.New() // Not in repository
	ctx := context.Background()

	// Get entry options
	options, err := service.GetEntryOptions(ctx, worldID)

	// Assert resilience (defaults used)
	require.NoError(t, err)
	require.NotNil(t, options)
	assert.True(t, options.CanEnterAsWatcher)
	assert.NotEmpty(t, options.AvailableNPCs)
}

// TestGenerateRandomNPCs_CreatesUniqueNPCs tests NPC generation
func TestGenerateRandomNPCs_CreatesUniqueNPCs(t *testing.T) {
	repo := interview.NewMockRepository()
	service := NewService(repo)

	config := &interview.WorldConfiguration{
		SentientSpecies: []string{"Human", "Elf"},
	}

	// Generate NPCs
	npcs := service.generateRandomNPCs(config, 5)

	// Assert
	assert.Len(t, npcs, 5)

	// Verify each NPC has required fields
	ids := make(map[string]bool)
	for _, npc := range npcs {
		assert.NotEmpty(t, npc.ID, "NPC should have ID")
		assert.NotEmpty(t, npc.Name, "NPC should have name")
		assert.NotEmpty(t, npc.Species, "NPC should have species")
		assert.NotEmpty(t, npc.DNA, "NPC should have DNA")
		assert.Contains(t, config.SentientSpecies, npc.Species, "Species should be from config")

		// Check ID uniqueness
		assert.False(t, ids[npc.ID], "NPC IDs should be unique")
		ids[npc.ID] = true
	}
}

// TestGenerateRandomNPCs_DefaultsToHuman tests fallback species
func TestGenerateRandomNPCs_DefaultsToHuman(t *testing.T) {
	repo := interview.NewMockRepository()
	service := NewService(repo)

	config := &interview.WorldConfiguration{
		SentientSpecies: []string{}, // Empty species list
	}

	// Generate NPCs
	npcs := service.generateRandomNPCs(config, 3)

	// Assert defaults to Human
	for _, npc := range npcs {
		assert.Equal(t, "Human", npc.Species, "Should default to Human when no species configured")
	}
}

// TestGenerateRandomDNA_CreatesValidDNA tests DNA generation
func TestGenerateRandomDNA_CreatesValidDNA(t *testing.T) {
	service := NewService(nil)

	// Generate DNA
	dna := service.generateRandomDNA()

	// Assert genes are present
	assert.NotEmpty(t, dna.Genes, "DNA should have genes")

	// Check for expected genes
	expectedGenes := []string{
		"height", "build", "muscle", "hair", "pigment",
		"eye", "melanin", "strength", "reflex", "stamina",
		"health", "cognition", "perception",
	}

	for _, geneName := range expectedGenes {
		gene, exists := dna.Genes[geneName]
		assert.True(t, exists, "Should have gene: %s", geneName)
		assert.Equal(t, geneName, gene.TraitName)
		assert.NotEmpty(t, gene.Allele1, "Allele1 should be set")
		assert.NotEmpty(t, gene.Allele2, "Allele2 should be set")
	}
}

// TestNPCPreview_HasValidDNAEncoding tests DNA can be encoded and decoded
func TestNPCPreview_HasValidDNAEncoding(t *testing.T) {
	repo := interview.NewMockRepository()
	service := NewService(repo)

	worldID := uuid.New()
	config := &interview.WorldConfiguration{
		WorldID:         &worldID,
		SentientSpecies: []string{"Human"},
	}
	repo.SaveConfiguration(context.Background(), config)

	// Get entry options
	options, err := service.GetEntryOptions(context.Background(), worldID)
	require.NoError(t, err)

	// Verify DNA encoding is valid JSON
	for _, npc := range options.AvailableNPCs {
		assert.NotEmpty(t, npc.DNA)
		// DNA should be valid JSON (lowercase "genes" due to struct tags)
		assert.Contains(t, npc.DNA, "genes", "DNA should contain genes field")
	}
}
