// Package population provides speciation mechanics for the ecosystem simulation.
// This implements both allopatric (geographic) and sympatric (same-place) speciation.
package population

import (
	"math"
	"math/rand"

	"github.com/google/uuid"
)

// SpeciationThreshold is the genetic distance required for speciation
const SpeciationThreshold float32 = 0.35

// SympatricSpeciationRarity is the relative probability of sympatric vs allopatric speciation
// Real sympatric speciation is rare - about 5% of speciation events
const SympatricSpeciationRarity = 0.05

// MinimumPopulationForSpeciation prevents very small populations from speciating
const MinimumPopulationForSpeciation int64 = 100

// MinimumIsolationYearsForAllopatric is the time needed for allopatric speciation
const MinimumIsolationYearsForAllopatric int64 = 50000

// SpeciationEvent represents a speciation event that occurred
type SpeciationEvent struct {
	Year            int64          `json:"year"`
	ParentSpeciesID uuid.UUID      `json:"parent_species_id"`
	NewSpeciesID    uuid.UUID      `json:"new_species_id"`
	NewSpeciesName  string         `json:"new_species_name"`
	Type            SpeciationType `json:"speciation_type"`
	GeneticDistance float32        `json:"genetic_distance"`
	Cause           string         `json:"cause"` // e.g., "mountain uplift", "island isolation", "niche divergence"
	RegionID        *uuid.UUID     `json:"region_id,omitempty"`
}

// SpeciationType indicates how the speciation occurred
type SpeciationType string

const (
	SpeciationAllopatric    SpeciationType = "allopatric"    // Geographic isolation
	SpeciationPeripatric    SpeciationType = "peripatric"    // Small peripheral population
	SpeciationParapatric    SpeciationType = "parapatric"    // Adjacent populations, partial barriers
	SpeciationSympatric     SpeciationType = "sympatric"     // Same place, niche divergence
	SpeciationPolyploidy    SpeciationType = "polyploidy"    // Chromosome doubling (mostly plants)
	SpeciationHybridization SpeciationType = "hybridization" // Two species merge
)

// SpeciationChecker determines when speciation events should occur
type SpeciationChecker struct {
	rng              *rand.Rand
	speciationEvents []SpeciationEvent
	highMutationRate bool    // Increased during recovery periods
	radiationBonus   float64 // Multiplier during adaptive radiation
}

// NewSpeciationChecker creates a new speciation checker
func NewSpeciationChecker(seed int64) *SpeciationChecker {
	return &SpeciationChecker{
		rng:              rand.New(rand.NewSource(seed)),
		speciationEvents: make([]SpeciationEvent, 0),
		radiationBonus:   1.0,
	}
}

// SetRadiationBonus sets the adaptive radiation multiplier (e.g., post-extinction)
func (sc *SpeciationChecker) SetRadiationBonus(bonus float64) {
	sc.radiationBonus = bonus
}

// SetHighMutationRate enables high mutation rate for recovery periods
func (sc *SpeciationChecker) SetHighMutationRate(enabled bool) {
	sc.highMutationRate = enabled
}

// GetEvents returns all recorded speciation events
func (sc *SpeciationChecker) GetEvents() []SpeciationEvent {
	return sc.speciationEvents
}

// ClearEvents clears the event log
func (sc *SpeciationChecker) ClearEvents() {
	sc.speciationEvents = make([]SpeciationEvent, 0)
}

