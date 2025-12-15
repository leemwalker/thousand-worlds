package processor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/ai/behaviortree"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/ecosystem/population"
	"tw-backend/internal/ecosystem/state"
	"tw-backend/internal/worldgen/geography"
	"tw-backend/internal/worldgen/weather"

	"github.com/google/uuid"
)

// handleWorld handles world-level commands including simulation
func (p *GameProcessor) handleWorld(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		client.SendGameMessage("error", "Usage: world <action> [args]\nTry: world simulate <years>", nil)
		return nil
	}

	subCmd := strings.ToLower(*cmd.Target)

	switch subCmd {
	case "simulate", "sim":
		if cmd.Message == nil {
			client.SendGameMessage("error", "Usage: world simulate <years>\nExample: world simulate 1000000", nil)
			return nil
		}
		return p.handleWorldSimulate(ctx, client, *cmd.Message)
	case "info":
		return p.handleWorldInfo(ctx, client)
	case "reset":
		return p.handleWorldReset(ctx, client)
	default:
		client.SendGameMessage("error", "Unknown world command. Try: 'simulate', 'info', 'reset'", nil)
		return nil
	}
}

// handleWorldSimulate runs a fast-forward simulation of the world
func (p *GameProcessor) handleWorldSimulate(ctx context.Context, client websocket.GameClient, yearsStr string) error {
	years, err := strconv.ParseInt(strings.TrimSpace(yearsStr), 10, 64)
	if err != nil || years <= 0 {
		client.SendGameMessage("error", "Invalid years. Please provide a positive number.", nil)
		return nil
	}

	// Get current world for context
	char, _ := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if char == nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	// Get world for circumference/seed
	world, err := p.worldRepo.GetWorld(ctx, char.WorldID)
	if err != nil {
		client.SendGameMessage("error", "Could not get world info", nil)
		return nil
	}

	// Initialize geology if not exists
	geology, exists := p.worldGeology[char.WorldID]
	if !exists {
		// Default circumference if not set (Earth-like: 40,000 km = 40,000,000 m)
		circumference := 40_000_000.0
		if world.Circumference != nil {
			circumference = *world.Circumference
		}

		// Use world ID bytes as seed for determinism
		seed := int64(char.WorldID[0])<<56 | int64(char.WorldID[1])<<48 |
			int64(char.WorldID[2])<<40 | int64(char.WorldID[3])<<32 |
			int64(char.WorldID[4])<<24 | int64(char.WorldID[5])<<16 |
			int64(char.WorldID[6])<<8 | int64(char.WorldID[7])

		geology = ecosystem.NewWorldGeology(char.WorldID, seed, circumference)
		p.worldGeology[char.WorldID] = geology
	}

	// Initialize terrain if first simulation
	if !geology.IsInitialized() {
		client.SendGameMessage("system", "Initializing world geology...", nil)
		geology.InitializeGeology()
		client.SendGameMessage("system", "Geology initialized with tectonic plates and terrain.", nil)

		// Spawn initial creatures based on generated biomes
		if len(geology.Biomes) > 0 {
			client.SendGameMessage("system", "Spawning initial life forms...", nil)
			p.ecosystemService.SpawnBiomes(geology.Biomes)
			client.SendGameMessage("system", fmt.Sprintf("Spawned %d entities across %d biomes.", len(p.ecosystemService.Entities), len(geology.Biomes)), nil)
		}
	}

	// Use population-based simulation for efficiency
	client.SendGameMessage("system", fmt.Sprintf("Starting population simulation of %d years...", years), nil)

	// Create seed from world ID
	seed := int64(char.WorldID[0])<<56 | int64(char.WorldID[1])<<48 |
		int64(char.WorldID[2])<<40 | int64(char.WorldID[3])<<32 |
		int64(char.WorldID[4])<<24 | int64(char.WorldID[5])<<16 |
		int64(char.WorldID[6])<<8 | int64(char.WorldID[7])

	// Initialize population simulator
	popSim := population.NewPopulationSimulator(char.WorldID, seed)

	// Group biomes by type to ensure diversity
	biomesByType := make(map[geography.BiomeType][]*geography.Biome)
	for i := range geology.Biomes {
		biome := &geology.Biomes[i]
		biomesByType[biome.Type] = append(biomesByType[biome.Type], biome)
	}

	// Create populations for each biome type (sample up to 2 per type)
	for biomeType, biomes := range biomesByType {
		// Take up to 2 biomes of each type
		count := 2
		if len(biomes) < count {
			count = len(biomes)
		}

		for i := 0; i < count; i++ {
			bp := population.NewBiomePopulation(uuid.New(), biomeType)

			// Flora with biome-specific growth type
			floraTraits := population.DefaultTraitsForDiet(population.DietPhotosynthetic)
			floraTraits.FloraGrowth = population.GetFloraGrowthForBiome(biomeType)
			floraTraits.Covering = population.GetCoveringForDiet(population.DietPhotosynthetic, biomeType)
			floraSpecies := &population.SpeciesPopulation{
				SpeciesID:     uuid.New(),
				Name:          population.GenerateSpeciesName(floraTraits, population.DietPhotosynthetic, biomeType),
				Count:         500,
				Traits:        floraTraits,
				TraitVariance: 0.3,
				Diet:          population.DietPhotosynthetic,
				Generation:    0,
				CreatedYear:   0,
			}
			bp.AddSpecies(floraSpecies)

			// Herbivore with biome-specific covering
			herbTraits := population.DefaultTraitsForDiet(population.DietHerbivore)
			herbTraits.Covering = population.GetCoveringForDiet(population.DietHerbivore, biomeType)
			herbSpecies := &population.SpeciesPopulation{
				SpeciesID:     uuid.New(),
				Name:          population.GenerateSpeciesName(herbTraits, population.DietHerbivore, biomeType),
				Count:         200,
				Traits:        herbTraits,
				TraitVariance: 0.3,
				Diet:          population.DietHerbivore,
				Generation:    0,
				CreatedYear:   0,
			}
			bp.AddSpecies(herbSpecies)

			// Carnivore with biome-specific covering
			carnTraits := population.DefaultTraitsForDiet(population.DietCarnivore)
			carnTraits.Covering = population.GetCoveringForDiet(population.DietCarnivore, biomeType)
			carnSpecies := &population.SpeciesPopulation{
				SpeciesID:     uuid.New(),
				Name:          population.GenerateSpeciesName(carnTraits, population.DietCarnivore, biomeType),
				Count:         50,
				Traits:        carnTraits,
				TraitVariance: 0.3,
				Diet:          population.DietCarnivore,
				Generation:    0,
				CreatedYear:   0,
			}
			bp.AddSpecies(carnSpecies)

			popSim.Biomes[bp.BiomeID] = bp
		}
	}

	client.SendGameMessage("system", fmt.Sprintf("Simulating %d biome types with %d total biome instances...", len(biomesByType), len(popSim.Biomes)), nil)

	// Track statistics
	geologicalEvents := 0
	geoManager := ecosystem.NewGeologicalEventManager()
	progressInterval := years / 10
	lastProgress := int64(0)

	// Run simulation year by year (fast!)
	for year := int64(0); year < years; year++ {
		// Progress reporting
		if year-lastProgress >= progressInterval && progressInterval > 0 {
			percent := (year * 100) / years
			totalPop, totalSpecies, totalExtinct := popSim.GetStats()
			client.SendGameMessage("system", fmt.Sprintf("⏳ Progress: %d%% (Year %d, Pop: %d, Species: %d, Extinct: %d)",
				percent, year, totalPop, totalSpecies, totalExtinct), nil)
			lastProgress = year
		}

		// Simulate population dynamics + evolution + speciation
		popSim.SimulateYear()

		// Apply evolution every 1000 years
		if popSim.CurrentYear%1000 == 0 {
			popSim.ApplyEvolution()
		}

		// Check for speciation every 10000 years
		if popSim.CurrentYear%10000 == 0 {
			popSim.CheckSpeciation()
		}

		// Check for geological events (every 10000 years)
		if year%10000 == 0 {
			tick := year * 365 // Convert to ticks for geo manager
			previousEventCount := len(geoManager.ActiveEvents)
			geoManager.CheckForNewEvents(tick, 365*10000)
			newEvents := len(geoManager.ActiveEvents) - previousEventCount
			geologicalEvents += newEvents

			if newEvents > 0 {
				for i := len(geoManager.ActiveEvents) - newEvents; i < len(geoManager.ActiveEvents); i++ {
					e := geoManager.ActiveEvents[i]
					client.SendGameMessage("system", fmt.Sprintf("⚠️ GEOLOGICAL EVENT: %s (severity: %.0f%%)", e.Type, e.Severity*100), nil)
					geology.ApplyEvent(e)

					// Apply extinction event to populations based on event type
					eventType := population.ExtinctionEventType(e.Type)
					deaths := popSim.ApplyExtinctionEvent(eventType, e.Severity)
					if deaths > 100 {
						client.SendGameMessage("system", fmt.Sprintf("   💀 %d organisms perished", deaths), nil)
					}
				}
			}

			// Update geology
			geology.SimulateGeology(10000)
		}
	}

	// Get final statistics
	geoStats := geology.GetStats()
	totalPop, totalSpecies, totalExtinct := popSim.GetStats()

	// Build summary
	var sb strings.Builder
	sb.WriteString("=== Simulation Complete ===\n")
	sb.WriteString(fmt.Sprintf("Years Simulated: %d\n", years))
	sb.WriteString(fmt.Sprintf("Total Population: %d\n", totalPop))
	sb.WriteString(fmt.Sprintf("Living Species: %d\n", totalSpecies))
	sb.WriteString(fmt.Sprintf("Extinct Species: %d\n", totalExtinct))
	sb.WriteString(fmt.Sprintf("Geological Events: %d\n", geologicalEvents))
	sb.WriteString("--- Terrain Stats ---\n")
	sb.WriteString(fmt.Sprintf("Tectonic Plates: %d\n", geoStats.PlateCount))
	sb.WriteString(fmt.Sprintf("Avg Elevation: %.0fm\n", geoStats.AverageElevation))
	sb.WriteString(fmt.Sprintf("Max Elevation: %.0fm\n", geoStats.MaxElevation))
	sb.WriteString(fmt.Sprintf("Sea Level: %.0fm\n", geoStats.SeaLevel))
	sb.WriteString(fmt.Sprintf("Land Coverage: %.1f%%\n", geoStats.LandPercent))

	// Species breakdown grouped by biome type
	sb.WriteString("--- Species by Biome Type ---\n")

	// Aggregate by biome type
	type biomeTypeStats struct {
		count      int
		population int64
		species    map[string]struct {
			count      int64
			generation int64
		}
	}
	biomeTypeMap := make(map[string]*biomeTypeStats)

	for _, biome := range popSim.Biomes {
		biomeTypeName := string(biome.BiomeType)
		if _, exists := biomeTypeMap[biomeTypeName]; !exists {
			biomeTypeMap[biomeTypeName] = &biomeTypeStats{
				species: make(map[string]struct {
					count      int64
					generation int64
				}),
			}
		}
		stats := biomeTypeMap[biomeTypeName]
		stats.count++
		stats.population += biome.TotalPopulation()

		for _, sp := range biome.Species {
			// Use base species name (without biome prefix for cleaner display)
			existing := stats.species[sp.Name]
			existing.count += sp.Count
			if sp.Generation > existing.generation {
				existing.generation = sp.Generation
			}
			stats.species[sp.Name] = existing
		}
	}

	// Output grouped stats
	for biomeType, stats := range biomeTypeMap {
		sb.WriteString(fmt.Sprintf("%s (%d biomes, Pop: %d):\n", biomeType, stats.count, stats.population))
		speciesShown := 0
		for name, sp := range stats.species {
			if speciesShown >= 5 {
				sb.WriteString(fmt.Sprintf("  ...and %d more species\n", len(stats.species)-5))
				break
			}
			sb.WriteString(fmt.Sprintf("  %s: %d (Gen %d)\n", name, sp.count, sp.generation))
			speciesShown++
		}
	}

	// Fossil record
	if len(popSim.FossilRecord.Extinct) > 0 {
		sb.WriteString("--- Fossil Record ---\n")
		shown := 0
		for _, ext := range popSim.FossilRecord.Extinct {
			if shown >= 5 {
				sb.WriteString(fmt.Sprintf("...and %d more extinct species\n", len(popSim.FossilRecord.Extinct)-5))
				break
			}
			duration := ext.ExistedUntil - ext.ExistedFrom
			sb.WriteString(fmt.Sprintf("† %s (existed %d years, cause: %s)\n", ext.Name, duration, ext.ExtinctionCause))
			shown++
		}
	}

	client.SendGameMessage("system", sb.String(), nil)
	return nil
}

