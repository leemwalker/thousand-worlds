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
func (g *DescriptionGenerator) GenerateDescription(ctx context.Context, user *auth.User, char *auth.Character, currentPlayers []WebsocketClient) (string, error) {
	var sb strings.Builder

	// 1. Base Description & Atmosphere based on position
	var baseDesc string

	// Lobby is 1000m long (X axis).
	// Zones:
	// 0-333: West Wing (Quiet, meditation)
	// 333-666: Central Hub (Statue, portals)
	// 666-1000: East Wing (Gathering, planning)

	posX := char.PositionX

	if posX < 333 {
		baseDesc = "You are in the West Wing of the Grand Lobby. It is quieter here, with alcoves for rest and meditation. Soft light filters from above, illuminating the smooth stone floor."
	} else if posX > 666 {
		baseDesc = "You are in the East Wing of the Grand Lobby. Several long tables are arranged here for travelers to gather and plan their journeys. The air buzzes with faint conversation from distant echoes."
	} else {
		baseDesc = "You are in the Central Hub of the Grand Lobby. The massive statue dominates the space here, surrounded by shimmering portals to other worlds. The area is bustling with potential energy."
	}

	// Add atmosphere from last world if applicable
	if char.LastWorldVisited != nil && *char.LastWorldVisited != uuid.Nil {
		// Fetch last world to get atmospheric details
		lastWorld, err := g.worldRepo.GetWorld(ctx, *char.LastWorldVisited)
		if err == nil {
			// Apply atmospheric overrides based on world metadata/theme
			desc, _ := lastWorld.Metadata["description"].(string)
			theme := strings.ToLower(lastWorld.Name + " " + desc)

			if strings.Contains(theme, "desert") || strings.Contains(theme, "sand") {
				baseDesc += " A lingering warmth from the desert clings to your clothes."
			} else if strings.Contains(theme, "ocean") || strings.Contains(theme, "sea") || strings.Contains(theme, "water") {
				baseDesc += " You can still smell the faint tang of salt in the air."
			} else if strings.Contains(theme, "forest") || strings.Contains(theme, "wood") || strings.Contains(theme, "tree") {
				baseDesc += " A fresh scent of pine follows you."
			}
		}
	}

	sb.WriteString(baseDesc)
	sb.WriteString("\n\n")

	// 2. World Portals (Only visible in Central Hub?)
	// Let's make them visible everywhere but described differently maybe, or just keep as is for now implies they are central.
	// If in Central Hub, describe them prominently.
	if posX >= 333 && posX <= 666 {
		worlds, err := g.worldRepo.ListWorlds(ctx)
		if err == nil && len(worlds) > 0 {
			var portalDescs []string
			for _, w := range worlds {
				if IsLobby(w.ID) {
					continue
				}
				portalDesc := fmt.Sprintf("A portal to %s shimmers nearby.", w.Name)
				portalDescs = append(portalDescs, portalDesc)
			}
			if len(portalDescs) > 0 {
				sb.WriteString(strings.Join(portalDescs, " "))
				sb.WriteString("\n\n")
			}
		}
	} else {
		sb.WriteString("In the distance, toward the center of the hall, you can see the shimmering lights of the world portals.\n\n")
	}

	// 3. Other Players
	if len(currentPlayers) > 0 {
		var playerNames []string
		count := 0
		for _, p := range currentPlayers {
			if p.GetUserID() != user.UserID {
				// TODO: Calculate distance and only show nearby players?
				// For now, list all but maybe indicate distance if we had their positions.
				// Since we don't have their positions easily accessible here (players slice doesn't have pos?),
				// we will just list them.
				name := p.GetUsername()
				if name == "" {
					name = "A spirit"
				}
				playerNames = append(playerNames, name)
				count++
				if count >= 10 { // Limit list
					playerNames = append(playerNames, "others")
					break
				}
			}
		}

		if len(playerNames) > 0 {
			if len(playerNames) == 1 {
				sb.WriteString(fmt.Sprintf("%s stands nearby.", playerNames[0]))
			} else {
				sb.WriteString(fmt.Sprintf("You see %s nearby.", strings.Join(playerNames, ", ")))
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
