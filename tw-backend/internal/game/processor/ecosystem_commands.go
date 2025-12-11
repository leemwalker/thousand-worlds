package processor

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/ai/behaviortree"
	"tw-backend/internal/ecosystem/state"
)

// handleEcosystem handles ecosystem debug and interaction commands
func (p *GameProcessor) handleEcosystem(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	if cmd.Target == nil {
		client.SendGameMessage("error", "Usage: ecosystem <action> [args]", nil)
		return nil
	}

	subCmd := strings.ToLower(*cmd.Target)

	switch subCmd {
	case "status":
		return p.handleEcosystemStatus(ctx, client)
	case "spawn":
		// Example: ecosystem spawn rabbit
		if cmd.Message == nil {
			client.SendGameMessage("error", "Usage: ecosystem spawn <species>", nil)
			return nil
		}
		return p.handleEcosystemSpawn(ctx, client, *cmd.Message)
	case "log", "logs":
		// Example: ecosystem log <id>
		if cmd.Message == nil {
			client.SendGameMessage("error", "Usage: ecosystem log <entity_id>", nil)
			return nil
		}
		return p.handleEcosystemLog(ctx, client, *cmd.Message)
	case "lineage":
		// Example: ecosystem lineage <id>
		if cmd.Message == nil {
			client.SendGameMessage("error", "Usage: ecosystem lineage <entity_id>", nil)
			return nil
		}
		return p.handleEcosystemLineage(ctx, client, *cmd.Message)
	case "breed":
		// Example: ecosystem breed <id1> <id2>
		if cmd.Message == nil {
			client.SendGameMessage("error", "Usage: ecosystem breed <id1> <id2>", nil)
			return nil
		}
		return p.handleEcosystemBreed(ctx, client, *cmd.Message)
	default:
		client.SendGameMessage("error", "Unknown ecosystem command. Try 'status', 'spawn', 'log', 'lineage', or 'breed'.", nil)
		return nil
	}
}

func (p *GameProcessor) handleEcosystemStatus(ctx context.Context, client websocket.GameClient) error {
	// For MVP, just dump all entities irrespective of location
	// Later filter by client.GetWorldID()

	count := len(p.ecosystemService.Entities)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== Ecosystem Status (%d entities) ===\n", count))

	for _, e := range p.ecosystemService.Entities {
		// Basic info
		info := fmt.Sprintf("[%s] %s (Age: %d) - Hunger: %.1f\n",
			e.Species, e.EntityID.String()[:8], e.Age, e.Needs.Hunger)
		sb.WriteString(info)
	}

	client.SendGameMessage("system", sb.String(), nil)
	return nil
}

func (p *GameProcessor) handleEcosystemSpawn(ctx context.Context, client websocket.GameClient, speciesStr string) error {
	// Special case: spawn all Precambrian life
	if strings.ToLower(speciesStr) == "precambrian" {
		return p.handleEcosystemSpawnPrecambrian(ctx, client)
	}

	// Validate Species
	var sp state.Species
	switch strings.ToLower(speciesStr) {
	case "rabbit":
		sp = state.SpeciesRabbit
	case "wolf":
		sp = state.SpeciesWolf
	case "deer":
		sp = state.SpeciesDeer
	case "cactus":
		sp = state.SpeciesCactus
	case "oak":
		sp = state.SpeciesOak
	case "fern":
		sp = state.SpeciesFern
	case "grass":
		sp = state.SpeciesGrass
	// Precambrian species
	case "cyanobacteria":
		sp = state.SpeciesCyanobacteria
	case "stromatolite":
		sp = state.SpeciesStromatolite
	case "ediacaran":
		sp = state.SpeciesEdiacaran
	case "dickinsonia":
		sp = state.SpeciesDickinsonia
	case "charnia":
		sp = state.SpeciesCharnia
	default:
		client.SendGameMessage("error", fmt.Sprintf("Unknown species '%s'. Try: precambrian, cyanobacteria, stromatolite, ediacaran, dickinsonia, charnia, rabbit, wolf, deer", speciesStr), nil)
		return nil
	}

	// Create entity
	ent := p.ecosystemService.Spawner.CreateEntity(sp, 1)

	// Assign location from player
	char, err := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if err == nil && char != nil {
		ent.WorldID = char.WorldID
		ent.PositionX = char.PositionX
		ent.PositionY = char.PositionY
	}

	// Add to entities
	p.ecosystemService.Entities[ent.EntityID] = ent

	// Assign behavior tree based on diet
	switch ent.Diet {
	case state.DietPhotosynthetic:
		p.ecosystemService.Behaviors[ent.EntityID] = behaviortree.NewFloraTree()
	default:
		p.ecosystemService.Behaviors[ent.EntityID] = behaviortree.NewHerbivoreTree()
	}

	client.SendGameMessage("system", fmt.Sprintf("Spawned %s at your location.", sp), nil)
	return nil
}

