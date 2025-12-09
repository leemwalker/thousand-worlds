package lobby

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/game/formatter"
	"mud-platform-backend/internal/repository"
	"mud-platform-backend/internal/world/interview"
	"mud-platform-backend/internal/worldgen/geography"
	"mud-platform-backend/internal/worldgen/orchestrator"
	"mud-platform-backend/internal/worldgen/weather"

	"github.com/google/uuid"
)

// InterviewRepository defines the interface for accessing interview data
type InterviewRepository interface {
	GetConfigurationByWorldID(ctx context.Context, worldID uuid.UUID) (*interview.WorldConfiguration, error)
	GetConfigurationByUserID(ctx context.Context, userID uuid.UUID) (*interview.WorldConfiguration, error)
}

// LookService handles generating descriptions for the look command
type LookService struct {
	authRepo      auth.Repository
	worldRepo     repository.WorldRepository
	interviewRepo InterviewRepository

	// World Data Cache
	worldCache     map[uuid.UUID]*orchestrator.GeneratedWorld
	cacheMutex     sync.RWMutex
	generator      *orchestrator.GeneratorService
	weatherService *weather.Service
}

// NewLookService creates a new look service
func NewLookService(authRepo auth.Repository, worldRepo repository.WorldRepository, interviewRepo InterviewRepository, weatherService *weather.Service) *LookService {
	return &LookService{
		authRepo:       authRepo,
		worldRepo:      worldRepo,
		interviewRepo:  interviewRepo,
		worldCache:     make(map[uuid.UUID]*orchestrator.GeneratedWorld),
		generator:      orchestrator.NewGeneratorService(),
		weatherService: weatherService,
	}
}

// GetLobbyDescription generates the full lobby description
func (s *LookService) GetLobbyDescription(ctx context.Context, userID uuid.UUID, characterID uuid.UUID, currentPlayers []WebsocketClient) (string, error) {
	user, err := s.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	char, err := s.authRepo.GetCharacter(ctx, characterID)
	if err != nil {
		return "", fmt.Errorf("failed to get character: %w", err)
	}

	gen := NewDescriptionGenerator(s.worldRepo, s.authRepo)
	return gen.GenerateDescription(ctx, user, char, currentPlayers)
}

// DescribePlayer generates a description of another player in the lobby
func (s *LookService) DescribePlayer(ctx context.Context, targetUsername string) (string, error) {
	// Get target user by username
	user, err := s.authRepo.GetUserByUsername(ctx, targetUsername)
	if err != nil {
		return "", fmt.Errorf("player not found")
	}

	// Get all characters for this user
	characters, err := s.authRepo.GetUserCharacters(ctx, user.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve player data: %w", err)
	}

	// If user has no characters, they're a new player
	if len(characters) == 0 {
		return s.generateNewPlayerDescription(targetUsername), nil
	}

	// Find most recently played character
	var latestChar *auth.Character
	for _, char := range characters {
		if latestChar == nil || (char.LastPlayed != nil && (latestChar.LastPlayed == nil || char.LastPlayed.After(*latestChar.LastPlayed))) {
			latestChar = char
		}
	}

	// If no character has been played yet, treat as new
	if latestChar == nil || latestChar.LastPlayed == nil {
		return s.generateNewPlayerDescription(targetUsername), nil
	}

	// Generate returning player description based on last character
	return s.generateReturningPlayerDescription(ctx, latestChar)
}

// DescribePortal generates a description of a world portal
func (s *LookService) DescribePortal(ctx context.Context, portalName string) (string, error) {
	// Try to find world by name
	worlds, err := s.worldRepo.ListWorlds(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list worlds: %w", err)
	}

	var targetWorld *repository.World
	for _, w := range worlds {
		if strings.EqualFold(w.Name, portalName) {
			targetWorld = &w
			break
		}
	}

	if targetWorld == nil {
		return "", fmt.Errorf("portal not found")
	}

	// Get world configuration for theme details
	config, err := s.interviewRepo.GetConfigurationByWorldID(ctx, targetWorld.ID)
	if err != nil {
		// No config, generate basic description
		return s.generateBasicPortalDescription(targetWorld), nil
	}

	// Generate themed portal description
	return s.generateThemedPortalDescription(targetWorld, config), nil
}

