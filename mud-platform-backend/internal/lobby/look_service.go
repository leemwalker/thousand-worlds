package lobby

import (
	"context"
	"fmt"
	"strings"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/repository"
	"mud-platform-backend/internal/world/interview"

	"github.com/google/uuid"
)

// LookService handles generating descriptions for the look command
type LookService struct {
	authRepo      auth.Repository
	worldRepo     repository.WorldRepository
	interviewRepo *interview.Repository
}

// NewLookService creates a new look service
func NewLookService(authRepo auth.Repository, worldRepo repository.WorldRepository, interviewRepo *interview.Repository) *LookService {
	return &LookService{
		authRepo:      authRepo,
		worldRepo:     worldRepo,
		interviewRepo: interviewRepo,
	}
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
	config, err := s.interviewRepo.GetConfigurationByWorldID(targetWorld.ID)
	if err != nil {
		// No config, generate basic description
		return s.generateBasicPortalDescription(targetWorld), nil
	}

	// Generate themed portal description
	return s.generateThemedPortalDescription(targetWorld, config), nil
}

// DescribeStatue generates a description of the central statue
func (s *LookService) DescribeStatue(ctx context.Context, userID uuid.UUID) (string, error) {
	// List all worlds and check if user created any
	worlds, err := s.worldRepo.ListWorlds(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to check world creation status: %w", err)
	}

	// Check if any world was created by this user
	var userWorld *repository.World
	var worldName string
	var worldConfig *interview.WorldConfiguration

	for _, w := range worlds {
		// Try to get config for this world
		config, err := s.interviewRepo.GetConfigurationByWorldID(w.ID)
		if err == nil && config.CreatedBy == userID {
			userWorld = &w
			worldName = config.WorldName
			worldConfig = config
			break
		}
	}

	if userWorld == nil {
		// User hasn't created a world - show neutral statue
		return s.generateNeutralStatueDescription(), nil
	}

	// User has created a world - show themed statue
	return s.generateThemedStatueDescription(worldName, worldConfig), nil
}

// Helper methods for generating descriptions

func (s *LookService) generateNewPlayerDescription(username string) string {
	return fmt.Sprintf("A shapeless gray spirit drifts nearby, their form indistinct and barely visible in the fog. They seem new to this place, not yet having taken on any particular shape. You sense this is %s, not yet bound to any world.", username)
}

func (s *LookService) generateReturningPlayerDescription(ctx context.Context, char *auth.Character) (string, error) {
	// Get world to understand theme
	world, err := s.worldRepo.GetWorld(ctx, char.WorldID)
	if err != nil {
		return fmt.Sprintf("You see %s, a traveler who has walked other realms.", char.Name), nil
	}

	// Try to get world config for theme
	config, err := s.interviewRepo.GetConfigurationByWorldID(world.ID)

	baseDesc := ""
	if char.Description != "" {
		baseDesc = char.Description
	} else if char.Appearance != "" {
		// Parse appearance JSON if available
		baseDesc = fmt.Sprintf("%s appears before you", char.Name)
	} else {
		baseDesc = fmt.Sprintf("You see %s", char.Name)
	}

	// Add atmospheric suffix based on world theme
	if config != nil {
		atmosphere := s.getWorldAtmosphere(config)
		return fmt.Sprintf("%s. They carry with them the %s of %s.", baseDesc, atmosphere, world.Name), nil
	}

	return fmt.Sprintf("%s, recently returned from %s.", baseDesc, world.Name), nil
}

func (s *LookService) generateBasicPortalDescription(world *repository.World) string {
	return fmt.Sprintf("The portal to %s shimmers before you, its surface rippling like water. Beyond it, you can sense another realm waiting to be explored.", world.Name)
}

