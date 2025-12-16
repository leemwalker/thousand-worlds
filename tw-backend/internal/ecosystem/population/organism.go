// Package population provides the unified Organism type for ecosystem simulation.
// This replaces the separate Flora/Fauna distinction with continuous traits.
package population

import (
	"math"

	"github.com/google/uuid"
)

// Organism represents any living entity in the simulation
// Uses continuous traits instead of discrete Flora/Fauna categories
type Organism struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	GeneticCode    *GeneticCode    `json:"genetic_code"`
	Traits         *OrganismTraits `json:"traits"`
	AncestorID     *uuid.UUID      `json:"ancestor_id,omitempty"`
	OriginYear     int64           `json:"origin_year"`
	ExtinctionYear int64           `json:"extinction_year,omitempty"` // 0 if extant
}

// OrganismTraits represents the phenotypic traits derived from genetic code
// All values are normalized to 0.0-1.0 (or 0.0-10.0 for some)
type OrganismTraits struct {
	// Core organism classification (continuous values 0.0-1.0)
	Autotrophy float64 `json:"autotrophy"` // 0=heterotroph, 1=autotroph
	Complexity float64 `json:"complexity"` // 0=prokaryote, 1=complex multicellular
	Motility   float64 `json:"motility"`   // 0=sessile, 1=highly mobile

	// Physical traits (0.0-10.0 scale matching existing EvolvableTraits)
	Size     float64 `json:"size"`     // Body size
	Speed    float64 `json:"speed"`    // Movement speed
	Strength float64 `json:"strength"` // Physical power

	// Behavioral traits
	Aggression float64 `json:"aggression"` // 0=passive, 10=aggressive
	Social     float64 `json:"social"`     // 0=solitary, 10=highly social

	// Cognitive traits
	Intelligence  float64 `json:"intelligence"`  // General cognitive ability
	ToolUse       float64 `json:"tool_use"`      // Tool manipulation ability
	Communication float64 `json:"communication"` // Communication complexity

	// Environmental adaptation
	ColdResist  float64 `json:"cold_resist"`  // Cold tolerance
	HeatResist  float64 `json:"heat_resist"`  // Heat tolerance
	NightVision float64 `json:"night_vision"` // Low-light vision
	Camouflage  float64 `json:"camouflage"`   // Natural camouflage

	// Reproduction
	Fertility  float64 `json:"fertility"`   // Reproductive rate
	Lifespan   float64 `json:"lifespan"`    // Expected lifespan (scaled)
	Maturity   float64 `json:"maturity"`    // Time to maturity
	LitterSize float64 `json:"litter_size"` // Offspring per reproduction

	// Diet and defense
	Carnivore     float64 `json:"carnivore"`      // 0=herbivore, 1=carnivore
	Venom         float64 `json:"venom"`          // Venomous capability
	PoisonResist  float64 `json:"poison_resist"`  // Resistance to toxins
	DiseaseResist float64 `json:"disease_resist"` // Immune system strength
	Armor         float64 `json:"armor"`          // Natural armor/shell

	// Special abilities
	Echolocation   float64 `json:"echolocation"`   // Sonar capability
	Regeneration   float64 `json:"regeneration"`   // Healing ability
	Photosynthesis float64 `json:"photosynthesis"` // Light energy conversion
	Chemosynthesis float64 `json:"chemosynthesis"` // Chemical energy conversion
	Display        float64 `json:"display"`        // Sexual selection display
	MagicAffinity  float64 `json:"magic_affinity"` // Magic capability (if enabled)
}