func (p *GameProcessor) handleEcosystemLog(ctx context.Context, client websocket.GameClient, targetID string) error {
	// Find entity by partial ID
	var targetEntity *state.LivingEntityState
	for id, e := range p.ecosystemService.Entities {
		idStr := id.String()
		if strings.HasPrefix(idStr, targetID) {
			targetEntity = e
			break
		}
	}

	if targetEntity == nil {
		client.SendGameMessage("error", fmt.Sprintf("Entity not found: %s", targetID), nil)
		return nil
	}

	// Format logs
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== Decision Logs for %s (%s) ===\n", targetEntity.EntityID.String()[:8], targetEntity.Species))

	if len(targetEntity.Logs) == 0 {
		sb.WriteString("No logs recorded.\n")
	} else {
		for _, l := range targetEntity.Logs {
			// Format time as rough relative or simple timestamp
			// For simplicity just using HH:MM:SS
			t := time.Unix(l.Timestamp, 0)
			timeStr := t.Format("15:04:05")
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", timeStr, l.Action, l.Reason))
		}
	}

	client.SendGameMessage("system", sb.String(), nil)
	return nil
}

func (p *GameProcessor) handleEcosystemLineage(ctx context.Context, client websocket.GameClient, targetID string) error {
	// Find entity by partial ID
	var targetEntity *state.LivingEntityState
	for id, e := range p.ecosystemService.Entities {
		idStr := id.String()
		if strings.HasPrefix(idStr, targetID) {
			targetEntity = e
			break
		}
	}

	if targetEntity == nil {
		client.SendGameMessage("error", fmt.Sprintf("Entity not found: %s", targetID), nil)
		return nil
	}

	// Format lineage
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== Lineage for %s (%s) ===\n", targetEntity.EntityID.String()[:8], targetEntity.Species))
	sb.WriteString(fmt.Sprintf("Generation: %d\n", targetEntity.Generation))

	if targetEntity.Parent1ID == nil && targetEntity.Parent2ID == nil {
		sb.WriteString("Parents: None (First Generation)\n")
	} else {
		if targetEntity.Parent1ID != nil {
			parent1Str := targetEntity.Parent1ID.String()[:8]
			// Check if parent still exists
			if p1, ok := p.ecosystemService.Entities[*targetEntity.Parent1ID]; ok {
				sb.WriteString(fmt.Sprintf("Parent 1: %s (%s)\n", parent1Str, p1.Species))
			} else {
				sb.WriteString(fmt.Sprintf("Parent 1: %s (deceased)\n", parent1Str))
			}
		}
		if targetEntity.Parent2ID != nil {
			parent2Str := targetEntity.Parent2ID.String()[:8]
			if p2, ok := p.ecosystemService.Entities[*targetEntity.Parent2ID]; ok {
				sb.WriteString(fmt.Sprintf("Parent 2: %s (%s)\n", parent2Str, p2.Species))
			} else {
				sb.WriteString(fmt.Sprintf("Parent 2: %s (deceased)\n", parent2Str))
			}
		}
	}

	client.SendGameMessage("system", sb.String(), nil)
	return nil
}