// CheckAllopatricSpeciation checks if a population should speciate due to geographic isolation
// Returns the new species if speciation occurs, nil otherwise
func (sc *SpeciationChecker) CheckAllopatricSpeciation(
	species *SpeciesPopulation,
	isolationYears int64,
	regionID uuid.UUID,
	currentYear int64,
) *SpeciesPopulation {
	// Check prerequisites
	if species.Count < MinimumPopulationForSpeciation {
		return nil
	}
	if isolationYears < MinimumIsolationYearsForAllopatric {
		return nil
	}

	// Check if species has genetic code (V2 system)
	if species.GeneticCode == nil {
		return nil
	}

	// Base probability increases with isolation time
	// 50,000 years = 1% chance, 500,000 years = 10% chance
	baseProbability := float64(isolationYears) / 5000000.0
	if baseProbability > 0.15 {
		baseProbability = 0.15 // Cap at 15% per check
	}

	// Apply radiation bonus
	probability := baseProbability * sc.radiationBonus

	// Roll for speciation
	if sc.rng.Float64() > probability {
		return nil
	}

	// Speciation occurs! Create divergent offspring
	newSpecies := sc.createDaughterSpecies(species, currentYear, SpeciationAllopatric)

	// Apply extra mutations due to isolation (founder effect)
	mutationRate := float32(0.1)
	if sc.highMutationRate {
		mutationRate = 0.2
	}
	newSpecies.GeneticCode = newSpecies.GeneticCode.Mutate(sc.rng, mutationRate, 0.15)

	// Recalculate traits from mutated genetic code
	if newSpecies.OrganismTraits != nil {
		em := DefaultExpressionMatrix()
		newSpecies.OrganismTraits = OrganismTraitsFromEvolvableTraitsWithDefaults(
			newSpecies.GeneticCode.ToPhenotypeAsTraits(em),
			newSpecies.Diet,
		)
	}

	// Record event
	regID := regionID
	sc.speciationEvents = append(sc.speciationEvents, SpeciationEvent{
		Year:            currentYear,
		ParentSpeciesID: species.SpeciesID,
		NewSpeciesID:    newSpecies.SpeciesID,
		NewSpeciesName:  newSpecies.Name,
		Type:            SpeciationAllopatric,
		GeneticDistance: CalculateGeneticDistance(species.GeneticCode, newSpecies.GeneticCode),
		Cause:           "geographic isolation",
		RegionID:        &regID,
	})

	return newSpecies
}

// CheckSympatricSpeciation checks if a population should speciate due to niche divergence
// This is rare and requires specific conditions (strong disruptive selection)
func (sc *SpeciationChecker) CheckSympatricSpeciation(
	species *SpeciesPopulation,
	competitionPressure float64, // 0.0-1.0, how much competition for resources
	nicheDiversity float64, // 0.0-1.0, how many different niches available
	currentYear int64,
) *SpeciesPopulation {
	// Check prerequisites
	if species.Count < MinimumPopulationForSpeciation*2 {
		return nil // Need larger population for sympatric
	}
	if species.GeneticCode == nil {
		return nil
	}
	if species.TraitVariance < 0.3 {
		return nil // Need high genetic diversity
	}

	// Sympatric speciation requires strong disruptive selection
	// High competition + high niche diversity = opportunity
	selectionStrength := competitionPressure * nicheDiversity

	// Very low base probability (sympatric is rare)
	baseProbability := SympatricSpeciationRarity * selectionStrength * float64(species.TraitVariance)

	// Apply radiation bonus
	probability := baseProbability * sc.radiationBonus

	// Roll for speciation
	if sc.rng.Float64() > probability {
		return nil
	}

	// Sympatric speciation - niche divergence
	newSpecies := sc.createDaughterSpecies(species, currentYear, SpeciationSympatric)

	// Key trait divergence for niche separation
	// Modify dietary or habitat-related genes
	divergeNicheTraits(newSpecies.GeneticCode, sc.rng)

	// Record event
	sc.speciationEvents = append(sc.speciationEvents, SpeciationEvent{
		Year:            currentYear,
		ParentSpeciesID: species.SpeciesID,
		NewSpeciesID:    newSpecies.SpeciesID,
		NewSpeciesName:  newSpecies.Name,
		Type:            SpeciationSympatric,
		GeneticDistance: CalculateGeneticDistance(species.GeneticCode, newSpecies.GeneticCode),
		Cause:           "niche divergence",
	})

	return newSpecies
}

