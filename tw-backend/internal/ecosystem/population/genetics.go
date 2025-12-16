// Package population provides genetic code functionality for organisms.
// The genetic system uses float32 for memory efficiency and implements
// a genotype-to-phenotype mapping via an expression matrix.
package population

import (
	"math"
	"math/rand"
)

// GeneCategory represents the importance category of a gene for speciation distance
type GeneCategory int

const (
	// GeneBodyPlan genes (0-5) affect fundamental body structure - 10x weight
	GeneBodyPlan GeneCategory = iota
	// GeneMorphology genes (6-20) affect major physical features - 5x weight
	GeneMorphology
	// GeneBehavior genes (21-50) affect behavior and metabolism - 2x weight
	GeneBehavior
	// GeneMinor genes (51-99) affect minor traits - 1x weight
	GeneMinor
)

const (
	// DefinedGeneCount is the number of genes with defined phenotypic mappings
	DefinedGeneCount = 100
	// BlankGeneCount is the number of unlockable exotic/fantastical genes
	BlankGeneCount = 100
	// TotalGeneCount is the total genetic code length
	TotalGeneCount = DefinedGeneCount + BlankGeneCount
	// PhenotypeCount is the number of derived phenotypic traits
	PhenotypeCount = 30
)

// Gene weight multipliers for speciation distance calculation
var geneWeights = []float32{
	// Body plan genes (0-5): 10x weight
	10, 10, 10, 10, 10, 10,
	// Major morphology genes (6-20): 5x weight
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	// Behavior/metabolism genes (21-50): 2x weight
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	// Minor trait genes (51-99): 1x weight
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
}

// GeneticCode represents an organism's complete genome
// Uses float32 for 50% memory savings over float64
type GeneticCode struct {
	// DefinedGenes map to phenotypic traits via expression matrix
	DefinedGenes [DefinedGeneCount]float32
	// BlankGenes are unlockable exotic/fantastical traits
	BlankGenes [BlankGeneCount]float32
	// ActiveBlanks tracks which blank genes have been activated
	ActiveBlanks []int
}

// ExpressionMatrix maps genotype to phenotype
// P = G × E where P is phenotype vector, G is genotype, E is expression matrix
type ExpressionMatrix struct {
	// Weights maps each gene to each phenotypic trait
	// [gene_index][trait_index] = contribution weight
	Weights [DefinedGeneCount][PhenotypeCount]float32
	// Thresholds for non-linear expression (punctuated equilibrium)
	Thresholds [PhenotypeCount]float32
}

// PhenotypeIndex constants for accessing trait values
const (
	TraitSize           = 0
	TraitSpeed          = 1
	TraitStrength       = 2
	TraitAggression     = 3
	TraitSocial         = 4
	TraitIntelligence   = 5
	TraitColdResist     = 6
	TraitHeatResist     = 7
	TraitNightVision    = 8
	TraitCamouflage     = 9
	TraitFertility      = 10
	TraitLifespan       = 11
	TraitMaturity       = 12
	TraitLitterSize     = 13
	TraitCarnivore      = 14
	TraitVenom          = 15
	TraitPoisonResist   = 16
	TraitDiseaseResist  = 17
	TraitAutotrophy     = 18 // 0=heterotroph, 1=autotroph
	TraitComplexity     = 19 // 0=simple, 1=complex multicellular
	TraitMotility       = 20 // 0=sessile, 1=highly mobile
	TraitToolUse        = 21
	TraitCommunication  = 22
	TraitMagicAffinity  = 23 // Requires magic-enabled world
	TraitArmor          = 24
	TraitDisplay        = 25 // Sexual selection display
	TraitEcholocation   = 26
	TraitRegeneration   = 27
	TraitPhotosynthesis = 28 // Light energy conversion
	TraitChemosynthesis = 29 // Chemical energy conversion
)

// NewGeneticCode creates a new random genetic code
func NewGeneticCode(rng *rand.Rand) *GeneticCode {
	gc := &GeneticCode{
		ActiveBlanks: make([]int, 0),
	}

	// Initialize defined genes with random values
	for i := 0; i < DefinedGeneCount; i++ {
		gc.DefinedGenes[i] = rng.Float32()
	}

	// Blank genes start at 0 (inactive)
	for i := 0; i < BlankGeneCount; i++ {
		gc.BlankGenes[i] = 0
	}

	return gc
}