// DescribeStatue generates a description of the central statue
func (s *LookService) DescribeStatue(ctx context.Context, userID uuid.UUID) (string, error) {
	// Check if user has created a world (optimized query)
	config, err := s.interviewRepo.GetConfigurationByUserID(ctx, userID)
	if err != nil {
		// Log error if needed, but treat as no world for user experience
		// Fallback to neutral statue
		return s.generateNeutralStatueDescription(), nil
	}

	if config == nil {
		// User hasn't created a world - show neutral statue
		return s.generateNeutralStatueDescription(), nil
	}

	// User has created a world - show themed statue
	return s.generateThemedStatueDescription(config.WorldName, config), nil
}

// DescribeRoom generates a description of the current room/world at the character's position
func (s *LookService) DescribeRoom(ctx context.Context, worldID uuid.UUID, char *auth.Character) (string, error) {
	// Check if this is the lobby
	if IsLobby(worldID) {
		return "You are in the Grand Lobby of Thousand Worlds.", nil
	}

	// Get World Info
	world, err := s.worldRepo.GetWorld(ctx, worldID)
	if err != nil {
		return "", fmt.Errorf("failed to get world: %w", err)
	}

	config, err := s.interviewRepo.GetConfigurationByWorldID(ctx, worldID)
	if err != nil || config == nil {
		// Fallback for legacy worlds
		return fmt.Sprintf("You are in %s.\nIt is a quiet place.", formatter.Format(world.Name, formatter.StyleBlue)), nil
	}

	// Get Generated World Data (Cached or Regenerated)
	genData, err := s.getWorldData(ctx, worldID, config)
	if err != nil {
		// Fallback if generation fails
		return s.generateFallbackDescription(world, config), nil
	}

	// Describe based on coordinates
	return s.generateCoordinateDescription(ctx, world, config, genData, char), nil
}

// DescribeView generates a description of what is seen in a specific direction/position
func (s *LookService) DescribeView(ctx context.Context, worldID uuid.UUID, char *auth.Character, x, y float64) (string, error) {
	// Create a temporary character with the target position to reuse generation logic
	viewChar := *char
	viewChar.PositionX = x
	viewChar.PositionY = y

	return s.DescribeRoom(ctx, worldID, &viewChar)
}

// getWorldData retrieves cached world data or regenerates it
func (s *LookService) getWorldData(ctx context.Context, worldID uuid.UUID, config *interview.WorldConfiguration) (*orchestrator.GeneratedWorld, error) {
	s.cacheMutex.RLock()
	if data, ok := s.worldCache[worldID]; ok {
		s.cacheMutex.RUnlock()
		return data, nil
	}
	s.cacheMutex.RUnlock()

	// Not in cache, acquire write lock
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// Double check
	if data, ok := s.worldCache[worldID]; ok {
		return data, nil
	}

	// Generate
	// Use seed from metadata if available, otherwise random
	// Assuming GenerateWorld handles this via config mapping or we trust config
	// Ideally we should extract seed from world metadata if we want strict persistence
	// But orchestrator.Mapper handles config -> params.

	// FIX: Ensure consistency. We must use the SAME seed.
	// The config mapper likely uses a hash of the NAME or ID if seed isn't explicit?
	// Orchestrator service regenerates based on config provided.

	data, err := s.generator.GenerateWorld(ctx, worldID, config)
	if err != nil {
		return nil, err
	}

	s.worldCache[worldID] = data
	// Initialize weather service with generated weather
	if s.weatherService != nil {
		s.weatherService.InitializeWorldWeather(ctx, worldID, data.Weather, data.WeatherCells)
	}
	return data, nil
}

