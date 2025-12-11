package processor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/worldgen/weather"
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
	weatherUpdateInterval := int64(100) // Every 100 ticks (1 year)
	lastWeatherUpdate := int64(0)

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

		// Report major events
		if newEvents > 0 {
			for i := len(geoManager.ActiveEvents) - newEvents; i < len(geoManager.ActiveEvents); i++ {
				e := geoManager.ActiveEvents[i]
				client.SendGameMessage("system", fmt.Sprintf("⚠️ GEOLOGICAL EVENT: %s (severity: %.0f%%)", e.Type, e.Severity*100), nil)
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
					p.weatherService.UpdateWorldWeather(ctx, char.WorldID, time.Now(), season)
				}
				lastWeatherUpdate = tick + i
			}
		}

		// Update max generation
		for _, e := range p.ecosystemService.Entities {
			if e.Generation > generations {
				generations = e.Generation
			}
		}

		// Update active events (expire old ones)
		geoManager.UpdateActiveEvents(tick + currentBatch)
	}

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
