package interview

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestExtractConfiguration_Success(t *testing.T) {
	playerID := uuid.New()
	interviewID := uuid.New()

	// Mock LLM that returns valid JSON
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return `{
				"theme": "high fantasy",
				"tone": "epic and hopeful",
				"inspirations": ["Lord of the Rings", "Wheel of Time"],
				"uniqueAspect": "magic is dying out",
				"techLevel": "medieval",
				"magicLevel": "rare",
				"planetSize": "Earth-sized",
				"climateRange": "varied, mostly temperate",
				"landWaterRatio": "60% land, 40% water",
				"sentientSpecies": ["humans", "elves", "dwarves"],
				"politicalStructure": "feudal kingdoms",
				"culturalValues": ["honor", "tradition", "courage"],
				"economicSystem": "barter and coin",
				"religions": ["multiple pantheons"],
				"conflicts": ["kingdoms at war", "magic fading"]
			}`, nil
		},
	}

	service := NewExtractionService(mockLLM)

	// Create a completed interview session
	session := &InterviewSession{
		ID:       interviewID,
		PlayerID: playerID,
		State: InterviewState{
			CurrentCategory:   CategoryCulture,
			CurrentTopicIndex: len(AllTopics),
			Answers: map[string]string{
				"World Type":   "High fantasy",
				"Tone":         "Epic and hopeful",
				"Inspirations": "Lord of the Rings, Wheel of Time",
				"Species":      "Humans, elves, dwarves",
			},
			IsComplete: true,
		},
		History:   []ConversationTurn{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	config, err := service.ExtractConfiguration(session, playerID)
	if err != nil {
		t.Fatalf("Failed to extract configuration: %v", err)
	}

	if config == nil {
		t.Fatal("Expected configuration, got nil")
	}

	// Verify extracted fields
	if config.Theme != "high fantasy" {
		t.Errorf("Expected theme 'high fantasy', got '%s'", config.Theme)
	}

	if config.Tone != "epic and hopeful" {
		t.Errorf("Expected tone 'epic and hopeful', got '%s'", config.Tone)
	}

	if config.TechLevel != "medieval" {
		t.Errorf("Expected tech level 'medieval', got '%s'", config.TechLevel)
	}

	if config.MagicLevel != "rare" {
		t.Errorf("Expected magic level 'rare', got '%s'", config.MagicLevel)
	}

	if len(config.SentientSpecies) != 3 {
		t.Errorf("Expected 3 sentient species, got %d", len(config.SentientSpecies))
	}

	if config.InterviewID != interviewID {
		t.Error("Expected InterviewID to be set")
	}

	if config.CreatedBy != playerID {
		t.Error("Expected CreatedBy to be set")
	}
}

func TestExtractConfiguration_ParsesTheme(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return `{
				"theme": "cyberpunk",
				"tone": "grim and dark",
				"inspirations": ["Blade Runner", "Neuromancer"],
				"uniqueAspect": "corporations rule everything",
				"conflicts": ["class warfare", "AI uprising"],
				"techLevel": "futuristic",
				"magicLevel": "none",
				"planetSize": "Earth-sized",
				"sentientSpecies": ["humans", "androids"]
			}`, nil
		},
	}

	service := NewExtractionService(mockLLM)
	session := createMockSession()

	config, err := service.ExtractConfiguration(session, uuid.New())
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	if config.Theme != "cyberpunk" {
		t.Errorf("Expected theme 'cyberpunk', got '%s'", config.Theme)
	}

	if config.Tone != "grim and dark" {
		t.Errorf("Expected tone 'grim and dark', got '%s'", config.Tone)
	}

	if len(config.Inspirations) != 2 {
		t.Errorf("Expected 2 inspirations, got %d", len(config.Inspirations))
	}

	if len(config.MajorConflicts) != 2 {
		t.Errorf("Expected 2 conflicts, got %d", len(config.MajorConflicts))
	}
}

func TestExtractConfiguration_ParsesTechLevel(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return `{
				"theme": "steampunk",
				"techLevel": "industrial",
				"magicLevel": "rare",
				"advancedTech": "steam-powered automata",
				"magicImpact": "magic used for industrial processes",
				"planetSize": "Earth-sized",
				"sentientSpecies": ["humans"]
			}`, nil
		},
	}

	service := NewExtractionService(mockLLM)
	session := createMockSession()

	config, err := service.ExtractConfiguration(session, uuid.New())
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	if config.TechLevel != "industrial" {
		t.Errorf("Expected tech level 'industrial', got '%s'", config.TechLevel)
	}

	if config.MagicLevel != "rare" {
		t.Errorf("Expected magic level 'rare', got '%s'", config.MagicLevel)
	}

	if config.AdvancedTech != "steam-powered automata" {
		t.Errorf("Expected advanced tech, got '%s'", config.AdvancedTech)
	}
}