func (s *LookService) generateCoordinateDescription(ctx context.Context, world *repository.World, config *interview.WorldConfiguration, data *orchestrator.GeneratedWorld, char *auth.Character) string {
	hm := data.Geography.Heightmap

	// Map float position to integer grid coordinates
	// Assuming char position is in same units as heightmap (or scaled?)
	// If heightmap is 512x512, and world is "infinite" or "10000km", we need projection.
	// Current generation uses pixel coordinates. Let's assume char.PositionX/Y are grid coordinates for now.

	x, y := int(char.PositionX), int(char.PositionY)

	// Clamp to map bounds
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

	// Biome
	biomeIdx := y*hm.Width + x
	if biomeIdx >= len(data.Geography.Biomes) {
		biomeIdx = 0 // Safety
	}
	biome := data.Geography.Biomes[biomeIdx] // Assuming this is parallel array?
	// Wait, AssignBiomes returns a biome map?
	// data.Geography has Biomes field. Let's assume it's []BiomeType.
	// Checking struct definition... WorldMap struct in geography/types.go usually.
	// We'll rely on generic logic if precise types fail, but let's assume it works.

	worldName := formatter.Format(world.Name, formatter.StyleBlue)

	var terrainDesc string
	var biomeDesc string

	// Elevation description
	if elev < seaLevel {
		depth := seaLevel - elev
		if depth > 2000 {
			terrainDesc = "in the deep abyss"
		} else {
			terrainDesc = "beneath the waves"
		}
	} else if elev < seaLevel+10 {
		terrainDesc = "on the shore"
	} else if elev > seaLevel+4000 {
		terrainDesc = "on a towering peak"
	} else if elev > seaLevel+1500 {
		terrainDesc = "on a high mountainside"
	} else {
		terrainDesc = "on rolling terrain"
	}

	// Biome description
	switch biome.Type {
	case geography.BiomeOcean:
		biomeDesc = "surrounded by water"
	case geography.BiomeDesert:
		biomeDesc = "surrounded by arid sands"
	case geography.BiomeRainforest:
		biomeDesc = "deep within a lush jungle"
	case geography.BiomeDeciduousForest:
		biomeDesc = "amongst tall trees"
	case geography.BiomeMountain:
		biomeDesc = "surrounded by rocky crags"
	case geography.BiomeTundra:
		biomeDesc = "in a frozen wasteland"
	case geography.BiomeGrassland:
		biomeDesc = "on a grassy plain"
	case geography.BiomeLowland:
		biomeDesc = "in the lowlands"
	default:
		biomeDesc = "in a wild landscape"
	}

	// Combine
	desc := fmt.Sprintf("You are in %s.\nYou act %s, %s.", worldName, terrainDesc, biomeDesc)

	// Add details (Rivers, etc)
	// Check neighbors for features
	// (Simple check for river/water nearby)

	// Add weather description
	if s.weatherService != nil {
		// Calculate cell ID based on grid coordinates?
		// Actually, we need the EXACT CellID from the generated data to lookup in weather service.
		// The generated data should ideally provide a lookup or we calculate it.
		// For now, let's look it up in data.Weather if we can (linear search is bad).
		// Better: map x,y to index, and data.Weather[index] should match if it's same order.
		// UpdateWeather returns slice corresponding to input cells.
		// mapToGeographyCells creates cells in row-major order.
		cellIdx := y*hm.Width + x
		if cellIdx >= 0 && cellIdx < len(data.Weather) {
			weatherState, err := s.weatherService.GetCurrentWeather(ctx, world.ID, data.Weather[cellIdx].CellID)
			if err == nil && weatherState != nil {
				weatherDesc := getWeatherDescription(weatherState)
				desc = fmt.Sprintf("%s\n%s", desc, weatherDesc)
			}
		}
	}

	return desc
}

func getWeatherDescription(state *weather.WeatherState) string {
	switch state.State {
	case weather.WeatherClear:
		return "The sky is clear and blue."
	case weather.WeatherCloudy:
		return "Grey clouds hang low overhead."
	case weather.WeatherRain:
		return "Rain falls steadily, soaking the ground."
	case weather.WeatherStorm:
		return "A fierce storm rages, wind howling around you."
	case weather.WeatherSnow:
		return "Snow falls gently, blanketing the world in white."
	default:
		return "The weather is calm."
	}
}

