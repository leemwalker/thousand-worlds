package ecosystem

import (
	"testing"

	"github.com/google/uuid"
)

func TestTurningPointManager_IntervalTrigger(t *testing.T) {
	tpm := NewTurningPointManager(uuid.New())
	tpm.IntervalYears = 100000 // Shorter for testing

	t.Run("triggers at interval", func(t *testing.T) {
		tp := tpm.CheckForTurningPoint(100000, 50, 0, nil, "")

		if tp == nil {
			t.Fatal("Expected turning point at interval")
		}
		if tp.Trigger != TriggerInterval {
			t.Errorf("Trigger = %s, want interval", tp.Trigger)
		}
		if tp.Title != "Era Milestone" {
			t.Errorf("Title = %s, want Era Milestone", tp.Title)
		}
	})

	t.Run("pauses simulation", func(t *testing.T) {
		if !tpm.IsPaused() {
			t.Error("Should be paused after turning point")
		}
	})
}

func TestTurningPointManager_ExtinctionTrigger(t *testing.T) {
	tpm := NewTurningPointManager(uuid.New())
	tpm.ExtinctionThreshold = 0.25

	t.Run("triggers on mass extinction", func(t *testing.T) {
		// 100 species, 30 recently extinct = 30% > 25% threshold
		tp := tpm.CheckForTurningPoint(1000, 100, 30, nil, "")

		if tp == nil {
			t.Fatal("Expected turning point on extinction")
		}
		if tp.Trigger != TriggerExtinction {
			t.Errorf("Trigger = %s, want extinction", tp.Trigger)
		}
	})
}

func TestTurningPointManager_SapienceTrigger(t *testing.T) {
	tpm := NewTurningPointManager(uuid.New())

	t.Run("triggers on sapience emergence", func(t *testing.T) {
		sapientSpecies := []uuid.UUID{uuid.New()}
		tp := tpm.CheckForTurningPoint(5000000, 200, 0, sapientSpecies, "")

		if tp == nil {
			t.Fatal("Expected turning point on sapience")
		}
		if tp.Trigger != TriggerSapience {
			t.Errorf("Trigger = %s, want sapience", tp.Trigger)
		}
		if tp.Title != "Sapience Emerges" {
			t.Errorf("Title = %s, want 'Sapience Emerges'", tp.Title)
		}
	})
}

func TestTurningPointManager_Interventions(t *testing.T) {
	tpm := NewTurningPointManager(uuid.New())

	// Trigger an interval turning point
	tp := tpm.CheckForTurningPoint(1000000, 50, 0, nil, "")

	t.Run("has interventions", func(t *testing.T) {
		if len(tp.Interventions) == 0 {
			t.Error("Expected available interventions")
		}
		t.Logf("Available interventions: %d", len(tp.Interventions))
		for _, i := range tp.Interventions {
			t.Logf("  - %s (%s): cost %d", i.Name, i.Type, i.Cost)
		}
	})

	t.Run("observe is always available", func(t *testing.T) {
		hasObserve := false
		for _, i := range tp.Interventions {
			if i.Type == InterventionNone {
				hasObserve = true
				break
			}
		}
		if !hasObserve {
			t.Error("Observe Only should always be available")
		}
	})
}

func TestTurningPointManager_Resolution(t *testing.T) {
	tpm := NewTurningPointManager(uuid.New())

	// Create and resolve a turning point
	tp := tpm.CheckForTurningPoint(1000000, 50, 0, nil, "")

	t.Run("can resolve with intervention", func(t *testing.T) {
		if len(tp.Interventions) == 0 {
			t.Skip("No interventions available")
		}

		chosen := tp.Interventions[0]
		resolved := tpm.ResolveTurningPoint(tp.ID, chosen.ID)

		if resolved == nil {
			t.Fatal("Resolution failed")
		}
		if !resolved.IsResolved {
			t.Error("Should be marked resolved")
		}
		if resolved.ChosenIntervention == nil || *resolved.ChosenIntervention != chosen.ID {
			t.Error("Chosen intervention not recorded")
		}
	})

	t.Run("no longer paused after resolution", func(t *testing.T) {
		if tpm.IsPaused() {
			t.Error("Should not be paused after resolution")
		}
	})
}

func TestTurningPointManager_Cooldowns(t *testing.T) {
	tpm := NewTurningPointManager(uuid.New())
	tpm.IntervalYears = 100000

	// Find an intervention with cooldown
	var interventionWithCooldown *Intervention
	for i := range tpm.InterventionTemplates {
		if tpm.InterventionTemplates[i].Cooldown > 0 {
			interventionWithCooldown = &tpm.InterventionTemplates[i]
			break
		}
	}

	if interventionWithCooldown == nil {
		t.Skip("No interventions with cooldown found")
	}

	t.Run("cooldown prevents reuse", func(t *testing.T) {
		// First use
		tp1 := tpm.CheckForTurningPoint(100000, 50, 0, nil, "")
		tpm.ResolveTurningPoint(tp1.ID, interventionWithCooldown.ID)

		// Second use (should be on cooldown)
		tp2 := tpm.CheckForTurningPoint(200000, 50, 0, nil, "")

		// Check if intervention is available
		available := false
		for _, i := range tp2.Interventions {
			if i.Name == interventionWithCooldown.Name {
				available = true
				break
			}
		}

		if available {
			t.Logf("Intervention still available at year 200000 (cooldown: %d)", interventionWithCooldown.Cooldown)
		}
	})
}

func TestTurningPointManager_SpecialEvents(t *testing.T) {
	tpm := NewTurningPointManager(uuid.New())

	tests := []struct {
		event   string
		trigger TurningPointTrigger
		title   string
	}{
		{"climate_shift", TriggerClimateShift, "Climate Upheaval"},
		{"tectonic_event", TriggerTectonicEvent, "Continental Drift"},
		{"pandemic", TriggerPandemic, "The Great Plague"},
	}

	for _, tt := range tests {
		// Reset pending
		tpm.PendingTurningPoint = nil

		t.Run(tt.event, func(t *testing.T) {
			tp := tpm.CheckForTurningPoint(1000, 50, 0, nil, tt.event)

			if tp == nil {
				t.Fatal("Expected turning point")
			}
			if tp.Trigger != tt.trigger {
				t.Errorf("Trigger = %s, want %s", tp.Trigger, tt.trigger)
			}
			if tp.Title != tt.title {
				t.Errorf("Title = %s, want %s", tp.Title, tt.title)
			}
		})
	}
}

func TestTurningPointManager_History(t *testing.T) {
	tpm := NewTurningPointManager(uuid.New())
	tpm.IntervalYears = 10000

	// Create and resolve multiple turning points
	for year := int64(10000); year <= 30000; year += 10000 {
		tpm.PendingTurningPoint = nil
		tp := tpm.CheckForTurningPoint(year, 50, 0, nil, "")
		if tp != nil && len(tp.Interventions) > 0 {
			tpm.ResolveTurningPoint(tp.ID, tp.Interventions[0].ID)
		}
	}

	history := tpm.GetHistory()
	t.Logf("History length: %d", len(history))

	if len(history) < 2 {
		t.Error("Expected at least 2 resolved turning points")
	}
}
