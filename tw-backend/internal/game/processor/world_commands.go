package processor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/ecosystem/state"
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
	default:
		client.SendGameMessage("error", "Unknown world command. Try: 'simulate', 'info'", nil)
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

	// Conversion: 100 ticks = 1 year
	ticksPerYear := int64(100)
	totalTicks := years * ticksPerYear

	// Cap at 100M ticks to prevent infinite runs (1M years max)
	maxTicks := int64(100_000_000)
	if totalTicks > maxTicks {
		totalTicks = maxTicks
		client.SendGameMessage("system", fmt.Sprintf("Capping simulation to %d years (limit).", maxTicks/ticksPerYear), nil)
	}

	client.SendGameMessage("system", fmt.Sprintf("Starting simulation of %d years (%d ticks)...", years, totalTicks), nil)

	// Track statistics
	startEntities := len(p.ecosystemService.Entities)
	extinctions := 0
	generations := 0
	geologicalEvents := 0

	// Create geological event manager for this simulation
	geoManager := ecosystem.NewGeologicalEventManager()

	// Simulation loop
	batchSize := int64(10000)
	weatherUpdateInterval := int64(100)     // Every 100 ticks (1 year)
	geologyUpdateInterval := int64(1000000) // Every 1M ticks (10,000 years)
	lastWeatherUpdate := int64(0)
	lastGeologyUpdate := int64(0)

	for tick := int64(0); tick < totalTicks; tick += batchSize {
		remaining := totalTicks - tick
		currentBatch := batchSize
		if remaining < currentBatch {
			currentBatch = remaining
		}

		// Check for geological events at batch level
		previousEventCount := len(geoManager.ActiveEvents)
		geoManager.CheckForNewEvents(tick, currentBatch)
		newEvents := len(geoManager.ActiveEvents) - previousEventCount
		geologicalEvents += newEvents

		// Report and apply terrain effects from new events
		if newEvents > 0 {
			for i := len(geoManager.ActiveEvents) - newEvents; i < len(geoManager.ActiveEvents); i++ {
				e := geoManager.ActiveEvents[i]
				client.SendGameMessage("system", fmt.Sprintf("⚠️ GEOLOGICAL EVENT: %s (severity: %.0f%%)", e.Type, e.Severity*100), nil)

				// Apply terrain effects from event
				geology.ApplyEvent(e)
			}
		}

		// Get environment modifiers from active events
		tempMod, sunlightMod, _ := geoManager.GetEnvironmentModifiers()

		// Run ecosystem ticks with environmental pressure
		for i := int64(0); i < currentBatch; i++ {
			// TODO: Apply tempMod and sunlightMod to ecosystem tick
			// For now, just track temperature affecting death rates
			_ = tempMod
			_ = sunlightMod

			p.ecosystemService.Tick()

			// Track extinctions
			currentCount := len(p.ecosystemService.Entities)
			if currentCount < startEntities {
				extinctions += startEntities - currentCount
				startEntities = currentCount
			}

			// Update weather periodically (every simulated year)
			if tick+i-lastWeatherUpdate >= weatherUpdateInterval {
				if p.weatherService != nil {
					// Calculate simulated season based on tick
					simulatedYear := (tick + i) / 100
					season := p.getSeasonFromYear(simulatedYear)
					_, _ = p.weatherService.UpdateWorldWeather(ctx, char.WorldID, time.Now(), season)
				}
				lastWeatherUpdate = tick + i
			}
		}

		// Simulate geology periodically (every 10,000 simulated years)
		if tick+currentBatch-lastGeologyUpdate >= geologyUpdateInterval {
			yearsElapsed := (tick + currentBatch - lastGeologyUpdate) / ticksPerYear
			geology.SimulateGeology(yearsElapsed)
			lastGeologyUpdate = tick + currentBatch
		}

		// Reproduction pass: entities with high reproduction urge can mate
		// Process at batch level for performance
		p.processReproduction()

		// Update max generation
		for _, e := range p.ecosystemService.Entities {
			if e.Generation > generations {
				generations = e.Generation
			}
		}

		// Update active events (expire old ones)
		geoManager.UpdateActiveEvents(tick + currentBatch)
	}

	// Get geology stats for summary
	geoStats := geology.GetStats()

	// Summary
	finalCount := len(p.ecosystemService.Entities)
	var sb strings.Builder
	sb.WriteString("=== Simulation Complete ===\n")
	sb.WriteString(fmt.Sprintf("Years Simulated: %d\n", years))
	sb.WriteString(fmt.Sprintf("Ticks Processed: %d\n", totalTicks))
	sb.WriteString(fmt.Sprintf("Remaining Entities: %d\n", finalCount))
	sb.WriteString(fmt.Sprintf("Extinctions: %d\n", extinctions))
	sb.WriteString(fmt.Sprintf("Max Generation: %d\n", generations))
	sb.WriteString(fmt.Sprintf("Geological Events: %d\n", geologicalEvents))
	sb.WriteString("--- Terrain Stats ---\n")
	sb.WriteString(fmt.Sprintf("Tectonic Plates: %d\n", geoStats.PlateCount))
	sb.WriteString(fmt.Sprintf("Avg Elevation: %.0fm\n", geoStats.AverageElevation))
	sb.WriteString(fmt.Sprintf("Max Elevation: %.0fm\n", geoStats.MaxElevation))
	sb.WriteString(fmt.Sprintf("Sea Level: %.0fm\n", geoStats.SeaLevel))
	sb.WriteString(fmt.Sprintf("Land Coverage: %.1f%%\n", geoStats.LandPercent))

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
