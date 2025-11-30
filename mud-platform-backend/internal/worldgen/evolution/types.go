package evolution

import (
	"github.com/google/uuid"
)

// SpeciesType categorizes species by trophic level
type SpeciesType string

const (
	SpeciesFlora     SpeciesType = "flora"
	SpeciesHerbivore SpeciesType = "herbivore"
	SpeciesCarnivore SpeciesType = "carnivore"
	SpeciesOmnivore  SpeciesType = "omnivore"
)

// DietType specifies how a species obtains energy
type DietType string

const (
	DietPhotosynthesis DietType = "photosynthesis"
	DietHerbivore      DietType = "herbivore"
	DietCarnivore      DietType = "carnivore"
	DietOmnivore       DietType = "omnivore"
)

// TemperatureRange defines acceptable temperature bounds
type TemperatureRange struct {
	Min     float64 // °C
	Max     float64 // °C
	Optimal float64 // °C
}

// MoistureRange defines acceptable moisture/precipitation bounds
type MoistureRange struct {
	Min     float64 // mm/year
	Max     float64 // mm/year
	Optimal float64 // mm/year
}

// ElevationRange defines acceptable elevation bounds
type ElevationRange struct {
	Min     float64 // meters
	Max     float64 // meters
	Optimal float64 // meters
}

// Species represents a distinct species with traits and population
type Species struct {
	SpeciesID  uuid.UUID
	Name       string
	Type       SpeciesType
	Generation int // Evolutionary generation number

	// Physical Traits
	Size       float64 // kg for fauna, height in m for flora
	Speed      float64 // m/s for fauna, 0 for flora
	Armor      float64 // 0-100, defense rating
	Camouflage float64 // 0-100, stealth rating

	// Dietary Needs
	Diet            DietType
	CaloriesPerDay  int
	PreferredPrey   []uuid.UUID // For carnivores
	PreferredPlants []uuid.UUID // For herbivores

	// Habitat
	PreferredBiomes      []string
	TemperatureTolerance TemperatureRange
	MoistureTolerance    MoistureRange
	ElevationTolerance   ElevationRange

	// Reproduction
	ReproductionRate float64 // Offspring per year
	MaturityAge      int     // Years to adulthood
	Lifespan         int     // Years

	// Population
	Population        int
	PopulationDensity float64 // Per km²
	ExtinctionRisk    float64 // 0-1, higher = more at risk
	PeakPopulation    int     // Historical peak for bottleneck detection

	// Evolution
	MutationRate    float64
	FitnessScore    float64
	ParentSpeciesID *uuid.UUID // null for initial species
}

// Environment represents environmental conditions in a region
type Environment struct {
	Temperature   float64
	Moisture      float64 // Precipitation mm/year
	Elevation     float64
	Sunlight      float64 // 0-1 scale
	BiomeName     string
	FoodAvailable float64 // 0-1 scale of available food
}

// IsInTolerance checks if species can survive in environment
func (s *Species) IsInTolerance(env *Environment) bool {
	tempOK := env.Temperature >= s.TemperatureTolerance.Min &&
		env.Temperature <= s.TemperatureTolerance.Max

	moistOK := env.Moisture >= s.MoistureTolerance.Min &&
		env.Moisture <= s.MoistureTolerance.Max

	elevOK := env.Elevation >= s.ElevationTolerance.Min &&
		env.Elevation <= s.ElevationTolerance.Max

	return tempOK && moistOK && elevOK
}

// IsFlora returns true if species is a plant
func (s *Species) IsFlora() bool {
	return s.Type == SpeciesFlora
}

// IsHerbivore returns true if species is an herbivore
func (s *Species) IsHerbivore() bool {
	return s.Type == SpeciesHerbivore || s.Diet == DietHerbivore
}

// IsCarnivore returns true if species is a carnivore
func (s *Species) IsCarnivore() bool {
	return s.Type == SpeciesCarnivore || s.Diet == DietCarnivore
}

// IsBottlenecked returns true if population is in a genetic bottleneck
func (s *Species) IsBottlenecked() bool {
	if s.PeakPopulation == 0 {
		return false
	}
	return s.Population < int(float64(s.PeakPopulation)*0.2)
}

// Clone creates a deep copy of the species
func (s *Species) Clone() *Species {
	clone := *s
	clone.SpeciesID = uuid.New()

	// Deep copy slices
	clone.PreferredPrey = make([]uuid.UUID, len(s.PreferredPrey))
	copy(clone.PreferredPrey, s.PreferredPrey)

	clone.PreferredPlants = make([]uuid.UUID, len(s.PreferredPlants))
	copy(clone.PreferredPlants, s.PreferredPlants)

	clone.PreferredBiomes = make([]string, len(s.PreferredBiomes))
	copy(clone.PreferredBiomes, s.PreferredBiomes)

	return &clone
}