// NewOrganismFromGeneticCode creates an organism with traits derived from genetic code
func NewOrganismFromGeneticCode(id uuid.UUID, name string, gc *GeneticCode, em *ExpressionMatrix, originYear int64, ancestorID *uuid.UUID) *Organism {
	phenotype := gc.ToPhenotype(em)

	traits := &OrganismTraits{
		// Map phenotype array to trait struct
		// Core classification (0-1 scale)
		Autotrophy: float64(phenotype[TraitAutotrophy]),
		Complexity: float64(phenotype[TraitComplexity]),
		Motility:   float64(phenotype[TraitMotility]),

		// Physical (scale to 0-10)
		Size:     float64(phenotype[TraitSize]) * 10,
		Speed:    float64(phenotype[TraitSpeed]) * 10,
		Strength: float64(phenotype[TraitStrength]) * 10,

		// Behavioral (scale to 0-10)
		Aggression: float64(phenotype[TraitAggression]) * 10,
		Social:     float64(phenotype[TraitSocial]) * 10,

		// Cognitive (scale to 0-10)
		Intelligence:  float64(phenotype[TraitIntelligence]) * 10,
		ToolUse:       float64(phenotype[TraitToolUse]) * 10,
		Communication: float64(phenotype[TraitCommunication]) * 10,

		// Environmental (scale to 0-10)
		ColdResist:  float64(phenotype[TraitColdResist]) * 10,
		HeatResist:  float64(phenotype[TraitHeatResist]) * 10,
		NightVision: float64(phenotype[TraitNightVision]) * 10,
		Camouflage:  float64(phenotype[TraitCamouflage]) * 10,

		// Reproduction (scale to 0-10)
		Fertility:  float64(phenotype[TraitFertility]) * 10,
		Lifespan:   float64(phenotype[TraitLifespan]) * 10,
		Maturity:   float64(phenotype[TraitMaturity]) * 10,
		LitterSize: float64(phenotype[TraitLitterSize]) * 10,

		// Diet and defense (scale to 0-10)
		Carnivore:     float64(phenotype[TraitCarnivore]) * 10,
		Venom:         float64(phenotype[TraitVenom]) * 10,
		PoisonResist:  float64(phenotype[TraitPoisonResist]) * 10,
		DiseaseResist: float64(phenotype[TraitDiseaseResist]) * 10,
		Armor:         float64(phenotype[TraitArmor]) * 10,

		// Special abilities (scale to 0-10)
		Echolocation:   float64(phenotype[TraitEcholocation]) * 10,
		Regeneration:   float64(phenotype[TraitRegeneration]) * 10,
		Photosynthesis: float64(phenotype[TraitPhotosynthesis]) * 10,
		Chemosynthesis: float64(phenotype[TraitChemosynthesis]) * 10,
		Display:        float64(phenotype[TraitDisplay]) * 10,
		MagicAffinity:  float64(phenotype[TraitMagicAffinity]) * 10,
	}

	return &Organism{
		ID:          id,
		Name:        name,
		GeneticCode: gc,
		Traits:      traits,
		AncestorID:  ancestorID,
		OriginYear:  originYear,
	}
}

// IsAutotroph returns true if the organism is primarily autotrophic (plant-like)
func (o *Organism) IsAutotroph() bool {
	return o.Traits.Autotrophy > 0.5
}

// IsHeterotroph returns true if the organism is primarily heterotrophic (animal-like)
func (o *Organism) IsHeterotroph() bool {
	return o.Traits.Autotrophy <= 0.5
}

// IsSessile returns true if the organism is sessile (non-mobile)
func (o *Organism) IsSessile() bool {
	return o.Traits.Motility < 0.2
}

// IsMobile returns true if the organism is mobile
func (o *Organism) IsMobile() bool {
	return o.Traits.Motility >= 0.2
}

// IsCarnivore returns true if the organism is primarily carnivorous
func (o *Organism) IsCarnivore() bool {
	return o.Traits.Carnivore > 7.0
}

// IsHerbivore returns true if the organism is primarily herbivorous
func (o *Organism) IsHerbivore() bool {
	return o.Traits.Carnivore < 3.0 && o.IsHeterotroph()
}

// IsOmnivore returns true if the organism is omnivorous
func (o *Organism) IsOmnivore() bool {
	return o.Traits.Carnivore >= 3.0 && o.Traits.Carnivore <= 7.0 && o.IsHeterotroph()
}

// IsComplex returns true if the organism is complex multicellular
func (o *Organism) IsComplex() bool {
	return o.Traits.Complexity > 0.5
}

// IsProtoSapient returns true if the organism meets proto-sapience thresholds
func (o *Organism) IsProtoSapient() bool {
	return o.Traits.Intelligence > 7.0 &&
		o.Traits.Social > 6.0 &&
		o.Traits.ToolUse > 3.0 &&
		o.Traits.Communication > 3.0
}

// IsProtoSapientWithMagic returns true if meets magic-assisted sapience threshold
func (o *Organism) IsProtoSapientWithMagic() bool {
	return o.Traits.MagicAffinity > 5.0 &&
		o.Traits.Intelligence > 4.0 &&
		o.Traits.Social > 4.0
}