// NewGeneticCodeWithBias creates a genetic code biased toward certain traits
func NewGeneticCodeWithBias(rng *rand.Rand, traitBiases map[int]float32) *GeneticCode {
	gc := NewGeneticCode(rng)

	// Apply biases by adjusting genes that map to those traits
	// This is a simplified approach - real implementation would use expression matrix
	for traitIdx, bias := range traitBiases {
		if traitIdx < DefinedGeneCount {
			// Adjust genes in the region that typically affects this trait
			startGene := (traitIdx * DefinedGeneCount) / PhenotypeCount
			endGene := ((traitIdx + 1) * DefinedGeneCount) / PhenotypeCount
			for g := startGene; g < endGene && g < DefinedGeneCount; g++ {
				gc.DefinedGenes[g] = gc.DefinedGenes[g]*0.5 + bias*0.5
			}
		}
	}

	return gc
}

// Clone creates a deep copy of the genetic code
func (gc *GeneticCode) Clone() *GeneticCode {
	clone := &GeneticCode{
		ActiveBlanks: make([]int, len(gc.ActiveBlanks)),
	}
	copy(clone.DefinedGenes[:], gc.DefinedGenes[:])
	copy(clone.BlankGenes[:], gc.BlankGenes[:])
	copy(clone.ActiveBlanks, gc.ActiveBlanks)
	return clone
}

// Mutate creates a mutated copy of the genetic code
// mutationRate is probability of each gene mutating (0.0-1.0)
// mutationStrength is the magnitude of change (0.0-1.0)
func (gc *GeneticCode) Mutate(rng *rand.Rand, mutationRate, mutationStrength float32) *GeneticCode {
	mutated := gc.Clone()

	// Mutate defined genes
	for i := 0; i < DefinedGeneCount; i++ {
		if rng.Float32() < mutationRate {
			delta := (rng.Float32() - 0.5) * 2.0 * mutationStrength
			mutated.DefinedGenes[i] = clamp32(mutated.DefinedGenes[i]+delta, 0, 1)
		}
	}

	// Mutate active blank genes
	for _, idx := range mutated.ActiveBlanks {
		if rng.Float32() < mutationRate {
			delta := (rng.Float32() - 0.5) * 2.0 * mutationStrength
			mutated.BlankGenes[idx] = clamp32(mutated.BlankGenes[idx]+delta, 0, 1)
		}
	}

	return mutated
}

// ActivateBlankGene activates a blank gene slot
// Returns true if successful, false if already active or invalid index
func (gc *GeneticCode) ActivateBlankGene(index int, initialValue float32) bool {
	if index < 0 || index >= BlankGeneCount {
		return false
	}

	// Check if already active
	for _, active := range gc.ActiveBlanks {
		if active == index {
			return false
		}
	}

	gc.ActiveBlanks = append(gc.ActiveBlanks, index)
	gc.BlankGenes[index] = clamp32(initialValue, 0, 1)
	return true
}

// IsBlankActive returns true if the blank gene at index is active
func (gc *GeneticCode) IsBlankActive(index int) bool {
	for _, active := range gc.ActiveBlanks {
		if active == index {
			return true
		}
	}
	return false
}

// CalculateGeneticDistance computes weighted genetic distance between two codes
// Body plan genes count 10x, morphology 5x, behavior 2x, minor 1x
func CalculateGeneticDistance(g1, g2 *GeneticCode) float32 {
	var weightedSumSq float32
	var totalWeight float32

	for i := 0; i < DefinedGeneCount; i++ {
		weight := geneWeights[i]
		diff := g1.DefinedGenes[i] - g2.DefinedGenes[i]
		weightedSumSq += weight * diff * diff
		totalWeight += weight
	}

	// Normalize by total weight
	if totalWeight == 0 {
		return 0
	}

	return float32(math.Sqrt(float64(weightedSumSq / totalWeight)))
}

// Crossover creates offspring genetic code from two parents
func Crossover(parent1, parent2 *GeneticCode, rng *rand.Rand) *GeneticCode {
	offspring := &GeneticCode{
		ActiveBlanks: make([]int, 0),
	}

	// Single-point crossover for defined genes
	crossoverPoint := rng.Intn(DefinedGeneCount)
	for i := 0; i < DefinedGeneCount; i++ {
		if i < crossoverPoint {
			offspring.DefinedGenes[i] = parent1.DefinedGenes[i]
		} else {
			offspring.DefinedGenes[i] = parent2.DefinedGenes[i]
		}
	}

	// Inherit active blank genes from both parents
	seenBlanks := make(map[int]bool)
	for _, idx := range parent1.ActiveBlanks {
		if !seenBlanks[idx] {
			offspring.ActiveBlanks = append(offspring.ActiveBlanks, idx)
			// Average the values from both parents if both have it active
			offspring.BlankGenes[idx] = parent1.BlankGenes[idx]
			seenBlanks[idx] = true
		}
	}
	for _, idx := range parent2.ActiveBlanks {
		if !seenBlanks[idx] {
			offspring.ActiveBlanks = append(offspring.ActiveBlanks, idx)
			offspring.BlankGenes[idx] = parent2.BlankGenes[idx]
			seenBlanks[idx] = true
		} else {
			// Average with parent1's value
			offspring.BlankGenes[idx] = (offspring.BlankGenes[idx] + parent2.BlankGenes[idx]) / 2
		}
	}

	return offspring
}