// handleWorldInfo shows current world state
func (p *GameProcessor) handleWorldInfo(ctx context.Context, client websocket.GameClient) error {
	char, err := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if err != nil {
		client.SendGameMessage("error", "Could not get character info", nil)
		return nil
	}

	world, err := p.worldRepo.GetWorld(ctx, char.WorldID)
	if err != nil {
		client.SendGameMessage("error", "Could not get world info", nil)
		return nil
	}

	var sb strings.Builder
	sb.WriteString("=== World Info ===\n")
	sb.WriteString(fmt.Sprintf("Name: %s\n", world.Name))
	sb.WriteString(fmt.Sprintf("ID: %s\n", world.ID))
	if world.Circumference != nil {
		circumKm := *world.Circumference / 1000
		sb.WriteString(fmt.Sprintf("Circumference: %.0f km\n", circumKm))
	}
	sb.WriteString(fmt.Sprintf("Entities: %d\n", len(p.ecosystemService.Entities)))

	// Show terrain stats if geology has been simulated
	if geology, exists := p.worldGeology[char.WorldID]; exists && geology.IsInitialized() {
		geoStats := geology.GetStats()
		sb.WriteString("--- Terrain ---\n")
		sb.WriteString(fmt.Sprintf("Tectonic Plates: %d\n", geoStats.PlateCount))
		sb.WriteString(fmt.Sprintf("Avg Elevation: %.0fm\n", geoStats.AverageElevation))
		sb.WriteString(fmt.Sprintf("Max Elevation: %.0fm\n", geoStats.MaxElevation))
		sb.WriteString(fmt.Sprintf("Min Elevation: %.0fm\n", geoStats.MinElevation))
		sb.WriteString(fmt.Sprintf("Sea Level: %.0fm\n", geoStats.SeaLevel))
		sb.WriteString(fmt.Sprintf("Land Coverage: %.1f%%\n", geoStats.LandPercent))
		sb.WriteString(fmt.Sprintf("Years Simulated: %d\n", geoStats.YearsSimulated))
	} else {
		sb.WriteString("--- Terrain ---\n")
		sb.WriteString("Not yet simulated. Use 'world simulate <years>' to generate terrain.\n")
	}

	client.SendGameMessage("system", sb.String(), nil)
	return nil
}

