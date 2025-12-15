package population

import (
	"math"
	"math/rand"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// PopulationSimulator handles macro-level population dynamics
type PopulationSimulator struct {
	Biomes                   map[uuid.UUID]*BiomePopulation
	FossilRecord             *FossilRecord
	CurrentYear              int64
	OxygenLevel              float64 // Atmospheric O2 as fraction (0.21 = 21%, modern baseline)
	ContinentalFragmentation float64 // 0.0 = supercontinent (Pangaea), 1.0 = fully fragmented
	RecoveryPhase            bool    // True if recovering from mass extinction
	RecoveryCounter          int64   // Years remaining in recovery phase
	rng                      *rand.Rand
}

// CalculateMetabolicRate returns the metabolic rate based on size using Kleiber's Law
// Metabolic rate scales with mass^0.75 (larger animals are more efficient per kg)
// Returns a multiplier for food requirements (1.0 = baseline at size 5)
func CalculateMetabolicRate(size float64) float64 {
	if size <= 0 {
		size = 1
	}
	// Kleiber's Law: B = B0 * M^0.75
	// Normalize to size 5 as baseline (medium animal)
	return math.Pow(size/5.0, 0.75)
}

// CalculateReproductionModifier returns reproduction rate modifier based on size
// Smaller animals reproduce faster (r-strategy), larger slower (K-strategy)
// Returns a multiplier for reproduction rate (1.0 = baseline at size 5)
func CalculateReproductionModifier(size float64) float64 {
	if size <= 0 {
		size = 1
	}
	// Inverse square root relationship - smaller = faster reproduction
	// Normalize to size 5 as baseline
	return math.Sqrt(5.0 / size)
}

// CalculateJuvenileSurvival returns the probability of a juvenile surviving to adulthood
// K-strategists (large, intelligent) have higher juvenile survival due to parental care
// r-strategists (small, high litter) have lower survival but compensate with numbers
func CalculateJuvenileSurvival(traits EvolvableTraits) float64 {
	// Base survival rate
	baseSurvival := 0.3

	// Size increases survival (larger animals protect offspring better)
	sizeBonus := math.Min(0.3, traits.Size*0.03)

	// Intelligence increases survival (better parenting, teaching)
	intelligenceBonus := traits.Intelligence * 0.2

	// Social animals have group protection for young
	socialBonus := traits.Social * 0.1

	// High fertility (r-strategy) typically means lower per-offspring investment
	fertilityPenalty := math.Min(0.2, (traits.Fertility-1.0)*0.1)
	if fertilityPenalty < 0 {
		fertilityPenalty = 0
	}

	survival := baseSurvival + sizeBonus + intelligenceBonus + socialBonus - fertilityPenalty

	// Clamp to reasonable range
	if survival < 0.1 {
		survival = 0.1
	}
	if survival > 0.9 {
		survival = 0.9
	}

	return survival
}

// CalculateMaturationRate returns the fraction of juveniles that mature to adult each year
// Based on maturity age - longer maturation = lower rate
func CalculateMaturationRate(maturityAge float64) float64 {
	if maturityAge < 0.5 {
		maturityAge = 0.5
	}
	// Rate is inverse of maturity age (1 year maturity = 100% mature per year)
	return 1.0 / maturityAge
}

// NewPopulationSimulator creates a new simulator
func NewPopulationSimulator(worldID uuid.UUID, seed int64) *PopulationSimulator {
	return &PopulationSimulator{
		Biomes:                   make(map[uuid.UUID]*BiomePopulation),
		FossilRecord:             &FossilRecord{WorldID: worldID, Extinct: []*ExtinctSpecies{}},
		CurrentYear:              0,
		OxygenLevel:              0.21, // Modern Earth baseline (21%)
		ContinentalFragmentation: 0.5,  // Start at medium fragmentation
		rng:                      rand.New(rand.NewSource(seed)),
	}
}

// UpdateContinentalConfiguration gradually changes continental fragmentation
// Continental drift events can trigger rapid changes
// Returns the new fragmentation level
func (ps *PopulationSimulator) UpdateContinentalConfiguration(driftEvent bool, severity float64) float64 {
	if driftEvent {
		// Drift events cause significant fragmentation changes
		change := ps.rng.NormFloat64() * 0.1 * severity
		ps.ContinentalFragmentation += change
	} else {
		// Very slow random walk over time
		ps.ContinentalFragmentation += ps.rng.NormFloat64() * 0.001
	}

	// Mean reversion toward 0.5 (equilibrium)
	ps.ContinentalFragmentation += (0.5 - ps.ContinentalFragmentation) * 0.0001

	// Clamp to valid range
	if ps.ContinentalFragmentation < 0 {
		ps.ContinentalFragmentation = 0
	}
	if ps.ContinentalFragmentation > 1 {
		ps.ContinentalFragmentation = 1
	}

	return ps.ContinentalFragmentation
}

// ApplyContinentalEffects applies effects based on continental configuration
// Supercontinent (0): Uniform climate, easy migration, lower endemism
// Fragmented (1): Diverse climates, species isolation, high endemism
func (ps *PopulationSimulator) ApplyContinentalEffects() int {
	affectedSpecies := 0
	frag := ps.ContinentalFragmentation

	for _, biome := range ps.Biomes {
		for _, species := range biome.Species {
			if species.Count == 0 {
				continue
			}

			// SUPERCONTINENT EFFECTS (low fragmentation)
			if frag < 0.3 {
				// Uniform climate reduces trait variance (genetic homogenization)
				if ps.rng.Float64() < 0.1 {
					species.TraitVariance *= 0.999
					if species.TraitVariance < 0.01 {
						species.TraitVariance = 0.01
					}
				}
				// Easier competition - slight population pressure
				if ps.rng.Float64() < 0.05 {
					species.Count = int64(float64(species.Count) * 0.999)
				}
				affectedSpecies++
			}

			// FRAGMENTED CONTINENT EFFECTS (high fragmentation)
			if frag > 0.7 {
				// Isolation increases trait variance (allopatric speciation driver)
				if ps.rng.Float64() < 0.1 {
					species.TraitVariance = math.Min(1.0, species.TraitVariance*1.002)
				}
				// Endemic adaptation to specific biomes
				if ps.rng.Float64() < 0.05 {
					// Boost traits that match biome
					applyBiomeSelection(species, biome.BiomeType)
				}
				affectedSpecies++
			}

			// MODERATE FRAGMENTATION (0.3-0.7): Balanced effects
			// This is the "Goldilocks zone" for biodiversity
		}
	}

	return affectedSpecies
}

// ApplyHabitatFragmentation applies effects of local habitat fragmentation
// Fragmentation reduces gene flow, increases genetic drift, stresses large species
func (ps *PopulationSimulator) ApplyHabitatFragmentation() int {
	affectedSpecies := 0

	for _, biome := range ps.Biomes {
		frag := biome.Fragmentation

		for _, species := range biome.Species {
			if species.Count == 0 {
				continue
			}

			// HIGHLY FRAGMENTED (>0.6): Stress effects
			if frag > 0.6 {
				// Large species suffer most from fragmentation (need large ranges)
				if species.Traits.Size > 4.0 {
					stress := (frag - 0.6) * (species.Traits.Size / 10.0)
					if ps.rng.Float64() < stress*0.1 {
						// Population penalty for large species
						species.Count = int64(float64(species.Count) * (1.0 - stress*0.05))
						if species.Count < 1 {
							species.Count = 1
						}
						affectedSpecies++
					}
				}

				// Increased genetic drift in fragmented habitats
				if ps.rng.Float64() < frag*0.1 {
					driftAmount := ps.rng.NormFloat64() * frag * 0.01
					traitIndex := ps.rng.Intn(5)
					switch traitIndex {
					case 0:
						species.Traits.Size += driftAmount * 2
					case 1:
						species.Traits.Speed += driftAmount * 2
					case 2:
						species.Traits.Strength += driftAmount * 2
					case 3:
						species.TraitVariance += driftAmount * 0.5
					case 4:
						species.Traits.Intelligence += driftAmount
					}
					species.Traits = clampTraits(species.Traits)
					affectedSpecies++
				}
			}

			// CONNECTED HABITAT (<0.3): Genetic homogenization
			if frag < 0.3 && ps.rng.Float64() < 0.05 {
				// Well-connected habitats slowly reduce extreme traits
				species.TraitVariance *= 0.999
				if species.TraitVariance < 0.01 {
					species.TraitVariance = 0.01
				}
			}
		}
	}

	return affectedSpecies
}

// ApplyRecoveryEffects applies dynamics specific to post-extinction recovery
// - Lilliput Effect: Large species are disadvantaged, small species thrive
// - Disaster Taxa: Generalists (high variance) get a boost
func (ps *PopulationSimulator) ApplyRecoveryEffects() {
	if !ps.RecoveryPhase {
		return
	}

	for _, biome := range ps.Biomes {
		for _, species := range biome.Species {
			if species.Count == 0 {
				continue
			}

			// Lilliput Effect: Large species suffer higher mortality during recovery instability
			if species.Traits.Size > 3.0 {
				sizePenalty := (species.Traits.Size - 3.0) * 0.05
				if ps.rng.Float64() < sizePenalty {
					species.Count = int64(float64(species.Count) * 0.9) // 10% reduction
				}
			} else if species.Traits.Size < 1.0 {
				// Small species ("disaster taxa") reproduce faster in empty niches
				if ps.rng.Float64() < 0.2 {
					species.Count = int64(float64(species.Count) * 1.1)
				}
			}

			// Generalists (high variance) survive better in unstable recovery period
			if species.TraitVariance > 0.5 {
				if ps.rng.Float64() < 0.1 {
					species.Count = int64(float64(species.Count) * 1.05)
				}
			}
		}
	}
}

// UpdateBiomeFragmentation changes fragmentation based on population and events
func (ps *PopulationSimulator) UpdateBiomeFragmentation() {
	for _, biome := range ps.Biomes {
		// Fragmentation naturally increases slightly over time (habitat loss)
		biome.Fragmentation += ps.rng.NormFloat64() * 0.001

		// High population density can fragment habitat
		totalPop := biome.TotalPopulation()
		density := float64(totalPop) / float64(biome.CarryingCapacity+1)
		if density > 0.8 {
			biome.Fragmentation += 0.001 * (density - 0.8)
		}

		// Mean reversion toward 0.3 (equilibrium)
		biome.Fragmentation += (0.3 - biome.Fragmentation) * 0.0001

		// Clamp
		if biome.Fragmentation < 0 {
			biome.Fragmentation = 0
		}
		if biome.Fragmentation > 1.0 {
			biome.Fragmentation = 1.0
		}
	}
}

// UpdateOxygenLevel slowly varies atmospheric oxygen over geological time
// Based on historical Earth data: O2 varied from 15% to 35% over billions of years
// Returns the new oxygen level
func (ps *PopulationSimulator) UpdateOxygenLevel() float64 {
	// Slow random walk with mean reversion to 0.21
	meanReversion := (0.21 - ps.OxygenLevel) * 0.0001
	randomChange := ps.rng.NormFloat64() * 0.0005 // Very slow change

	ps.OxygenLevel += meanReversion + randomChange

	// Clamp to historical extremes (15% - 35%)
	if ps.OxygenLevel < 0.15 {
		ps.OxygenLevel = 0.15
	}
	if ps.OxygenLevel > 0.35 {
		ps.OxygenLevel = 0.35
	}

	return ps.OxygenLevel
}

// CalculateOxygenSizeModifier returns how oxygen affects maximum viable size
// High O2 (>25%): Allows giant insects and larger organisms
// Low O2 (<18%): Limits maximum size, stresses large organisms
func CalculateOxygenSizeModifier(oxygenLevel float64) float64 {
	// Normalize around 21% baseline
	// At 35% O2: 1.67x size bonus (Carboniferous-like giant insects)
	// At 21% O2: 1.0x (baseline)
	// At 15% O2: 0.71x (size penalty)
	return oxygenLevel / 0.21
}

// ApplyOxygenEffects applies oxygen-based selection pressure to all species
// High O2 benefits large organisms, low O2 penalizes them
func (ps *PopulationSimulator) ApplyOxygenEffects() int {
	affectedSpecies := 0
	oxygenModifier := CalculateOxygenSizeModifier(ps.OxygenLevel)

	for _, biome := range ps.Biomes {
		for _, species := range biome.Species {
			if species.Count == 0 || species.Diet == DietPhotosynthetic {
				continue // Plants don't breathe O2
			}

			// Calculate size stress based on oxygen levels
			sizeRatio := species.Traits.Size / 5.0 // Normalize to medium size

			// Large organisms (size > 5) are stressed by low oxygen
			// Small organisms are less affected
			if sizeRatio > 1.0 && oxygenModifier < 1.0 {
				// Stress proportional to how oversized for current O2
				stress := (sizeRatio - 1.0) * (1.0 - oxygenModifier)
				// Slight population penalty and evolutionary pressure toward smaller size
				if stress > 0.1 && ps.rng.Float64() < stress {
					species.Traits.Size -= 0.01 * stress // Evolve smaller
					species.Count = int64(float64(species.Count) * (1.0 - stress*0.01))
					affectedSpecies++
				}
			}

			// High oxygen allows larger sizes - reduce size pressure
			if oxygenModifier > 1.0 && ps.rng.Float64() < 0.1 {
				// Slight pressure toward larger sizes when O2 is high
				species.Traits.Size += 0.005 * (oxygenModifier - 1.0)
				if species.Traits.Size > 10 {
					species.Traits.Size = 10
				}
			}
		}
	}

	return affectedSpecies
}

// ApplyAgeStructure processes age-related population dynamics each year
// - Juveniles mature to adults based on maturity age
// - Juveniles have higher mortality (predation targets young)
// - Adults die based on lifespan
// - Births add to juvenile population (not adult)
func (ps *PopulationSimulator) ApplyAgeStructure() {
	for _, biome := range ps.Biomes {
		// Count predators for juvenile predation modifier
		var predatorPop int64
		for _, sp := range biome.Species {
			if sp.Diet == DietCarnivore || sp.Diet == DietOmnivore {
				predatorPop += sp.Count
			}
		}
		predatorDensity := float64(predatorPop) / float64(biome.CarryingCapacity+1)

		for _, species := range biome.Species {
			if species.Count == 0 || species.Diet == DietPhotosynthetic {
				continue // Flora don't have juveniles in this model
			}

			// Initialize age structure if not set (migration from old data)
			if species.JuvenileCount == 0 && species.AdultCount == 0 && species.Count > 0 {
				// Assume 70% adults, 30% juveniles at steady state
				species.AdultCount = int64(float64(species.Count) * 0.7)
				species.JuvenileCount = species.Count - species.AdultCount
			}

			// Calculate survival and maturation rates
			juvenileSurvival := CalculateJuvenileSurvival(species.Traits)
			maturationRate := CalculateMaturationRate(species.Traits.Maturity)

			// Predators preferentially hunt juveniles (easier prey)
			predatorPenalty := predatorDensity * 0.2
			juvenileSurvival *= (1.0 - predatorPenalty)

			// Apply juvenile mortality
			survivingJuveniles := int64(float64(species.JuvenileCount) * juvenileSurvival)

			// Juveniles mature to adults
			maturing := int64(float64(survivingJuveniles) * maturationRate)
			if maturing > survivingJuveniles {
				maturing = survivingJuveniles
			}
			survivingJuveniles -= maturing

			// Adult mortality based on lifespan
			adultSurvival := 1.0 - (1.0 / species.Traits.Lifespan)
			if adultSurvival < 0.5 {
				adultSurvival = 0.5
			}
			survivingAdults := int64(float64(species.AdultCount) * adultSurvival)

			// Update counts
			species.JuvenileCount = survivingJuveniles
			species.AdultCount = survivingAdults + maturing
			species.Count = species.JuvenileCount + species.AdultCount

			// Ensure minimums
			if species.Count < 1 && (species.JuvenileCount > 0 || species.AdultCount > 0) {
				species.Count = 1
			}
		}
	}
}

// ApplySexualSelection implements Fisher's runaway selection and the handicap principle
// High display traits boost reproduction success but increase predation vulnerability
// Returns the number of species where sexual selection affected evolution
func (ps *PopulationSimulator) ApplySexualSelection() int {
	affectedSpecies := 0

	for _, biome := range ps.Biomes {
		// Count predator presence for handicap cost calculation
		var predatorPop int64
		for _, sp := range biome.Species {
			if sp.Diet == DietCarnivore || sp.Diet == DietOmnivore {
				predatorPop += sp.Count
			}
		}
		predatorDensity := float64(predatorPop) / float64(biome.CarryingCapacity+1)

		for _, species := range biome.Species {
			if species.Count == 0 || species.Diet == DietPhotosynthetic {
				continue // Plants don't have sexual selection in this model
			}

			display := species.Traits.Display

			// BENEFIT: High display increases reproductive success
			// (Fisherian runaway selection - females prefer showy males)
			if display > 0.2 && ps.rng.Float64() < display*0.3 {
				// Display boosts effective fertility
				fertilityBonus := display * 0.1
				species.Traits.Fertility = math.Min(3.0, species.Traits.Fertility+fertilityBonus*0.01)

				// Runaway selection: tendency to evolve even more display
				if ps.rng.Float64() < 0.2 {
					species.Traits.Display = math.Min(1.0, species.Traits.Display+0.005)
				}
				affectedSpecies++
			}

			// COST: High display increases predation risk (handicap principle)
			// Bright colors, large antlers make animals easier to spot
			if display > 0.3 && predatorDensity > 0.1 {
				predationCost := display * predatorDensity * 0.02
				// Slight population penalty from increased predation
				if ps.rng.Float64() < predationCost {
					species.Count = int64(float64(species.Count) * (1.0 - predationCost*0.1))
					if species.Count < 1 {
						species.Count = 1
					}
				}

				// Selection pressure to reduce display when predators are abundant
				if ps.rng.Float64() < predatorDensity*0.1 {
					species.Traits.Display = math.Max(0, species.Traits.Display-0.002)
				}
			}

			// EQUILIBRIUM: Display evolves based on predator-reproduction balance
			// Low predator environments allow extravagant displays (island effect)
			if predatorDensity < 0.05 && display < 0.5 && ps.rng.Float64() < 0.1 {
				species.Traits.Display = math.Min(1.0, species.Traits.Display+0.003)
			}
		}
	}

	return affectedSpecies
}

// SimulateYear advances the simulation by one year using Lotka-Volterra dynamics
func (ps *PopulationSimulator) SimulateYear() {
	ps.CurrentYear++

	for _, biome := range ps.Biomes {
		biome.YearsSimulated++
		ps.simulateBiomeYear(biome)
	}

	// Apply age structure transitions (juveniles mature, mortality by age)
	ps.ApplyAgeStructure()

	// Check for mass extinction events (triggers recovery phase)
	ps.CheckForMassExtinction()

	// Apply post-extinction recovery dynamics (Lilliput effect, etc.)
	ps.ApplyRecoveryEffects()
}

// simulateBiomeYear runs population dynamics for a single biome
func (ps *PopulationSimulator) simulateBiomeYear(biome *BiomePopulation) {
	// Calculate current season and modifiers
	season := GetSeasonFromYear(ps.CurrentYear)
	foodModifier := SeasonalFoodModifier(season, biome.BiomeType)
	breedingModifier := SeasonalBreedingModifier(season, biome.BiomeType)

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
			// Apply seasonal growth modifier - plants grow more in summer, less in winter
			growthRate := 0.5 * species.Traits.Fertility * fitness * foodModifier
			k := float64(biome.CarryingCapacity) * 0.4 // Flora takes 40% of capacity
			p := float64(oldCount)
			growth := growthRate * p * (1 - p/k)
			// Reduction from herbivore grazing
			grazingRate := 0.001 * float64(herbivoreCount) * (1 - species.Traits.Camouflage*0.3)
			// Seeds always survive - minimum population of 10
			newCount = int64(math.Max(10, p+growth-grazingRate*p))

		case DietHerbivore:
			// Herbivores: prey dynamics with Kleiber's Law
			// dH/dt = (birth_rate * H) - (predation_rate * H * C)
			fitness := CalculateBiomeFitness(species.Traits, biome.BiomeType)

			// Kleiber's Law: smaller animals reproduce faster, larger need more food
			reproModifier := CalculateReproductionModifier(species.Traits.Size)
			metabolicRate := CalculateMetabolicRate(species.Traits.Size)

			// Apply seasonal breeding modifier - animals breed more in spring/summer
			birthRate := 0.25 * species.Traits.Fertility * fitness * reproModifier * breedingModifier
			deathRate := (0.05 / species.Traits.Lifespan * 10) / fitness

			// Predation scales with predator count but herbivores get defensive bonuses
			// Larger herbivores are harder to take down
			sizeDefense := 1.0 - math.Min(0.3, species.Traits.Size*0.03)
			predationRate := 0.0001 * (1 - species.Traits.Speed*0.05) * (1 - species.Traits.Camouflage*0.3) * sizeDefense

			p := float64(oldCount)
			// Food availability scaled by metabolic rate and season - larger animals need more
			foodAvailability := math.Min(1.0, float64(floraCount)/float64(oldCount+1)*0.3/metabolicRate*foodModifier)
			if floraCount > 100 {
				foodAvailability = math.Max(0.3, foodAvailability) // Minimum 30% if flora exists
			}
			effectiveBirth := birthRate * foodAvailability

			predationLoss := predationRate * p * float64(carnivoreCount)
			growth := effectiveBirth*p - deathRate*p - predationLoss
			newCount = int64(math.Max(1, p+growth)) // Don't drop below 1 from dynamics alone

		case DietCarnivore, DietOmnivore:
			// Carnivores: predator dynamics with Kleiber's Law
			// dC/dt = (efficiency * predation * H * C) - (death_rate * C)
			fitness := CalculateBiomeFitness(species.Traits, biome.BiomeType)

			// Kleiber's Law: smaller predators reproduce faster, larger are better hunters
			reproModifier := CalculateReproductionModifier(species.Traits.Size)
			metabolicRate := CalculateMetabolicRate(species.Traits.Size)

			// Larger predators are more effective hunters but need more food
			sizeHuntingBonus := 1.0 + math.Min(0.3, species.Traits.Size*0.03)
			efficiency := 0.3 * (1 + species.Traits.Intelligence*0.3) * fitness * sizeHuntingBonus
			predationRate := 0.002 * (0.5 + species.Traits.Speed*0.1) * (0.5 + species.Traits.Strength*0.1)
			deathRate := (0.05 / species.Traits.Lifespan * 10) / fitness

			p := float64(oldCount)
			preyCount := herbivoreCount
			if species.Diet == DietOmnivore {
				preyCount += floraCount / 5 // Omnivores get more calories from flora
			}

			// Prey ratio scaled by metabolic rate - larger predators need more prey
			preyRatio := math.Min(1.0, float64(preyCount)/float64(oldCount+1)*0.2/metabolicRate)
			// Apply seasonal breeding modifier to growth
			growth := efficiency * predationRate * float64(preyCount) * p * preyRatio * reproModifier * breedingModifier
			death := deathRate * p * (1 - preyRatio*0.5)  // Less death when prey available
			newCount = int64(math.Max(1, p+growth-death)) // Don't go below 1 unless truly extinct
		}

		// Apply carrying capacity limit (biome-level)
		if biome.TotalPopulation() > biome.CarryingCapacity {
			excess := float64(biome.TotalPopulation() - biome.CarryingCapacity)
			reduction := excess * float64(oldCount) / float64(biome.TotalPopulation())
			newCount = int64(math.Max(0, float64(newCount)-reduction))
		}

		// Apply trophic pyramid limits (ecological carrying capacity)
		trophicLevel := GetTrophicLevel(species.Diet)
		var trophicCapacity int64
		switch trophicLevel {
		case TrophicProducer:
			trophicCapacity = biome.CarryingCapacity // Limited by biome
		case TrophicPrimaryConsumer:
			trophicCapacity = CalculateTrophicCapacity(trophicLevel, floraCount)
		case TrophicSecondaryConsumer, TrophicApexPredator:
			trophicCapacity = CalculateTrophicCapacity(trophicLevel, herbivoreCount)
		}
		// If this species exceeds its share of trophic capacity, reduce it
		if trophicCapacity > 0 && newCount > trophicCapacity {
			newCount = trophicCapacity
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

// ApplyGeneticDrift applies random allele frequency changes based on population size
// Smaller populations experience stronger drift (founder effect / bottleneck)
// Returns the number of species affected by significant drift
func (ps *PopulationSimulator) ApplyGeneticDrift() int {
	driftEvents := 0

	for _, biome := range ps.Biomes {
		for _, species := range biome.Species {
			if species.Count == 0 {
				continue
			}

			// Drift strength inversely proportional to population size
			// Formula: drift = 1 / sqrt(2 * N) (Wright-Fisher model approximation)
			// At N=10: drift=0.22, At N=100: drift=0.07, At N=1000: drift=0.02
			driftStrength := 1.0 / math.Sqrt(2.0*float64(species.Count))

			// Only apply significant drift to small populations (< 500)
			if species.Count > 500 {
				driftStrength *= 0.1 // Much weaker for large populations
			}

			// Very small populations (< 50) experience founder effect
			founderBonus := 0.0
			if species.Count < 50 {
				founderBonus = 0.5 * (1.0 - float64(species.Count)/50.0)
			}

			totalDrift := (driftStrength + founderBonus) * 0.1

			// Apply random drift to traits
			if ps.rng.Float64() < totalDrift*5 { // 5x chance for drift check
				// Pick a random trait to drift
				traitIndex := ps.rng.Intn(8)
				driftAmount := ps.rng.NormFloat64() * totalDrift

				switch traitIndex {
				case 0:
					species.Traits.Size += driftAmount * 2
				case 1:
					species.Traits.Speed += driftAmount * 2
				case 2:
					species.Traits.Strength += driftAmount * 2
				case 3:
					species.Traits.ColdResistance += driftAmount
				case 4:
					species.Traits.HeatResistance += driftAmount
				case 5:
					species.Traits.Camouflage += driftAmount
				case 6:
					species.Traits.Intelligence += driftAmount
				case 7:
					// Variance can drift too - diversity can be lost or gained
					species.TraitVariance += driftAmount * 0.5
				}

				species.Traits = clampTraits(species.Traits)
				if species.TraitVariance < 0.01 {
					species.TraitVariance = 0.01 // Minimum variance
				}
				if species.TraitVariance > 1.0 {
					species.TraitVariance = 1.0
				}
				driftEvents++
			}
		}
	}

	return driftEvents
}

// ApplyCoEvolution applies the Red Queen effect - predator-prey arms race
// Predators drive prey to evolve escape traits, prey drive predators to evolve hunting traits
func (ps *PopulationSimulator) ApplyCoEvolution() int {
	coevolutionEvents := 0

	for _, biome := range ps.Biomes {
		// Count populations by trophic level
		var preyPop, predatorPop int64
		var preySpecies, predatorSpecies []*SpeciesPopulation

		for _, species := range biome.Species {
			switch species.Diet {
			case DietHerbivore:
				preyPop += species.Count
				preySpecies = append(preySpecies, species)
			case DietCarnivore, DietOmnivore:
				predatorPop += species.Count
				predatorSpecies = append(predatorSpecies, species)
			}
		}

		if preyPop == 0 || predatorPop == 0 {
			continue // No arms race without both sides
		}

		// Calculate predation pressure (how much predators threaten prey)
		predationPressure := float64(predatorPop) / float64(preyPop+predatorPop)

		// Calculate escape pressure (how hard prey are to catch)
		escapePressure := float64(preyPop) / float64(preyPop+predatorPop)

		// Prey evolve escape traits when predator pressure is high
		if predationPressure > 0.2 && ps.rng.Float64() < predationPressure {
			for _, prey := range preySpecies {
				if prey.Count < 10 {
					continue
				}
				// Prey evolve toward speed, camouflage, or size (harder to catch)
				traitBoost := predationPressure * 0.01
				if ps.rng.Float64() < 0.5 {
					prey.Traits.Speed = math.Min(10, prey.Traits.Speed+traitBoost)
				} else {
					prey.Traits.Camouflage = math.Min(1.0, prey.Traits.Camouflage+traitBoost*0.5)
				}
				prey.TraitVariance = math.Min(1.0, prey.TraitVariance+0.01)
				coevolutionEvents++
			}
		}

		// Predators evolve hunting traits when escape pressure is high
		if escapePressure > 0.6 && ps.rng.Float64() < escapePressure*0.5 {
			for _, predator := range predatorSpecies {
				if predator.Count < 5 {
					continue
				}
				// Predators evolve toward speed, strength, or intelligence
				traitBoost := escapePressure * 0.01
				roll := ps.rng.Float64()
				if roll < 0.33 {
					predator.Traits.Speed = math.Min(10, predator.Traits.Speed+traitBoost)
				} else if roll < 0.66 {
					predator.Traits.Strength = math.Min(10, predator.Traits.Strength+traitBoost)
				} else {
					predator.Traits.Intelligence = math.Min(10, predator.Traits.Intelligence+traitBoost*0.5)
				}
				predator.TraitVariance = math.Min(1.0, predator.TraitVariance+0.01)
				coevolutionEvents++
			}
		}
	}

	return coevolutionEvents
}

// CheckForMassExtinction detects if a mass extinction event has occurred
// Triggers recovery phase if >50% of species went extinct recently
func (ps *PopulationSimulator) CheckForMassExtinction() {
	if ps.RecoveryPhase {
		ps.RecoveryCounter--
		if ps.RecoveryCounter <= 0 {
			ps.RecoveryPhase = false
		}
		return
	}

	// Only check periodically to save performance (every 100 years)
	if ps.CurrentYear%100 != 0 {
		return
	}

	if ps.FossilRecord == nil || len(ps.FossilRecord.Extinct) == 0 {
		return
	}

	// Count extinctions in last 500 years
	recentExtinctions := 0
	checkWindow := int64(500)
	for i := len(ps.FossilRecord.Extinct) - 1; i >= 0; i-- {
		fossil := ps.FossilRecord.Extinct[i]
		if ps.CurrentYear-fossil.ExistedUntil < checkWindow {
			recentExtinctions++
		} else {
			break // Sorted by time usually, or just optimization
		}
	}

	if recentExtinctions == 0 {
		return
	}

	// Estimate total species count checkWindow years ago
	currentSpecies := 0
	for _, biome := range ps.Biomes {
		currentSpecies += len(biome.Species)
	}

	totalThen := currentSpecies + recentExtinctions
	if totalThen < 10 {
		return // Too few species to call it a mass extinction
	}

	extinctionRate := float64(recentExtinctions) / float64(totalThen)

	// Mass extinction threshold: >50% loss
	if extinctionRate > 0.5 {
		ps.RecoveryPhase = true
		ps.RecoveryCounter = 20000 // 20k years of recovery/adaptive radiation
	}
}

// CheckSpeciation checks if any species should split based on trait divergence
// Returns the number of new species created
func (ps *PopulationSimulator) CheckSpeciation() int {
	newSpeciesCount := 0

	// Adaptive radiation bonus
	adaptiveRadiationBonus := 0.0

	// Bonus from recovery phase (post-mass extinction)
	if ps.RecoveryPhase {
		adaptiveRadiationBonus = 0.4 // High speciation during recovery
	} else if ps.FossilRecord != nil && len(ps.FossilRecord.Extinct) > 0 {
		// Minor bonus from recent local extinctions
		recentExtinctions := 0
		for i := len(ps.FossilRecord.Extinct) - 1; i >= 0; i-- {
			fossil := ps.FossilRecord.Extinct[i]
			if ps.CurrentYear-fossil.ExistedUntil < 50000 {
				recentExtinctions++
			} else {
				if ps.CurrentYear-fossil.ExistedUntil > 50000 {
					break // Optimization
				}
			}
		}
		adaptiveRadiationBonus = float64(recentExtinctions) * 0.01
		if adaptiveRadiationBonus > 0.1 {
			adaptiveRadiationBonus = 0.1
		}
	}

	for _, biome := range ps.Biomes {
		var newSpecies []*SpeciesPopulation

		for _, species := range biome.Species {
			// Base speciation chance: 10%
			speciationChance := 0.1 + adaptiveRadiationBonus

			// Large populations with high variance may speciate
			if species.Count > 500 && species.TraitVariance > 0.3 && ps.rng.Float64() < speciationChance {
				// Create mutated traits for the new species
				newTraits := mutateTraits(species.Traits, 0.15, ps.rng)

				// Generate a proper name based on the new traits
				newName := GenerateSpeciesName(newTraits, species.Diet, biome.BiomeType)
				// Prefix with biome type for clarity
				newName = string(biome.BiomeType) + " " + newName

				// Split into two species
				child := &SpeciesPopulation{
					SpeciesID:     uuid.New(),
					Name:          newName,
					AncestorID:    &species.SpeciesID,
					Count:         species.Count / 3, // 1/3 goes to new species
					Traits:        newTraits,
					TraitVariance: species.TraitVariance * 0.8,
					Diet:          species.Diet,
					Generation:    species.Generation + 1,
					CreatedYear:   ps.CurrentYear,
				}

				species.Count -= child.Count
				species.TraitVariance *= 0.8 // Reduce variance after split

				newSpecies = append(newSpecies, child)
				newSpeciesCount++
			}
		}

		// Add new species to biome
		for _, sp := range newSpecies {
			biome.AddSpecies(sp)
		}
	}

	return newSpeciesCount
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