func (s *LookService) generateThemedPortalDescription(world *repository.World, config *interview.WorldConfiguration) string {
	theme := strings.ToLower(config.Theme)

	var portalDesc string

	if strings.Contains(theme, "desert") || strings.Contains(theme, "sand") {
		portalDesc = fmt.Sprintf("The portal to %s is framed by sun-bleached stone, heat shimmering across its surface. You hear the whisper of sand-laden winds and feel dry warmth emanating from within. Through the haze, you glimpse endless golden dunes beneath a burning sky.", config.WorldName)
	} else if strings.Contains(theme, "ocean") || strings.Contains(theme, "sea") || strings.Contains(theme, "water") {
		portalDesc = fmt.Sprintf("The portal to %s appears as a frame of coral and driftwood, its surface rippling like ocean waves. The cry of seabirds echoes through the opening, and you smell salt on the air. Glimpses of azure waters and white caps flash through the translucent barrier.", config.WorldName)
	} else if strings.Contains(theme, "forest") || strings.Contains(theme, "wood") || strings.Contains(theme, "tree") {
		portalDesc = fmt.Sprintf("The portal to %s is wreathed in living vines and moss-covered bark. The scent of pine and earth drifts through, accompanied by distant birdsong. Through gaps in the foliage, you see towering trees and dappled sunlight filtering through emerald canopy.", config.WorldName)
	} else if strings.Contains(theme, "mountain") || strings.Contains(theme, "stone") {
		portalDesc = fmt.Sprintf("The portal to %s is carved from ancient granite, its edges sharp and weathered. Cold mountain air flows through the opening, carrying distant echoes. Beyond, you glimpse snow-capped peaks reaching toward gray skies.", config.WorldName)
	} else if strings.Contains(theme, "tech") || strings.Contains(theme, "cyber") || strings.Contains(theme, "futur") {
		portalDesc = fmt.Sprintf("The portal to %s hums with energy, its frame made of sleek alloy and pulsing lights. You hear the rhythmic thrum of machinery and smell ozone. Through the shimmering field, you see gleaming structures and neon-lit pathways.", config.WorldName)
	} else if strings.Contains(theme, "magic") || strings.Contains(theme, "arcane") {
		portalDesc = fmt.Sprintf("The portal to %s glows with otherworldly light, arcane symbols dancing across its ethereal frame. Power crackles in the air around it, and you hear whispered incantations. Through the luminous veil, you see impossible geometries and shifting colors.", config.WorldName)
	} else {
		// Generic themed description
		portalDesc = fmt.Sprintf("The portal to %s stands before you, its frame adorned with symbols reflecting its nature. An atmosphere of %s emanates from within, drawing your attention. Through its surface, you catch glimpses of the realm beyond.", config.WorldName, theme)
	}

	return portalDesc
}

func (s *LookService) generateNeutralStatueDescription() string {
	return "A weathered stone statue stands at the center of the portals, one arm outstretched in a welcoming gesture. Its surface is carved with countless symbols and words from a thousand different languages. One phrase stands out clearly, seeming to glow with faint light: 'Type *create world* to forge your own realm.'"
}

func (s *LookService) generateThemedStatueDescription(worldName string, config *interview.WorldConfiguration) string {
	if config == nil {
		return fmt.Sprintf("The statue at the center has transformed, now pointing steadily toward the portal of %s, the world you created.", worldName)
	}

	theme := strings.ToLower(config.Theme)

	var statueDesc string

	if strings.Contains(theme, "forest") || strings.Contains(theme, "wood") {
		statueDesc = fmt.Sprintf("A statue of intertwined vines and living wood stands at the center, its branches pointing toward the verdant portal of %s, your world. Moss grows across its surface, and small flowers bloom from cracks in the bark. It pulses with gentle life, a testament to your creation.", worldName)
	} else if strings.Contains(theme, "desert") || strings.Contains(theme, "sand") {
		statueDesc = fmt.Sprintf("A sun-bleached sandstone statue rises from the center, its weathered surface carved with desert winds. One arm points unerringly toward the portal of %s, your realm. Grain by grain, it seems to shift and reform, eternally shaped by the sands of time.", worldName)
	} else if strings.Contains(theme, "ocean") || strings.Contains(theme, "sea") {
		statueDesc = fmt.Sprintf("A statue of water-smoothed coral and shell stands at the center, its surface still damp with salt spray. It points toward the ocean portal of %s, your creation. Small pools of water collect at its base, and you can hear the faint sound of waves within the stone.", worldName)
	} else if strings.Contains(theme, "mountain") || strings.Contains(theme, "stone") {
		statueDesc = fmt.Sprintf("A towering statue of granite and ice dominates the center, its peaks pointing toward the portal of %s, your world. Snow clings to its highest points, and you feel the cold majesty of mountain heights radiating from its surface.", worldName)
	} else if strings.Contains(theme, "tech") || strings.Contains(theme, "cyber") {
		statueDesc = fmt.Sprintf("A sleek statue of chrome and circuitry stands at the center, LED displays running across its surface. Its outstretched limb points with mechanical precision toward the portal of %s, your creation. Soft electronic tones emanate from within its frame.", worldName)
	} else if strings.Contains(theme, "magic") || strings.Contains(theme, "arcane") {
		statueDesc = fmt.Sprintf("A crystalline statue shimmers at the center, arcane energy flowing through its translucent form. It gestures toward the glowing portal of %s, your world. Symbols of power drift around it like fireflies, and reality seems to bend in its presence.", worldName)
	} else {
		statueDesc = fmt.Sprintf("The statue at the center has transformed to reflect the essence of %s, your created world. It stands as a monument to your imagination, one arm pointing steadily toward your realm's portal.", worldName)
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
