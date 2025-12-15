package population

import (
	"math"
	"math/rand"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// PopulationSimulator handles macro-level population dynamics
type PopulationSimulator struct {
	Biomes       map[uuid.UUID]*BiomePopulation
	FossilRecord *FossilRecord
	CurrentYear  int64
	rng          *rand.Rand
}

// NewPopulationSimulator creates a new simulator
func NewPopulationSimulator(worldID uuid.UUID, seed int64) *PopulationSimulator {
	return &PopulationSimulator{
		Biomes:       make(map[uuid.UUID]*BiomePopulation),
		FossilRecord: &FossilRecord{WorldID: worldID, Extinct: []*ExtinctSpecies{}},
		CurrentYear:  0,
		rng:          rand.New(rand.NewSource(seed)),
	}
}

// SimulateYear advances the simulation by one year using Lotka-Volterra dynamics
func (ps *PopulationSimulator) SimulateYear() {
	ps.CurrentYear++

	for _, biome := range ps.Biomes {
		biome.YearsSimulated++
		ps.simulateBiomeYear(biome)
	}
}

// simulateBiomeYear runs population dynamics for a single biome
func (ps *PopulationSimulator) simulateBiomeYear(biome *BiomePopulation) {
	// Count populations by diet type
	var floraCount, herbivoreCount, carnivoreCount int64
	for _, sp := range biome.Species {
		switch sp.Diet {
		case DietPhotosynthetic:
			floraCount += sp.Count
		case DietHerbivore:
			herbivoreCount += sp.Count
		case DietCarnivore, DietOmnivore:
			carnivoreCount += sp.Count
		}
	}

	// Track extinctions this year
	var toExtinct []uuid.UUID

	for speciesID, species := range biome.Species {
		oldCount := species.Count
		newCount := oldCount

		switch species.Diet {
		case DietPhotosynthetic:
			// Flora: logistic growth limited by carrying capacity
			// dP/dt = r * P * (1 - P/K)
			fitness := CalculateBiomeFitness(species.Traits, biome.BiomeType)
			growthRate := 0.5 * species.Traits.Fertility * fitness // Increased from 0.3 for faster recovery
			k := float64(biome.CarryingCapacity) * 0.4             // Flora takes 40% of capacity
			p := float64(oldCount)
			growth := growthRate * p * (1 - p/k)
			// Reduction from herbivore grazing
			grazingRate := 0.001 * float64(herbivoreCount) * (1 - species.Traits.Camouflage*0.3)
			// Seeds always survive - minimum population of 10
			newCount = int64(math.Max(10, p+growth-grazingRate*p))

		case DietHerbivore:
			// Herbivores: prey dynamics
			// dH/dt = (birth_rate * H) - (predation_rate * H * C)
			fitness := CalculateBiomeFitness(species.Traits, biome.BiomeType)
			birthRate := 0.25 * species.Traits.Fertility * fitness       // Slightly higher birth rate
			deathRate := (0.05 / species.Traits.Lifespan * 10) / fitness // Lower base death rate

			// Predation scales with predator count but herbivores get defensive bonuses
			predationRate := 0.0001 * (1 - species.Traits.Speed*0.05) * (1 - species.Traits.Camouflage*0.3)

			p := float64(oldCount)
			// More forgiving food availability - herbivores are efficient grazers
			foodAvailability := math.Min(1.0, float64(floraCount)/float64(oldCount+1)*0.3)
			if floraCount > 100 {
				foodAvailability = math.Max(0.5, foodAvailability) // Minimum 50% if flora exists
			}
			effectiveBirth := birthRate * foodAvailability

			predationLoss := predationRate * p * float64(carnivoreCount)
			growth := effectiveBirth*p - deathRate*p - predationLoss
			newCount = int64(math.Max(1, p+growth)) // Don't drop below 1 from dynamics alone

		case DietCarnivore, DietOmnivore:
			// Carnivores: predator dynamics with improved survival
			// dC/dt = (efficiency * predation * H * C) - (death_rate * C)
			fitness := CalculateBiomeFitness(species.Traits, biome.BiomeType)
			efficiency := 0.3 * (1 + species.Traits.Intelligence*0.3) * fitness
			predationRate := 0.002 * (0.5 + species.Traits.Speed*0.1) * (0.5 + species.Traits.Strength*0.1)
			deathRate := (0.05 / species.Traits.Lifespan * 10) / fitness

			p := float64(oldCount)
			preyCount := herbivoreCount
			if species.Diet == DietOmnivore {
				preyCount += floraCount / 5 // Omnivores get more calories from flora
			}

			// Ensure minimum prey availability for survival
			preyRatio := math.Min(1.0, float64(preyCount)/float64(oldCount+1)*0.2)
			growth := efficiency * predationRate * float64(preyCount) * p * preyRatio
			death := deathRate * p * (1 - preyRatio*0.5)  // Less death when prey available
			newCount = int64(math.Max(1, p+growth-death)) // Don't go below 1 unless truly extinct
		}

		// Apply carrying capacity limit
		if biome.TotalPopulation() > biome.CarryingCapacity {
			excess := float64(biome.TotalPopulation() - biome.CarryingCapacity)
			reduction := excess * float64(oldCount) / float64(biome.TotalPopulation())
			newCount = int64(math.Max(0, float64(newCount)-reduction))
		}

		// Add some randomness
		variance := float64(newCount) * 0.05
		newCount += int64(ps.rng.NormFloat64() * variance)
		if newCount < 0 {
			newCount = 0
		}

		species.Count = newCount

		// Check for extinction
		if species.Count <= 0 {
			toExtinct = append(toExtinct, speciesID)
		}
	}

	// Process extinctions
	for _, speciesID := range toExtinct {
		ps.recordExtinction(biome, speciesID, "population_collapse")
	}
}

// recordExtinction records a species going extinct
func (ps *PopulationSimulator) recordExtinction(biome *BiomePopulation, speciesID uuid.UUID, cause string) {
	species := biome.RemoveSpecies(speciesID)
	if species == nil {
		return
	}

	extinct := &ExtinctSpecies{
		SpeciesID:       species.SpeciesID,
		Name:            species.Name,
		Traits:          species.Traits,
		Diet:            species.Diet,
		PeakPopulation:  species.Count, // This would be tracked better in practice
		ExistedFrom:     species.CreatedYear,
		ExistedUntil:    ps.CurrentYear,
		ExtinctionCause: cause,
		FossilBiomes:    []uuid.UUID{biome.BiomeID},
	}

	ps.FossilRecord.Extinct = append(ps.FossilRecord.Extinct, extinct)
}

// SimulateYears runs the simulation for multiple years
func (ps *PopulationSimulator) SimulateYears(years int64) {
	for i := int64(0); i < years; i++ {
		ps.SimulateYear()

		// Every 1000 years, apply evolution
		if ps.CurrentYear%1000 == 0 {
			ps.ApplyEvolution()
		}

		// Every 10000 years, check for speciation
		if ps.CurrentYear%10000 == 0 {
			ps.CheckSpeciation()
		}
	}
}

// ApplyEvolution applies trait drift and selection pressure based on species-specific rates
// Species with earlier maturity and larger litter sizes evolve faster
func (ps *PopulationSimulator) ApplyEvolution() {
	for _, biome := range ps.Biomes {
		for _, species := range biome.Species {
			if species.Count == 0 {
				continue
			}

			// Calculate evolution rate based on breeding traits
			// Lower maturity = faster generations, higher litter size = more genetic variation
			maturity := species.Traits.Maturity
			if maturity < 0.5 {
				maturity = 0.5 // Minimum breeding age
			}
			litterSize := species.Traits.LitterSize
			if litterSize < 1 {
				litterSize = 1
			}

			// Evolution rate: higher litter size and lower maturity = faster evolution
			// Formula: (litter_size / maturity) gives "generations per year" effectively
			evolutionRate := litterSize / maturity

			// Apply multiple generations of evolution based on rate
			// In 1000 years with maturity=1 and litter=2, that's 2000 generations worth of selection
			generationsToApply := int64(evolutionRate * 1000) // Scale for 1000-year evolution cycles
			species.Generation += generationsToApply

			// Trait mutation (scaled by number of generations and variance)
			mutationStrength := 0.002 * species.TraitVariance * float64(generationsToApply)
			species.Traits.Size += ps.rng.NormFloat64() * mutationStrength * 0.5
			species.Traits.Speed += ps.rng.NormFloat64() * mutationStrength * 0.5
			species.Traits.Strength += ps.rng.NormFloat64() * mutationStrength * 0.5
			species.Traits.Aggression += ps.rng.NormFloat64() * mutationStrength * 0.1
			species.Traits.ColdResistance += ps.rng.NormFloat64() * mutationStrength * 0.1
			species.Traits.HeatResistance += ps.rng.NormFloat64() * mutationStrength * 0.1
			species.Traits.NightVision += ps.rng.NormFloat64() * mutationStrength * 0.1
			species.Traits.Camouflage += ps.rng.NormFloat64() * mutationStrength * 0.1
			species.Traits.Fertility += ps.rng.NormFloat64() * mutationStrength * 0.05
			species.Traits.Intelligence += ps.rng.NormFloat64() * mutationStrength * 0.05
			species.Traits.Maturity += ps.rng.NormFloat64() * mutationStrength * 0.02
			species.Traits.LitterSize += ps.rng.NormFloat64() * mutationStrength * 0.1

			// Clamp values
			species.Traits = clampTraits(species.Traits)

			// Selection pressure based on biome
			applyBiomeSelection(species, biome.BiomeType)
		}
	}
}

// CheckSpeciation checks if any species should split based on trait divergence
func (ps *PopulationSimulator) CheckSpeciation() {
	for _, biome := range ps.Biomes {
		var newSpecies []*SpeciesPopulation

		for _, species := range biome.Species {
			// Large populations with high variance may speciate
			if species.Count > 500 && species.TraitVariance > 0.3 && ps.rng.Float64() < 0.1 {
				// Split into two species
				child := &SpeciesPopulation{
					SpeciesID:     uuid.New(),
					Name:          species.Name + " (variant)",
					AncestorID:    &species.SpeciesID,
					Count:         species.Count / 3, // 1/3 goes to new species
					Traits:        mutateTraits(species.Traits, 0.15, ps.rng),
					TraitVariance: species.TraitVariance * 0.8,
					Diet:          species.Diet,
					Generation:    species.Generation + 1,
					CreatedYear:   ps.CurrentYear,
				}

				species.Count -= child.Count
				species.TraitVariance *= 0.8 // Reduce variance after split

				newSpecies = append(newSpecies, child)
			}
		}

		// Add new species to biome
		for _, sp := range newSpecies {
			biome.AddSpecies(sp)
		}
	}
}

// clampTraits ensures all trait values are within valid ranges
func clampTraits(t EvolvableTraits) EvolvableTraits {
	clamp := func(v, min, max float64) float64 {
		if v < min {
			return min
		}
		if v > max {
			return max
		}
		return v
	}
	t.Size = clamp(t.Size, 0.1, 10.0)
	t.Speed = clamp(t.Speed, 0.0, 10.0)
	t.Strength = clamp(t.Strength, 0.1, 10.0)
	t.Aggression = clamp(t.Aggression, 0.0, 1.0)
	t.Social = clamp(t.Social, 0.0, 1.0)
	t.Intelligence = clamp(t.Intelligence, 0.0, 1.0)
	t.ColdResistance = clamp(t.ColdResistance, 0.0, 1.0)
	t.HeatResistance = clamp(t.HeatResistance, 0.0, 1.0)
	t.NightVision = clamp(t.NightVision, 0.0, 1.0)
	t.Camouflage = clamp(t.Camouflage, 0.0, 1.0)
	t.Fertility = clamp(t.Fertility, 0.1, 3.0)
	t.Lifespan = clamp(t.Lifespan, 1.0, 200.0)
	t.Maturity = clamp(t.Maturity, 0.25, 20.0)    // Breeding age: 3 months to 20 years
	t.LitterSize = clamp(t.LitterSize, 1.0, 50.0) // 1 to 50 offspring
	t.CarnivoreTendency = clamp(t.CarnivoreTendency, 0.0, 1.0)
	t.VenomPotency = clamp(t.VenomPotency, 0.0, 1.0)
	t.PoisonResistance = clamp(t.PoisonResistance, 0.0, 1.0)
	return t
}

// mutateTraits creates a mutated copy of traits
func mutateTraits(t EvolvableTraits, rate float64, rng *rand.Rand) EvolvableTraits {
	mutate := func(v float64) float64 {
		return v + rng.NormFloat64()*rate
	}
	t.Size = mutate(t.Size)
	t.Speed = mutate(t.Speed)
	t.Strength = mutate(t.Strength)
	t.Aggression = mutate(t.Aggression)
	t.Social = mutate(t.Social)
	t.Intelligence = mutate(t.Intelligence)
	t.ColdResistance = mutate(t.ColdResistance)
	t.HeatResistance = mutate(t.HeatResistance)
	t.NightVision = mutate(t.NightVision)
	t.Camouflage = mutate(t.Camouflage)
	t.Fertility = mutate(t.Fertility)
	t.Lifespan = mutate(t.Lifespan)
	return clampTraits(t)
}

// applyBiomeSelection applies selection pressure based on biome type
func applyBiomeSelection(species *SpeciesPopulation, biomeType geography.BiomeType) {
	// Biome-specific selection pressure (small trait adjustments)
	switch biomeType {
	case geography.BiomeTundra:
		// Cold environments favor cold resistance
		species.Traits.ColdResistance += 0.01
		species.Traits.HeatResistance -= 0.005
	case geography.BiomeDesert:
		// Hot dry environments
		species.Traits.HeatResistance += 0.01
		species.Traits.ColdResistance -= 0.005
	case geography.BiomeRainforest:
		// Dense vegetation favors camouflage
		species.Traits.Camouflage += 0.005
	case geography.BiomeGrassland:
		// Open terrain favors speed
		species.Traits.Speed += 0.005
	}
	species.Traits = clampTraits(species.Traits)
}

// GetStats returns summary statistics for the simulation
func (ps *PopulationSimulator) GetStats() (totalPop, totalSpecies, totalExtinct int64) {
	for _, biome := range ps.Biomes {
		totalPop += biome.TotalPopulation()
		totalSpecies += int64(len(biome.Species))
	}
	totalExtinct = int64(len(ps.FossilRecord.Extinct))
	return
}

// ExtinctionEventType represents types of mass extinction events
type ExtinctionEventType string

const (
	EventVolcanicWinter   ExtinctionEventType = "volcanic_winter"
	EventAsteroidImpact   ExtinctionEventType = "asteroid_impact"
	EventIceAge           ExtinctionEventType = "ice_age"
	EventOceanAnoxia      ExtinctionEventType = "ocean_anoxia"
	EventFloodBasalt      ExtinctionEventType = "flood_basalt"
	EventContinentalDrift ExtinctionEventType = "continental_drift"
)

// ApplyExtinctionEvent reduces populations based on event type and severity
// severity is 0.0-1.0, where 1.0 is catastrophic
func (ps *PopulationSimulator) ApplyExtinctionEvent(eventType ExtinctionEventType, severity float64) int64 {
	var totalDeaths int64

	for _, biome := range ps.Biomes {
		var toExtinct []uuid.UUID

		for speciesID, species := range biome.Species {
			if species.Count == 0 {
				continue
			}

			// Base mortality from event type
			var mortalityRate float64
			switch eventType {
			case EventVolcanicWinter:
				// Blocks sunlight - flora affected but seeds/roots survive
				if species.Diet == DietPhotosynthetic {
					mortalityRate = 0.10 * severity // Reduced from 0.3 - plants can regrow from seeds
				} else {
					mortalityRate = 0.15 * severity
				}
				// Cold resistance helps survive
				mortalityRate *= (1.0 - species.Traits.ColdResistance*0.3)

			case EventAsteroidImpact:
				// Catastrophic - affects everything
				mortalityRate = 0.7 * severity
				// Small species survive better (less food needs)
				if species.Traits.Size < 2.0 {
					mortalityRate *= 0.6
				}
				// Intelligence helps find shelter
				mortalityRate *= (1.0 - species.Traits.Intelligence*0.2)

			case EventIceAge:
				// Cold kills tropical species, helps cold-adapted
				if biome.BiomeType == geography.BiomeRainforest || biome.BiomeType == geography.BiomeDesert {
					mortalityRate = 0.4 * severity
				} else {
					mortalityRate = 0.1 * severity
				}
				// Cold resistance dramatically reduces mortality
				mortalityRate *= (1.0 - species.Traits.ColdResistance*0.5)

			case EventOceanAnoxia:
				// Only affects ocean biomes
				if biome.BiomeType == geography.BiomeOcean {
					mortalityRate = 0.5 * severity
					// Larger species need more oxygen
					mortalityRate += species.Traits.Size * 0.02
				}

			case EventFloodBasalt:
				// Toxic gases kill land species
				if biome.BiomeType != geography.BiomeOcean {
					mortalityRate = 0.25 * severity
					// Poison resistance helps
					mortalityRate *= (1.0 - species.Traits.PoisonResistance*0.4)
				}

			case EventContinentalDrift:
				// Minor direct impact but increases speciation
				mortalityRate = 0.05 * severity
			}

			// Apply mortality
			deaths := int64(float64(species.Count) * mortalityRate)
			species.Count -= deaths
			totalDeaths += deaths

			// Check for extinction
			if species.Count <= 0 {
				species.Count = 0
				toExtinct = append(toExtinct, speciesID)
			}
		}

		// Process extinctions
		for _, speciesID := range toExtinct {
			ps.recordExtinction(biome, speciesID, string(eventType))
		}
	}

	return totalDeaths
}
