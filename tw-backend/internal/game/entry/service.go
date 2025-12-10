package entry

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"tw-backend/internal/lobby"
	"tw-backend/internal/npc/appearance"
	"tw-backend/internal/npc/genetics"
	"tw-backend/internal/world/interview"

	"github.com/google/uuid"
)

type Service struct {
	interviewRepo interview.Repository
}

func NewService(interviewRepo interview.Repository) *Service {
	return &Service{
		interviewRepo: interviewRepo,
	}
}

type EntryOptions struct {
	CanEnterAsWatcher bool         `json:"can_enter_as_watcher"`
	AvailableNPCs     []NPCPreview `json:"available_npcs"`
	CanCreateCustom   bool         `json:"can_create_custom"`
}

type NPCPreview struct {
	ID          string `json:"id"` // Temporary ID for selection
	Name        string `json:"name"`
	Species     string `json:"species"`
	Description string `json:"description"`
	Occupation  string `json:"occupation"`
	DNA         string `json:"dna"` // Encoded DNA to recreate/save
}

// GetEntryOptions returns available entry modes for a world
func (s *Service) GetEntryOptions(ctx context.Context, worldID uuid.UUID) (*EntryOptions, error) {
	if lobby.IsLobby(worldID) {
		return nil, fmt.Errorf("cannot get entry options for lobby")
	}

	// Get world configuration
	config, err := s.interviewRepo.GetConfigurationByWorldID(ctx, worldID)
	if err != nil {
		return nil, fmt.Errorf("failed to get world config: %w", err)
	}

	// If config is nil (legacy world or missing config), use default
	if config == nil {
		config = &interview.WorldConfiguration{
			WorldName:       "Unknown World",
			Theme:           "Generic",
			SentientSpecies: []string{"Human"},
		}
	}

	// Generate random NPCs
	npcs := s.generateRandomNPCs(config, 5)

	return &EntryOptions{
		CanEnterAsWatcher: true,
		AvailableNPCs:     npcs,
		CanCreateCustom:   true,
	}, nil
}

func (s *Service) generateRandomNPCs(config *interview.WorldConfiguration, count int) []NPCPreview {
	npcs := make([]NPCPreview, count)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	speciesList := config.SentientSpecies
	if len(speciesList) == 0 {
		speciesList = []string{"Human"}
	}

	for i := 0; i < count; i++ {
		species := speciesList[r.Intn(len(speciesList))]
		dna := s.generateRandomDNA()

		// Generate appearance
		age := 20 + r.Intn(30)
		lifespan := 80 // Default
		appDesc := appearance.GenerateAppearance(dna, age, lifespan, species)

		// Simple name generation (placeholder)
		name := fmt.Sprintf("NPC-%d", r.Intn(1000))

		// Encode DNA
		dnaBytes, _ := json.Marshal(dna)

		npcs[i] = NPCPreview{
			ID:          uuid.New().String(),
			Name:        name,
			Species:     species,
			Description: appDesc.FullDescription,
			Occupation:  "Commoner", // Placeholder
			DNA:         string(dnaBytes),
		}
	}

	return npcs
}

func (s *Service) generateRandomDNA() genetics.DNA {
	genes := make(map[string]genetics.Gene)

	// List of genes to generate
	geneNames := []string{
		genetics.GeneHeight, genetics.GeneBuild, genetics.GeneMuscle,
		genetics.GeneHair, genetics.GenePigment,
		genetics.GeneEye, genetics.GeneMelanin,
		genetics.GeneStrength, genetics.GeneReflex, genetics.GeneStamina, genetics.GeneHealth,
		genetics.GeneCognition, genetics.GenePerception,
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, name := range geneNames {
		// Random alleles
		a1 := "A"
		if r.Intn(2) == 0 {
			a1 = "a"
		}
		a2 := "A"
		if r.Intn(2) == 0 {
			a2 = "a"
		}

		genes[name] = genetics.Gene{
			TraitName:   name,
			Allele1:     a1,
			Allele2:     a2,
			IsDominant1: a1 == "A",
			IsDominant2: a2 == "A",
		}
	}

	return genetics.DNA{Genes: genes}
}
