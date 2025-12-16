package population

import (
	"testing"

	"github.com/google/uuid"
)

func TestCascadeSimulator_BasicCascade(t *testing.T) {
	cs := NewCascadeSimulator()

	// Create a simple food chain: grass -> deer -> wolf
	grassID := uuid.New()
	deerID := uuid.New()
	wolfID := uuid.New()

	// Deer eats grass (obligate herbivore)
	cs.AddRelationship(EcologicalRelationship{
		SourceSpeciesID: deerID,
		TargetSpeciesID: grassID,
		Type:            RelationshipPredation,
		Strength:        0.8,
		IsObligate:      true,
	})

	// Wolf eats deer
	cs.AddRelationship(EcologicalRelationship{
		SourceSpeciesID: wolfID,
		TargetSpeciesID: deerID,
		Type:            RelationshipPredation,
		Strength:        0.9,
		IsObligate:      true,
	})

	cs.SetSpeciesRole(grassID, []EcologicalRole{RolePrimaryProducer})
	cs.SetSpeciesRole(deerID, []EcologicalRole{})
	cs.SetSpeciesRole(wolfID, []EcologicalRole{RoleApexPredator})

	t.Run("grass extinction cascades up food chain", func(t *testing.T) {
		result := cs.CalculateCascade(grassID, "Grass", 1000, 5)

		if len(result.Events) == 0 {
			t.Error("Expected cascade events")
		}

		// Deer should be affected
		deerAffected := false
		for _, event := range result.Events {
			if event.AffectedSpeciesID == deerID {
				deerAffected = true
				// When deer (predator/herbivore) loses grass (prey/food), it's food loss
				if event.CascadeType != CascadeFoodLoss && event.CascadeType != CascadePredatorLoss {
					t.Errorf("Deer cascade type = %s, want food_loss or predator_loss", event.CascadeType)
				}
				t.Logf("Deer cascade: %s - %s", event.CascadeType, event.Description)
			}
		}
		if !deerAffected {
			t.Error("Deer should be affected by grass extinction")
		}

		// Secondary extinction of deer should trigger wolf cascade
		t.Logf("Events: %d, Secondary extinctions: %d", len(result.Events), len(result.SecondaryExtinctions))
	})
}

func TestCascadeSimulator_MutualismCollapse(t *testing.T) {
	cs := NewCascadeSimulator()

	// Pollinator-plant mutualism
	beeID := uuid.New()
	flowerID := uuid.New()

	// Flower depends on bee for pollination (obligate)
	cs.AddRelationship(EcologicalRelationship{
		SourceSpeciesID: flowerID,
		TargetSpeciesID: beeID,
		Type:            RelationshipMutualism,
		Strength:        0.95,
		IsObligate:      true,
	})

	// Bee depends on flower for food (obligate)
	cs.AddRelationship(EcologicalRelationship{
		SourceSpeciesID: beeID,
		TargetSpeciesID: flowerID,
		Type:            RelationshipMutualism,
		Strength:        0.9,
		IsObligate:      true,
	})

	cs.SetSpeciesRole(beeID, []EcologicalRole{RolePollinator})
	cs.SetSpeciesRole(flowerID, []EcologicalRole{RolePrimaryProducer})

	t.Run("bee extinction causes flower co-extinction", func(t *testing.T) {
		result := cs.CalculateCascade(beeID, "Bee", 1000, 3)

		// Flower should suffer co-extinction
		flowerImpact := result.PopulationChanges[flowerID]
		if flowerImpact > 0.2 { // Should be near 0 (extinction)
			t.Errorf("Flower population = %f, expected near-extinction from obligate mutualism loss", flowerImpact)
		}
	})
}

func TestCascadeSimulator_PredatorRelease(t *testing.T) {
	cs := NewCascadeSimulator()

	// Wolf controls deer population
	wolfID := uuid.New()
	deerID := uuid.New()

	cs.AddRelationship(EcologicalRelationship{
		SourceSpeciesID: wolfID,
		TargetSpeciesID: deerID,
		Type:            RelationshipPredation,
		Strength:        0.7,
		IsObligate:      false,
	})

	cs.SetSpeciesRole(wolfID, []EcologicalRole{RoleApexPredator})

	t.Run("wolf extinction causes deer population surge", func(t *testing.T) {
		result := cs.CalculateCascade(wolfID, "Wolf", 1000, 3)

		deerImpact := result.PopulationChanges[deerID]
		t.Logf("Deer population impact: %f", deerImpact)

		// Deer should experience population increase (predator release)
		if deerImpact <= 1.0 {
			t.Errorf("Deer should experience population surge, got %f", deerImpact)
		}
	})
}