// CalculateMetabolicRate returns the metabolic rate using Kleiber's Law (M^0.75)
func (o *Organism) CalculateMetabolicRate() float64 {
	baseRate := math.Pow(o.Traits.Size, 0.75)

	// Add costs for enhanced traits (energy budget system)
	speedCost := (o.Traits.Speed - 5.0) * 0.08
	strengthCost := (o.Traits.Strength - 5.0) * 0.10
	armorCost := o.Traits.Armor * 0.05
	intelligenceCost := o.Traits.Intelligence * 0.15
	magicCost := o.Traits.MagicAffinity * 0.20

	totalCost := 1.0 + speedCost + strengthCost + armorCost + intelligenceCost + magicCost
	if totalCost < 0.1 {
		totalCost = 0.1 // Minimum metabolic multiplier
	}

	return baseRate * totalCost
}

// CalculateReproductionRate returns reproduction rate using M^-0.25 scaling
// This is the corrected formula (not M^-0.5 which penalizes megafauna too much)
func (o *Organism) CalculateReproductionRate() float64 {
	if o.Traits.Size <= 0 {
		return 1.0
	}
	// Quarter-power scaling for generation time
	return math.Pow(o.Traits.Size, -0.25) * (o.Traits.Fertility / 10.0)
}

// CalculateInbreedingPenalty returns fitness penalty for small populations
func CalculateInbreedingPenalty(populationSize int64) float64 {
	if populationSize >= 50 {
		return 1.0 // No penalty
	}
	if populationSize < 2 {
		return 0.1 // Near extinction
	}
	// Linear penalty from 50 down to 2
	return 0.1 + 0.9*(float64(populationSize-2)/48.0)
}

// Clone creates a deep copy of the organism
func (o *Organism) Clone() *Organism {
	clone := &Organism{
		ID:             o.ID,
		Name:           o.Name,
		GeneticCode:    o.GeneticCode.Clone(),
		OriginYear:     o.OriginYear,
		ExtinctionYear: o.ExtinctionYear,
	}

	if o.AncestorID != nil {
		ancestorCopy := *o.AncestorID
		clone.AncestorID = &ancestorCopy
	}

	// Clone traits
	if o.Traits != nil {
		traitsCopy := *o.Traits
		clone.Traits = &traitsCopy
	}

	return clone
}

// TropicalAffinity returns how well-adapted the organism is to tropical climates
func (o *Organism) TropicalAffinity() float64 {
	return o.Traits.HeatResist - o.Traits.ColdResist*0.5
}

// ArcticAffinity returns how well-adapted the organism is to arctic climates
func (o *Organism) ArcticAffinity() float64 {
	return o.Traits.ColdResist - o.Traits.HeatResist*0.5
}

// TemperateAffinity returns how well-adapted the organism is to temperate climates
func (o *Organism) TemperateAffinity() float64 {
	// Prefers moderate heat and cold resistance
	return (o.Traits.HeatResist + o.Traits.ColdResist) / 2
}

// AquaticAffinity returns how well-adapted the organism is to aquatic environments
// (inferred from traits - organisms with low heat resist and high speed often aquatic)
func (o *Organism) AquaticAffinity() float64 {
	// This is a simplification - in full implementation would have separate aquatic trait
	return math.Max(0, o.Traits.Speed-o.Traits.HeatResist*0.5)
}

// --- Conversion functions for gradual migration from EvolvableTraits ---

// ToEvolvableTraits converts OrganismTraits to the legacy EvolvableTraits format
// This allows new Organism types to work with existing code that uses EvolvableTraits
func (t *OrganismTraits) ToEvolvableTraits() EvolvableTraits {
	// Determine covering type from traits
	covering := CoveringSkin // Default
	if t.ColdResist > 6.0 {
		covering = CoveringFur
	} else if t.Armor > 5.0 {
		covering = CoveringScales
	}

	// Determine flora growth type for autotrophs
	floraGrowth := FloraPerennial
	if t.Autotrophy > 0.5 {
		if t.ColdResist < 3.0 {
			floraGrowth = FloraEvergreen
		} else {
			floraGrowth = FloraDeciduous
		}
	}

	return EvolvableTraits{
		// Physical traits (direct mapping)
		Size:     t.Size,
		Speed:    t.Speed,
		Strength: t.Strength,

		// Behavioral traits (scale from 0-10 to 0-1)
		Aggression:   t.Aggression / 10.0,
		Social:       t.Social / 10.0,
		Intelligence: t.Intelligence / 10.0,

		// Survival traits (scale from 0-10 to 0-1)
		ColdResistance: t.ColdResist / 10.0,
		HeatResistance: t.HeatResist / 10.0,
		NightVision:    t.NightVision / 10.0,
		Camouflage:     t.Camouflage / 10.0,

		// Reproduction traits
		Fertility:  t.Fertility / 5.0,  // Scale 0-10 to 0-2
		Lifespan:   t.Lifespan * 5.0,   // Scale 0-10 to years
		Maturity:   t.Maturity * 2.0,   // Scale 0-10 to years
		LitterSize: t.LitterSize * 2.0, // Scale 0-10 to count

		// Dietary traits
		CarnivoreTendency: t.Carnivore / 10.0,
		VenomPotency:      t.Venom / 10.0,
		PoisonResistance:  t.PoisonResist / 10.0,
		DiseaseResistance: t.DiseaseResist / 10.0,

		// Appearance
		Covering:    covering,
		FloraGrowth: floraGrowth,
		Display:     t.Display / 10.0,
	}
}