// CheckPeripatricSpeciation checks for speciation in small peripheral populations
// This is a form of allopatric speciation with stronger founder effects
func (sc *SpeciationChecker) CheckPeripatricSpeciation(
	species *SpeciesPopulation,
	peripheralPopulation int64,
	isolationYears int64,
	currentYear int64,
) *SpeciesPopulation {
	// Peripatric speciation works on small populations (50-500 individuals)
	if peripheralPopulation < 50 || peripheralPopulation > 500 {
		return nil
	}
	if isolationYears < 10000 {
		return nil // Need some isolation time
	}
	if species.GeneticCode == nil {
		return nil
	}

	// Higher probability than allopatric due to stronger genetic drift
	baseProbability := float64(isolationYears) / 1000000.0 * (500.0 / float64(peripheralPopulation))
	if baseProbability > 0.2 {
		baseProbability = 0.2
	}

	probability := baseProbability * sc.radiationBonus

	if sc.rng.Float64() > probability {
		return nil
	}

	// Create daughter species with strong founder effect
	newSpecies := sc.createDaughterSpecies(species, currentYear, SpeciationPeripatric)

	// Strong mutations due to small population genetic drift
	mutationRate := float32(0.15)
	newSpecies.GeneticCode = newSpecies.GeneticCode.Mutate(sc.rng, mutationRate, 0.2)

	// Apply inbreeding effects
	penalty := CalculateInbreedingPenalty(peripheralPopulation)
	newSpecies.TraitVariance *= penalty

	sc.speciationEvents = append(sc.speciationEvents, SpeciationEvent{
		Year:            currentYear,
		ParentSpeciesID: species.SpeciesID,
		NewSpeciesID:    newSpecies.SpeciesID,
		NewSpeciesName:  newSpecies.Name,
		Type:            SpeciationPeripatric,
		GeneticDistance: CalculateGeneticDistance(species.GeneticCode, newSpecies.GeneticCode),
		Cause:           "founder effect",
	})

	return newSpecies
}

// createDaughterSpecies creates a new species derived from a parent
func (sc *SpeciationChecker) createDaughterSpecies(
	parent *SpeciesPopulation,
	currentYear int64,
	speciationType SpeciationType,
) *SpeciesPopulation {
	newID := uuid.New()

	daughter := &SpeciesPopulation{
		SpeciesID:     newID,
		Name:          generateDaughterName(parent.Name, speciationType, sc.rng),
		AncestorID:    &parent.SpeciesID,
		Count:         parent.Count / 4, // Start with fraction of parent population
		JuvenileCount: parent.JuvenileCount / 4,
		AdultCount:    parent.AdultCount / 4,
		Traits:        parent.Traits,              // Copy traits
		TraitVariance: parent.TraitVariance * 0.8, // Slightly reduced variance (founder effect)
		Diet:          parent.Diet,
		Generation:    parent.Generation + 1,
		CreatedYear:   currentYear,
	}

	// Copy genetic code if present
	if parent.GeneticCode != nil {
		daughter.GeneticCode = parent.GeneticCode.Clone()
	}

	// Copy organism traits if present
	if parent.OrganismTraits != nil {
		traits := *parent.OrganismTraits
		daughter.OrganismTraits = &traits
	}

	return daughter
}