func (p *GameProcessor) handleEcosystemBreed(ctx context.Context, client websocket.GameClient, args string) error {
	// Parse args: "id1 id2"
	parts := strings.Fields(args)
	if len(parts) < 2 {
		client.SendGameMessage("error", "Usage: ecosystem breed <id1> <id2>", nil)
		return nil
	}

	id1Str := parts[0]
	id2Str := parts[1]

	// Find entities
	var parent1, parent2 *state.LivingEntityState
	for id, e := range p.ecosystemService.Entities {
		idStr := id.String()
		if strings.HasPrefix(idStr, id1Str) {
			parent1 = e
		}
		if strings.HasPrefix(idStr, id2Str) {
			parent2 = e
		}
		if parent1 != nil && parent2 != nil {
			break
		}
	}

	if parent1 == nil {
		client.SendGameMessage("error", fmt.Sprintf("Entity not found: %s", id1Str), nil)
		return nil
	}
	if parent2 == nil {
		client.SendGameMessage("error", fmt.Sprintf("Entity not found: %s", id2Str), nil)
		return nil
	}

	// Create offspring
	em := p.ecosystemService.GetEvolutionManager()
	child, err := em.Reproduce(parent1, parent2)
	if err != nil {
		client.SendGameMessage("error", fmt.Sprintf("Breeding failed: %v", err), nil)
		return nil
	}

	// Set location to average of parents
	child.WorldID = parent1.WorldID
	child.PositionX = (parent1.PositionX + parent2.PositionX) / 2
	child.PositionY = (parent1.PositionY + parent2.PositionY) / 2

	// Add to ecosystem
	p.ecosystemService.Entities[child.EntityID] = child

	client.SendGameMessage("system", fmt.Sprintf("Bred %s! Offspring ID: %s (Generation %d)", child.Species, child.EntityID.String()[:8], child.Generation), nil)
	return nil
}

// handleEcosystemSpawnPrecambrian spawns a balanced Precambrian ecosystem
// Population ratios based on ecological principles:
// - Producers (cyanobacteria, stromatolites) vastly outnumber consumers (~10:1)
// - Base population scales with world circumference
func (p *GameProcessor) handleEcosystemSpawnPrecambrian(ctx context.Context, client websocket.GameClient) error {
	char, err := p.authRepo.GetCharacter(ctx, client.GetCharacterID())
	if err != nil {
		client.SendGameMessage("error", "Could not get character", nil)
		return nil
	}

	// Get world circumference for population scaling
	// Default: ~1000 entities for default world
	basePopulation := 100
	world, err := p.worldRepo.GetWorld(ctx, char.WorldID)
	if err == nil && world != nil && world.Circumference != nil {
		// Scale: 1 entity per 50,000 meters of circumference
		// Half-moon (5.46M m) → ~109 base entities
		basePopulation = int(*world.Circumference / 50000)
		if basePopulation < 50 {
			basePopulation = 50
		}
		if basePopulation > 500 {
			basePopulation = 500
		}
	}

	// Precambrian ecosystem ratios (total = 100%)
	// Producers (photosynthetic): 80%
	// Filter feeders/detritivores: 20%
	populations := map[state.Species]int{
		state.SpeciesCyanobacteria: basePopulation * 50 / 100, // 50% - primary producers
		state.SpeciesStromatolite:  basePopulation * 30 / 100, // 30% - colony formers
		state.SpeciesEdiacaran:     basePopulation * 10 / 100, // 10% - filter feeders
		state.SpeciesDickinsonia:   basePopulation * 6 / 100,  // 6%  - detritivores
		state.SpeciesCharnia:       basePopulation * 4 / 100,  // 4%  - sessile filter feeders
	}

	// Spawn spread across world (random positions within reasonable range)
	spreadRadius := 1000.0 // meters from player
	totalSpawned := 0

	for species, count := range populations {
		if count < 1 {
			count = 1
		}
		for i := 0; i < count; i++ {
			ent := p.ecosystemService.Spawner.CreateEntity(species, 1)
			ent.WorldID = char.WorldID

			// Random position around player
			ent.PositionX = char.PositionX + (rand.Float64()*2-1)*spreadRadius
			ent.PositionY = char.PositionY + (rand.Float64()*2-1)*spreadRadius

			p.ecosystemService.Entities[ent.EntityID] = ent
			totalSpawned++
		}
	}

	var sb strings.Builder
	sb.WriteString("🌊 Precambrian Era Initialized!\n")
	sb.WriteString(fmt.Sprintf("Total organisms spawned: %d\n", totalSpawned))
	sb.WriteString(fmt.Sprintf("  Cyanobacteria: %d\n", populations[state.SpeciesCyanobacteria]))
	sb.WriteString(fmt.Sprintf("  Stromatolites: %d\n", populations[state.SpeciesStromatolite]))
	sb.WriteString(fmt.Sprintf("  Ediacaran: %d\n", populations[state.SpeciesEdiacaran]))
	sb.WriteString(fmt.Sprintf("  Dickinsonia: %d\n", populations[state.SpeciesDickinsonia]))
	sb.WriteString(fmt.Sprintf("  Charnia: %d\n", populations[state.SpeciesCharnia]))
	sb.WriteString("\nUse 'ecosystem status' to view, 'world simulate <years>' to evolve.")

	client.SendGameMessage("system", sb.String(), nil)
	return nil
}
