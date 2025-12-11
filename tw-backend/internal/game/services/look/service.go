package look

import (
	"context"
	"fmt"
	"strings"

	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/game/services/entity"
	"tw-backend/internal/repository"
	"tw-backend/internal/world/interview"
	"tw-backend/internal/worldentity"
	"tw-backend/internal/worldgen/orchestrator"
	"tw-backend/internal/worldgen/weather"

	"github.com/google/uuid"

	"tw-backend/internal/game/constants"
	"tw-backend/internal/game/formatter"
)

// LookService definition
type LookService struct {
	worldRepo          repository.WorldRepository
	weatherService     *weather.Service
	entityService      *entity.Service
	worldEntityService *worldentity.Service
	ecosystemService   *ecosystem.Service
	authRepo           auth.Repository

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
	authRepo auth.Repository,
	worldEntityService *worldentity.Service,
	ecosystemService *ecosystem.Service,
) *LookService {
	return &LookService{
		worldRepo:          worldRepo,
		weatherService:     weatherService,
		entityService:      entityService,
		worldEntityService: worldEntityService,
		ecosystemService:   ecosystemService,
		interviewRepo:      interviewRepo,
		authRepo:           authRepo,
		worldCache:         make(map[uuid.UUID]*orchestrator.GeneratedWorld),
		generator:          orchestrator.NewGeneratorService(),
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
	genData, ok := s.GetCachedWorldData(dc.WorldID)
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

	// 2. Check for Entities (in-memory)
	if s.entityService != nil {
		entities, err := s.entityService.GetEntitiesAt(ctx, char.WorldID, char.PositionX, char.PositionY, 20.0)
		if err == nil {
			for _, e := range entities {
				if strings.EqualFold(e.Name, targetName) {
					return s.describeEntityObject(e), nil
				}
			}
		}
	}

	// 3. Check for WorldEntity objects (database-persisted static objects)
	if s.worldEntityService != nil {
		worldEntity, err := s.worldEntityService.GetEntityByName(ctx, char.WorldID, targetName)
		if err == nil && worldEntity != nil {
			return s.describeWorldEntity(char, worldEntity), nil
		}
	}

	// 4. Check for Ecosystem Entities
	if s.ecosystemService != nil {
		// ecosystem entities don't have unique names yet, so we match by Species
		// Get nearby entities
		ecoEntities := s.ecosystemService.GetEntitiesAt(char.WorldID, char.PositionX, char.PositionY, 20.0)
		for _, e := range ecoEntities {
			// Check if target name matches species (e.g. "rabbit")
			if strings.EqualFold(string(e.Species), targetName) {
				return fmt.Sprintf("You see a %s.\nIt looks healthy and alert.", e.Species), nil
			}
		}
	}

	// 5. Check for Other Players
	// Look up user by username
	// Note: This matches any user in the DB, but we should strictly check if they are "here" (in the same world/pos).
	// However, without a "GetPlayersAt" service, we might just check generic user and see if they are in this world.
	if s.authRepo != nil {
		targetUser, err := s.authRepo.GetUserByUsername(ctx, targetName)
		if err == nil && targetUser != nil {
			// Check if they have a character in this world
			targetChar, err := s.authRepo.GetCharacterByUserAndWorld(ctx, targetUser.UserID, char.WorldID)
			if err == nil && targetChar != nil {
				// Check distance? For now assuming if they are in the world and you asked for them, you can see them or we describe them.
				// Ideally we verify distance < 20.0
				dx := targetChar.PositionX - char.PositionX
				dy := targetChar.PositionY - char.PositionY
				distSq := dx*dx + dy*dy
				if distSq <= 400 { // 20^2
					return s.describeCharacter(ctx, targetChar)
				}
			}
		}
	}

	return "", fmt.Errorf("I don't see any '%s' here.", targetName)
}

func (s *LookService) describeCharacter(ctx context.Context, char *auth.Character) (string, error) {
	world, err := s.worldRepo.GetWorld(ctx, char.WorldID)
	if err != nil {
		return fmt.Sprintf("You see %s, a traveler.", formatter.Format(char.Name, formatter.StyleCyan)), nil
	}

	_, err = s.interviewRepo.GetConfigurationByWorldID(ctx, world.ID)

	baseDesc := ""
	name := formatter.Format(char.Name, formatter.StyleCyan)
	if char.Description != "" {
		baseDesc = char.Description
	} else if char.Appearance != "" {
		baseDesc = fmt.Sprintf("%s appears before you", name)
	} else {
		baseDesc = fmt.Sprintf("You see %s", name)
	}

	// If in lobby, we might want to mention where they came from (LastWorldVisited)
	if constants.IsLobby(char.WorldID) {
		if char.LastWorldVisited != nil && *char.LastWorldVisited != uuid.Nil {
			lastWorld, err := s.worldRepo.GetWorld(ctx, *char.LastWorldVisited)
			if err == nil {
				lastWorldConfig, _ := s.interviewRepo.GetConfigurationByWorldID(ctx, *char.LastWorldVisited)
				worldName := formatter.Format(lastWorld.Name, formatter.StyleBlue)
				if lastWorldConfig != nil {
					atmosphere := s.getWorldAtmosphere(lastWorldConfig)
					return fmt.Sprintf("%s. They carry with them the %s of %s.", baseDesc, atmosphere, worldName), nil
				}
				return fmt.Sprintf("%s, recently returned from %s.", baseDesc, worldName), nil
			}
		}
	}

	return fmt.Sprintf("%s.", baseDesc), nil
}

func (s *LookService) getWorldAtmosphere(config *interview.WorldConfiguration) string {
	theme := strings.ToLower(config.Theme)

	if strings.Contains(theme, "desert") {
		return "scent of hot sand and dry winds"
	} else if strings.Contains(theme, "ocean") {
		return "salt spray and sound of waves"
	} else if strings.Contains(theme, "forest") {
		return "earthy scent and whisper of leaves"
	} else if strings.Contains(theme, "mountain") {
		return "crisp cold and echoing heights"
	} else if strings.Contains(theme, "tech") {
		return "hum of machinery and scent of ozone"
	} else if strings.Contains(theme, "magic") {
		return "crackle of arcane power"
	}

	return "essence"
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

// describeWorldEntity describes a database-persisted world entity (static objects like statues)
func (s *LookService) describeWorldEntity(char *auth.Character, e *worldentity.WorldEntity) string {
	// Calculate distance to entity
	dx := e.X - char.PositionX
	dy := e.Y - char.PositionY
	distSq := dx*dx + dy*dy

	// Format the name
	name := formatter.Format(e.Name, formatter.StyleYellow)

	// If very close (within 2 meters), show detailed description
	if distSq <= 4 && e.Details != "" {
		return fmt.Sprintf("%s\n\n%s", e.Description, e.Details)
	}

	// Otherwise just show the basic description
	if e.Description != "" {
		return e.Description
	}

	// Fallback if no description
	return fmt.Sprintf("You see %s here.", name)
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
	var descriptions []string

	if s.entityService != nil {
		entities, err := s.entityService.GetEntitiesAt(ctx, worldID, char.PositionX, char.PositionY, 20.0)
		if err == nil && len(entities) > 0 {
			for _, e := range entities {
				descriptions = append(descriptions, fmt.Sprintf("A %s is here.", e.Name))
			}
		}
	}

	if s.ecosystemService != nil {
		ecoEntities := s.ecosystemService.GetEntitiesAt(worldID, char.PositionX, char.PositionY, 20.0)
		for _, e := range ecoEntities {
			descriptions = append(descriptions, fmt.Sprintf("A %s is here.", e.Species))
		}
	}

	return strings.Join(descriptions, "\n")
}

// Helper methods from old service

func (s *LookService) getWorldData(ctx context.Context, worldID uuid.UUID, config *interview.WorldConfiguration) (*orchestrator.GeneratedWorld, error) {
	// Check cache
	if data, ok := s.GetCachedWorldData(worldID); ok {
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

// GetCachedWorldData returns cached generated world data if available
func (s *LookService) GetCachedWorldData(worldID uuid.UUID) (*orchestrator.GeneratedWorld, bool) {
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
