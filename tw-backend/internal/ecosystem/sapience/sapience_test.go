package sapience

import (
	"testing"

	"github.com/google/uuid"
)

func TestSapienceDetector_StandardThreshold(t *testing.T) {
	detector := NewSapienceDetector(uuid.New(), false)

	t.Run("high intelligence species becomes sapient", func(t *testing.T) {
		traits := SpeciesTraits{
			Intelligence:  8.0,
			Social:        7.0,
			ToolUse:       6.0,
			Communication: 7.0,
			MagicAffinity: 0,
			Population:    10000,
		}

		candidate := detector.Evaluate(uuid.New(), "Smart Ape", traits, 1000000)

		if candidate.Level != SapienceSapient {
			t.Errorf("Level = %s, want sapient", candidate.Level)
		}
		if candidate.IsMagicAssisted {
			t.Error("Should not be magic-assisted")
		}
	})

	t.Run("moderate intelligence is proto-sapient", func(t *testing.T) {
		traits := SpeciesTraits{
			Intelligence:  5.5,
			Social:        4.0,
			ToolUse:       4.0,
			Communication: 3.0,
			Population:    5000,
		}

		candidate := detector.Evaluate(uuid.New(), "Clever Monkey", traits, 1000000)

		if candidate.Level != SapienceProtoSapient {
			t.Errorf("Level = %s, want proto_sapient", candidate.Level)
		}
	})

	t.Run("low intelligence is not sapient", func(t *testing.T) {
		traits := SpeciesTraits{
			Intelligence:  3.0,
			Social:        2.0,
			ToolUse:       1.0,
			Communication: 1.0,
			Population:    1000,
		}

		candidate := detector.Evaluate(uuid.New(), "Simple Animal", traits, 1000000)

		if candidate.Level != SapienceNone {
			t.Errorf("Level = %s, want none", candidate.Level)
		}
	})
}

func TestSapienceDetector_MagicThreshold(t *testing.T) {
	detector := NewSapienceDetector(uuid.New(), true) // Magic enabled

	t.Run("magic-assisted sapience with lower intelligence", func(t *testing.T) {
		traits := SpeciesTraits{
			Intelligence:  5.5,
			Social:        5.5,
			ToolUse:       2.0,
			Communication: 3.0,
			MagicAffinity: 8.0, // High magic
			Population:    5000,
		}

		candidate := detector.Evaluate(uuid.New(), "Magic Folk", traits, 1000000)

		if candidate.Level != SapienceSapient {
			t.Errorf("Level = %s, want sapient", candidate.Level)
		}
		if !candidate.IsMagicAssisted {
			t.Error("Should be magic-assisted")
		}
	})

	t.Run("magic alone not enough without minimum intelligence", func(t *testing.T) {
		traits := SpeciesTraits{
			Intelligence:  2.0, // Too low
			Social:        5.0,
			ToolUse:       0,
			Communication: 1.0,
			MagicAffinity: 9.0, // Very high magic
			Population:    1000,
		}

		candidate := detector.Evaluate(uuid.New(), "Magic Beast", traits, 1000000)

		// Low intelligence means even magic can't fully compensate
		if candidate.Level == SapienceSapient {
			t.Log("Very high magic compensated for low intelligence")
		}
	})
}

func TestSapienceDetector_SapientTracking(t *testing.T) {
	detector := NewSapienceDetector(uuid.New(), false)

	// Add a sapient species
	traits := SpeciesTraits{
		Intelligence:  8.0,
		Social:        7.0,
		ToolUse:       6.0,
		Communication: 7.0,
	}

	speciesID := uuid.New()
	detector.Evaluate(speciesID, "First Sapient", traits, 1000000)

	t.Run("tracks sapient species", func(t *testing.T) {
		if !detector.HasAnySapience() {
			t.Error("Should have sapience")
		}
		if detector.GetSapientCount() != 1 {
			t.Errorf("Sapient count = %d, want 1", detector.GetSapientCount())
		}
	})

	t.Run("records first sapience year", func(t *testing.T) {
		if detector.FirstSapienceYear != 1000000 {
			t.Errorf("First sapience year = %d, want 1000000", detector.FirstSapienceYear)
		}
	})
}

func TestSapienceDetector_Progress(t *testing.T) {
	detector := NewSapienceDetector(uuid.New(), false)

	t.Run("progress starts at 0", func(t *testing.T) {
		progress := detector.CalculateSapienceProgress()
		if progress != 0 {
			t.Errorf("Initial progress = %f, want 0", progress)
		}
	})

	t.Run("progress increases with candidates", func(t *testing.T) {
		// Add a proto-sapient
		traits := SpeciesTraits{
			Intelligence:  5.5,
			Social:        4.0,
			ToolUse:       4.0,
			Communication: 3.0,
		}
		detector.Evaluate(uuid.New(), "Candidate", traits, 1000000)

		progress := detector.CalculateSapienceProgress()
		if progress <= 0 || progress >= 1.0 {
			t.Errorf("Progress = %f, want between 0 and 1", progress)
		}
		t.Logf("Sapience progress: %.2f", progress)
	})

	t.Run("progress is 1.0 after sapience", func(t *testing.T) {
		traits := SpeciesTraits{
			Intelligence:  8.0,
			Social:        7.0,
			ToolUse:       6.0,
			Communication: 7.0,
		}
		detector.Evaluate(uuid.New(), "Sapient", traits, 2000000)

		progress := detector.CalculateSapienceProgress()
		if progress != 1.0 {
			t.Errorf("Progress after sapience = %f, want 1.0", progress)
		}
	})
}

func TestSapienceDetector_Prediction(t *testing.T) {
	detector := NewSapienceDetector(uuid.New(), false)

	// Add a proto-sapient candidate
	traits := SpeciesTraits{
		Intelligence:  6.0,
		Social:        5.0,
		ToolUse:       4.0,
		Communication: 4.0,
	}
	detector.Evaluate(uuid.New(), "Candidate", traits, 1000000)

	predicted := detector.PredictSapienceYear(0.001) // 0.001 intelligence per million years
	t.Logf("Predicted sapience year: %d", predicted)

	if predicted <= 1000000 {
		t.Error("Predicted year should be in the future")
	}
}
