package interaction

import (
	"math/rand"
	"mud-platform-backend/internal/npc/personality"
)

// Initiation Constants
const (
	BaseInitiationChance = 0.1 // 10% per tick
	ProximityThreshold   = 5.0 // Meters
)

// ShouldInitiateConversation determines if a conversation should start
func ShouldInitiateConversation(ctx InteractionContext, p *personality.Personality, companionshipNeed float64) bool {
	// 1. Check Proximity
	if ctx.Distance > ProximityThreshold {
		return false
	}

	// 2. Check Cooldown (handled externally usually, but good to note)
	// Assuming cooldown check passed before calling this

	// 3. Calculate Probability
	// probability = baseChance × extraversionMultiplier × relationshipBonus × needUrgency

	// Extraversion Multiplier: 1.0 + (extraversion / 100)
	extraversionMult := 1.0 + (p.Extraversion.Value / 100.0)

	// Relationship Bonus: 1.0 + (affection / 200)
	// Affection is -100 to 100.
	// If 100 -> 1.5. If -100 -> 0.5.
	relationshipBonus := 1.0 + (float64(ctx.InitiatorAffection) / 200.0)

	// Need Urgency: 1.0 + (companionship / 100)
	needUrgency := 1.0 + (companionshipNeed / 100.0)

	probability := BaseInitiationChance * extraversionMult * relationshipBonus * needUrgency

	// 4. Roll
	return rand.Float64() < probability
}
