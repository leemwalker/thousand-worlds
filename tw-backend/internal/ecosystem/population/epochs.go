package population

import (
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// EpochType represents geological epochs for simulation starting conditions
type EpochType string

const (
	EpochHadean        EpochType = "hadean"        // 4.5-4.0 Bya - No life
	EpochArchean       EpochType = "archean"       // 4.0-2.5 Bya - Single-cell ocean
	EpochProterozoic   EpochType = "proterozoic"   // 2.5-0.5 Bya - Multicellular
	EpochCambrian      EpochType = "cambrian"      // 541-485 Mya - Marine explosion
	EpochDevonian      EpochType = "devonian"      // 419-358 Mya - Age of fish
	EpochCarboniferous EpochType = "carboniferous" // 358-298 Mya - Giant insects
	EpochTriassic      EpochType = "triassic"      // 252-201 Mya - First dinosaurs
	EpochJurassic      EpochType = "jurassic"      // 201-145 Mya - Dinosaur peak
	EpochCretaceous    EpochType = "cretaceous"    // 145-66 Mya - Diverse dinos
	EpochCenozoic      EpochType = "cenozoic"      // 66-0 Mya - Mammals dominate
)

// EvolutionGoal represents player-directed evolution targets
type EvolutionGoal string

const (
	GoalSize           EvolutionGoal = "size"
	GoalSpeed          EvolutionGoal = "speed"
	GoalStrength       EvolutionGoal = "strength"
	GoalIntelligence   EvolutionGoal = "intelligence"
	GoalColdResistance EvolutionGoal = "cold_resistance"
	GoalHeatResistance EvolutionGoal = "heat_resistance"
	GoalCamouflage     EvolutionGoal = "camouflage"
	GoalNightVision    EvolutionGoal = "night_vision"
	GoalFertility      EvolutionGoal = "fertility"
	GoalLifespan       EvolutionGoal = "lifespan"
	GoalSocial         EvolutionGoal = "social"
	GoalVenom          EvolutionGoal = "venom"
)

// InitializeFromEpoch creates species appropriate for the given epoch and biome
func InitializeFromEpoch(epoch EpochType, biome geography.BiomeType) []*SpeciesPopulation {
	var species []*SpeciesPopulation

	switch epoch {
	case EpochHadean:
		// No life in Hadean epoch
		return nil

	case EpochArchean:
		// Only primitive single-cell life in oceans
		if biome == geography.BiomeOcean {
			species = append(species, createPrimitiveLife(biome))
		}

	case EpochProterozoic:
		// Multicellular life, mostly ocean
		if biome == geography.BiomeOcean {
			species = append(species, createSimpleFlora(biome))
			species = append(species, createSimpleFauna(biome, DietHerbivore))
		}

	case EpochCambrian:
		// Marine explosion - diverse ocean life
		if biome == geography.BiomeOcean {
			species = append(species, createDiverseMarineLife(biome)...)
		}

	case EpochDevonian:
		// Age of fish, first land plants
		if biome == geography.BiomeOcean {
			species = append(species, createDiverseMarineLife(biome)...)
		} else {
			species = append(species, createSimpleFlora(biome))
		}

	case EpochCarboniferous:
		// Giant insects, swamp flora, early reptiles
		species = append(species, createCarboniferousLife(biome)...)

	case EpochTriassic, EpochJurassic, EpochCretaceous:
		// Dinosaur ages
		species = append(species, createMesozoicLife(epoch, biome)...)

	case EpochCenozoic:
		// Mammals dominate - modern ecosystem
		species = append(species, createCenozoicLife(biome)...)
	}

	return species
}

// createPrimitiveLife creates single-cell organisms
func createPrimitiveLife(biome geography.BiomeType) *SpeciesPopulation {
	traits := EvolvableTraits{
		Size: 0.01, Speed: 0.1, Strength: 0.1,
		Fertility: 3.0, Lifespan: 0.1, Maturity: 0.01, LitterSize: 1000,
		Covering: CoveringNone, FloraGrowth: FloraAquatic,
	}
	return &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Primitive Microbe",
		Count:         10000,
		Traits:        traits,
		TraitVariance: 0.5,
		Diet:          DietPhotosynthetic,
		Generation:    0,
		CreatedYear:   0,
	}
}

