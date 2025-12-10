package look

import (
	"context"
	"fmt"
	"strings"

	"tw-backend/internal/auth"
	"tw-backend/internal/game/services/entity"
	"tw-backend/internal/repository"
	"tw-backend/internal/world/interview"
	"tw-backend/internal/worldgen/orchestrator"
	"tw-backend/internal/worldgen/weather"

	"github.com/google/uuid"
)

// LookService definition
type LookService struct {
	worldRepo      repository.WorldRepository
	weatherService *weather.Service
	entityService  *entity.Service

	// We might need to keep the orchestrator/cache logic here or genericize it
	// For now, let's keep the cache logic as it was essential for description generation
	worldCache    map[uuid.UUID]*orchestrator.GeneratedWorld
	generator     *orchestrator.GeneratorService
	interviewRepo InterviewRepository
}

// InterviewRepository interface (same as before to decouple)
type InterviewRepository interface {
	GetConfigurationByWorldID(ctx context.Context, worldID uuid.UUID) (*interview.WorldConfiguration, error)
	GetConfigurationByUserID(ctx context.Context, userID uuid.UUID) (*interview.WorldConfiguration, error)
}

// NewLookService constructor
func NewLookService(
	worldRepo repository.WorldRepository,
	weatherService *weather.Service,
	entityService *entity.Service,
	interviewRepo InterviewRepository,
) *LookService {
	return &LookService{
		worldRepo:      worldRepo,
		weatherService: weatherService,
		entityService:  entityService,
		interviewRepo:  interviewRepo,
		worldCache:     make(map[uuid.UUID]*orchestrator.GeneratedWorld),
		generator:      orchestrator.NewGeneratorService(),
	}
}

// DescribeContext holds all data needed for a look operation
type DescribeContext struct {
	WorldID     uuid.UUID
	Character   *auth.Character
	Orientation string // "North", "South", etc.
	DetailLevel int    // 1=Basic, 2=Detailed, 3=Deep
}

// Describe generates the description
func (s *LookService) Describe(ctx context.Context, dc DescribeContext) (string, error) {
	// 1. Get Base Room Description (Terrain/Biome)
	baseDesc, err := s.generateBaseDescription(ctx, dc.WorldID, dc.Character)
	if err != nil {
		// Fallback
		baseDesc = "You are in a mysterious place. The mist conceals everything."
	}

	// 2. Get Orientation Description
	orientDesc := s.generateOrientationDescription(dc.Character, dc.Orientation)

	// 3. Get Environmental Details (Weather, Time)
	// We pass the generated data if available to be more precise
	envDesc := ""
	genData, ok := s.getCachedWorldData(dc.WorldID)
	if ok && genData != nil {
		envDesc = s.generateEnvironmentDescription(ctx, dc.WorldID, genData, dc.Character)
	}

	// 4. Get Entities (NPCs, Items)
	entityDesc := s.generateEntityDescription(ctx, dc.WorldID, dc.Character)

	// Combine
	fullDesc := baseDesc
	if orientDesc != "" {
		fullDesc += "\n" + orientDesc
	}
	if envDesc != "" {
		fullDesc += "\n" + envDesc
	}
	if entityDesc != "" {
		fullDesc += "\n\n" + entityDesc
	}

	return fullDesc, nil
}

// DescribeEntity generates a description for a specific target (self or entity)
func (s *LookService) DescribeEntity(ctx context.Context, char *auth.Character, targetName string) (string, error) {
	// 1. Check for Self
	targetLower := strings.ToLower(targetName)
	if targetLower == "self" || targetLower == "me" || targetLower == "myself" || strings.EqualFold(char.Name, targetName) {
		return s.describeSelf(char), nil
	}

	// 2. Check for Entities
	if s.entityService == nil {
		return "", fmt.Errorf("I don't see any '%s' here.", targetName)
	}

	// Search radius - let's agree on a reasonable visibility range, say 20 meters
	entities, err := s.entityService.GetEntitiesAt(ctx, char.WorldID, char.PositionX, char.PositionY, 20.0)
	if err != nil {
		return "", fmt.Errorf("failed to look for entities: %w", err)
	}

	for _, e := range entities {
		if strings.EqualFold(e.Name, targetName) {
			return s.describeEntityObject(e), nil
		}
	}

	return "", fmt.Errorf("I don't see any '%s' here.", targetName)
}

func (s *LookService) describeSelf(char *auth.Character) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s", char.Name))
	if char.Occupation != "" {
		sb.WriteString(fmt.Sprintf(", the %s", char.Occupation))
	}
	sb.WriteString(".\n")

	if char.Description != "" {
		sb.WriteString(char.Description + "\n")
	}

	if char.Appearance != "" {
		// Appearance is stored as JSON string but might be raw text in some cases or simple string in early dev
		// For now, let's just print it if it looks like a description, or ignore if strict JSON structure isn't parsed
		// Assuming for now it's a simple string or we print it as is if it's not complex JSON
		if !strings.HasPrefix(char.Appearance, "{") {
			sb.WriteString(char.Appearance + "\n")
		} else {
			// If it is JSON, we might want to parse it, but for now let's skip complex parsing
			// and assume if it's JSON it's primarily for frontend, unless we want to flatten it here.
			// Let's print a generic message if we can't parse it easily or it's complex
		}
	}

	// Add stats or other info if desired later
	return strings.TrimSpace(sb.String())
}

func (s *LookService) describeEntityObject(e *entity.Entity) string {
	desc := e.Description
	if desc == "" {
		desc = fmt.Sprintf("You see a %s.", e.Name)
	}
	return desc
}

