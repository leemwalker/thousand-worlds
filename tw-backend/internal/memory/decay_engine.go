package memory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"

	"tw-backend/internal/repository"
)

const (
	BaseDecayRate      = 0.05 // 5% per day? Adjust as needed.
	ImportanceThreshold = 0.2
	SummaryTemplate    = "Memory faded into a vague recollection."
)

type DecayEngine struct {
	nc   *nats.Conn
	repo *repository.NPCMemoryRepository
}

func NewDecayEngine(nc *nats.Conn, repo *repository.NPCMemoryRepository) *DecayEngine {
	return &DecayEngine{
		nc:   nc,
		repo: repo,
	}
}

type DayTick struct {
	WorldID     string  `json:"worldID"`
	VirtualTime float64 `json:"virtualTime"`
}

// StartDecayJob listens for the daily tick and runs the decay simulation.
func (e *DecayEngine) StartDecayJob(ctx context.Context) error {
	_, err := e.nc.Subscribe("world.tick.day", func(msg *nats.Msg) {
		var tick DayTick
		if err := json.Unmarshal(msg.Data, &tick); err != nil {
			log.Error().Err(err).Msg("DecayEngine: failed to unmarshal tick")
			return
		}

		log.Info().Str("worldID", tick.WorldID).Float64("time", tick.VirtualTime).Msg("Running memory decay simulation")

		if err := e.runDecayCycle(ctx, tick.WorldID, tick.VirtualTime); err != nil {
			log.Error().Err(err).Msg("DecayEngine: decay cycle failed")
		}
	})

	if err != nil {
		return fmt.Errorf("DecayEngine.StartDecayJob: subscribe failed: %w", err)
	}

	return nil
}

func (e *DecayEngine) runDecayCycle(ctx context.Context, worldID string, currentVirtualTime float64) error {
	memories, err := e.repo.GetMemoriesByWorldID(ctx, worldID)
	if err != nil {
		return fmt.Errorf("get memories failed: %w", err)
	}

	for _, mem := range memories {
		// Calculate time passed
		timePassed := currentVirtualTime - mem.LastAccessedVirtualTime
		if timePassed <= 0 {
			continue
		}

		// Determine decay rate
		rate := BaseDecayRate
		if mem.Source == "NPC" {
			rate *= 5
		}

		// Apply decay
		decayAmount := timePassed * rate
		mem.ImportanceScore -= decayAmount

		// Clamp score
		if mem.ImportanceScore < 0 {
			mem.ImportanceScore = 0
		}

		updated := false
		// Check for crystallization/downgrade
		if mem.ImportanceScore < ImportanceThreshold && mem.DetailLevel > 1 {
			mem.DetailLevel--
			mem.Content = SummaryTemplate // Placeholder for summary
			// Optionally bump score slightly so it doesn't immediately degrade again next tick if we had more levels?
			// But here we just check threshold.
			updated = true
			log.Info().Str("memoryID", mem.ID).Int("newLevel", mem.DetailLevel).Msg("Memory crystallized")
		} else if decayAmount > 0 {
			// Even if not crystallized, we need to save the new score
			updated = true
		}

		if updated {
			// In a real system, we might batch updates.
			if err := e.repo.UpdateMemory(ctx, mem); err != nil {
				log.Error().Err(err).Str("id", mem.ID).Msg("Failed to update memory")
			}
		}
	}

	return nil
}
