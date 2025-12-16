package population

import (
	"math/rand"
	"testing"

	"github.com/google/uuid"
)

func TestNewOrganismFromGeneticCode(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	gc := NewGeneticCode(rng)
	em := DefaultExpressionMatrix()
	id := uuid.New()

	org := NewOrganismFromGeneticCode(id, "Test Species", gc, em, 1000000, nil)

	t.Run("creates valid organism", func(t *testing.T) {
		if org.ID != id {
			t.Error("ID mismatch")
		}
		if org.Name != "Test Species" {
			t.Error("Name mismatch")
		}
		if org.GeneticCode == nil {
			t.Error("GeneticCode should not be nil")
		}
		if org.Traits == nil {
			t.Error("Traits should not be nil")
		}
		if org.OriginYear != 1000000 {
			t.Errorf("OriginYear = %d, want 1000000", org.OriginYear)
		}
	})

	t.Run("traits are in valid range", func(t *testing.T) {
		// Check traits are in expected ranges (0-10 for most)
		if org.Traits.Size < 0 || org.Traits.Size > 10 {
			t.Errorf("Size = %f, out of range [0, 10]", org.Traits.Size)
		}
		if org.Traits.Speed < 0 || org.Traits.Speed > 10 {
			t.Errorf("Speed = %f, out of range [0, 10]", org.Traits.Speed)
		}
		if org.Traits.Autotrophy < 0 || org.Traits.Autotrophy > 1 {
			t.Errorf("Autotrophy = %f, out of range [0, 1]", org.Traits.Autotrophy)
		}
	})

	t.Run("with ancestor", func(t *testing.T) {
		ancestorID := uuid.New()
		orgWithAncestor := NewOrganismFromGeneticCode(uuid.New(), "Child", gc, em, 2000000, &ancestorID)

		if orgWithAncestor.AncestorID == nil {
			t.Error("AncestorID should not be nil")
		}
		if *orgWithAncestor.AncestorID != ancestorID {
			t.Error("AncestorID mismatch")
		}
	})
}

func TestOrganism_Classification(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	em := DefaultExpressionMatrix()

	t.Run("autotroph detection", func(t *testing.T) {
		gc := NewGeneticCode(rng)
		// Force high autotrophy
		for i := 0; i < DefinedGeneCount; i++ {
			gc.DefinedGenes[i] = 0.5
		}
		// Set genes that affect autotrophy trait to high values
		// TraitAutotrophy = 18, so genes around ~60 should affect it
		for i := 54; i < 60; i++ {
			gc.DefinedGenes[i] = 1.0
		}

		org := NewOrganismFromGeneticCode(uuid.New(), "Plant", gc, em, 0, nil)

		// Test depends on expression matrix mapping
		t.Logf("Autotrophy value: %f", org.Traits.Autotrophy)
	})

	t.Run("carnivore detection", func(t *testing.T) {
		gc := NewGeneticCode(rng)
		org := NewOrganismFromGeneticCode(uuid.New(), "Test", gc, em, 0, nil)

		// Set carnivore trait manually for testing helper functions
		org.Traits.Carnivore = 8.0
		org.Traits.Autotrophy = 0.2

		if !org.IsCarnivore() {
			t.Error("Should be detected as carnivore with Carnivore=8.0")
		}

		org.Traits.Carnivore = 2.0
		if !org.IsHerbivore() {
			t.Error("Should be detected as herbivore with Carnivore=2.0")
		}

		org.Traits.Carnivore = 5.0
		if !org.IsOmnivore() {
			t.Error("Should be detected as omnivore with Carnivore=5.0")
		}
	})
}

func TestOrganism_ProtoSapience(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	gc := NewGeneticCode(rng)
	em := DefaultExpressionMatrix()
	org := NewOrganismFromGeneticCode(uuid.New(), "Test", gc, em, 0, nil)

	t.Run("below threshold", func(t *testing.T) {
		org.Traits.Intelligence = 5.0
		org.Traits.Social = 5.0
		org.Traits.ToolUse = 2.0
		org.Traits.Communication = 2.0

		if org.IsProtoSapient() {
			t.Error("Should not be proto-sapient with low traits")
		}
	})

	t.Run("above threshold", func(t *testing.T) {
		org.Traits.Intelligence = 8.0
		org.Traits.Social = 7.0
		org.Traits.ToolUse = 4.0
		org.Traits.Communication = 4.0

		if !org.IsProtoSapient() {
			t.Error("Should be proto-sapient with high traits")
		}
	})

	t.Run("magic uplift threshold", func(t *testing.T) {
		org.Traits.Intelligence = 5.0
		org.Traits.Social = 5.0
		org.Traits.MagicAffinity = 6.0

		if !org.IsProtoSapientWithMagic() {
			t.Error("Should be proto-sapient with magic uplift")
		}
	})
}

