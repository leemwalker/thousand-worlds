package area

import (
	"context"
	"fmt"
	"strings"

	"tw-backend/internal/ai/ollama"
	"tw-backend/internal/npc/memory" // For Location struct
)

// LLMClient defines interface for generating text
type LLMClient interface {
	Generate(prompt string) (string, error)
}

// AreaDescriptionService generates descriptions for locations
type AreaDescriptionService struct {
	client LLMClient
	cache  *AreaCache
}

// NewAreaDescriptionService creates a new service
func NewAreaDescriptionService(client LLMClient, cache *AreaCache) *AreaDescriptionService {
	return &AreaDescriptionService{
		client: client,
		cache:  cache,
	}
}

// ContextData holds all necessary info for generation
type ContextData struct {
	Location    memory.Location
	WorldName   string
	Biome       string
	Terrain     string
	Weather     string
	Temperature float64
	TimeOfDay   string
	Season      string
	Entities    []string
	Structures  []string
	Perception  int
}

// GenerateAreaDescription generates or retrieves a description
func (s *AreaDescriptionService) GenerateAreaDescription(ctx context.Context, data ContextData) (string, error) {
	// 1. Check Cache
	key := GenerateKey(data.Location.WorldID, data.Location.X, data.Location.Y, data.Location.Z,
		data.Weather, data.TimeOfDay, data.Season, data.Perception)

	if desc, ok := s.cache.Get(key); ok {
		return desc, nil
	}

	// 2. Build Prompt
	prompt := s.buildPrompt(data)

	// 3. Call Ollama
	// Note: In a real scenario, we might want to use the Queue here too, but for now direct call or separate queue
	// The plan mentions priority queue, so eventually this should go through that.
	// For this step, we'll use the client directly but keep in mind integration.
	resp, err := s.client.Generate(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate description: %w", err)
	}

	// 4. Parse/Clean
	cleaned := ollama.SanitizeResponse(resp)

	// 5. Cache
	s.cache.Set(key, cleaned)

	return cleaned, nil
}

func (s *AreaDescriptionService) buildPrompt(data ContextData) string {
	var sb strings.Builder

	sb.WriteString("Generate a description of this location:\n\n")
	sb.WriteString(fmt.Sprintf("COORDINATES: %.1f, %.1f, %.1f in %s\n", data.Location.X, data.Location.Y, data.Location.Z, data.WorldName))
	sb.WriteString(fmt.Sprintf("BIOME: %s\n", data.Biome))
	sb.WriteString(fmt.Sprintf("TERRAIN: %s\n", data.Terrain))
	sb.WriteString(fmt.Sprintf("WEATHER: %s (%.1f°C)\n", data.Weather, data.Temperature))
	sb.WriteString(fmt.Sprintf("TIME: %s, %s\n\n", data.TimeOfDay, data.Season))

	sb.WriteString("NEARBY ENTITIES:\n")
	if len(data.Entities) > 0 {
		for _, e := range data.Entities {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	} else {
		sb.WriteString("None\n")
	}
	sb.WriteString("\n")

	sb.WriteString("NEARBY STRUCTURES:\n")
	if len(data.Structures) > 0 {
		for _, st := range data.Structures {
			sb.WriteString(fmt.Sprintf("- %s\n", st))
		}
	} else {
		sb.WriteString("None\n")
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("OBSERVER PERCEPTION: %d/100\n\n", data.Perception))

	sb.WriteString("Describe what the observer sees, hears, and smells. Adjust detail level based on perception skill:\n")
	sb.WriteString("- Low (0-25): Basic, vague description\n")
	sb.WriteString("- Medium (26-75): Standard detail\n")
	sb.WriteString("- High (76-100): Rich, nuanced detail\n\n")
	sb.WriteString("Keep description to 3-5 sentences.")

	return sb.String()
}
