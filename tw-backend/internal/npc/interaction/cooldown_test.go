package interaction

import (
	"testing"
	"time"

	"tw-backend/internal/npc/personality"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewCooldownTracker(t *testing.T) {
	tracker := NewCooldownTracker()
	assert.NotNil(t, tracker)
	assert.NotNil(t, tracker.LastConversation)
}

func TestCheckCooldown_NoPreviousConversation(t *testing.T) {
	tracker := NewCooldownTracker()
	npcID := uuid.New()
	p := personality.NewPersonality()

	// Should be ready (no previous conversation)
	ready := tracker.CheckCooldown(npcID, p, time.Now())
	assert.True(t, ready)
}

func TestCheckCooldown_WithinCooldown(t *testing.T) {
	tracker := NewCooldownTracker()
	npcID := uuid.New()
	p := personality.NewPersonality()
	p.Extraversion.Value = 50 // Average

	now := time.Now()
	tracker.SetCooldown(npcID, now)

	// Check 2 minutes later (within 5min cooldown)
	laterTime := now.Add(2 * time.Minute)
	ready := tracker.CheckCooldown(npcID, p, laterTime)
	assert.False(t, ready, "Should still be in cooldown")
}

func TestCheckCooldown_AfterCooldown(t *testing.T) {
	tracker := NewCooldownTracker()
	npcID := uuid.New()
	p := personality.NewPersonality()
	p.Extraversion.Value = 0 // Low E = 5min cooldown

	now := time.Now()
	tracker.SetCooldown(npcID, now)

	// Check 6 minutes later (after 5min cooldown)
	laterTime := now.Add(6 * time.Minute)
	ready := tracker.CheckCooldown(npcID, p, laterTime)
	assert.True(t, ready, "Should be ready after cooldown expires")
}

func TestCheckCooldown_HighExtraversion(t *testing.T) {
	tracker := NewCooldownTracker()
	npcID := uuid.New()
	p := personality.NewPersonality()
	p.Extraversion.Value = 100 // High E = 2.5min cooldown

	now := time.Now()
	tracker.SetCooldown(npcID, now)

	// Check 3 minutes later (after 2.5min cooldown)
	laterTime := now.Add(3 * time.Minute)
	ready := tracker.CheckCooldown(npcID, p, laterTime)
	assert.True(t, ready, "High extraversion should have shorter cooldown")

	// But not ready at 2 minutes
	earlyTime := now.Add(2 * time.Minute)
	readyEarly := tracker.CheckCooldown(npcID, p, earlyTime)
	assert.False(t, readyEarly, "Should not be ready before cooldown expires")
}

func TestCheckCooldown_LowExtraversion(t *testing.T) {
	tracker := NewCooldownTracker()
	npcID := uuid.New()
	p := personality.NewPersonality()
	p.Extraversion.Value = 0 // Low E = 5min cooldown

	now := time.Now()
	tracker.SetCooldown(npcID, now)

	// Check 4 minutes later (before 5min cooldown)
	laterTime := now.Add(4 * time.Minute)
	ready := tracker.CheckCooldown(npcID, p, laterTime)
	assert.False(t, ready, "Low extraversion should have longer cooldown")
}

func TestCheckCooldown_ExactThreshold(t *testing.T) {
	tracker := NewCooldownTracker()
	npcID := uuid.New()
	p := personality.NewPersonality()
	p.Extraversion.Value = 0 // 5min cooldown

	now := time.Now()
	tracker.SetCooldown(npcID, now)

	// Check exactly 5 minutes later
	exactTime := now.Add(5 * time.Minute)
	ready := tracker.CheckCooldown(npcID, p, exactTime)
	assert.True(t, ready, "Should be ready at exact cooldown threshold")
}

func TestSetCooldown(t *testing.T) {
	tracker := NewCooldownTracker()
	npcID := uuid.New()
	testTime := time.Now()

	tracker.SetCooldown(npcID, testTime)

	// Verify it was set
	lastTime, ok := tracker.LastConversation[npcID]
	assert.True(t, ok)
	assert.Equal(t, testTime, lastTime)
}

func TestCheckCooldown_MultiplecNPCs(t *testing.T) {
	tracker := NewCooldownTracker()
	npc1 := uuid.New()
	npc2 := uuid.New()
	p := personality.NewPersonality()
	p.Extraversion.Value = 50

	now := time.Now()

	// Set cooldown for npc1 only
	tracker.SetCooldown(npc1, now)

	laterTime := now.Add(3 * time.Minute)

	// npc1 should not be ready
	ready1 := tracker.CheckCooldown(npc1, p, laterTime)
	assert.False(t, ready1)

	// npc2 should be ready (no cooldown)
	ready2 := tracker.CheckCooldown(npc2, p, laterTime)
	assert.True(t, ready2)
}

func TestCooldownFormula(t *testing.T) {
	// Test the exact formula: 5min * (1 - extraversion/200)
	tests := []struct {
		extraversion float64
		expectedMin  float64
	}{
		{0, 5.0},   // 5 * (1 - 0/200) = 5.0
		{100, 2.5}, // 5 * (1 - 100/200) = 2.5
		{50, 3.75}, // 5 * (1 - 50/200) = 3.75
		{200, 0.0}, // 5 * (1 - 200/200) = 0.0 (edge case)
	}

	for _, tt := range tests {
		tracker := NewCooldownTracker()
		npcID := uuid.New()
		p := personality.NewPersonality()
		p.Extraversion.Value = tt.extraversion

		now := time.Now()
		tracker.SetCooldown(npcID, now)

		// Check just before expected cooldown
		beforeTime := now.Add(time.Duration((tt.expectedMin - 0.1) * float64(time.Minute)))
		readyBefore := tracker.CheckCooldown(npcID, p, beforeTime)

		// Check just after expected cooldown
		afterTime := now.Add(time.Duration((tt.expectedMin + 0.1) * float64(time.Minute)))
		readyAfter := tracker.CheckCooldown(npcID, p, afterTime)

		if tt.expectedMin > 0 {
			assert.False(t, readyBefore, "E=%v should not be ready before %.1fmin", tt.extraversion, tt.expectedMin)
		}
		assert.True(t, readyAfter, "E=%v should be ready after %.1fmin", tt.extraversion, tt.expectedMin)
	}
}