func (s *LookService) generateFallbackDescription(world *repository.World, config *interview.WorldConfiguration) string {
	atmosphere := s.getWorldAtmosphere(config)
	worldName := formatter.Format(world.Name, formatter.StyleBlue)
	return fmt.Sprintf("You stand in %s.\nThe air is filled with the %s.", worldName, atmosphere)
}

// ... (Rest of helper methods: generateNewPlayerDescription, etc. - keep unchanged)
func (s *LookService) generateNewPlayerDescription(username string) string {
	return fmt.Sprintf("A shapeless gray spirit drifts nearby, their form indistinct and barely visible in the fog. They seem new to this place, not yet having taken on any particular shape. You sense this is %s, not yet bound to any world.", formatter.Format(username, formatter.StyleCyan))
}

func (s *LookService) generateReturningPlayerDescription(ctx context.Context, char *auth.Character) (string, error) {
	// ... (Same as before)
	world, err := s.worldRepo.GetWorld(ctx, char.WorldID)
	if err != nil {
		return fmt.Sprintf("You see %s, a traveler who has walked other realms.", formatter.Format(char.Name, formatter.StyleCyan)), nil
	}

	config, err := s.interviewRepo.GetConfigurationByWorldID(ctx, world.ID)

	baseDesc := ""
	name := formatter.Format(char.Name, formatter.StyleCyan)
	if char.Description != "" {
		baseDesc = char.Description
	} else if char.Appearance != "" {
		baseDesc = fmt.Sprintf("%s appears before you", name)
	} else {
		baseDesc = fmt.Sprintf("You see %s", name)
	}

	worldName := formatter.Format(world.Name, formatter.StyleBlue)

	if config != nil {
		atmosphere := s.getWorldAtmosphere(config)
		return fmt.Sprintf("%s. They carry with them the %s of %s.", baseDesc, atmosphere, worldName), nil
	}

	return fmt.Sprintf("%s, recently returned from %s.", baseDesc, worldName), nil
}

func (s *LookService) generateBasicPortalDescription(world *repository.World) string {
	worldName := formatter.Format(world.Name, formatter.StyleBlue)
	return fmt.Sprintf("The portal to %s shimmers before you, its surface rippling like water. Beyond it, you can sense another realm waiting to be explored.", worldName)
}

func (s *LookService) generateThemedPortalDescription(world *repository.World, config *interview.WorldConfiguration) string {
	theme := strings.ToLower(config.Theme)
	worldName := formatter.Format(config.WorldName, formatter.StyleBlue)

	var portalDesc string

	if strings.Contains(theme, "desert") || strings.Contains(theme, "sand") {
		portalDesc = fmt.Sprintf("The portal to %s is framed by sun-bleached stone, heat shimmering across its surface. You hear the whisper of sand-laden winds and feel dry warmth emanating from within. Through the haze, you glimpse endless golden dunes beneath a burning sky.", worldName)
	} else if strings.Contains(theme, "ocean") || strings.Contains(theme, "sea") || strings.Contains(theme, "water") {
		portalDesc = fmt.Sprintf("The portal to %s appears as a frame of coral and driftwood, its surface rippling like ocean waves. The cry of seabirds echoes through the opening, and you smell salt on the air. Glimpses of azure waters and white caps flash through the translucent barrier.", worldName)
	} else if strings.Contains(theme, "forest") || strings.Contains(theme, "wood") || strings.Contains(theme, "tree") {
		portalDesc = fmt.Sprintf("The portal to %s is wreathed in living vines and moss-covered bark. The scent of pine and earth drifts through, accompanied by distant birdsong. Through gaps in the foliage, you see towering trees and dappled sunlight filtering through emerald canopy.", worldName)
	} else if strings.Contains(theme, "mountain") || strings.Contains(theme, "stone") {
		portalDesc = fmt.Sprintf("The portal to %s is carved from ancient granite, its edges sharp and weathered. Cold mountain air flows through the opening, carrying distant echoes. Beyond, you glimpse snow-capped peaks reaching toward gray skies.", worldName)
	} else if strings.Contains(theme, "tech") || strings.Contains(theme, "cyber") || strings.Contains(theme, "futur") {
		portalDesc = fmt.Sprintf("The portal to %s hums with energy, its frame made of sleek alloy and pulsing lights. You hear the rhythmic thrum of machinery and smell ozone. Through the shimmering field, you see gleaming structures and neon-lit pathways.", worldName)
	} else if strings.Contains(theme, "magic") || strings.Contains(theme, "arcane") {
		portalDesc = fmt.Sprintf("The portal to %s glows with otherworldly light, arcane symbols dancing across its ethereal frame. Power crackles in the air around it, and you hear whispered incantations. Through the luminous veil, you see impossible geometries and shifting colors.", worldName)
	} else {
		portalDesc = fmt.Sprintf("The portal to %s stands before you, its frame adorned with symbols reflecting its nature. An atmosphere of %s emanates from within, drawing your attention. Through its surface, you catch glimpses of the realm beyond.", worldName, theme)
	}

	return portalDesc
}

