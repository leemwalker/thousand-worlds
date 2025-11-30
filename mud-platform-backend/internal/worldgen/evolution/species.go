package evolution

import (
	"math/rand"

	"github.com/google/uuid"
)

// GenerateInitialSpecies creates starting species diversity based on biomes
func GenerateInitialSpecies(biomes []string) []*Species {
	species := []*Species{}
	idCounter := 0

	for _, biome := range biomes {
		// Generate 5-10 species per biome
		numSpecies := 5 + rand.Intn(6)

		for i := 0; i < numSpecies; i++ {
			idCounter++
			s := generateSpeciesForBiome(biome, idCounter)
			species = append(species, s)
		}
	}

	return species
}

func generateSpeciesForBiome(biome string, idNum int) *Species {
	// Determine species type
	r := rand.Float64()
	var speciesType SpeciesType
	var diet DietType

	if r < 0.4 {
		speciesType = SpeciesFlora
		diet = DietPhotosynthesis
	} else if r < 0.7 {
		speciesType = SpeciesHerbivore
		diet = DietHerbivore
	} else {
		speciesType = SpeciesCarnivore
		diet = DietCarnivore
	}

	s := &Species{
		SpeciesID:         uuid.New(),
		Name:              biome + "_species_" + string(rune(idNum)),
		Type:              speciesType,
		Generation:        0,
		Size:              10 + rand.Float64()*90,
		Speed:             rand.Float64() * 20,
		Armor:             rand.Float64() * 50,
		Camouflage:        rand.Float64() * 50,
		Diet:              diet,
		CaloriesPerDay:    1000 + rand.Intn(9000),
		PreferredBiomes:   []string{biome},
		ReproductionRate:  0.1 + rand.Float64()*0.9,
		MaturityAge:       1 + rand.Intn(10),
		Lifespan:          5 + rand.Intn(95),
		Population:        1000 + rand.Intn(9000),
		PopulationDensity: rand.Float64() * 100,
		ExtinctionRisk:    rand.Float64() * 0.3,
		MutationRate:      BaseMutationRateMin,
		FitnessScore:      0.5,
	}

	// Set temperature tolerance based on biome
	s.TemperatureTolerance = getTemperatureToleranceForBiome(biome)
	s.MoistureTolerance = getMoistureToleranceForBiome(biome)
	s.ElevationTolerance = ElevationRange{Min: 0, Max: 5000, Optimal: 1000}

	s.PeakPopulation = s.Population

	return s
}

func getTemperatureToleranceForBiome(biome string) TemperatureRange {
	// Simplified biome temperature mappings
	switch biome {
	case "tropical":
		return TemperatureRange{Min: 20, Max: 35, Optimal: 27}
	case "temperate":
		return TemperatureRange{Min: 0, Max: 30, Optimal: 15}
	case "arctic":
		return TemperatureRange{Min: -40, Max: 10, Optimal: -10}
	case "desert":
		return TemperatureRange{Min: 0, Max: 50, Optimal: 25}
	default:
		return TemperatureRange{Min: -10, Max: 40, Optimal: 15}
	}
}

func getMoistureToleranceForBiome(biome string) MoistureRange {
	switch biome {
	case "tropical":
		return MoistureRange{Min: 2000, Max: 4000, Optimal: 3000}
	case "temperate":
		return MoistureRange{Min: 500, Max: 2000, Optimal: 1000}
	case "arctic":
		return MoistureRange{Min: 100, Max: 500, Optimal: 300}
	case "desert":
		return MoistureRange{Min: 0, Max: 250, Optimal: 100}
	default:
		return MoistureRange{Min: 300, Max: 1500, Optimal: 800}
	}
}