// createSimpleFlora creates basic plant life
func createSimpleFlora(biome geography.BiomeType) *SpeciesPopulation {
	traits := DefaultTraitsForDiet(DietPhotosynthetic)
	traits.Size = 0.5
	traits.FloraGrowth = GetFloraGrowthForBiome(biome)
	traits.Covering = GetCoveringForDiet(DietPhotosynthetic, biome)
	return &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          GenerateSpeciesName(traits, DietPhotosynthetic, biome),
		Count:         1000,
		Traits:        traits,
		TraitVariance: 0.3,
		Diet:          DietPhotosynthetic,
		Generation:    0,
		CreatedYear:   0,
	}
}

// createSimpleFauna creates basic animal life
func createSimpleFauna(biome geography.BiomeType, diet DietType) *SpeciesPopulation {
	traits := DefaultTraitsForDiet(diet)
	traits.Size = 0.3
	traits.Covering = GetCoveringForDiet(diet, biome)
	return &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          GenerateSpeciesName(traits, diet, biome),
		Count:         500,
		Traits:        traits,
		TraitVariance: 0.3,
		Diet:          diet,
		Generation:    0,
		CreatedYear:   0,
	}
}

// createDiverseMarineLife creates Cambrian-style ocean life
func createDiverseMarineLife(biome geography.BiomeType) []*SpeciesPopulation {
	var species []*SpeciesPopulation

	// Algae/kelp
	species = append(species, createSimpleFlora(biome))

	// Various marine invertebrates
	herbTraits := DefaultTraitsForDiet(DietHerbivore)
	herbTraits.Size = 0.5
	herbTraits.Covering = CoveringShell
	species = append(species, &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Armored Grazer",
		Count:         800,
		Traits:        herbTraits,
		TraitVariance: 0.3,
		Diet:          DietHerbivore,
	})

	// Early predators
	carnTraits := DefaultTraitsForDiet(DietCarnivore)
	carnTraits.Size = 1.0
	carnTraits.Covering = CoveringShell
	species = append(species, &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Small Armored Hunter",
		Count:         200,
		Traits:        carnTraits,
		TraitVariance: 0.3,
		Diet:          DietCarnivore,
	})

	return species
}

// createCarboniferousLife creates coal age life
func createCarboniferousLife(biome geography.BiomeType) []*SpeciesPopulation {
	var species []*SpeciesPopulation

	// Giant flora
	floraTraits := DefaultTraitsForDiet(DietPhotosynthetic)
	floraTraits.Size = 5.0
	floraTraits.FloraGrowth = FloraPerennial
	floraTraits.Covering = CoveringBark
	species = append(species, &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Towering Hardy Tree",
		Count:         500,
		Traits:        floraTraits,
		TraitVariance: 0.3,
		Diet:          DietPhotosynthetic,
	})

	if biome != geography.BiomeOcean {
		// Giant insects
		herbTraits := DefaultTraitsForDiet(DietHerbivore)
		herbTraits.Size = 2.0
		herbTraits.Covering = CoveringShell
		species = append(species, &SpeciesPopulation{
			SpeciesID:     uuid.New(),
			Name:          "Large Armored Grazer",
			Count:         300,
			Traits:        herbTraits,
			TraitVariance: 0.3,
			Diet:          DietHerbivore,
		})

		// Early reptiles
		carnTraits := DefaultTraitsForDiet(DietCarnivore)
		carnTraits.Size = 1.5
		carnTraits.Covering = CoveringScales
		species = append(species, &SpeciesPopulation{
			SpeciesID:     uuid.New(),
			Name:          "Swift Scaled Hunter",
			Count:         100,
			Traits:        carnTraits,
			TraitVariance: 0.3,
			Diet:          DietCarnivore,
		})
	}

	return species
}

