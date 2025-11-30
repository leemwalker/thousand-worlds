package interview

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ExtractionService extracts structured configuration from interview
type ExtractionService struct {
	client LLMClient
}

// NewExtractionService creates a new extraction service
func NewExtractionService(client LLMClient) *ExtractionService {
	return &ExtractionService{
		client: client,
	}
}

// ExtractConfiguration uses LLM to parse conversation into WorldConfiguration
func (e *ExtractionService) ExtractConfiguration(session *InterviewSession, playerID uuid.UUID) (*WorldConfiguration, error) {
	// Build prompt for extraction
	prompt := buildExtractionPrompt(session)

	// Call LLM to extract structured data
	response, err := e.client.Generate(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate extraction: %w", err)
	}

	// Parse JSON response
	config, err := parseExtractionResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extraction response: %w", err)
	}

	// Set metadata
	config.ID = uuid.New()
	config.InterviewID = session.ID
	config.CreatedBy = playerID
	config.CreatedAt = time.Now()

	// Derive generation parameters
	if err := deriveGenerationParameters(config); err != nil {
		return nil, fmt.Errorf("failed to derive generation parameters: %w", err)
	}

	return config, nil
}

// buildExtractionPrompt creates prompt for structured data extraction
func buildExtractionPrompt(session *InterviewSession) string {
	// Build conversation history
	var conversationHistory strings.Builder
	conversationHistory.WriteString("Interview Conversation:\n\n")

	for topicName, answer := range session.State.Answers {
		conversationHistory.WriteString(fmt.Sprintf("Q: %s\nA: %s\n\n", topicName, answer))
	}

	template := `You are a data extraction assistant. Extract structured world parameters from this interview conversation.

%s

Return ONLY a valid JSON object with these exact fields (use null for missing data):
{
  "theme": "string - world type (e.g., 'high fantasy', 'sci-fi')",
  "tone": "string - overall tone (e.g., 'grim', 'hopeful')",
  "inspirations": ["array of inspiration strings"],
  "uniqueAspect": "string - what makes this world unique",
  "conflicts": ["array of major conflicts/tensions"],
  "techLevel": "string - one of: stone_age, medieval, renaissance, industrial, modern, futuristic, mixed",
  "magicLevel": "string - one of: none, rare, common, dominant (or null)",
  "advancedTech": "string - most advanced technology",
  "magicImpact": "string - how magic affects daily life",
  "planetSize": "string - planet size description",
  "climateRange": "string - climate description",
  "landWaterRatio": "string - land/water distribution",
  "uniqueFeatures": ["array of unique geographical features"],
  "extremeEnvironments": ["array of extreme environments"],
  "sentientSpecies": ["array of sentient species - REQUIRED, at least one"],
  "politicalStructure": "string - political system",
  "culturalValues": ["array of main cultural values"],
  "economicSystem": "string - economic system",
  "religions": ["array of religions/belief systems"],
  "taboos": ["array of taboos/forbidden things"]
}

CRITICAL: Return ONLY the JSON object, no explanatory text before or after.`

	return fmt.Sprintf(template, conversationHistory.String())
}