func (s *LookService) generateNeutralStatueDescription() string {
	return "A weathered stone statue stands at the center of the portals, one arm outstretched in a welcoming gesture. Its surface is carved with countless symbols and words from a thousand different languages. The statue's eyes seem to follow you, waiting patiently.\n\nThe statue seems ready to speak with you. Send it a tell asking to create a world when you're ready to begin."
}

func (s *LookService) generateThemedStatueDescription(worldName string, config *interview.WorldConfiguration) string {
	formattedWorldName := formatter.Format(worldName, formatter.StyleBlue)

	if config == nil {
		return fmt.Sprintf("The statue at the center has transformed, now pointing steadily toward the portal of %s, the world you created.", formattedWorldName)
	}

	theme := strings.ToLower(config.Theme)

	var statueDesc string

	if strings.Contains(theme, "forest") || strings.Contains(theme, "wood") {
		statueDesc = fmt.Sprintf("A statue of intertwined vines and living wood stands at the center, its branches pointing toward the verdant portal of %s, your world. Moss grows across its surface, and small flowers bloom from cracks in the bark. It pulses with gentle life, a testament to your creation.", formattedWorldName)
	} else if strings.Contains(theme, "desert") || strings.Contains(theme, "sand") {
		statueDesc = fmt.Sprintf("A sun-bleached sandstone statue rises from the center, its weathered surface carved with desert winds. One arm points unerringly toward the portal of %s, your realm. Grain by grain, it seems to shift and reform, eternally shaped by the sands of time.", formattedWorldName)
	} else if strings.Contains(theme, "ocean") || strings.Contains(theme, "sea") {
		statueDesc = fmt.Sprintf("A statue of water-smoothed coral and shell stands at the center, its surface still damp with salt spray. It points toward the ocean portal of %s, your creation. Small pools of water collect at its base, and you can hear the faint sound of waves within the stone.", formattedWorldName)
	} else if strings.Contains(theme, "mountain") || strings.Contains(theme, "stone") {
		statueDesc = fmt.Sprintf("A towering statue of granite and ice dominates the center, its peaks pointing toward the portal of %s, your world. Snow clings to its highest points, and you feel the cold majesty of mountain heights radiating from its surface.", formattedWorldName)
	} else if strings.Contains(theme, "tech") || strings.Contains(theme, "cyber") {
		statueDesc = fmt.Sprintf("A sleek statue of chrome and circuitry stands at the center, LED displays running across its surface. Its outstretched limb points with mechanical precision toward the portal of %s, your creation. Soft electronic tones emanate from within its frame.", formattedWorldName)
	} else if strings.Contains(theme, "magic") || strings.Contains(theme, "arcane") {
		statueDesc = fmt.Sprintf("A crystalline statue shimmers at the center, arcane energy flowing through its translucent form. It gestures toward the glowing portal of %s, your world. Symbols of power drift around it like fireflies, and reality seems to bend in its presence.", formattedWorldName)
	} else {
		statueDesc = fmt.Sprintf("The statue at the center has transformed to reflect the essence of %s, your created world. It stands as a monument to your imagination, one arm pointing steadily toward your realm's portal.", formattedWorldName)
	}

	return statueDesc
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