// handleWorldReset resets the world simulation to default state
func (p *GameProcessor) handleWorldReset(ctx context.Context, client websocket.GameClient) error {
	char, err := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if err != nil {
		client.SendGameMessage("error", "Could not get character info", nil)
		return nil
	}

	// Clear geology for this world
	delete(p.worldGeology, char.WorldID)

	// Clear all entities
	p.ecosystemService.Entities = make(map[uuid.UUID]*state.LivingEntityState)
	p.ecosystemService.Behaviors = make(map[uuid.UUID]behaviortree.Node)

	client.SendGameMessage("system", "🔄 World reset complete. Geology and entities cleared.\nUse 'world simulate <years>' to start fresh.", nil)
	return nil
}

// getSeasonFromYear calculates season from simulated year for weather simulation
func (p *GameProcessor) getSeasonFromYear(simulatedYear int64) weather.Season {
	// Cycle through seasons: 4 seasons per year
	seasonIndex := simulatedYear % 4
	switch seasonIndex {
	case 0:
		return weather.SeasonSpring
	case 1:
		return weather.SeasonSummer
	case 2:
		return weather.SeasonFall
	default:
		return weather.SeasonWinter
	}
}

// processReproduction handles reproduction for entities with high reproduction urge
func (p *GameProcessor) processReproduction() {
	em := p.ecosystemService.GetEvolutionManager()
	if em == nil {
		return
	}

	// Population cap - no reproduction if at capacity
	const maxPopulation = 2000
	if len(p.ecosystemService.Entities) >= maxPopulation {
		return
	}

	// Collect entities ready to reproduce (urge > 80)
	var readyToMate []*state.LivingEntityState
	for _, e := range p.ecosystemService.Entities {
		if e.Needs.ReproductionUrge > 80 {
			readyToMate = append(readyToMate, e)
		}
	}

	// Match pairs of same species
	mated := make(map[uuid.UUID]bool)
	for i, e1 := range readyToMate {
		if mated[e1.EntityID] {
			continue
		}
		for j := i + 1; j < len(readyToMate); j++ {
			e2 := readyToMate[j]
			if mated[e2.EntityID] {
				continue
			}
			// Same species required
			if e1.Species != e2.Species {
				continue
			}

			// Reproduce
			child, err := em.Reproduce(e1, e2)
			if err != nil {
				continue
			}

			// Add child to ecosystem
			p.ecosystemService.Entities[child.EntityID] = child

			// Reset parents' urge
			e1.Needs.ReproductionUrge = 0
			e2.Needs.ReproductionUrge = 0

			mated[e1.EntityID] = true
			mated[e2.EntityID] = true
			break
		}
	}
}