// divergeNicheTraits modifies genes related to niche separation
func divergeNicheTraits(gc *GeneticCode, rng *rand.Rand) {
	// Key traits for niche divergence - we'll modify genes that affect these traits
	// Each trait is influenced by ~3 genes based on expression matrix layout
	nicheTraits := []int{
		TraitSize,        // Size differentiation
		TraitSpeed,       // Behavioral niche
		TraitCarnivore,   // Diet shift
		TraitNightVision, // Temporal niche
		TraitColdResist,  // Habitat preference
		TraitHeatResist,  // Habitat preference
	}

	// Calculate which gene range affects each trait
	genesPerTrait := DefinedGeneCount / PhenotypeCount

	// Strongly shift one or two key traits
	numShifts := 1 + rng.Intn(2)
	for i := 0; i < numShifts; i++ {
		traitIdx := nicheTraits[rng.Intn(len(nicheTraits))]

		// Find genes that affect this trait
		startGene := traitIdx * genesPerTrait
		endGene := (traitIdx + 1) * genesPerTrait
		if endGene > DefinedGeneCount {
			endGene = DefinedGeneCount
		}

		// Shift these genes toward extremes
		for g := startGene; g < endGene; g++ {
			current := gc.DefinedGenes[g]
			if current > 0.5 {
				gc.DefinedGenes[g] = float32(math.Min(1.0, float64(current)+0.15+rng.Float64()*0.15))
			} else {
				gc.DefinedGenes[g] = float32(math.Max(0.0, float64(current)-0.15-rng.Float64()*0.15))
			}
		}
	}
}

// generateDaughterName creates a name for a newly speciated species
func generateDaughterName(parentName string, speciationType SpeciationType, rng *rand.Rand) string {
	prefixes := []string{
		"Lesser", "Greater", "Island", "Mountain", "Coastal",
		"Northern", "Southern", "Dwarf", "Giant", "Spotted",
		"Striped", "Crested", "Long-tailed", "Short-horned",
	}

	// Pick a prefix that suggests the speciation type
	var prefix string
	switch speciationType {
	case SpeciationAllopatric, SpeciationPeripatric:
		geoPrefixes := []string{"Island", "Mountain", "Coastal", "Northern", "Southern"}
		prefix = geoPrefixes[rng.Intn(len(geoPrefixes))]
	case SpeciationSympatric:
		nichePrefixes := []string{"Lesser", "Greater", "Dwarf", "Giant", "Spotted", "Striped"}
		prefix = nichePrefixes[rng.Intn(len(nichePrefixes))]
	default:
		prefix = prefixes[rng.Intn(len(prefixes))]
	}

	return prefix + " " + parentName
}

// ToPhenotypeAsTraits is a helper that converts genetic code to EvolvableTraits
func (gc *GeneticCode) ToPhenotypeAsTraits(em *ExpressionMatrix) EvolvableTraits {
	phenotype := gc.ToPhenotype(em)

	return EvolvableTraits{
		Size:              float64(phenotype[TraitSize]) * 10,
		Speed:             float64(phenotype[TraitSpeed]) * 10,
		Strength:          float64(phenotype[TraitStrength]) * 10,
		Aggression:        float64(phenotype[TraitAggression]),
		Social:            float64(phenotype[TraitSocial]),
		Intelligence:      float64(phenotype[TraitIntelligence]),
		ColdResistance:    float64(phenotype[TraitColdResist]),
		HeatResistance:    float64(phenotype[TraitHeatResist]),
		NightVision:       float64(phenotype[TraitNightVision]),
		Camouflage:        float64(phenotype[TraitCamouflage]),
		Fertility:         0.5 + float64(phenotype[TraitFertility]),
		Lifespan:          float64(phenotype[TraitLifespan]) * 50,
		Maturity:          0.5 + float64(phenotype[TraitMaturity])*10,
		LitterSize:        1 + float64(phenotype[TraitLitterSize])*10,
		CarnivoreTendency: float64(phenotype[TraitCarnivore]),
		VenomPotency:      float64(phenotype[TraitVenom]),
		PoisonResistance:  float64(phenotype[TraitPoisonResist]),
		DiseaseResistance: float64(phenotype[TraitDiseaseResist]),
		Display:           float64(phenotype[TraitDisplay]),
	}
}

// HasSufficientDivergence returns true if two populations have diverged enough for speciation
func HasSufficientDivergence(gc1, gc2 *GeneticCode) bool {
	if gc1 == nil || gc2 == nil {
		return false
	}
	return CalculateGeneticDistance(gc1, gc2) >= SpeciationThreshold
}