func TestExtractConfiguration_ParsesGeography(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return `{
				"theme": "ocean world",
				"techLevel": "medieval",
				"planetSize": "super-Earth",
				"climateRange": "tropical",
				"landWaterRatio": "10% land, 90% water",
				"uniqueFeatures": ["floating islands", "underwater cities"],
				"extremeEnvironments": ["deep ocean trenches", "permanent storms"],
				"sentientSpecies": ["merfolk", "humans"]
			}`, nil
		},
	}

	service := NewExtractionService(mockLLM)
	session := createMockSession()

	config, err := service.ExtractConfiguration(session, uuid.New())
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	if config.PlanetSize != "super-Earth" {
		t.Errorf("Expected planet size 'super-Earth', got '%s'", config.PlanetSize)
	}

	if config.ClimateRange != "tropical" {
		t.Errorf("Expected climate range 'tropical', got '%s'", config.ClimateRange)
	}

	if len(config.UniqueFeatures) != 2 {
		t.Errorf("Expected 2 unique features, got %d", len(config.UniqueFeatures))
	}

	if len(config.ExtremeEnvironments) != 2 {
		t.Errorf("Expected 2 extreme environments, got %d", len(config.ExtremeEnvironments))
	}
}

func TestExtractConfiguration_ParsesCulture(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return `{
				"theme": "tribal",
				"techLevel": "stone_age",
				"planetSize": "Earth-sized",
				"sentientSpecies": ["humans", "neanderthals", "spirits"],
				"politicalStructure": "tribal councils",
				"culturalValues": ["harmony with nature", "ancestor worship"],
				"economicSystem": "gift economy",
				"religions": ["animism", "shamanism"],
				"taboos": ["harming sacred groves", "speaking names of the dead"]
			}`, nil
		},
	}

	service := NewExtractionService(mockLLM)
	session := createMockSession()

	config, err := service.ExtractConfiguration(session, uuid.New())
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	if len(config.SentientSpecies) != 3 {
		t.Errorf("Expected 3 sentient species, got %d", len(config.SentientSpecies))
	}

	if config.PoliticalStructure != "tribal councils" {
		t.Errorf("Expected political structure 'tribal councils', got '%s'", config.PoliticalStructure)
	}

	if len(config.CulturalValues) != 2 {
		t.Errorf("Expected 2 cultural values, got %d", len(config.CulturalValues))
	}

	if len(config.Religions) != 2 {
		t.Errorf("Expected 2 religions, got %d", len(config.Religions))
	}

	if len(config.Taboos) != 2 {
		t.Errorf("Expected 2 taboos, got %d", len(config.Taboos))
	}
}

func TestParseExtractionResponse_ValidJSON(t *testing.T) {
	validJSON := `{
		"theme": "fantasy",
		"techLevel": "medieval",
		"planetSize": "Earth-sized",
		"sentientSpecies": ["humans", "elves"]
	}`

	config, err := parseExtractionResponse(validJSON)
	if err != nil {
		t.Fatalf("Failed to parse valid JSON: %v", err)
	}

	if config.Theme != "fantasy" {
		t.Errorf("Expected theme 'fantasy', got '%s'", config.Theme)
	}
}

func TestParseExtractionResponse_InvalidJSON(t *testing.T) {
	invalidJSON := `{
		"theme": "fantasy",
		"techLevel": "medieval",
		"planetSize": "Earth-sized"
		// Missing closing brace and comma errors
	`

	_, err := parseExtractionResponse(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParseExtractionResponse_MalformedJSON(t *testing.T) {
	malformed := "This is not JSON at all!"

	_, err := parseExtractionResponse(malformed)
	if err == nil {
		t.Error("Expected error for malformed JSON, got nil")
	}
}

func TestDeriveGenerationParameters(t *testing.T) {
	config := &WorldConfiguration{
		Theme:          "fantasy",
		ClimateRange:   "varied",
		LandWaterRatio: "60% land, 40% water",
	}

	err := deriveGenerationParameters(config)
	if err != nil {
		t.Fatalf("Failed to derive generation parameters: %v", err)
	}

	// Check that parameters were populated
	if config.BiomeWeights == nil {
		t.Error("Expected BiomeWeights to be populated")
	}

	if config.ResourceDistribution == nil {
		t.Error("Expected ResourceDistribution to be populated")
	}

	if config.SpeciesStartAttributes == nil {
		t.Error("Expected SpeciesStartAttributes to be populated")
	}
}

// Helper function to create a mock session for testing
func createMockSession() *InterviewSession {
	return &InterviewSession{
		ID:       uuid.New(),
		PlayerID: uuid.New(),
		State: InterviewState{
			CurrentCategory:   CategoryCulture,
			CurrentTopicIndex: len(AllTopics),
			Answers: map[string]string{
				"World Type": "Fantasy",
				"Species":    "Humans, elves",
			},
			IsComplete: true,
		},
		History:   []ConversationTurn{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Helper to verify JSON marshaling
func TestWorldConfiguration_JSONMarshaling(t *testing.T) {
	config := &WorldConfiguration{
		ID:              uuid.New(),
		InterviewID:     uuid.New(),
		CreatedBy:       uuid.New(),
		Theme:           "fantasy",
		TechLevel:       "medieval",
		PlanetSize:      "Earth-sized",
		SentientSpecies: []string{"humans", "elves"},
		CreatedAt:       time.Now(),
	}

	// Marshal to JSON
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal back
	var decoded WorldConfiguration
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if decoded.Theme != config.Theme {
		t.Errorf("Theme mismatch after marshal/unmarshal")
	}
}
