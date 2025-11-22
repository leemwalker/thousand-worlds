package memory

import (
	"time"
)

// CalculateRehearsalBonus returns the decay reduction factor based on access count
// Formula: min(accessCount / 20, 0.5)
func CalculateRehearsalBonus(accessCount int) float64 {
	bonus := float64(accessCount) / 20.0
	if bonus > 0.5 {
		bonus = 0.5
	}
	return bonus
}

// RecordAccess updates the memory's access stats
func RecordAccess(memory *Memory, now time.Time) {
	memory.AccessCount++
	memory.LastAccessed = now
}