// parseExtractionResponse parses JSON response from LLM
func parseExtractionResponse(jsonStr string) (*WorldConfiguration, error) {
	// Clean up the response
	jsonStr = strings.TrimSpace(jsonStr)

	// Remove markdown code blocks if present
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimPrefix(jsonStr, "```")
	jsonStr = strings.TrimSuffix(jsonStr, "```")
	jsonStr = strings.TrimSpace(jsonStr)

	// Parse JSON into a temporary structure
	var raw struct {
		Theme               string   `json:"theme"`
		Tone                string   `json:"tone"`
		Inspirations        []string `json:"inspirations"`
		UniqueAspect        string   `json:"uniqueAspect"`
		Conflicts           []string `json:"conflicts"`
		TechLevel           string   `json:"techLevel"`
		MagicLevel          string   `json:"magicLevel"`
		AdvancedTech        string   `json:"advancedTech"`
		MagicImpact         string   `json:"magicImpact"`
		PlanetSize          string   `json:"planetSize"`
		ClimateRange        string   `json:"climateRange"`
		LandWaterRatio      string   `json:"landWaterRatio"`
		UniqueFeatures      []string `json:"uniqueFeatures"`
		ExtremeEnvironments []string `json:"extremeEnvironments"`
		SentientSpecies     []string `json:"sentientSpecies"`
		PoliticalStructure  string   `json:"politicalStructure"`
		CulturalValues      []string `json:"culturalValues"`
		EconomicSystem      string   `json:"economicSystem"`
		Religions           []string `json:"religions"`
		Taboos              []string `json:"taboos"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Map to WorldConfiguration
	config := &WorldConfiguration{
		Theme:               raw.Theme,
		Tone:                raw.Tone,
		Inspirations:        raw.Inspirations,
		UniqueAspect:        raw.UniqueAspect,
		MajorConflicts:      raw.Conflicts,
		TechLevel:           raw.TechLevel,
		MagicLevel:          raw.MagicLevel,
		AdvancedTech:        raw.AdvancedTech,
		MagicImpact:         raw.MagicImpact,
		PlanetSize:          raw.PlanetSize,
		ClimateRange:        raw.ClimateRange,
		LandWaterRatio:      raw.LandWaterRatio,
		UniqueFeatures:      raw.UniqueFeatures,
		ExtremeEnvironments: raw.ExtremeEnvironments,
		SentientSpecies:     raw.SentientSpecies,
		PoliticalStructure:  raw.PoliticalStructure,
		CulturalValues:      raw.CulturalValues,
		EconomicSystem:      raw.EconomicSystem,
		Religions:           raw.Religions,
		Taboos:              raw.Taboos,
	}

	return config, nil
}

// deriveGenerationParameters calculates biome weights, resource distribution from config
func deriveGenerationParameters(config *WorldConfiguration) error {
	// Initialize maps
	config.BiomeWeights = make(map[string]float64)
	config.ResourceDistribution = make(map[string]float64)
	config.SpeciesStartAttributes = make(map[string]interface{})

	// Derive biome weights from climate and geography
	if config.ClimateRange != "" {
		climate := strings.ToLower(config.ClimateRange)

		// Simple biome weight derivation based on keywords
		if strings.Contains(climate, "tropical") {
			config.BiomeWeights["tropical_rainforest"] = 0.4
			config.BiomeWeights["tropical_savanna"] = 0.3
			config.BiomeWeights["jungle"] = 0.3
		} else if strings.Contains(climate, "frozen") || strings.Contains(climate, "arctic") {
			config.BiomeWeights["tundra"] = 0.5
			config.BiomeWeights["ice_sheet"] = 0.3
			config.BiomeWeights["taiga"] = 0.2
		} else if strings.Contains(climate, "desert") {
			config.BiomeWeights["desert"] = 0.6
			config.BiomeWeights["arid_scrubland"] = 0.4
		} else if strings.Contains(climate, "temperate") || strings.Contains(climate, "varied") {
			// Balanced distribution for temperate/varied
			config.BiomeWeights["temperate_forest"] = 0.25
			config.BiomeWeights["grassland"] = 0.25
			config.BiomeWeights["temperate_rainforest"] = 0.15
			config.BiomeWeights["woodland"] = 0.15
			config.BiomeWeights["mediterranean"] = 0.1
			config.BiomeWeights["mountain"] = 0.1
		} else {
			// Default balanced distribution
			config.BiomeWeights["temperate_forest"] = 0.3
			config.BiomeWeights["grassland"] = 0.3
			config.BiomeWeights["mountain"] = 0.2
			config.BiomeWeights["desert"] = 0.2
		}
	}

	// Derive resource distribution from tech level
	if config.TechLevel != "" {
		switch config.TechLevel {
		case "stone_age":
			config.ResourceDistribution["stone"] = 0.5
			config.ResourceDistribution["wood"] = 0.3
			config.ResourceDistribution["food"] = 0.2
		case "medieval":
			config.ResourceDistribution["iron"] = 0.3
			config.ResourceDistribution["wood"] = 0.3
			config.ResourceDistribution["stone"] = 0.2
			config.ResourceDistribution["food"] = 0.2
		case "industrial", "modern":
			config.ResourceDistribution["coal"] = 0.25
			config.ResourceDistribution["iron"] = 0.25
			config.ResourceDistribution["oil"] = 0.2
			config.ResourceDistribution["copper"] = 0.15
			config.ResourceDistribution["food"] = 0.15
		case "futuristic":
			config.ResourceDistribution["energy_crystals"] = 0.3
			config.ResourceDistribution["rare_metals"] = 0.3
			config.ResourceDistribution["quantum_materials"] = 0.2
			config.ResourceDistribution["synthesized_food"] = 0.2
		default:
			// Mixed or unknown - balanced distribution
			config.ResourceDistribution["basic_materials"] = 0.4
			config.ResourceDistribution["food"] = 0.3
			config.ResourceDistribution["advanced_materials"] = 0.3
		}
	}

	// Derive species start attributes
	for _, species := range config.SentientSpecies {
		speciesLower := strings.ToLower(species)
		attrs := make(map[string]interface{})

		// Set default attributes based on common species archetypes
		if strings.Contains(speciesLower, "human") {
			attrs["adaptability"] = 0.8
			attrs["social"] = 0.7
			attrs["technology_affinity"] = 0.7
		} else if strings.Contains(speciesLower, "elf") || strings.Contains(speciesLower, "elves") {
			attrs["longevity"] = 0.9
			attrs["magic_affinity"] = 0.8
			attrs["agility"] = 0.8
		} else if strings.Contains(speciesLower, "dwarf") || strings.Contains(speciesLower, "dwarves") {
			attrs["strength"] = 0.8
			attrs["crafting"] = 0.9
			attrs["resilience"] = 0.8
		} else if strings.Contains(speciesLower, "orc") {
			attrs["strength"] = 0.9
			attrs["endurance"] = 0.8
			attrs["aggression"] = 0.7
		} else {
			// Unknown species - balanced attributes
			attrs["adaptability"] = 0.6
			attrs["intelligence"] = 0.6
			attrs["strength"] = 0.6
		}

		config.SpeciesStartAttributes[species] = attrs
	}

	return nil
}
