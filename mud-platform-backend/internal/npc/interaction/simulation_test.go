package interaction

import (
	"mud-platform-backend/internal/npc/personality"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInteractionSimulation(t *testing.T) {
	// Setup 10 NPCs
	count := 10
	npcs := make([]*personality.Personality, count)
	ids := make([]uuid.UUID, count)
	cooldowns := NewCooldownTracker()

	for i := 0; i < count; i++ {
		npcs[i] = personality.NewPersonality()
		// Vary Extraversion
		npcs[i].Extraversion.Value = float64(i * 10) // 0 to 90
		ids[i] = uuid.New()
	}

	// Simulate 1 Hour (60 ticks of 1 minute)
	// Just simulate logic loops
	interactions := 0

	currentTime := time.Now()

	for tick := 0; tick < 60; tick++ {
		currentTime = currentTime.Add(time.Minute)

		// Check pairs
		for i := 0; i < count; i++ {
			for j := i + 1; j < count; j++ {
				// Context
				ctx := InteractionContext{
					Distance:           2.0, // Close
					InitiatorAffection: 50,
				}

				// Check Cooldown
				if !cooldowns.CheckCooldown(ids[i], npcs[i], currentTime) {
					continue
				}

				// Check Initiation
				if ShouldInitiateConversation(ctx, npcs[i], 50.0) {
					interactions++
					// Set Cooldown
					cooldowns.SetCooldown(ids[i], currentTime)
					cooldowns.SetCooldown(ids[j], currentTime)
				}
			}
		}
	}

	// Verify interactions occurred
	if interactions == 0 {
		t.Error("No interactions occurred in simulation")
	}

	// High Extraversion should have more?
	// Hard to track individual counts in this simple loop without map
	// But total interactions should be reasonable.
	t.Logf("Total interactions in 1 hour: %d", interactions)
}
