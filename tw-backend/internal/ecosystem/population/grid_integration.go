// Package population provides grid integration for connecting hex grid to population simulation.
package population

import (
	"math/rand"

	"tw-backend/internal/ecosystem/geography"
	wgeo "tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// InitializeGeographicSystems creates and links the hex grid, region, and tectonic systems
// This should be called after biomes are populated to enable geographic isolation tracking
func (sim *PopulationSimulator) InitializeGeographicSystems(worldID uuid.UUID, seed int64) {
	// Create hex grid with size based on biome count
	// Each biome gets approximately 10-20 hex cells
	gridSize := 20 + len(sim.Biomes)/2
	if gridSize > 100 {
		gridSize = 100 // Cap at reasonable size
	}
	sim.HexGrid = geography.NewHexGrid(worldID, gridSize, gridSize, 1.0)

	// Create tectonic system
	sim.Tectonics = geography.NewTectonicSystem(worldID, seed)

	// Create region system
	sim.RegionSystem = geography.NewRegionSystem(worldID)

	// Map biomes to hex cells
	sim.mapBiomesToGrid(seed)

	// Identify regions from connected landmasses
	sim.RegionSystem.IdentifyRegions(sim.HexGrid)
}

// mapBiomesToGrid assigns biomes to hex cells on the grid
func (sim *PopulationSimulator) mapBiomesToGrid(seed int64) {
	rng := rand.New(rand.NewSource(seed))

	biomeList := make([]*BiomePopulation, 0, len(sim.Biomes))
	for _, b := range sim.Biomes {
		biomeList = append(biomeList, b)
	}

	if len(biomeList) == 0 {
		return
	}

	// Assign cells to biomes based on grid position
	for _, cell := range sim.HexGrid.Cells {
		// Use cell position to determine biome assignment
		// This creates clustered biome regions
		biomeIdx := (abs(cell.Coord.Q) + abs(cell.Coord.R)) % len(biomeList)
		biome := biomeList[biomeIdx]

		// Set cell properties based on biome type
		cell.IsLand = biome.BiomeType != wgeo.BiomeOcean
		cell.Terrain = biomeTypeToTerrain(biome.BiomeType)
		cell.Elevation = 0.3 + rng.Float32()*0.4 // Random elevation for land

		// Link biome to cell using BiomeID field
		cell.BiomeID = &biome.BiomeID
	}
}

// biomeTypeToTerrain converts worldgen biome types to hex terrain types
func biomeTypeToTerrain(biomeType wgeo.BiomeType) geography.TerrainType {
	switch biomeType {
	case wgeo.BiomeOcean:
		return geography.TerrainOcean
	case wgeo.BiomeDesert:
		return geography.TerrainPlains // Desert is still walkable land
	case wgeo.BiomeAlpine:
		return geography.TerrainMountain
	default:
		return geography.TerrainPlains
	}
}

// UpdateGeographicSystems advances the geographic systems by the given years
// This should be called periodically (e.g., every 10,000 years)
func (sim *PopulationSimulator) UpdateGeographicSystems(years int64) {
	if sim.Tectonics == nil || sim.RegionSystem == nil || sim.HexGrid == nil {
		return // Not initialized
	}

	// Update tectonic plates
	sim.Tectonics.Update(years)

	// Update region system with current tectonic state
	sim.RegionSystem.Update(years, sim.HexGrid, sim.Tectonics)

	// Update continental fragmentation from tectonics
	sim.ContinentalFragmentation = float64(sim.Tectonics.CalculateFragmentation())
}

// ApplyIsolationEffects applies island effects (gigantism/dwarfism) to isolated populations
// Returns the number of populations affected
func (sim *PopulationSimulator) ApplyIsolationEffects() int {
	if sim.RegionSystem == nil {
		return 0
	}

	affected := 0

	// For each region, check if isolated and apply effects
	for _, region := range sim.RegionSystem.Regions {
		if !region.IsIsolated() {
			continue
		}

		modifier := region.GetIsolationModifier()
		if modifier.Strength < 0.1 {
			continue // Isolation not significant yet
		}

		// Find species in this region and apply size modification
		for _, biome := range sim.Biomes {
			// Check if this biome is in the isolated region
			if sim.isBiomeInRegion(biome, region) {
				for _, sp := range biome.Species {
					if sp.Count > 0 {
						// Apply island rule size modification
						newSize := modifier.ApplyToSize(sp.Traits.Size)
						if newSize != sp.Traits.Size {
							sp.Traits.Size = newSize
							affected++
						}
					}
				}
			}
		}
	}

	return affected
}

// isBiomeInRegion checks if a biome is located within a region
func (sim *PopulationSimulator) isBiomeInRegion(biome *BiomePopulation, region *geography.Region) bool {
	if sim.HexGrid == nil {
		return false
	}

	// Check if any cells with this biome's ID are in the region
	for _, cell := range sim.HexGrid.Cells {
		if cell.BiomeID != nil && *cell.BiomeID == biome.BiomeID && region.ContainsCell(cell.Coord) {
			return true
		}
	}
	return false
}

// ApplyRegionalMigration allows species to spread between connected regions
// Returns the number of successful migrations
func (sim *PopulationSimulator) ApplyRegionalMigration() int {
	if sim.RegionSystem == nil {
		return 0
	}

	migrations := 0

	// For each region, check connections and allow species spread
	for _, region := range sim.RegionSystem.Regions {
		for _, connection := range region.Connections {
			// Difficulty affects migration rate (0 = easy, 1 = impossible)
			if connection.Difficulty > 0.8 {
				continue // Too difficult to cross
			}

			targetRegion := sim.RegionSystem.GetRegion(connection.TargetRegionID)
			if targetRegion == nil {
				continue
			}

			// Find species that could migrate
			for _, biome := range sim.Biomes {
				if !sim.isBiomeInRegion(biome, region) {
					continue
				}

				for _, sp := range biome.Species {
					if sp.Count < 100 {
						continue // Not enough population to send migrants
					}

					// Check if target region already has this species
					targetHasSpecies := false
					for _, targetBiome := range sim.Biomes {
						if sim.isBiomeInRegion(targetBiome, targetRegion) {
							if _, exists := targetBiome.Species[sp.SpeciesID]; exists {
								targetHasSpecies = true
								break
							}
						}
					}

					// If target doesn't have this species and connection allows, migrate
					if !targetHasSpecies && sim.rng.Float64() > float64(connection.Difficulty) {
						// Find a target biome in the target region
						for _, targetBiome := range sim.Biomes {
							if sim.isBiomeInRegion(targetBiome, targetRegion) {
								// Create founding population in target
								migrants := sp.Count / 100 // 1% migrate
								if migrants < 10 {
									migrants = 10
								}

								newPop := &SpeciesPopulation{
									SpeciesID:     sp.SpeciesID,
									Name:          sp.Name,
									Count:         migrants,
									Traits:        sp.Traits,
									TraitVariance: sp.TraitVariance * 1.2, // Increased variance from founder effect
									Diet:          sp.Diet,
									Generation:    sp.Generation,
									CreatedYear:   sim.CurrentYear,
								}
								targetBiome.Species[sp.SpeciesID] = newPop
								sp.Count -= migrants
								migrations++
								break
							}
						}
					}
				}
			}
		}
	}

	return migrations
}

// abs returns absolute value of int
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
