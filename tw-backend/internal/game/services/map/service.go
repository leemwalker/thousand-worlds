package gamemap

import (
	"context"
	"math"

	"tw-backend/internal/auth"
	"tw-backend/internal/game/services/entity"
	"tw-backend/internal/game/services/look"
	"tw-backend/internal/repository"
	"tw-backend/internal/skills"
	"tw-backend/internal/worldgen/orchestrator"
)

const (
	// MapGridRadius is the number of tiles in each direction from the player (4 = 9x9 grid)
	MapGridRadius = 4
	// MapGridSize is the total grid dimension
	MapGridSize = MapGridRadius*2 + 1
)

// Service provides map data generation for the mini-map feature
type Service struct {
	worldRepo     repository.WorldRepository
	skillsRepo    skills.Repository
	entityService *entity.Service
	lookService   *look.LookService
}

// NewService creates a new map service
func NewService(
	worldRepo repository.WorldRepository,
	skillsRepo skills.Repository,
	entityService *entity.Service,
	lookService *look.LookService,
) *Service {
	return &Service{
		worldRepo:     worldRepo,
		skillsRepo:    skillsRepo,
		entityService: entityService,
		lookService:   lookService,
	}
}

// GetMapData returns visible tiles in a 9x9 grid centered on the player
func (s *Service) GetMapData(ctx context.Context, char *auth.Character) (*MapData, error) {
	// Default to max perception (100) for lobby users who don't have skills yet
	// This shows high-quality map in lobby
	perception := 100
	if s.skillsRepo != nil {
		skillList, err := s.skillsRepo.GetSkills(ctx, char.CharacterID)
		if err == nil && len(skillList) > 0 {
			// Only override if skills exist
			for _, skill := range skillList {
				if skill.Name == skills.SkillPerception {
					perception = skill.Level
					break
				}
			}
		}
	}

	quality := GetRenderQuality(perception)

	// Get world data for biome/elevation info via look service's cached world data
	var hasWorldData bool
	var worldData *orchestrator.GeneratedWorld
	if s.lookService != nil {
		worldData, hasWorldData = s.lookService.GetCachedWorldData(char.WorldID)
	}

	tiles := make([]MapTile, 0, MapGridSize*MapGridSize)

	// Generate 9x9 grid centered on player
	for dy := -MapGridRadius; dy <= MapGridRadius; dy++ {
		for dx := -MapGridRadius; dx <= MapGridRadius; dx++ {
			tileX := int(math.Round(char.PositionX)) + dx
			tileY := int(math.Round(char.PositionY)) + dy

			tile := MapTile{
				X:        tileX,
				Y:        tileY,
				IsPlayer: dx == 0 && dy == 0,
			}

			// Get biome and elevation from world data if available
			if hasWorldData && worldData.Geography != nil {
				hm := worldData.Geography.Heightmap
				// Clamp to world bounds
				gridX := tileX
				gridY := tileY
				if gridX < 0 {
					gridX = 0
				}
				if gridX >= hm.Width {
					gridX = hm.Width - 1
				}
				if gridY < 0 {
					gridY = 0
				}
				if gridY >= hm.Height {
					gridY = hm.Height - 1
				}

				tile.Elevation = hm.Get(gridX, gridY)

				// Get biome
				idx := gridY*hm.Width + gridX
				if idx >= 0 && idx < len(worldData.Geography.Biomes) {
					tile.Biome = string(worldData.Geography.Biomes[idx].Type)
				}
			}

			// Get entities at this tile if entity service is available
			if s.entityService != nil {
				entities, err := s.entityService.GetEntitiesAt(ctx, char.WorldID, float64(tileX), float64(tileY), 1.0)
				if err == nil && len(entities) > 0 {
					for _, e := range entities {
						tile.Entities = append(tile.Entities, MapEntity{
							ID:   e.ID,
							Type: string(e.Type),
							Name: e.Name,
						})
					}
				}
			}

			tiles = append(tiles, tile)
		}
	}

	return &MapData{
		Tiles:         tiles,
		PlayerX:       char.PositionX,
		PlayerY:       char.PositionY,
		RenderQuality: quality,
		GridSize:      MapGridSize,
		WorldID:       char.WorldID,
	}, nil
}
