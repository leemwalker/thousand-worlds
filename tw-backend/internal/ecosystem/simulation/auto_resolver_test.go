package simulation

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAutoResolver(t *testing.T) {
	ar := NewAutoResolver(12345)
	require.NotNil(t, ar)
	assert.Equal(t, int64(12345), ar.seed)
	assert.NotNil(t, ar.rng)
}

func TestAutoResolver_Resolve_NoInterventions(t *testing.T) {
	ar := NewAutoResolver(42)
	tp := TurningPointInfo{
		ID:            uuid.New(),
		Trigger:       "extinction",
		Interventions: []InterventionOption{},
	}

	result := ar.Resolve(tp)
	assert.Equal(t, -1, result)
}

func TestAutoResolver_Resolve_ExtinctionPreference(t *testing.T) {
	ar := NewAutoResolver(42)
	tp := TurningPointInfo{
		ID:      uuid.New(),
		Trigger: "extinction",
		Interventions: []InterventionOption{
			{ID: uuid.New(), Name: "Cataclysm", Type: "cataclysm", RiskLevel: 0.9},
			{ID: uuid.New(), Name: "Protection", Type: "protection", RiskLevel: 0.3},
			{ID: uuid.New(), Name: "Nudge", Type: "nudge", RiskLevel: 0.1},
		},
	}

	result := ar.Resolve(tp)
	// Should prefer protection for extinction events
	assert.Equal(t, 1, result) // Index of "protection"
}

func TestAutoResolver_Resolve_SapiencePreference(t *testing.T) {
	ar := NewAutoResolver(42)
	tp := TurningPointInfo{
		ID:      uuid.New(),
		Trigger: "sapience",
		Interventions: []InterventionOption{
			{ID: uuid.New(), Name: "Protection", Type: "protection", RiskLevel: 0.3},
			{ID: uuid.New(), Name: "Accelerate", Type: "accelerate", RiskLevel: 0.5},
			{ID: uuid.New(), Name: "Observe", Type: "none", RiskLevel: 0.0},
		},
	}

	result := ar.Resolve(tp)
	// Should prefer accelerate for sapience events
	assert.Equal(t, 1, result) // Index of "accelerate"
}

func TestAutoResolver_Resolve_IntervalPreference(t *testing.T) {
	ar := NewAutoResolver(42)
	tp := TurningPointInfo{
		ID:      uuid.New(),
		Trigger: "interval",
		Interventions: []InterventionOption{
			{ID: uuid.New(), Name: "Direct", Type: "direct", RiskLevel: 0.7},
			{ID: uuid.New(), Name: "Observe", Type: "none", RiskLevel: 0.0},
			{ID: uuid.New(), Name: "Nudge", Type: "nudge", RiskLevel: 0.1},
		},
	}

	result := ar.Resolve(tp)
	// Should prefer "none" for interval events (observe)
	assert.Equal(t, 1, result) // Index of "none"
}

func TestAutoResolver_Resolve_UnknownTrigger(t *testing.T) {
	ar := NewAutoResolver(42)
	tp := TurningPointInfo{
		ID:      uuid.New(),
		Trigger: "unknown_event",
		Interventions: []InterventionOption{
			{ID: uuid.New(), Name: "High Risk", Type: "cataclysm", RiskLevel: 0.9},
			{ID: uuid.New(), Name: "Low Risk", Type: "nudge", RiskLevel: 0.1},
			{ID: uuid.New(), Name: "Medium Risk", Type: "direct", RiskLevel: 0.5},
		},
	}

	result := ar.Resolve(tp)
	// Should select lowest risk for unknown triggers
	assert.Equal(t, 1, result) // Index of "Low Risk" (lowest RiskLevel)
}

func TestAutoResolver_selectByPreference_FallsBackToRandom(t *testing.T) {
	ar := NewAutoResolver(42)
	interventions := []InterventionOption{
		{ID: uuid.New(), Name: "A", Type: "typeA", RiskLevel: 0.5},
		{ID: uuid.New(), Name: "B", Type: "typeB", RiskLevel: 0.5},
	}

	// None of the preferences match, so should fall back to random
	result := ar.selectByPreference(interventions, []string{"typeX", "typeY"})
	assert.GreaterOrEqual(t, result, 0)
	assert.Less(t, result, len(interventions))
}

func TestAutoResolver_selectLowestRisk(t *testing.T) {
	ar := NewAutoResolver(42)

	tests := []struct {
		name          string
		interventions []InterventionOption
		expectedIdx   int
	}{
		{
			name:          "empty list",
			interventions: []InterventionOption{},
			expectedIdx:   -1,
		},
		{
			name: "single item",
			interventions: []InterventionOption{
				{Name: "Only", RiskLevel: 0.5},
			},
			expectedIdx: 0,
		},
		{
			name: "multiple items - lowest first",
			interventions: []InterventionOption{
				{Name: "Lowest", RiskLevel: 0.1},
				{Name: "Medium", RiskLevel: 0.5},
				{Name: "Highest", RiskLevel: 0.9},
			},
			expectedIdx: 0,
		},
		{
			name: "multiple items - lowest last",
			interventions: []InterventionOption{
				{Name: "Medium", RiskLevel: 0.5},
				{Name: "Highest", RiskLevel: 0.9},
				{Name: "Lowest", RiskLevel: 0.1},
			},
			expectedIdx: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ar.selectLowestRisk(tt.interventions)
			assert.Equal(t, tt.expectedIdx, result)
		})
	}
}

func TestAutoResolver_ResolveDeterministic(t *testing.T) {
	seed := int64(12345)
	tpID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tp := TurningPointInfo{
		ID:      tpID,
		Trigger: "unknown",
		Interventions: []InterventionOption{
			{ID: uuid.New(), Name: "A", Type: "typeA", RiskLevel: 0.5},
			{ID: uuid.New(), Name: "B", Type: "typeB", RiskLevel: 0.5},
			{ID: uuid.New(), Name: "C", Type: "typeC", RiskLevel: 0.5},
		},
	}

	// Run multiple times with identical inputs - results should be deterministic
	results := make([]int, 5)
	for i := 0; i < 5; i++ {
		ar := NewAutoResolver(seed)
		results[i] = ar.ResolveDeterministic(tp)
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0], results[i], "ResolveDeterministic should be deterministic")
	}
}

func TestAutoResolver_DifferentSeeds_DifferentResults(t *testing.T) {
	// Test the random fallback in selectByPreference when no preference matches
	interventions := []InterventionOption{
		{ID: uuid.New(), Name: "A", Type: "typeA", RiskLevel: 0.5},
		{ID: uuid.New(), Name: "B", Type: "typeB", RiskLevel: 0.5},
		{ID: uuid.New(), Name: "C", Type: "typeC", RiskLevel: 0.5},
		{ID: uuid.New(), Name: "D", Type: "typeD", RiskLevel: 0.5},
		{ID: uuid.New(), Name: "E", Type: "typeE", RiskLevel: 0.5},
	}

	// Test that different seeds can produce different results for random selection
	// With 5 options and many seeds, at least some should differ
	resultCounts := make(map[int]int)
	for seed := int64(1); seed <= 100; seed++ {
		ar := NewAutoResolver(seed)
		// Use preferences that don't match any intervention type
		result := ar.selectByPreference(interventions, []string{"nonexistent"})
		resultCounts[result]++
	}

	// With 100 seeds and 5 options, we should see multiple different results
	assert.Greater(t, len(resultCounts), 1, "Different seeds should produce different random results")
}
