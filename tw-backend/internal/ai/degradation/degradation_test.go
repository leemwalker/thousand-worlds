package degradation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMonitorHealth(t *testing.T) {
	m := NewFallbackManager()

	// Healthy
	m.MonitorHealth(50, 100*time.Millisecond)
	assert.Equal(t, Tier1_Healthy, m.GetTier())
	assert.True(t, m.ShouldUseLLM())

	// Slow (Tier 2)
	m.MonitorHealth(85, 100*time.Millisecond)
	assert.Equal(t, Tier2_Slow, m.GetTier())
	assert.False(t, m.ShouldUseLLM())

	// Unavailable (Tier 3)
	m.MonitorHealth(95, 100*time.Millisecond)
	assert.Equal(t, Tier3_Unavailable, m.GetTier())

	// Recover
	m.MonitorHealth(50, 100*time.Millisecond)
	assert.Equal(t, Tier1_Healthy, m.GetTier())
}

func TestGetFallbackTemplate(t *testing.T) {
	m := NewFallbackManager()
	assert.NotEmpty(t, m.GetFallbackTemplate("area"))
	assert.NotEmpty(t, m.GetFallbackTemplate("dialogue"))
	assert.NotEmpty(t, m.GetFallbackTemplate("unknown"))
}

func TestSetTierConcurrent(t *testing.T) {
	m := NewFallbackManager()

	// Just verify no race conditions
	go m.SetTier(Tier2_Slow)
	go m.GetTier()

	time.Sleep(10 * time.Millisecond)
	// Success if no panic/race detector failure (requires -race flag)
}