// FromEvolvableTraits creates OrganismTraits from the legacy EvolvableTraits format
// This allows existing species data to be converted to the new format
func FromEvolvableTraits(et EvolvableTraits) *OrganismTraits {
	// Determine autotrophy from diet-related traits and covering
	autotrophy := 0.0
	if et.CarnivoreTendency == 0 && et.Covering == CoveringBark {
		autotrophy = 1.0
	} else if et.FloraGrowth != "" {
		autotrophy = 0.8
	}

	// Determine motility (flora is sessile)
	motility := 0.8 // Default for fauna
	if autotrophy > 0.5 {
		motility = 0.0 // Flora is sessile
	} else if et.Speed > 0 {
		motility = math.Min(1.0, et.Speed/10.0)
	}

	return &OrganismTraits{
		// Core classification
		Autotrophy: autotrophy,
		Complexity: 0.7, // Default to complex (existing species are multicellular)
		Motility:   motility,

		// Physical traits (direct mapping)
		Size:     et.Size,
		Speed:    et.Speed,
		Strength: et.Strength,

		// Behavioral traits (scale from 0-1 to 0-10)
		Aggression: et.Aggression * 10.0,
		Social:     et.Social * 10.0,

		// Cognitive traits
		Intelligence:  et.Intelligence * 10.0,
		ToolUse:       0.0,             // Not in old format
		Communication: et.Social * 5.0, // Infer from social

		// Environmental adaptation (scale from 0-1 to 0-10)
		ColdResist:  et.ColdResistance * 10.0,
		HeatResist:  et.HeatResistance * 10.0,
		NightVision: et.NightVision * 10.0,
		Camouflage:  et.Camouflage * 10.0,

		// Reproduction (scale appropriately)
		Fertility:  et.Fertility * 5.0,  // Scale 0-2 to 0-10
		Lifespan:   et.Lifespan / 5.0,   // Scale years to 0-10
		Maturity:   et.Maturity / 2.0,   // Scale years to 0-10
		LitterSize: et.LitterSize / 2.0, // Scale count to 0-10

		// Diet and defense (scale from 0-1 to 0-10)
		Carnivore:     et.CarnivoreTendency * 10.0,
		Venom:         et.VenomPotency * 10.0,
		PoisonResist:  et.PoisonResistance * 10.0,
		DiseaseResist: et.DiseaseResistance * 10.0,
		Armor:         0.0, // Not in old format, infer from covering later

		// Special abilities
		Echolocation:   0.0,
		Regeneration:   0.0,
		Photosynthesis: autotrophy * 10.0,
		Chemosynthesis: 0.0,
		Display:        et.Display * 10.0,
		MagicAffinity:  0.0,
	}
}

// OrganismTraitsFromEvolvableTraitsWithDefaults creates OrganismTraits with intelligent defaults
func OrganismTraitsFromEvolvableTraitsWithDefaults(et EvolvableTraits, diet DietType) *OrganismTraits {
	traits := FromEvolvableTraits(et)

	// Set autotrophy based on diet
	if diet == DietPhotosynthetic {
		traits.Autotrophy = 1.0
		traits.Motility = 0.0
		traits.Photosynthesis = 8.0
	}

	// Infer armor from covering
	switch et.Covering {
	case CoveringShell:
		traits.Armor = 8.0
	case CoveringScales:
		traits.Armor = 4.0
	}

	return traits
}

// GetDietType returns the DietType that best matches the OrganismTraits
func (t *OrganismTraits) GetDietType() DietType {
	if t.Autotrophy > 0.5 {
		return DietPhotosynthetic
	}
	if t.Carnivore > 7.0 {
		return DietCarnivore
	}
	if t.Carnivore < 3.0 {
		return DietHerbivore
	}
	return DietOmnivore
}
