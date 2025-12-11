package ecosystem

import (
	"math/rand"
	"time"
	"tw-backend/internal/ecosystem/state"
	"tw-backend/internal/npc/genetics"
	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

// Spawner handles entity generation
type Spawner struct {
	Seed int64
}

func NewSpawner(seed int64) *Spawner {
	return &Spawner{Seed: seed}
}

// SpawnEntitiesForBiome generates a list of entities appropriate for the given biome
func (s *Spawner) SpawnEntitiesForBiome(biome geography.BiomeType, count int) []*state.LivingEntityState {
	candidates := getSpeciesForBiome(biome)
	if len(candidates) == 0 {
		return nil
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano())) // or use s.Seed
	entities := make([]*state.LivingEntityState, 0, count)

	for i := 0; i < count; i++ {
		species := candidates[rng.Intn(len(candidates))]
		entities = append(entities, s.CreateEntity(species, 1))
	}

	return entities
}

// CreateEntity initializes a new living entity with default stats for its species
func (s *Spawner) CreateEntity(species state.Species, generation int) *state.LivingEntityState {
	// Basic default
	return &state.LivingEntityState{
		EntityID:   uuid.New(),
		Species:    species,
		Diet:       getDietForSpecies(species),
		Age:        0,
		Generation: generation,
		Needs: state.NeedState{
			Hunger:           0,
			Thirst:           0,
			Energy:           100,
			ReproductionUrge: 0,
			Safety:           100,
		},
		DNA: genetics.NewDNA(), // To be populated with traits later
	}
}

func getSpeciesForBiome(b geography.BiomeType) []state.Species {
	switch b {
	case geography.BiomeDesert:
		return []state.Species{state.SpeciesCactus, state.SpeciesLizard, state.SpeciesScorpion, state.SpeciesVulture}
	case geography.BiomeRainforest, geography.BiomeDeciduousForest:
		return []state.Species{state.SpeciesFern, state.SpeciesOak, state.SpeciesWolf, state.SpeciesDeer, state.SpeciesBear}
	case geography.BiomeGrassland:
		return []state.Species{state.SpeciesGrass, state.SpeciesRabbit, state.SpeciesHawk, state.SpeciesBison}
	case geography.BiomeOcean:
		return []state.Species{state.SpeciesKelp} // Add Fish later
	default:
		return []state.Species{state.SpeciesGrass, state.SpeciesRabbit} // Generic fallback
	}
}

func getDietForSpecies(s state.Species) state.DietType {
	switch s {
	case state.SpeciesLizard, state.SpeciesHawk, state.SpeciesWolf, state.SpeciesScorpion:
		return state.DietCarnivore
	case state.SpeciesBear, state.SpeciesVulture:
		return state.DietOmnivore
	case state.SpeciesCactus, state.SpeciesFern, state.SpeciesOak, state.SpeciesGrass, state.SpeciesKelp:
		return state.DietPhotosynthetic
	// Precambrian
	case state.SpeciesCyanobacteria, state.SpeciesStromatolite:
		return state.DietPhotosynthetic
	case state.SpeciesEdiacaran, state.SpeciesDickinsonia, state.SpeciesCharnia:
		return state.DietHerbivore // Filter feeders/detritivores
	default:
		return state.DietHerbivore
	}
}