// createMesozoicLife creates dinosaur-age life
func createMesozoicLife(epoch EpochType, biome geography.BiomeType) []*SpeciesPopulation {
	var species []*SpeciesPopulation

	// Flora
	floraTraits := DefaultTraitsForDiet(DietPhotosynthetic)
	floraTraits.Size = 4.0
	floraTraits.FloraGrowth = GetFloraGrowthForBiome(biome)
	floraTraits.Covering = CoveringBark
	species = append(species, &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          GenerateSpeciesName(floraTraits, DietPhotosynthetic, biome),
		Count:         600,
		Traits:        floraTraits,
		TraitVariance: 0.3,
		Diet:          DietPhotosynthetic,
	})

	if biome != geography.BiomeOcean {
		// Herbivorous dinosaurs
		herbTraits := DefaultTraitsForDiet(DietHerbivore)
		herbTraits.Size = 6.0
		herbTraits.Covering = CoveringScales
		herbTraits.Social = 0.8
		species = append(species, &SpeciesPopulation{
			SpeciesID:     uuid.New(),
			Name:          "Giant Herd Scaled Grazer",
			Count:         200,
			Traits:        herbTraits,
			TraitVariance: 0.3,
			Diet:          DietHerbivore,
		})

		// Predatory dinosaurs
		carnTraits := DefaultTraitsForDiet(DietCarnivore)
		carnTraits.Size = 5.0
		carnTraits.Speed = 7.0
		carnTraits.Covering = CoveringScales
		species = append(species, &SpeciesPopulation{
			SpeciesID:     uuid.New(),
			Name:          "Massive Swift Scaled Hunter",
			Count:         50,
			Traits:        carnTraits,
			TraitVariance: 0.3,
			Diet:          DietCarnivore,
		})

		// In Jurassic/Cretaceous, add early birds
		if epoch == EpochJurassic || epoch == EpochCretaceous {
			birdTraits := DefaultTraitsForDiet(DietOmnivore)
			birdTraits.Size = 0.5
			birdTraits.Speed = 8.0
			birdTraits.Covering = CoveringFeathers
			species = append(species, &SpeciesPopulation{
				SpeciesID:     uuid.New(),
				Name:          "Small Swift Feathered Forager",
				Count:         150,
				Traits:        birdTraits,
				TraitVariance: 0.3,
				Diet:          DietOmnivore,
			})
		}
	}

	return species
}

// createCenozoicLife creates modern mammal-dominated ecosystem
func createCenozoicLife(biome geography.BiomeType) []*SpeciesPopulation {
	var species []*SpeciesPopulation

	// Flora
	floraTraits := DefaultTraitsForDiet(DietPhotosynthetic)
	floraTraits.Size = 3.0
	floraTraits.FloraGrowth = GetFloraGrowthForBiome(biome)
	floraTraits.Covering = GetCoveringForDiet(DietPhotosynthetic, biome)
	species = append(species, &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          GenerateSpeciesName(floraTraits, DietPhotosynthetic, biome),
		Count:         700,
		Traits:        floraTraits,
		TraitVariance: 0.3,
		Diet:          DietPhotosynthetic,
	})

	if biome != geography.BiomeOcean {
		// Mammals - herbivores
		herbTraits := DefaultTraitsForDiet(DietHerbivore)
		herbTraits.Size = 3.0
		herbTraits.Social = 0.8
		herbTraits.Intelligence = 0.4
		herbTraits.Covering = CoveringFur
		species = append(species, &SpeciesPopulation{
			SpeciesID:     uuid.New(),
			Name:          "Large Herd Woolly Grazer",
			Count:         300,
			Traits:        herbTraits,
			TraitVariance: 0.3,
			Diet:          DietHerbivore,
		})

		// Mammals - predators
		carnTraits := DefaultTraitsForDiet(DietCarnivore)
		carnTraits.Size = 2.5
		carnTraits.Speed = 7.0
		carnTraits.Social = 0.7
		carnTraits.Intelligence = 0.6
		carnTraits.Covering = CoveringFur
		species = append(species, &SpeciesPopulation{
			SpeciesID:     uuid.New(),
			Name:          "Large Swift Pack Woolly Hunter",
			Count:         80,
			Traits:        carnTraits,
			TraitVariance: 0.3,
			Diet:          DietCarnivore,
		})
	} else {
		// Marine mammals
		herbTraits := DefaultTraitsForDiet(DietHerbivore)
		herbTraits.Size = 4.0
		herbTraits.Covering = CoveringSkin
		species = append(species, &SpeciesPopulation{
			SpeciesID:     uuid.New(),
			Name:          "Large Smooth Grazer",
			Count:         200,
			Traits:        herbTraits,
			TraitVariance: 0.3,
			Diet:          DietHerbivore,
		})
	}

	return species
}

