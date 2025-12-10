package relationship

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateAverageBaseline(t *testing.T) {
	interactions := []Interaction{
		{BehavioralContext: BehavioralProfile{Aggression: 0.5, Honesty: 1.0}},
		{BehavioralContext: BehavioralProfile{Aggression: 0.7, Honesty: 0.0}},
	}

	avg := CalculateAverageBaseline(interactions)
	assert.InDelta(t, 0.6, avg.Aggression, 0.01)
	assert.InDelta(t, 0.5, avg.Honesty, 0.01)
}

func TestUpdateBaseline(t *testing.T) {
	current := BehavioralProfile{Aggression: 0.5}
	recent := BehavioralProfile{Aggression: 1.0}

	// New = 0.5*0.9 + 1.0*0.1 = 0.45 + 0.1 = 0.55
	updated := UpdateBaseline(current, recent)
	assert.InDelta(t, 0.55, updated.Aggression, 0.01)
}

func TestCalculateDrift(t *testing.T) {
	baseline := BehavioralProfile{Aggression: 0.2, Honesty: 0.8}
	recent := BehavioralProfile{Aggression: 0.8, Honesty: 0.8} // High aggression drift (0.6)

	metrics := CalculateDrift(baseline, recent)

	assert.InDelta(t, 0.6, metrics.DriftScore, 0.01)
	assert.Equal(t, "Moderate", metrics.DriftLevel)
	assert.Contains(t, metrics.AffectedTraits, "aggression")
	assert.Equal(t, 1, metrics.DriftDirection) // Increased
}

func TestGetDriftLevel(t *testing.T) {
	assert.Equal(t, "None", GetDriftLevel(0.1))
	assert.Equal(t, "Subtle", GetDriftLevel(0.35))
	assert.Equal(t, "Moderate", GetDriftLevel(0.6))
	assert.Equal(t, "Severe", GetDriftLevel(0.8))
}