// DefaultExpressionMatrix creates a default genotype-to-phenotype mapping
func DefaultExpressionMatrix() *ExpressionMatrix {
	em := &ExpressionMatrix{}

	// Initialize with identity-like mapping for simplicity
	// In reality, this would be more complex with polygenic effects
	genesPerTrait := DefinedGeneCount / PhenotypeCount

	for t := 0; t < PhenotypeCount; t++ {
		// Each trait is influenced by a range of genes
		startGene := t * genesPerTrait
		endGene := (t + 1) * genesPerTrait
		if endGene > DefinedGeneCount {
			endGene = DefinedGeneCount
		}

		// Assign weights to genes that affect this trait
		for g := startGene; g < endGene; g++ {
			em.Weights[g][t] = 1.0 / float32(endGene-startGene)
		}

		// Add some polygenic effects (genes affect multiple traits)
		// Body plan genes affect multiple traits
		for g := 0; g < 6; g++ {
			em.Weights[g][t] += 0.05 // Small contribution from body plan genes
		}

		// Default threshold for expression (no threshold = linear)
		em.Thresholds[t] = 0.0
	}

	return em
}

// ToPhenotype converts a genetic code to phenotypic trait values
func (gc *GeneticCode) ToPhenotype(em *ExpressionMatrix) []float32 {
	traits := make([]float32, PhenotypeCount)

	// Matrix multiplication: P = G × E
	for t := 0; t < PhenotypeCount; t++ {
		var sum float32
		for g := 0; g < DefinedGeneCount; g++ {
			sum += gc.DefinedGenes[g] * em.Weights[g][t]
		}

		// Apply expression threshold for non-linear effects (punctuated equilibrium)
		if em.Thresholds[t] > 0 {
			sum = applyExpressionCurve(sum, em.Thresholds[t])
		}

		// Clamp to valid range
		traits[t] = clamp32(sum, 0, 1)
	}

	// Add contributions from active blank genes
	// Blank genes can boost specific traits when active
	for _, idx := range gc.ActiveBlanks {
		traitIdx := idx % PhenotypeCount
		traits[traitIdx] = clamp32(traits[traitIdx]+gc.BlankGenes[idx]*0.2, 0, 1)
	}

	return traits
}

// applyExpressionCurve applies a sigmoid activation for threshold expression
// Enables punctuated equilibrium: genes don't express until threshold crossed
func applyExpressionCurve(value, threshold float32) float32 {
	if value < threshold {
		return 0.0
	}
	// Sigmoid activation above threshold
	x := 10 * (value - threshold)
	return float32(1.0 / (1.0 + math.Exp(-float64(x))))
}

// clamp32 clamps a float32 value to the given range
func clamp32(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// GetGeneCategory returns the category of importance for a gene index
func GetGeneCategory(geneIndex int) GeneCategory {
	switch {
	case geneIndex < 6:
		return GeneBodyPlan
	case geneIndex < 21:
		return GeneMorphology
	case geneIndex < 51:
		return GeneBehavior
	default:
		return GeneMinor
	}
}

// GetGeneWeight returns the speciation distance weight for a gene
func GetGeneWeight(geneIndex int) float32 {
	if geneIndex < 0 || geneIndex >= len(geneWeights) {
		return 1.0
	}
	return geneWeights[geneIndex]
}

// ExoticTraitIndex constants for blank gene slots
const (
	ExoticBioluminescence  = 0
	ExoticEcholocationAdv  = 1
	ExoticRegenerationAdv  = 2
	ExoticActiveCamouflage = 3
	ExoticVenomEnhanced    = 4
	ExoticElectricSense    = 5
	ExoticMagneticSense    = 6
	ExoticInfraredVision   = 7
	ExoticUltrasonicComm   = 8
	ExoticPhotomemory      = 9
	// ... more exotic traits 10-49

	// Fantastical traits (50-99) - require magic-enabled world
	FantasticMagicAffinity    = 50
	FantasticTeleportation    = 51
	FantasticShieldGeneration = 52
	FantasticTelekinesis      = 53
	FantasticShadowy          = 54
	FantasticFireBreath       = 55
	FantasticAcidBlood        = 56
	FantasticPsychic          = 57
	FantasticShapeshift       = 58
	FantasticTimeSense        = 59
	// ... more fantastical traits 60-99
)

// IsFantasticalTrait returns true if the blank gene index is a fantastical trait
func IsFantasticalTrait(blankIndex int) bool {
	return blankIndex >= 50
}