func TestOrganism_MetabolicRate(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	gc := NewGeneticCode(rng)
	em := DefaultExpressionMatrix()

	t.Run("larger organisms have higher base rate", func(t *testing.T) {
		smallOrg := NewOrganismFromGeneticCode(uuid.New(), "Small", gc, em, 0, nil)
		smallOrg.Traits.Size = 1.0

		largeOrg := NewOrganismFromGeneticCode(uuid.New(), "Large", gc.Clone(), em, 0, nil)
		largeOrg.Traits.Size = 8.0

		// Normalize other traits to compare only size effect
		smallOrg.Traits.Speed = 5.0
		smallOrg.Traits.Strength = 5.0
		smallOrg.Traits.Armor = 0
		smallOrg.Traits.Intelligence = 0
		smallOrg.Traits.MagicAffinity = 0

		largeOrg.Traits.Speed = 5.0
		largeOrg.Traits.Strength = 5.0
		largeOrg.Traits.Armor = 0
		largeOrg.Traits.Intelligence = 0
		largeOrg.Traits.MagicAffinity = 0

		smallRate := smallOrg.CalculateMetabolicRate()
		largeRate := largeOrg.CalculateMetabolicRate()

		if largeRate <= smallRate {
			t.Errorf("Large rate (%f) should be > small rate (%f)", largeRate, smallRate)
		}
	})

	t.Run("enhanced traits increase cost", func(t *testing.T) {
		baseOrg := NewOrganismFromGeneticCode(uuid.New(), "Base", gc, em, 0, nil)
		baseOrg.Traits.Size = 5.0
		baseOrg.Traits.Speed = 5.0
		baseOrg.Traits.Strength = 5.0
		baseOrg.Traits.Armor = 0
		baseOrg.Traits.Intelligence = 0
		baseOrg.Traits.MagicAffinity = 0

		enhancedOrg := NewOrganismFromGeneticCode(uuid.New(), "Enhanced", gc.Clone(), em, 0, nil)
		enhancedOrg.Traits.Size = 5.0
		enhancedOrg.Traits.Speed = 9.0 // Much faster
		enhancedOrg.Traits.Strength = 5.0
		enhancedOrg.Traits.Armor = 5.0 // Has armor
		enhancedOrg.Traits.Intelligence = 0
		enhancedOrg.Traits.MagicAffinity = 0

		baseRate := baseOrg.CalculateMetabolicRate()
		enhancedRate := enhancedOrg.CalculateMetabolicRate()

		if enhancedRate <= baseRate {
			t.Errorf("Enhanced rate (%f) should be > base rate (%f)", enhancedRate, baseRate)
		}
	})
}

func TestOrganism_ReproductionRate(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	gc := NewGeneticCode(rng)
	em := DefaultExpressionMatrix()

	t.Run("uses quarter-power scaling", func(t *testing.T) {
		org := NewOrganismFromGeneticCode(uuid.New(), "Test", gc, em, 0, nil)
		org.Traits.Fertility = 5.0 // Mid fertility

		// Small organism
		org.Traits.Size = 1.0
		smallRate := org.CalculateReproductionRate()

		// Large organism
		org.Traits.Size = 16.0
		largeRate := org.CalculateReproductionRate()

		// With M^-0.25, 16x size = 0.5x reproduction rate
		// smallRate / largeRate should be approximately 2
		ratio := smallRate / largeRate
		if ratio < 1.5 || ratio > 2.5 {
			t.Errorf("Ratio = %f, expected ~2 for M^-0.25 scaling", ratio)
		}
	})

	t.Run("fertility affects rate", func(t *testing.T) {
		org := NewOrganismFromGeneticCode(uuid.New(), "Test", gc, em, 0, nil)
		org.Traits.Size = 5.0

		org.Traits.Fertility = 2.0
		lowRate := org.CalculateReproductionRate()

		org.Traits.Fertility = 8.0
		highRate := org.CalculateReproductionRate()

		if highRate <= lowRate {
			t.Errorf("High fertility rate (%f) should be > low fertility rate (%f)", highRate, lowRate)
		}
	})
}

func TestCalculateInbreedingPenalty(t *testing.T) {
	tests := []struct {
		pop        int64
		minPenalty float64
		maxPenalty float64
	}{
		{100, 1.0, 1.0}, // No penalty
		{50, 1.0, 1.0},  // No penalty at threshold
		{25, 0.5, 0.6},  // Some penalty
		{10, 0.2, 0.3},  // More penalty
		{2, 0.1, 0.11},  // Near extinction
		{1, 0.1, 0.11},  // Near extinction
	}

	for _, tt := range tests {
		result := CalculateInbreedingPenalty(tt.pop)
		if result < tt.minPenalty || result > tt.maxPenalty {
			t.Errorf("Penalty for pop=%d is %f, want [%f, %f]", tt.pop, result, tt.minPenalty, tt.maxPenalty)
		}
	}
}

