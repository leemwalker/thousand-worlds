package skills

import (
	"time"
)

// DiminishingReturnsTracker tracks recent actions to reduce XP gain
type DiminishingReturnsTracker struct {
	History map[string][]time.Time // Key: ActionID (e.g., "MineRock:123"), Value: Timestamps
}

func NewDiminishingReturnsTracker() *DiminishingReturnsTracker {
	return &DiminishingReturnsTracker{
		History: make(map[string][]time.Time),
	}
}

// CalculateDiminishingReturn calculates the XP multiplier based on recent history
// Example: 10 uses in last minute = 10% XP
func (drt *DiminishingReturnsTracker) CalculateDiminishingReturn(actionID string) float64 {
	now := time.Now()
	window := 1 * time.Minute

	// Clean up old history
	if _, exists := drt.History[actionID]; !exists {
		drt.History[actionID] = []time.Time{}
	}

	// Filter timestamps within window
	validTimestamps := []time.Time{}
	for _, ts := range drt.History[actionID] {
		if now.Sub(ts) < window {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	drt.History[actionID] = validTimestamps

	count := len(validTimestamps)

	// Add current action
	drt.History[actionID] = append(drt.History[actionID], now)

	// Calculate multiplier
	// 0-1 uses: 1.0
	// 2-5 uses: 0.8
	// 6-9 uses: 0.5
	// 10+ uses: 0.1

	if count < 2 {
		return 1.0
	} else if count < 6 {
		return 0.8
	} else if count < 10 {
		return 0.5
	} else {
		return 0.1
	}
}
