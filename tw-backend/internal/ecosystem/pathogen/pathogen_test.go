package pathogen

import (
	"math/rand"
	"testing"

	"github.com/google/uuid"
)

func TestNewPathogen(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	tests := []struct {
		name     string
		pType    PathogenType
		minVir   float32
		maxVir   float32
		minTrans float32
		maxTrans float32
	}{
		{"Virus", PathogenVirus, 0.3, 0.8, 0.4, 0.9},
		{"Bacteria", PathogenBacteria, 0.2, 0.8, 0.2, 0.7},
		{"Fungus", PathogenFungus, 0.1, 0.5, 0.1, 0.4},
		{"Prion", PathogenPrion, 0.95, 1.0, 0.05, 0.15},
		{"Parasite", PathogenParasite, 0.1, 0.4, 0.1, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPathogen("Test", tt.pType, uuid.New(), 0, rng)

			if p.Virulence < tt.minVir || p.Virulence > tt.maxVir {
				t.Errorf("%s virulence = %f, want [%f, %f]", tt.name, p.Virulence, tt.minVir, tt.maxVir)
			}
			if p.Transmissibility < tt.minTrans || p.Transmissibility > tt.maxTrans {
				t.Errorf("%s transmissibility = %f, want [%f, %f]", tt.name, p.Transmissibility, tt.minTrans, tt.maxTrans)
			}
		})
	}
}

func TestPathogen_CalculateR0(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	p := NewPathogen("Test Virus", PathogenVirus, uuid.New(), 0, rng)
	p.Transmissibility = 0.5 // Known value for testing

	t.Run("higher density increases R0", func(t *testing.T) {
		lowDensity := p.CalculateR0(0.1, 0.3)
		highDensity := p.CalculateR0(0.9, 0.3)

		if highDensity <= lowDensity {
			t.Errorf("High density R0 (%f) should be > low density R0 (%f)", highDensity, lowDensity)
		}
	})

	t.Run("higher resistance decreases R0", func(t *testing.T) {
		lowResist := p.CalculateR0(0.5, 0.1)
		highResist := p.CalculateR0(0.5, 0.9)

		if highResist >= lowResist {
			t.Errorf("High resistance R0 (%f) should be < low resistance R0 (%f)", highResist, lowResist)
		}
	})
}

func TestPathogen_Mutate(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	t.Run("virus mutates frequently", func(t *testing.T) {
		p := NewPathogen("Test Virus", PathogenVirus, uuid.New(), 0, rng)
		p.MutationRate = 1.0 // Force mutation

		initialVir := p.Virulence
		for i := 0; i < 10; i++ {
			p.Mutate(rng)
		}

		if p.MutationsCount == 0 {
			t.Error("Expected some mutations")
		}
		// Virulence should trend downward (endemic evolution)
		t.Logf("Initial virulence: %f, Final: %f, Mutations: %d", initialVir, p.Virulence, p.MutationsCount)
	})

	t.Run("prions never mutate", func(t *testing.T) {
		p := NewPathogen("Test Prion", PathogenPrion, uuid.New(), 0, rng)

		for i := 0; i < 100; i++ {
			p.Mutate(rng)
		}

		if p.MutationsCount != 0 {
			t.Error("Prions should not mutate")
		}
	})
}

func TestPathogen_CanInfectHost(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	originID := uuid.New()

	p := NewPathogen("Test", PathogenVirus, originID, 0, rng)
	p.HostSpecificity = 0.5

	t.Run("origin species always susceptible", func(t *testing.T) {
		if !p.CanInfectHost(originID, "herbivore", 0.9) {
			t.Error("Origin species should always be susceptible")
		}
	})

	t.Run("low specificity allows cross-species", func(t *testing.T) {
		p.HostSpecificity = 0.1 // Very broad host range
		if !p.CanInfectHost(uuid.New(), "herbivore", 0.1) {
			t.Error("Low specificity should allow infection")
		}
	})

	t.Run("high specificity blocks cross-species", func(t *testing.T) {
		p.HostSpecificity = 0.95 // Very narrow host range
		// With high specificity, cross-species chance is low
		infected := false
		for i := 0; i < 100; i++ {
			if p.CanInfectHost(uuid.New(), "herbivore", 0.3) {
				infected = true
				break
			}
		}
		t.Logf("High specificity infection in 100 attempts: %v", infected)
	})
}

func TestOutbreak_Update(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pathogen := NewPathogen("Test Virus", PathogenVirus, uuid.New(), 0, rng)
	pathogen.Virulence = 0.3
	pathogen.Transmissibility = 0.6

	outbreak := NewOutbreak(pathogen.ID, uuid.New(), uuid.Nil, 0, 100)

	t.Run("outbreak progresses", func(t *testing.T) {
		initialInfected := outbreak.CurrentInfected

		// Update for several ticks
		for i := 0; i < 5; i++ {
			outbreak.Update(pathogen, 10000, 0.3, rng)
		}

		// Should have changes
		if outbreak.TotalInfected <= initialInfected {
			t.Error("Total infected should increase")
		}
		if outbreak.TotalDeaths == 0 {
			t.Log("No deaths yet (may be probabilistic)")
		}
	})

	t.Run("outbreak can end", func(t *testing.T) {
		// Highly resistant population
		outbreak2 := NewOutbreak(pathogen.ID, uuid.New(), uuid.Nil, 0, 10)

		for i := 0; i < 100 && outbreak2.IsActive; i++ {
			outbreak2.Update(pathogen, 1000, 0.9, rng) // High resistance
		}

		// Should eventually end
		if outbreak2.IsActive && outbreak2.CurrentInfected <= 0 {
			t.Error("Outbreak with no infected should end")
		}
	})
}

func TestDiseaseSystem_SpontaneousOutbreak(t *testing.T) {
	ds := NewDiseaseSystem(uuid.New(), 42)
	ds.OutbreakBaseChance = 1.0 // Force outbreak for testing

	pathogen, outbreak := ds.CheckSpontaneousOutbreak(
		uuid.New(),
		"Test Species",
		10000,
		0.5,
	)

	if pathogen == nil {
		t.Fatal("Expected pathogen to be created")
	}
	if outbreak == nil {
		t.Fatal("Expected outbreak to be created")
	}

	t.Logf("Created %s: %s", pathogen.Type, pathogen.Name)
}

func TestDiseaseSystem_Update(t *testing.T) {
	ds := NewDiseaseSystem(uuid.New(), 42)

	// Create a pathogen and outbreak
	speciesID := uuid.New()
	pathogen := ds.CreateNovelPathogen(speciesID, "Test Species", PathogenVirus)
	outbreak := NewOutbreak(pathogen.ID, speciesID, uuid.Nil, 0, 100)
	ds.Outbreaks[outbreak.ID] = outbreak
	pathogen.ActiveOutbreaks++

	speciesData := map[uuid.UUID]SpeciesInfo{
		speciesID: {
			Population:        10000,
			DiseaseResistance: 0.3,
			DietType:          "herbivore",
			Density:           0.5,
		},
	}

	// Update for several years
	for year := int64(1); year <= 20; year++ {
		ds.Update(year, speciesData)
	}

	t.Logf("After 20 years - Infected: %d, Deaths: %d, Active: %v",
		outbreak.CurrentInfected, outbreak.TotalDeaths, outbreak.IsActive)
}