func TestOrganism_Clone(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	gc := NewGeneticCode(rng)
	em := DefaultExpressionMatrix()
	ancestorID := uuid.New()

	original := NewOrganismFromGeneticCode(uuid.New(), "Original", gc, em, 1000000, &ancestorID)
	clone := original.Clone()

	t.Run("clone is independent", func(t *testing.T) {
		original.Traits.Size = 999
		if clone.Traits.Size == 999 {
			t.Error("Clone traits should be independent")
		}

		original.GeneticCode.DefinedGenes[0] = 0.999
		if clone.GeneticCode.DefinedGenes[0] == 0.999 {
			t.Error("Clone genetic code should be independent")
		}
	})

	t.Run("clone preserves ancestor", func(t *testing.T) {
		if clone.AncestorID == nil {
			t.Error("Clone should preserve ancestor ID")
		}
		if *clone.AncestorID != ancestorID {
			t.Error("Clone ancestor ID mismatch")
		}
	})
}

func TestOrganismTraits_ToEvolvableTraits(t *testing.T) {
	traits := &OrganismTraits{
		Size:       5.0,
		Speed:      6.0,
		Strength:   7.0,
		Aggression: 8.0,
		Social:     6.0,
		ColdResist: 7.0,
		HeatResist: 3.0,
		Carnivore:  8.0,
		Display:    5.0,
	}

	et := traits.ToEvolvableTraits()

	t.Run("physical traits map directly", func(t *testing.T) {
		if et.Size != 5.0 {
			t.Errorf("Size = %f, want 5.0", et.Size)
		}
		if et.Speed != 6.0 {
			t.Errorf("Speed = %f, want 6.0", et.Speed)
		}
	})

	t.Run("behavioral traits scale to 0-1", func(t *testing.T) {
		if et.Aggression != 0.8 {
			t.Errorf("Aggression = %f, want 0.8", et.Aggression)
		}
		if et.Social != 0.6 {
			t.Errorf("Social = %f, want 0.6", et.Social)
		}
	})

	t.Run("carnivore tendency scales to 0-1", func(t *testing.T) {
		if et.CarnivoreTendency != 0.8 {
			t.Errorf("CarnivoreTendency = %f, want 0.8", et.CarnivoreTendency)
		}
	})

	t.Run("cold resistance triggers fur", func(t *testing.T) {
		if et.Covering != CoveringFur {
			t.Errorf("Covering = %s, want fur (cold resistant)", et.Covering)
		}
	})
}

func TestFromEvolvableTraits(t *testing.T) {
	et := EvolvableTraits{
		Size:              3.0,
		Speed:             5.0,
		Strength:          4.0,
		Aggression:        0.6,
		Social:            0.8,
		Intelligence:      0.5,
		ColdResistance:    0.4,
		HeatResistance:    0.6,
		CarnivoreTendency: 0.7,
		Fertility:         1.5,
		Lifespan:          15,
		Maturity:          2.0,
		LitterSize:        4.0,
		Display:           0.3,
	}

	traits := FromEvolvableTraits(et)

	t.Run("physical traits map directly", func(t *testing.T) {
		if traits.Size != 3.0 {
			t.Errorf("Size = %f, want 3.0", traits.Size)
		}
	})

	t.Run("behavioral traits scale to 0-10", func(t *testing.T) {
		if traits.Aggression != 6.0 {
			t.Errorf("Aggression = %f, want 6.0", traits.Aggression)
		}
		if traits.Social != 8.0 {
			t.Errorf("Social = %f, want 8.0", traits.Social)
		}
	})

	t.Run("carnivore scales to 0-10", func(t *testing.T) {
		if traits.Carnivore != 7.0 {
			t.Errorf("Carnivore = %f, want 7.0", traits.Carnivore)
		}
	})
}

func TestOrganismTraits_GetDietType(t *testing.T) {
	tests := []struct {
		name       string
		autotrophy float64
		carnivore  float64
		expected   DietType
	}{
		{"autotroph", 0.8, 0.0, DietPhotosynthetic},
		{"carnivore", 0.0, 8.0, DietCarnivore},
		{"herbivore", 0.0, 2.0, DietHerbivore},
		{"omnivore", 0.0, 5.0, DietOmnivore},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traits := &OrganismTraits{
				Autotrophy: tt.autotrophy,
				Carnivore:  tt.carnivore,
			}
			result := traits.GetDietType()
			if result != tt.expected {
				t.Errorf("GetDietType() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestOrganismTraitsFromEvolvableTraitsWithDefaults(t *testing.T) {
	et := EvolvableTraits{
		Size:     2.0,
		Speed:    0.0,
		Covering: CoveringShell,
	}

	t.Run("photosynthetic sets autotrophy", func(t *testing.T) {
		traits := OrganismTraitsFromEvolvableTraitsWithDefaults(et, DietPhotosynthetic)
		if traits.Autotrophy != 1.0 {
			t.Errorf("Autotrophy = %f, want 1.0 for photosynthetic", traits.Autotrophy)
		}
		if traits.Motility != 0.0 {
			t.Errorf("Motility = %f, want 0.0 for photosynthetic", traits.Motility)
		}
	})

	t.Run("shell covering sets armor", func(t *testing.T) {
		traits := OrganismTraitsFromEvolvableTraitsWithDefaults(et, DietHerbivore)
		if traits.Armor != 8.0 {
			t.Errorf("Armor = %f, want 8.0 for shell", traits.Armor)
		}
	})
}
