package gamemap

import (
	"context"
	"fmt"
	"log"
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

	// Cache for world map data (key: "worldID:gridSize")
	worldMapCache sync.Map
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
	if geo != nil && geo.IsInitialized() {
		log.Printf("[MAP] SetWorldGeology: Registered geology for world %s with %d biomes, heightmap %dx%d",
			worldID, len(geo.Biomes), geo.Heightmap.Width, geo.Heightmap.Height)
	} else if geo == nil {
		log.Printf("[MAP] SetWorldGeology: Cleared geology for world %s", worldID)
	}
}

// getWorldGeology retrieves cached geology data
func (s *Service) getWorldGeology(worldID uuid.UUID) *ecosystem.WorldGeology {
	s.worldGeologyMu.RLock()
	defer s.worldGeologyMu.RUnlock()
	return s.worldGeology[worldID]
}

// worldToGrid converts world coordinates to heightmap grid indices
// World coordinates can be very large (e.g., spherical world with circumference 17M)
// but heightmap is typically 512x512 or similar
func worldToGrid(worldX, worldY float64, minX, minY, maxX, maxY float64, gridWidth, gridHeight int) (int, int) {
	// Calculate world dimensions
	worldWidth := maxX - minX
	worldHeight := maxY - minY

	// Avoid division by zero
	if worldWidth <= 0 {
		worldWidth = 1
	}
	if worldHeight <= 0 {
		worldHeight = 1
	}

	// Normalize world position to 0..1
	normalizedX := (worldX - minX) / worldWidth
	normalizedY := (worldY - minY) / worldHeight

	// Handle wrapping for spherical worlds
	normalizedX = normalizedX - math.Floor(normalizedX) // Keep in 0..1
	normalizedY = normalizedY - math.Floor(normalizedY)

	// Scale to grid dimensions
	gridX := int(normalizedX * float64(gridWidth))
	gridY := int(normalizedY * float64(gridHeight))

	// Clamp to valid range
	if gridX < 0 {
		gridX = 0
	}
	if gridX >= gridWidth {
		gridX = gridWidth - 1
	}
	if gridY < 0 {
		gridY = 0
	}
	if gridY >= gridHeight {
		gridY = gridHeight - 1
	}

	return gridX, gridY
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

	// Check for geology fallback
	geo := s.getWorldGeology(char.WorldID)
	hasGeology := geo != nil && geo.IsInitialized()

	log.Printf("[MAP] GetMapData: world=%s hasWorldData=%v hasGeology=%v quality=%s",
		char.WorldID, hasWorldData, hasGeology, quality)

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

	// Dynamic grid sizing based on altitude
	// Use ODD grid sizes (2*radius + 1) for perfect centering on player
	// Base: radius 4 = 9x9 grid at ground level
	// Grows by 1 to radius for every 5m of altitude
	// Max radius 25 = 51x51 grid (2601 tiles) for performance
	gridRadius := 4 // Base 9x9 grid

	// Calculate stride (downsampling factor) for high altitudes
	// Stride increases with altitude to show larger area with same number of tiles
	stride := 1
	if char.IsFlying && char.PositionZ > 0 {
		// Add 1 to radius for every 5m of altitude
		additionalRadius := int(char.PositionZ / 5.0)
		gridRadius = 4 + additionalRadius

		// Cap radius at 25 (51x51 grid = 2601 tiles max)
		if gridRadius > 25 {
			gridRadius = 25
		}

		// Calculate stride: 1 for <100m, increases by 1 for every 100m
		// e.g. 100m -> stride 1, 200m -> stride 2, 500m -> stride 5
		if char.PositionZ >= 100 {
			stride = int(char.PositionZ / 100.0)
			if stride < 1 {
				stride = 1
			}
		}
	}
	gridSize := gridRadius*2 + 1 // Odd number for perfect centering

	// Create 2D grid for visibility calculations
	// Access: grid[dy + gridRadius][dx + gridRadius]
	grid := make([][]*MapTile, gridSize)
	for i := range grid {
		grid[i] = make([]*MapTile, gridSize)
	}

	// Generate grid centered on player (-radius to +radius * stride)
	// We iterate clearly using grid indices (-radius to +radius) and multiply by stride for world coordinates
	for dy := -gridRadius; dy <= gridRadius; dy++ {
		for dx := -gridRadius; dx <= gridRadius; dx++ {
			// Calculate world position with stride
			// tileX = playerX + (dx * stride)
			tileX := int(math.Round(char.PositionX)) + (dx * stride)
			tileY := int(math.Round(char.PositionY)) + (dy * stride)

			// Check if tile is out of bounds
			outOfBounds := float64(tileX) < minX || float64(tileX) > maxX ||
				float64(tileY) < minY || float64(tileY) > maxY

			tile := &MapTile{
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
				// Convert world coordinates to heightmap grid indices
				gridX, gridY := worldToGrid(float64(tileX), float64(tileY), minX, minY, maxX, maxY, hm.Width, hm.Height)

				tile.Elevation = hm.Get(gridX, gridY)

				// Get biome
				idx := gridY*hm.Width + gridX
				if idx >= 0 && idx < len(worldData.Geography.Biomes) {
					tile.Biome = string(worldData.Geography.Biomes[idx].Type)
				}
			} else if hasGeology {
				// Fallback: use worldGeology data from async runner or world simulate
				hm := geo.Heightmap
				if hm != nil {
					// Convert world coordinates to heightmap grid indices
					gridX, gridY := worldToGrid(float64(tileX), float64(tileY), minX, minY, maxX, maxY, hm.Width, hm.Height)

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
				// Use stride as search radius? No, just look at the specific tile point.
				// For high altitude, we might miss entities if they aren't exactly on the sampled tile.
				// But aggregating entities is complex. For now, just show ents at sample points.
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

			// Store in grid
			grid[dy+gridRadius][dx+gridRadius] = tile
		}
	}

	// Calculate Visibility (Horizon Culling)
	s.computeOcclusion(grid, gridRadius, char.PositionZ, stride)

	// Flatten grid to tiles slice
	tiles := make([]MapTile, 0, gridSize*gridSize)
	for _, row := range grid {
		for _, tile := range row {
			if tile != nil {
				tiles = append(tiles, *tile)
			}
		}
	}

	// Verify tile count
	expectedCount := gridSize * gridSize
	if len(tiles) != expectedCount {
		log.Printf("[MAP] WARNING: Generated %d tiles, expected %d (gridSize %d)", len(tiles), expectedCount, gridSize)
	}

	// Debug: Log biome distribution
	biomeCounts := make(map[string]int)
	for _, t := range tiles {
		biomeCounts[t.Biome]++
	}
	log.Printf("[MAP] Generated %d tiles, biome distribution: %v", len(tiles), biomeCounts)

	return &MapData{
		Tiles:         tiles,
		PlayerX:       char.PositionX,
		PlayerY:       char.PositionY,
		RenderQuality: quality,
		GridSize:      gridSize,
		Scale:         stride, // Send stride as Scale
		WorldID:       char.WorldID,
	}, nil
}

const (
	// EyeHeight is the height of the player's eyes above their position (feet)
	EyeHeight = 1.7
)

// computeOcclusion implements Horizon Culling for Line of Sight
func (s *Service) computeOcclusion(grid [][]*MapTile, radius int, playerAlt float64, stride int) {
	// Adjust player altitude to eye level
	startAlt := playerAlt + EyeHeight
	cx, cy := 0, 0

	// Helper to cast ray
	castRay := func(tx, ty int) {
		// Bresenham's Line Algorithm
		x0, y0 := cx, cy
		x1, y1 := tx, ty

		dx := x1 - x0
		if dx < 0 {
			dx = -dx
		}
		dy := y1 - y0
		if dy < 0 {
			dy = -dy
		}

		sx := -1
		if x0 < x1 {
			sx = 1
		}
		sy := -1
		if y0 < y1 {
			sy = 1
		}

		errVal := dx - dy

		maxSlope := -math.MaxFloat64 // Start with no horizon (everything visible)

		for {
			// Process tile at (x0, y0)
			// Convert to grid index: grid[y0 + radius][x0 + radius]
			row := y0 + radius
			col := x0 + radius

			if row >= 0 && row < len(grid) && col >= 0 && col < len(grid[0]) {
				tile := grid[row][col]
				if tile != nil {
					// Calculate distance and slope
					// Distance in WORLD UNITS (meters)
					// dx, dy are in tiles. World dist = sqrt(dx*dx + dy*dy) * stride
					distTiles := math.Sqrt(float64(x0*x0 + y0*y0))
					distMeters := distTiles * float64(stride)

					if distMeters < 0.1 {
						// At player position (dist ~0)
						// Always visible. Do not update maxSlope (let horizons form naturally)
						tile.Occluded = false
					} else {
						// Slope = (TileElev - EyeLevel) / Dist
						slope := (tile.Elevation - startAlt) / distMeters

						if slope >= maxSlope {
							// Visible (above horizon)
							tile.Occluded = false
							maxSlope = slope
						} else {
							// Hidden (below horizon)
							tile.Occluded = true
						}
					}
				}
			}

			if x0 == x1 && y0 == y1 {
				break
			}

			e2 := 2 * errVal
			if e2 > -dy {
				errVal -= dy
				x0 += sx
			}
			if e2 < dx {
				errVal += dx
				y0 += sy
			}
		}
	}

	// Cast rays to all perimeter tiles
	// Top and Bottom rows
	for x := -radius; x <= radius; x++ {
		castRay(x, -radius) // Top
		castRay(x, radius)  // Bottom
	}
	// Left and Right columns (excluding corners already done)
	for y := -radius + 1; y <= radius-1; y++ {
		castRay(-radius, y) // Left
		castRay(radius, y)  // Right
	}
}

// GetWorldMapData returns aggregated world map data for full world display
// The world is divided into a grid of regions, each with a dominant biome
func (s *Service) GetWorldMapData(ctx context.Context, char *auth.Character, gridSize int) (*WorldMapData, error) {
	if gridSize <= 0 {
		gridSize = 64 // Default to 64x64 grid
	}

	// Check cache first
	cacheKey := fmt.Sprintf("%s:%d", char.WorldID, gridSize)
	if cached, ok := s.worldMapCache.Load(cacheKey); ok {
		if data, ok := cached.(*WorldMapData); ok {
			// Update player position in cached data (position can change)
			data.PlayerX = char.PositionX
			data.PlayerY = char.PositionY
			return data, nil
		}
	}

	world, err := s.worldRepo.GetWorld(ctx, char.WorldID)
	if err != nil {
		return nil, err
	}

	// Determine world bounds
	var worldWidth, worldHeight float64
	if world.Circumference != nil && *world.Circumference > 0 {
		// Spherical world
		worldWidth = *world.Circumference
		worldHeight = *world.Circumference / 2 // -90 to +90 degrees = half circumference
	} else if world.BoundsMin != nil && world.BoundsMax != nil {
		// Bounded world
		worldWidth = world.BoundsMax.X - world.BoundsMin.X
		worldHeight = world.BoundsMax.Y - world.BoundsMin.Y
	} else {
		// Default fallback
		worldWidth = 1000
		worldHeight = 1000
	}

	// Calculate region size (world units per grid cell)
	regionWidth := worldWidth / float64(gridSize)
	regionHeight := worldHeight / float64(gridSize)

	// Get geology data for biome lookup
	geo := s.getWorldGeology(char.WorldID)

	tiles := make([]WorldMapTile, 0, gridSize*gridSize)
	playerGridX := int(char.PositionX / regionWidth)
	playerGridY := int(char.PositionY / regionHeight)

	// Generate aggregated tiles
	for gy := 0; gy < gridSize; gy++ {
		for gx := 0; gx < gridSize; gx++ {
			// Calculate center of this region in world coordinates
			centerX := (float64(gx) + 0.5) * regionWidth
			centerY := (float64(gy) + 0.5) * regionHeight

			biome := "default"
			elevation := 0.0

			// Look up biome from geology
			if geo != nil && geo.Heightmap != nil {
				hm := geo.Heightmap
				if hm.Width > 0 && hm.Height > 0 {
					// Convert grid position to heightmap indices
					hmX, hmY := worldToGrid(centerX, centerY, 0, 0, worldWidth, worldHeight, hm.Width, hm.Height)
					if hmX >= 0 && hmX < hm.Width && hmY >= 0 && hmY < hm.Height {
						elevation = hm.Get(hmX, hmY)
						// Get biome from geo.Biomes array using linear index
						idx := hmY*hm.Width + hmX
						if idx >= 0 && idx < len(geo.Biomes) {
							biome = string(geo.Biomes[idx].Type)
						}
					}
				}
			}

			tile := WorldMapTile{
				GridX:        gx,
				GridY:        gy,
				Biome:        biome,
				AvgElevation: elevation,
				IsPlayer:     gx == playerGridX && gy == playerGridY,
			}
			tiles = append(tiles, tile)
		}
	}

	result := &WorldMapData{
		Tiles:       tiles,
		GridWidth:   gridSize,
		GridHeight:  gridSize,
		WorldWidth:  worldWidth,
		WorldHeight: worldHeight,
		PlayerX:     char.PositionX,
		PlayerY:     char.PositionY,
		WorldID:     char.WorldID,
		WorldName:   world.Name,
	}

	// Store in cache
	s.worldMapCache.Store(cacheKey, result)

	return result, nil
}
