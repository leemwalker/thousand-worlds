package degradation

import (
	"sync"
	"time"
)

// DegradationTier represents system health
type DegradationTier int

const (
	Tier1_Healthy     DegradationTier = 1 // Full LLM
	Tier2_Slow        DegradationTier = 2 // Cached only, templates for new
	Tier3_Unavailable DegradationTier = 3 // All templates
)

// FallbackManager manages degradation state
type FallbackManager struct {
	currentTier DegradationTier
	mu          sync.RWMutex
}

// NewFallbackManager creates a manager
func NewFallbackManager() *FallbackManager {
	return &FallbackManager{
		currentTier: Tier1_Healthy,
	}
}

// SetTier updates the current tier
func (m *FallbackManager) SetTier(tier DegradationTier) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentTier = tier
}

// GetTier returns current tier
func (m *FallbackManager) GetTier() DegradationTier {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTier
}

// ShouldUseLLM returns true if LLM should be called
func (m *FallbackManager) ShouldUseLLM() bool {
	tier := m.GetTier()
	return tier == Tier1_Healthy
}

// GetFallbackTemplate returns a template based on context
func (m *FallbackManager) GetFallbackTemplate(contextType string) string {
	switch contextType {
	case "area":
		return "You see a generic area. Details are unclear."
	case "dialogue":
		return "...nods silently."
	default:
		return "..."
	}
}

// MonitorHealth simulates checking health (e.g., from metrics)
func (m *FallbackManager) MonitorHealth(cpuPercent float64, avgResponseTime time.Duration) {
	// Simple logic: if CPU > 90% or response > 5s, degrade
	if cpuPercent > 90 || avgResponseTime > 5*time.Second {
		m.SetTier(Tier3_Unavailable)
	} else if cpuPercent > 80 || avgResponseTime > 2*time.Second {
		m.SetTier(Tier2_Slow)
	} else {
		m.SetTier(Tier1_Healthy)
	}
}