func TestCascadeSimulator_KeystoneSpecies(t *testing.T) {
	cs := NewCascadeSimulator()

	// Beaver is a keystone species (ecosystem engineer)
	beaverID := uuid.New()
	fishID := uuid.New()
	birdID := uuid.New()
	plantID := uuid.New()

	// Set roles
	cs.SetSpeciesRole(beaverID, []EcologicalRole{RoleEcosystemEngineer})
	cs.SetSpeciesRole(fishID, []EcologicalRole{})
	cs.SetSpeciesRole(birdID, []EcologicalRole{})
	cs.SetSpeciesRole(plantID, []EcologicalRole{RolePrimaryProducer})

	// Mark beaver as keystone
	cs.SetKeystoneImportance(beaverID, 0.8)

	t.Run("keystone collapse affects many species", func(t *testing.T) {
		result := cs.CalculateCascade(beaverID, "Beaver", 1000, 3)

		t.Logf("Keystone cascade: %d events, %d species affected", len(result.Events), result.TotalAffected)

		// All other species should be negatively affected
		keystoneEvents := 0
		for _, event := range result.Events {
			if event.CascadeType == CascadeKeystone {
				keystoneEvents++
			}
		}

		if keystoneEvents == 0 {
			t.Error("Expected keystone cascade events")
		}
	})
}

func TestCascadeSimulator_IdentifyKeystoneSpecies(t *testing.T) {
	cs := NewCascadeSimulator()

	// Create species with many dependents
	keystoneID := uuid.New()
	for i := 0; i < 8; i++ {
		dependentID := uuid.New()
		cs.AddRelationship(EcologicalRelationship{
			SourceSpeciesID: dependentID,
			TargetSpeciesID: keystoneID,
			Type:            RelationshipMutualism,
			Strength:        0.5,
			IsObligate:      i < 3, // 3 obligate dependents
		})
		cs.SetSpeciesRole(dependentID, []EcologicalRole{})
	}
	cs.SetSpeciesRole(keystoneID, []EcologicalRole{})

	keystones := cs.IdentifyKeystoneSpecies()

	importance, found := keystones[keystoneID]
	if !found {
		t.Error("Expected species to be identified as keystone")
	}
	if importance < 0.5 {
		t.Errorf("Keystone importance = %f, expected >= 0.5", importance)
	}
	t.Logf("Keystone importance: %f", importance)
}

func TestCascadeSimulator_ExtinctionRisk(t *testing.T) {
	cs := NewCascadeSimulator()

	specialistID := uuid.New()
	generalistID := uuid.New()
	foodID := uuid.New()

	// Specialist has one obligate dependency
	cs.AddRelationship(EcologicalRelationship{
		SourceSpeciesID: specialistID,
		TargetSpeciesID: foodID,
		Type:            RelationshipPredation,
		Strength:        1.0,
		IsObligate:      true,
	})

	// Generalist has many non-obligate dependencies
	for i := 0; i < 5; i++ {
		altFoodID := uuid.New()
		cs.AddRelationship(EcologicalRelationship{
			SourceSpeciesID: generalistID,
			TargetSpeciesID: altFoodID,
			Type:            RelationshipPredation,
			Strength:        0.2,
			IsObligate:      false,
		})
		cs.SetSpeciesRole(altFoodID, []EcologicalRole{})
	}

	cs.SetSpeciesRole(specialistID, []EcologicalRole{})
	cs.SetSpeciesRole(generalistID, []EcologicalRole{})
	cs.SetSpeciesRole(foodID, []EcologicalRole{})

	specialistRisk := cs.GetExtinctionRisk(specialistID)
	generalistRisk := cs.GetExtinctionRisk(generalistID)

	t.Logf("Specialist risk: %f, Generalist risk: %f", specialistRisk, generalistRisk)

	if specialistRisk <= generalistRisk {
		t.Error("Specialist should have higher extinction risk than generalist")
	}
}

func TestGenerateExtinctionCause(t *testing.T) {
	tests := []struct {
		cascadeType CascadeType
		trigger     string
		contains    string
	}{
		{CascadeFoodLoss, "Rabbit", "starvation"},
		{CascadePredatorLoss, "Wolf", "overpopulation"},
		{CascadeCoExtinction, "Bee", "symbiotic"},
		{CascadeKeystone, "Beaver", "keystone"},
	}

	for _, tt := range tests {
		cause := GenerateExtinctionCause(tt.cascadeType, tt.trigger)
		if cause == "" {
			t.Errorf("Empty cause for %s", tt.cascadeType)
		}
		t.Logf("%s: %s", tt.cascadeType, cause)
	}
}
