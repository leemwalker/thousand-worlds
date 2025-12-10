package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetQualityTier(t *testing.T) {
	assert.Equal(t, QualityPoor, GetQualityTier(0))
	assert.Equal(t, QualityPoor, GetQualityTier(19))
	assert.Equal(t, QualityCommon, GetQualityTier(20))
	assert.Equal(t, QualityCommon, GetQualityTier(39))
	assert.Equal(t, QualityFine, GetQualityTier(40))
	assert.Equal(t, QualityMasterwork, GetQualityTier(60))
	assert.Equal(t, QualityLegendary, GetQualityTier(80))
	assert.Equal(t, QualityLegendary, GetQualityTier(100))
}

func TestDiminishingReturns(t *testing.T) {
	tracker := NewDiminishingReturnsTracker()
	actionID := "test_action"

	// 1st use
	assert.Equal(t, 1.0, tracker.CalculateDiminishingReturn(actionID))

	// 2nd use (Count was 1) -> < 2 is false? No, count is previous uses.
	// Logic: count = len(validTimestamps) BEFORE adding current.
	// 1st call: count 0 -> 1.0. History has 1.
	// 2nd call: count 1 -> 1.0. History has 2.
	// 3rd call: count 2 -> 0.8.

	assert.Equal(t, 1.0, tracker.CalculateDiminishingReturn(actionID)) // 2nd use

	// 3rd use
	assert.Equal(t, 0.8, tracker.CalculateDiminishingReturn(actionID))

	// Spam 3 more times (4, 5, 6) -> Total 6 in history
	tracker.CalculateDiminishingReturn(actionID)
	tracker.CalculateDiminishingReturn(actionID)
	tracker.CalculateDiminishingReturn(actionID)

	// 7th use (Count 6) -> 0.5
	assert.Equal(t, 0.5, tracker.CalculateDiminishingReturn(actionID))

	// Spam to 10
	tracker.CalculateDiminishingReturn(actionID)
	tracker.CalculateDiminishingReturn(actionID)
	tracker.CalculateDiminishingReturn(actionID)

	// 11th use (Count 10) -> 0.1
	assert.Equal(t, 0.1, tracker.CalculateDiminishingReturn(actionID))
}