// ApplyEvolutionGoal applies selection pressure toward a goal
func ApplyEvolutionGoal(traits EvolvableTraits, goal EvolutionGoal, strength float64) EvolvableTraits {
	switch goal {
	case GoalSize:
		traits.Size += strength
	case GoalSpeed:
		traits.Speed += strength
	case GoalStrength:
		traits.Strength += strength
	case GoalIntelligence:
		traits.Intelligence += strength * 0.1
	case GoalColdResistance:
		traits.ColdResistance += strength * 0.1
	case GoalHeatResistance:
		traits.HeatResistance += strength * 0.1
	case GoalCamouflage:
		traits.Camouflage += strength * 0.1
	case GoalNightVision:
		traits.NightVision += strength * 0.1
	case GoalFertility:
		traits.Fertility += strength * 0.1
	case GoalLifespan:
		traits.Lifespan += strength
	case GoalSocial:
		traits.Social += strength * 0.1
	case GoalVenom:
		traits.VenomPotency += strength * 0.1
	}
	return clampTraits(traits)
}

// GetEpochDescription returns a human-readable description of an epoch
func GetEpochDescription(epoch EpochType) string {
	descriptions := map[EpochType]string{
		EpochHadean:        "Hadean (4.5-4.0 Bya): No life, volcanic, toxic atmosphere",
		EpochArchean:       "Archean (4.0-2.5 Bya): Primitive single-cell life, no oxygen",
		EpochProterozoic:   "Proterozoic (2.5-0.5 Bya): Multicellular life, oxygen rising",
		EpochCambrian:      "Cambrian (541-485 Mya): Explosion of marine life",
		EpochDevonian:      "Devonian (419-358 Mya): Age of fish, first land plants",
		EpochCarboniferous: "Carboniferous (358-298 Mya): Giant insects, first reptiles",
		EpochTriassic:      "Triassic (252-201 Mya): First dinosaurs and mammals",
		EpochJurassic:      "Jurassic (201-145 Mya): Dinosaur dominance, first birds",
		EpochCretaceous:    "Cretaceous (145-66 Mya): Flowering plants, diverse dinosaurs",
		EpochCenozoic:      "Cenozoic (66-0 Mya): Mammals dominate, modern animals",
	}
	if desc, ok := descriptions[epoch]; ok {
		return desc
	}
	return "Unknown epoch"
}

// GetAllEpochs returns all available epochs in chronological order
func GetAllEpochs() []EpochType {
	return []EpochType{
		EpochHadean, EpochArchean, EpochProterozoic, EpochCambrian,
		EpochDevonian, EpochCarboniferous, EpochTriassic, EpochJurassic,
		EpochCretaceous, EpochCenozoic,
	}
}

// GetAllGoals returns all available evolution goals
func GetAllGoals() []EvolutionGoal {
	return []EvolutionGoal{
		GoalSize, GoalSpeed, GoalStrength, GoalIntelligence,
		GoalColdResistance, GoalHeatResistance, GoalCamouflage,
		GoalNightVision, GoalFertility, GoalLifespan, GoalSocial, GoalVenom,
	}
}
