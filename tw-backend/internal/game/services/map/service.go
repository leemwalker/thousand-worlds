package gamemap

import (
	"context"
	"math"
	"sync"

	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/game/services/entity"
	"tw-backend/internal/game/services/look"
	"tw-backend/internal/repository"
	"tw-backend/internal/skills"
	"tw-backend/internal/worldentity"
	"tw-backend/internal/worldgen/orchestrator"

	"github.com/google/uuid"
)

const (
	// MapGridRadius is the number of tiles in each direction from the player (4 = 9x9 grid)
	MapGridRadius = 4
	// MapGridSize is the total grid dimension
	MapGridSize = MapGridRadius*2 + 1
)

// Service provides map data generation for the mini-map feature
type Service struct {
	worldRepo          repository.WorldRepository
	skillsRepo         skills.Repository
	entityService      *entity.Service
	lookService        *look.LookService
	worldEntityService *worldentity.Service
	ecosystemService   *ecosystem.Service

	// Fallback geology data for biome rendering
	worldGeologyMu sync.RWMutex
	worldGeology   map[uuid.UUID]*ecosystem.WorldGeology
}

// NewService creates a new map service
func NewService(
	worldRepo repository.WorldRepository,
	skillsRepo skills.Repository,
	entityService *entity.Service,
	lookService *look.LookService,
	worldEntityService *worldentity.Service,
	ecosystemService *ecosystem.Service,
) *Service {
	return &Service{
		worldRepo:          worldRepo,
		skillsRepo:         skillsRepo,
		entityService:      entityService,
		lookService:        lookService,
		worldEntityService: worldEntityService,
		ecosystemService:   ecosystemService,
		worldGeology:       make(map[uuid.UUID]*ecosystem.WorldGeology),
	}
}

// SetWorldGeology registers geology data for a world to enable biome rendering
func (s *Service) SetWorldGeology(worldID uuid.UUID, geo *ecosystem.WorldGeology) {
	s.worldGeologyMu.Lock()
	defer s.worldGeologyMu.Unlock()
	s.worldGeology[worldID] = geo
}

// getWorldGeology retrieves cached geology data
func (s *Service) getWorldGeology(worldID uuid.UUID) *ecosystem.WorldGeology {
	s.worldGeologyMu.RLock()
	defer s.worldGeologyMu.RUnlock()
	return s.worldGeology[worldID]
}

// GetMapData returns visible tiles in a 9x9 grid centered on the player (15x15 when flying)
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

	// Get world bounds for boundary checking
	var minX, minY, maxX, maxY float64 = 0, 0, 10, 10 // Default lobby bounds
	if s.worldRepo != nil {
		world, err := s.worldRepo.GetWorld(ctx, char.WorldID)
		if err == nil && world != nil {
			if world.BoundsMin != nil {
				minX, minY = world.BoundsMin.X, world.BoundsMin.Y
			}
			if world.BoundsMax != nil {
				maxX, maxY = world.BoundsMax.X, world.BoundsMax.Y
			}
		}
	}

	// Use expanded view radius when flying
	gridRadius := MapGridRadius
	scale := 1

	if char.IsFlying {
		gridRadius = MapGridRadius * 2 // 8 tiles = 17x17 grid

		// Calculate scale based on altitude
		// Scale 1 = ground level (1 tile = 1 coordinate)
		// Scale increases by 1 for every 20m of altitude
		if char.PositionZ > 0 {
			scale = 1 + int(char.PositionZ/20.0)
		}

		// Cap scale to prevent ridiculous values
		if scale > 50 {
			scale = 50
		}
	}
	gridSize := gridRadius*2 + 1

	tiles := make([]MapTile, 0, gridSize*gridSize)

	// Generate grid centered on player
	for dy := -gridRadius; dy <= gridRadius; dy++ {
		for dx := -gridRadius; dx <= gridRadius; dx++ {
			// Apply scale to offsets
			offsetX := dx * scale
			offsetY := dy * scale

			tileX := int(math.Round(char.PositionX)) + offsetX
			tileY := int(math.Round(char.PositionY)) + offsetY

			// Check if tile is out of bounds
			outOfBounds := float64(tileX) < minX || float64(tileX) > maxX ||
				float64(tileY) < minY || float64(tileY) > maxY

			tile := MapTile{
				X:           tileX,
				Y:           tileY,
				IsPlayer:    dx == 0 && dy == 0,
				OutOfBounds: outOfBounds,
			}

			// Set biome based on bounds
			if outOfBounds {
				tile.Biome = "void" // Void/wall for out-of-bounds tiles
			} else {
				tile.Biome = "lobby" // Default biome for in-bounds tiles without world data
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
			} else if geo := s.getWorldGeology(char.WorldID); geo != nil && geo.IsInitialized() {
				// Fallback: use worldGeology data from async runner or world simulate
				hm := geo.Heightmap
				if hm != nil {
					// Clamp to heightmap bounds
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

					// Get biome from geology
					idx := gridY*hm.Width + gridX
					if idx >= 0 && idx < len(geo.Biomes) {
						tile.Biome = string(geo.Biomes[idx].Type)
					}
				}
			}

			// Get entities at this tile if entity service is available (in-memory entities)
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

			// Get world entities at this tile (database-backed static objects)
			// Use 0.5 radius to only match entities at this exact tile position
			if s.worldEntityService != nil {
				worldEntities, err := s.worldEntityService.GetEntitiesAt(ctx, char.WorldID, float64(tileX), float64(tileY), 0.5)
				if err == nil && len(worldEntities) > 0 {
					for _, we := range worldEntities {
						tile.Entities = append(tile.Entities, MapEntity{
							ID:    we.ID,
							Type:  string(we.EntityType),
							Name:  we.Name,
							Glyph: we.GetGlyph(),
						})
					}
				}
			}

			// Get ecosystem entities (living creatures)
			if s.ecosystemService != nil {
				ecoEntities := s.ecosystemService.GetEntitiesAt(char.WorldID, float64(tileX), float64(tileY), 0.5)
				for _, e := range ecoEntities {
					glyph := "❓"
					switch e.Species {
					case "rabbit":
						glyph = "🐇"
					case "wolf":
						glyph = "🐺"
					case "deer":
						glyph = "🦌"
					case "bear":
						glyph = "🐻"
					}

					tile.Entities = append(tile.Entities, MapEntity{
						ID:     e.EntityID,
						Type:   "creature",
						Name:   string(e.Species),
						Glyph:  glyph,
						Status: "neutral", // TODO: Determine based on AI
					})
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
		GridSize:      gridSize,
		WorldID:       char.WorldID,
	}, nil
}
