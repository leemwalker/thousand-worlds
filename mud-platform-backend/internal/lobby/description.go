package lobby

import (
	"context"
	"fmt"
	"strings"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/repository"

	"github.com/google/uuid"
)

// DescriptionGenerator handles dynamic lobby description generation
type DescriptionGenerator struct {
	worldRepo repository.WorldRepository
	authRepo  auth.Repository
}

// NewDescriptionGenerator creates a new description generator
func NewDescriptionGenerator(worldRepo repository.WorldRepository, authRepo auth.Repository) *DescriptionGenerator {
	return &DescriptionGenerator{
		worldRepo: worldRepo,
		authRepo:  authRepo,
	}
}

// GenerateDescription generates the full lobby description for a user
func (g *DescriptionGenerator) GenerateDescription(ctx context.Context, user *auth.User, currentPlayers []WebsocketClient) (string, error) {
	var sb strings.Builder

	// 1. Base Description & Atmosphere
	baseDesc := "You are surrounded by a low gray fog that fades into the distance. Empty portals and closed doors stand unused around you. A stone statue stands at the center of the portals, a hand is stretched out as if it beckons for you to approach."

	if user.LastWorldID != nil && *user.LastWorldID != uuid.Nil {
		// Fetch last world to get atmospheric details
		lastWorld, err := g.worldRepo.GetWorld(ctx, *user.LastWorldID)
		if err == nil {
			// Apply atmospheric overrides based on world metadata/theme
			// For now, we'll use simple keyword matching on the name or description
			// In a real system, we'd have structured "Biome" or "Theme" data
			desc, _ := lastWorld.Metadata["description"].(string)
			theme := strings.ToLower(lastWorld.Name + " " + desc)

			if strings.Contains(theme, "desert") || strings.Contains(theme, "sand") {
				baseDesc = "Hot desert winds blow sand about you, pushing away the gray fog and revealing a sandstone floor. The portals, doors, and statue remain, but the air is dry and warm."
			} else if strings.Contains(theme, "ocean") || strings.Contains(theme, "sea") || strings.Contains(theme, "water") {
				baseDesc = "The cries of sea birds and smell of salt air fills the lobby, the gray fog has been pushed into the distance, revealing a damp floor of wood planks. The statue appears slightly weathered by salt."
			} else if strings.Contains(theme, "forest") || strings.Contains(theme, "wood") || strings.Contains(theme, "tree") {
				baseDesc = "Tall trees stand in the distance, wisps of the gray fog can be seen in their branches. A grassy meadow has been revealed by a light breeze that carries the scent of leaves. The statue is covered in faint moss."
			}
		}
	}

	sb.WriteString(baseDesc)
	sb.WriteString("\n\n")

	// 2. World Portals
	worlds, err := g.worldRepo.ListWorlds(ctx)
	if err == nil && len(worlds) > 0 {
		// Group worlds (simplified for now, just listing them with flavor text)
		// In the future, we can group by tags
		var portalDescs []string
		for _, w := range worlds {
			if IsLobby(w.ID) {
				continue
			}
			// Generate a short description for the portal
			// Ideally this comes from world metadata, but we'll generate it
			portalDesc := fmt.Sprintf("A portal to %s shimmers nearby.", w.Name)
			portalDescs = append(portalDescs, portalDesc)
		}

		if len(portalDescs) > 0 {
			sb.WriteString(strings.Join(portalDescs, " "))
			sb.WriteString("\n\n")
		}
	}

	// 3. Other Players
	if len(currentPlayers) > 0 {
		var playerNames []string
		for _, p := range currentPlayers {
			if p.GetUserID() != user.UserID {
				name := p.GetUsername()
				if name == "" {
					name = "A spirit"
				}
				playerNames = append(playerNames, name)
			}
		}

		if len(playerNames) > 0 {
			if len(playerNames) == 1 {
				sb.WriteString(fmt.Sprintf("%s stands in the fog.", playerNames[0]))
			} else if len(playerNames) == 2 {
				sb.WriteString(fmt.Sprintf("%s and %s stand in the fog.", playerNames[0], playerNames[1]))
			} else if len(playerNames) == 3 {
				sb.WriteString(fmt.Sprintf("%s, %s, and %s stand in the fog.", playerNames[0], playerNames[1], playerNames[2]))
			} else {
				sb.WriteString("Other spirits stand about you in the fog.")
			}
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

// WebsocketClient interface to decouple from websocket package
type WebsocketClient interface {
	GetUserID() uuid.UUID
	GetUsername() string
}
