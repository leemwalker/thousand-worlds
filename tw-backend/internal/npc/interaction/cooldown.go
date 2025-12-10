package interaction

import (
	"tw-backend/internal/npc/personality"
	"time"

	"github.com/google/uuid"
)

// Cooldown Constants
const (
	BaseCooldownMinutes = 5.0
)

// CooldownTracker manages conversation cooldowns
type CooldownTracker struct {
	LastConversation map[uuid.UUID]time.Time
}

// NewCooldownTracker creates a new tracker
func NewCooldownTracker() *CooldownTracker {
	return &CooldownTracker{
		LastConversation: make(map[uuid.UUID]time.Time),
	}
}

// CheckCooldown returns true if the NPC is ready for a new conversation
func (ct *CooldownTracker) CheckCooldown(npcID uuid.UUID, p *personality.Personality, currentTime time.Time) bool {
	lastTime, ok := ct.LastConversation[npcID]
	if !ok {
		return true // No previous conversation
	}

	// Calculate Cooldown Duration
	// 5min * (1 - extraversion/200)
	// High E (100) -> 5 * 0.5 = 2.5 min
	// Low E (0) -> 5 * 1.0 = 5.0 min
	extraversionFactor := p.Extraversion.Value
	cooldownDuration := BaseCooldownMinutes * (1.0 - (extraversionFactor / 200.0))

	// Convert to Duration
	duration := time.Duration(cooldownDuration * float64(time.Minute))

	return currentTime.Sub(lastTime) >= duration
}

// SetCooldown updates the last conversation time
func (ct *CooldownTracker) SetCooldown(npcID uuid.UUID, currentTime time.Time) {
	ct.LastConversation[npcID] = currentTime
}