// generateBaseDescription uses the world gen logic
func (s *LookService) generateBaseDescription(ctx context.Context, worldID uuid.UUID, char *auth.Character) (string, error) {
	// 1. Get World Info
	world, err := s.worldRepo.GetWorld(ctx, worldID)
	if err != nil {
		return "", fmt.Errorf("failed to get world: %w", err)
	}

	// 2. Get Config
	config, err := s.interviewRepo.GetConfigurationByWorldID(ctx, worldID)
	if err != nil || config == nil {
		// Fallback for legacy worlds or unconfigured
		return fmt.Sprintf("You are in %s.", world.Name), nil
	}

	// 3. Get Generated World Data (Cached or Regenerated)
	genData, err := s.getWorldData(ctx, worldID, config)
	if err != nil {
		return "", err
	}

	// 4. Generate Description based on coordinates
	return s.generateCoordinateDescription(ctx, world, config, genData, char), nil
}

func (s *LookService) generateOrientationDescription(char *auth.Character, orientation string) string {
	if orientation == "" {
		return ""
	}
	return fmt.Sprintf("You are facing %s.", orientation)
}

func (s *LookService) generateEnvironmentDescription(ctx context.Context, worldID uuid.UUID, data *orchestrator.GeneratedWorld, char *auth.Character) string {
	if s.weatherService == nil {
		return ""
	}

	// Calculate cell index
	hm := data.Geography.Heightmap
	x, y := int(char.PositionX), int(char.PositionY)

	// Bounds check
	if x < 0 {
		x = 0
	}
	if x >= hm.Width {
		x = hm.Width - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= hm.Height {
		y = hm.Height - 1
	}

	cellIdx := y*hm.Width + x
	if cellIdx >= 0 && cellIdx < len(data.Weather) {
		weatherState, err := s.weatherService.GetCurrentWeather(ctx, worldID, data.Weather[cellIdx].CellID)
		if err == nil && weatherState != nil {
			return s.getWeatherDescription(weatherState)
		}
	}
	return ""
}

func (s *LookService) getWeatherDescription(state *weather.WeatherState) string {
	switch state.State {
	case weather.WeatherClear:
		return "The sky is clear and blue."
	case weather.WeatherCloudy:
		return "Grey clouds hang low overhead."
	case weather.WeatherRain:
		return "Rain falls steadily."
	case weather.WeatherStorm:
		return "A fierce storm rages."
	case weather.WeatherSnow:
		return "Snow falls gently."
	default:
		return "The weather is calm."
	}
}

func (s *LookService) generateEntityDescription(ctx context.Context, worldID uuid.UUID, char *auth.Character) string {
	if s.entityService == nil {
		return ""
	}

	entities, err := s.entityService.GetEntitiesAt(ctx, worldID, char.PositionX, char.PositionY, 20.0)
	if err != nil || len(entities) == 0 {
		return ""
	}

	var descriptions []string
	for _, e := range entities {
		descriptions = append(descriptions, fmt.Sprintf("A %s is here.", e.Name))
	}

	return strings.Join(descriptions, "\n")
}

// Helper methods from old service

func (s *LookService) getWorldData(ctx context.Context, worldID uuid.UUID, config *interview.WorldConfiguration) (*orchestrator.GeneratedWorld, error) {
	// Check cache
	if data, ok := s.getCachedWorldData(worldID); ok {
		return data, nil
	}

	// Generate
	data, err := s.generator.GenerateWorld(ctx, worldID, config)
	if err != nil {
		return nil, err
	}

	s.worldCache[worldID] = data
	// Initialize weather service
	if s.weatherService != nil {
		s.weatherService.InitializeWorldWeather(ctx, worldID, data.Weather, data.WeatherCells)
	}
	return data, nil
}

func (s *LookService) getCachedWorldData(worldID uuid.UUID) (*orchestrator.GeneratedWorld, bool) {
	data, ok := s.worldCache[worldID]
	return data, ok
}

// DescribeView generates a description of what is seen in a specific direction/position
func (s *LookService) DescribeView(ctx context.Context, worldID uuid.UUID, char *auth.Character, x, y float64) (string, error) {
	// Create a temporary character with the target position to reuse generation logic
	viewChar := *char
	viewChar.PositionX = x
	viewChar.PositionY = y

	dc := DescribeContext{
		WorldID:     worldID,
		Character:   &viewChar,
		Orientation: "", // Looking AT location, not FROM it facing somewhere
		DetailLevel: 1,
	}

	return s.Describe(ctx, dc)
}

// generateCoordinateDescription (Ported)
func (s *LookService) generateCoordinateDescription(ctx context.Context, world *repository.World, config *interview.WorldConfiguration, data *orchestrator.GeneratedWorld, char *auth.Character) string {
	// ... (Simplification of previous logic for brevity, but retaining functionality)
	// We need to import internal/worldgen/geography
	// For now, let's assume we have access or mock it

	hm := data.Geography.Heightmap
	x, y := int(char.PositionX), int(char.PositionY)

	// Clamp
	if x < 0 {
		x = 0
	}
	if x >= hm.Width {
		x = hm.Width - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= hm.Height {
		y = hm.Height - 1
	}

	elev := hm.Get(x, y)
	seaLevel := data.Metadata.SeaLevel

	var terrainDesc string
	if elev < seaLevel {
		terrainDesc = "beneath the waves"
	} else if elev < seaLevel+10 {
		terrainDesc = "on the shore"
	} else if elev > seaLevel+1500 {
		terrainDesc = "on a high mountainside"
	} else {
		terrainDesc = "on rolling terrain"
	}

	return fmt.Sprintf("You are in %s.\nYou stand %s.", world.Name, terrainDesc)
}
