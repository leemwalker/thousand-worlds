package processor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/repository"
	"tw-backend/internal/worldgen/orchestrator"
)

// handleEnterTropicalWorld handles the "enter_tropical_world" command
// It ensures a specific Tropical Test World exists and moves the player there.
func (p *GameProcessor) handleEnterTropicalWorld(ctx context.Context, client websocket.GameClient, cmd *websocket.CommandData) error {
	log.Printf("[PROCESSOR] Handling enter_tropical_world")

	// 1. Check if the Tropical World already exists
	// We'll use a specific name to identify it
	const tropicalWorldName = "Tropical Test World"

	// We might store the ID in a well-known place or look it up by name/owner
	// For simplicity, let's list worlds and look for the name
	worlds, err := p.worldRepo.ListWorlds(ctx)
	if err != nil {
		return fmt.Errorf("failed to list worlds: %w", err)
	}

	var tropicalWorld *repository.World
	for _, w := range worlds {
		if w.Name == tropicalWorldName {
			tropicalWorld = &w
			break
		}
	}

	// 2. If not found, generate it
	if tropicalWorld == nil {
		log.Printf("[PROCESSOR] Tropical world not found, generating...")
		client.SendGameMessage("system", "Generating Tropical Test World... this may take a moment.", nil)

		generatedWorld, err := p.generateTropicalWorld(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate tropical world: %w", err)
		}

		tropicalWorld = generatedWorld
		log.Printf("[PROCESSOR] Tropical world generated: %s", tropicalWorld.ID)
	} else {
		log.Printf("[PROCESSOR] Found existing Tropical world: %s", tropicalWorld.ID)

		// Ensure geology is loaded for existing world
		if _, exists := p.worldGeology[tropicalWorld.ID]; !exists {
			// In a real system we'd load from DB/Blob.
			// For this test command, if it's missing from memory, let's just regenerate it
			// (or handle it gracefully). Re-generation preserves ID but overwrites data.
			// Ideally we should persist to disk.
			// For now, let's regenerate to be safe if memory was lost (server restart).
			// This changes content but keeps the ID.
			// NOTE: This might be confusing if player is already there, but strictly better than empty map.
			log.Printf("[PROCESSOR] Restoring geology for existing world...")
			// Use the existing ID
			svc := orchestrator.NewGeneratorService()
			config := &TropicalWorldConfig{}
			// Use saved seed if available
			if seedVal, ok := tropicalWorld.Metadata["generation_seed"].(float64); ok {
				// config seed override logic needed? config.GetSeed() returns fixed value anyway.
				_ = seedVal
			}

			generated, err := svc.GenerateWorld(ctx, tropicalWorld.ID, config)
			if err == nil {
				geo := mapGeneratedToGeology(generated)
				// Re-apply physics params from metadata
				if dl, ok := tropicalWorld.Metadata["day_length_hours"].(float64); ok {
					geo.Params.DayLengthSec = dl * 3600.0
				}
				if tilt, ok := tropicalWorld.Metadata["axial_tilt"].(float64); ok {
					geo.Params.AxialTiltDeg = tilt
				}
				p.worldGeology[tropicalWorld.ID] = geo
				p.mapService.SetWorldGeology(tropicalWorld.ID, geo)
			}
		}
	}

	// 3. Move character to the tropical world
	charID := client.GetCharacterID()
	if charID == uuid.Nil {
		return ErrNoCharacter
	}

	char, err := p.authRepo.GetCharacter(ctx, charID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// Update location
	char.WorldID = tropicalWorld.ID
	char.PositionX = 0.0
	char.PositionY = 0.0
	char.OrientationZ = 0.0 // Reset vertical orientation if flying

	if err := p.authRepo.UpdateCharacter(ctx, char); err != nil {
		return fmt.Errorf("failed to update character location: %w", err)
	}

	client.SendGameMessage("system", fmt.Sprintf("Welcome to %s!", tropicalWorldName), nil)

	// Force a map refresh
	p.sendMapUpdate(ctx, client)

	return nil
}

// generateTropicalWorld orchestrates the creation of the test world
func (p *GameProcessor) generateTropicalWorld(ctx context.Context) (*repository.World, error) {
	// 1. Setup Config
	config := &TropicalWorldConfig{}

	// 2. Run Generation
	svc := orchestrator.NewGeneratorService() // Use defaults

	newWorldID := uuid.New()
	generated, err := svc.GenerateWorld(ctx, newWorldID, config)
	if err != nil {
		return nil, err
	}

	// 3. Convert to Repository World
	repoWorld := &repository.World{
		ID:        generated.WorldID,
		Name:      "Tropical Test World",
		OwnerID:   uuid.Nil, // System owned
		Shape:     repository.WorldShapeSphere,
		Radius:    float64Ptr(6371000), // Earth-like
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// 4. Set Metadata (DayLength, AxialTilt, etc.)
	repoWorld.Metadata["day_length_hours"] = 30.0
	repoWorld.Metadata["axial_tilt"] = 2.0 // Low tilt -> stable seasons
	repoWorld.Metadata["generation_seed"] = generated.Metadata.Seed
	repoWorld.Metadata["climate_profile"] = "tropical"

	// 5. Persist World Record
	if err := p.worldRepo.CreateWorld(ctx, repoWorld); err != nil {
		return nil, fmt.Errorf("failed to persist world: %w", err)
	}

	// 6. In-Memory State Update
	// Construct WorldGeology for runtime usage (map service, weather, etc.)
	geo := mapGeneratedToGeology(generated)

	// Apply physics parameters from metadata
	// Note: mapGeneratedToGeology creates default Earth params, we override here
	if dl, ok := repoWorld.Metadata["day_length_hours"].(float64); ok {
		geo.Params.DayLengthSec = dl * 3600.0
	}
	if tilt, ok := repoWorld.Metadata["axial_tilt"].(float64); ok {
		geo.Params.AxialTiltDeg = tilt
	}
	// Mass is default 1.0 for tropical test world
	geo.Params.MassKg = 5.972e24 // Earth mass
	geo.Params.RadiusM = 6371000 // Earth radius

	// Store in processor state
	p.worldGeology[generated.WorldID] = geo

	// register with map service for visualization
	p.mapService.SetWorldGeology(generated.WorldID, geo)

	return repoWorld, nil
}

// mapGeneratedToGeology converts orchestrator output to runtime ecosystem state
func mapGeneratedToGeology(gen *orchestrator.GeneratedWorld) *ecosystem.WorldGeology {
	// Create new geology instance
	// Circumference approx 40000km for earth-sized
	geo := ecosystem.NewWorldGeology(gen.WorldID, gen.Metadata.Seed, 40030000.0)

	// Copy geography data
	if gen.Geography != nil {
		geo.Heightmap = gen.Geography.Heightmap
		geo.BiomeIDs = gen.Geography.BiomeIDs
		geo.Temperatures = gen.Geography.Temperatures
		geo.Precipitations = gen.Geography.Precipitations
		geo.Rivers = gen.Geography.Rivers
		geo.Plates = gen.Geography.Plates
		geo.SeaLevel = gen.Metadata.SeaLevel
	}

	// Copy satellites
	geo.Satellites = gen.Satellites

	return geo
}

// TropicalWorldConfig implements orchestrator.WorldConfig
type TropicalWorldConfig struct{}

func (c *TropicalWorldConfig) GetPlanetSize() string                       { return "medium" }
func (c *TropicalWorldConfig) GetLandWaterRatio() string                   { return "30% land" }
func (c *TropicalWorldConfig) GetClimateRange() string                     { return "tropical, wet" } // High temp, high rain
func (c *TropicalWorldConfig) GetTechLevel() string                        { return "primitive" }
func (c *TropicalWorldConfig) GetMagicLevel() string                       { return "low" }
func (c *TropicalWorldConfig) GetGeologicalAge() string                    { return "mature" } // Stable
func (c *TropicalWorldConfig) GetSentientSpecies() []string                { return []string{} }
func (c *TropicalWorldConfig) GetResourceDistribution() map[string]float64 { return nil }
func (c *TropicalWorldConfig) GetSimulationFlags() map[string]bool {
	// Disable tectonics/catastrophes to keep it stable
	return map[string]bool{
		"simulate_geology": false,
		"simulate_life":    true,
	}
}
func (c *TropicalWorldConfig) GetSeaLevel() *float64 { return nil }
func (c *TropicalWorldConfig) GetSeed() *int64 {
	// Fixed seed for consistent "Tropical Test World" geography
	seed := int64(999999)
	return &seed
}
func (c *TropicalWorldConfig) GetNaturalSatellites() string { return "none" }

func float64Ptr(v float64) *float64 { return &v }
