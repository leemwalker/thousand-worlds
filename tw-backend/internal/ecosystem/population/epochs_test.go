package population

import (
	"testing"

	"tw-backend/internal/worldgen/geography"
)

func TestEpochType(t *testing.T) {
	epochs := []EpochType{
		EpochHadean, EpochArchean, EpochProterozoic, EpochCambrian,
		EpochDevonian, EpochCarboniferous, EpochTriassic, EpochJurassic,
		EpochCretaceous, EpochCenozoic,
	}

	for _, epoch := range epochs {
		t.Run(string(epoch), func(t *testing.T) {
			if epoch == "" {
				t.Error("Epoch type should not be empty")
			}
		})
	}
}

func TestInitializeFromEpoch_Hadean(t *testing.T) {
	species := InitializeFromEpoch(EpochHadean, geography.BiomeOcean)
	if len(species) != 0 {
		t.Errorf("Hadean epoch should have no life, got %d species", len(species))
	}
}

func TestInitializeFromEpoch_Archean(t *testing.T) {
	species := InitializeFromEpoch(EpochArchean, geography.BiomeOcean)
	if len(species) == 0 {
		t.Error("Archean ocean should have primitive life")
	}
	// Only ocean should have life
	landSpecies := InitializeFromEpoch(EpochArchean, geography.BiomeDesert)
	if len(landSpecies) != 0 {
		t.Error("Archean land should have no life")
	}
}

func TestInitializeFromEpoch_Cambrian(t *testing.T) {
	species := InitializeFromEpoch(EpochCambrian, geography.BiomeOcean)
	if len(species) < 3 {
		t.Errorf("Cambrian ocean should have multiple species, got %d", len(species))
	}
	// Check for flora and fauna diversity
	var hasFlora, hasFauna bool
	for _, sp := range species {
		if sp.Diet == DietPhotosynthetic {
			hasFlora = true
		} else {
			hasFauna = true
		}
	}
	if !hasFlora || !hasFauna {
		t.Error("Cambrian should have both flora and fauna")
	}
}

func TestInitializeFromEpoch_Cenozoic(t *testing.T) {
	species := InitializeFromEpoch(EpochCenozoic, geography.BiomeGrassland)
	if len(species) < 3 {
		t.Errorf("Cenozoic grassland should have rich ecosystem, got %d", len(species))
	}
	// Check for mammals (fur covering)
	var hasFur bool
	for _, sp := range species {
		if sp.Traits.Covering == CoveringFur {
			hasFur = true
			break
		}
	}
	if !hasFur {
		t.Error("Cenozoic should have mammals with fur")
	}
}

func TestEvolutionGoal(t *testing.T) {
	goals := []EvolutionGoal{
		GoalSize, GoalSpeed, GoalStrength, GoalIntelligence,
		GoalColdResistance, GoalHeatResistance, GoalCamouflage,
		GoalNightVision, GoalFertility, GoalLifespan, GoalSocial, GoalVenom,
	}

	for _, goal := range goals {
		t.Run(string(goal), func(t *testing.T) {
			if goal == "" {
				t.Error("Goal type should not be empty")
			}
		})
	}
}

func TestApplyEvolutionGoal(t *testing.T) {
	tests := []struct {
		name     string
		goal     EvolutionGoal
		trait    func(EvolvableTraits) float64
		initial  float64
		expectGT bool // Expect trait > initial after applying goal
	}{
		{
			name:     "Size goal increases size",
			goal:     GoalSize,
			trait:    func(t EvolvableTraits) float64 { return t.Size },
			initial:  2.0,
			expectGT: true,
		},
		{
			name:     "Speed goal increases speed",
			goal:     GoalSpeed,
			trait:    func(t EvolvableTraits) float64 { return t.Speed },
			initial:  5.0,
			expectGT: true,
		},
		{
			name:     "Intelligence goal increases intelligence",
			goal:     GoalIntelligence,
			trait:    func(t EvolvableTraits) float64 { return t.Intelligence },
			initial:  0.5,
			expectGT: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traits := EvolvableTraits{Size: 2.0, Speed: 5.0, Intelligence: 0.5}
			newTraits := ApplyEvolutionGoal(traits, tt.goal, 0.1)

			if tt.expectGT && tt.trait(newTraits) <= tt.initial {
				t.Errorf("Expected trait to increase, got %.2f (was %.2f)", tt.trait(newTraits), tt.initial)
			}
		})
	}
}

func TestGetEpochDescription(t *testing.T) {
	tests := []struct {
		epoch EpochType
		min   int // Minimum expected description length
	}{
		{EpochHadean, 10},
		{EpochCambrian, 10},
		{EpochJurassic, 10},
		{EpochCenozoic, 10},
	}

	for _, tt := range tests {
		t.Run(string(tt.epoch), func(t *testing.T) {
			desc := GetEpochDescription(tt.epoch)
			if len(desc) < tt.min {
				t.Errorf("Description too short: %q", desc)
			}
		})
	}
}
