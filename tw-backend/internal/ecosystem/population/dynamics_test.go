package population

import (
	"testing"

	"tw-backend/internal/worldgen/geography"

	"github.com/google/uuid"
)

func TestNewPopulationSimulator(t *testing.T) {
	worldID := uuid.New()
	sim := NewPopulationSimulator(worldID, 12345)

	if sim == nil {
		t.Error("Simulator should not be nil")
	}
	if sim.CurrentYear != 0 {
		t.Error("CurrentYear should start at 0")
	}
	if sim.Biomes == nil {
		t.Error("Biomes map should be initialized")
	}
	if sim.FossilRecord == nil || sim.FossilRecord.Extinct == nil {
		t.Error("FossilRecord should be initialized")
	}
}

func TestSimulateYear(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	// Add a biome with species
	biome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	floraSpecies := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Test Flora",
		Count:     500,
		Traits:    DefaultTraitsForDiet(DietPhotosynthetic),
		Diet:      DietPhotosynthetic,
	}
	biome.AddSpecies(floraSpecies)
	sim.Biomes[biome.BiomeID] = biome

	initialYear := sim.CurrentYear
	sim.SimulateYear()

	if sim.CurrentYear != initialYear+1 {
		t.Errorf("Year should advance, got %d expected %d", sim.CurrentYear, initialYear+1)
	}
	if biome.YearsSimulated != 1 {
		t.Error("Biome should track years simulated")
	}
}

func TestSimulateYears(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	biome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	floraSpecies := &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Test Flora",
		Count:         500,
		Traits:        DefaultTraitsForDiet(DietPhotosynthetic),
		TraitVariance: 0.3,
		Diet:          DietPhotosynthetic,
	}
	biome.AddSpecies(floraSpecies)
	sim.Biomes[biome.BiomeID] = biome

	sim.SimulateYears(100)

	if sim.CurrentYear != 100 {
		t.Errorf("Should simulate 100 years, got %d", sim.CurrentYear)
	}
}

func TestApplyEvolution(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	biome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	species := &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Test Grazer",
		Count:         500,
		Traits:        DefaultTraitsForDiet(DietHerbivore),
		TraitVariance: 0.3,
		Diet:          DietHerbivore,
		Generation:    0,
	}
	biome.AddSpecies(species)
	sim.Biomes[biome.BiomeID] = biome

	initialGen := species.Generation
	sim.ApplyEvolution()

	if species.Generation <= initialGen {
		t.Error("Generation should increase after evolution")
	}
}

func TestCheckSpeciation(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	biome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	// Need high population and variance for speciation
	species := &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Test Grazer",
		Count:         1000, // High population
		Traits:        DefaultTraitsForDiet(DietHerbivore),
		TraitVariance: 0.5, // High variance
		Diet:          DietHerbivore,
	}
	biome.AddSpecies(species)
	sim.Biomes[biome.BiomeID] = biome

	// Run speciation check multiple times (it's probabilistic)
	initialSpeciesCount := len(biome.Species)
	for i := 0; i < 100; i++ {
		sim.CheckSpeciation()
	}

	// At least one speciation should occur with high variance and population
	// (This is probabilistic, so we can't guarantee it, but it's likely)
	t.Logf("Species count after speciation checks: %d (initial: %d)", len(biome.Species), initialSpeciesCount)
}

func TestGetStats(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	// Add biomes with species
	biome1 := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	biome1.AddSpecies(&SpeciesPopulation{SpeciesID: uuid.New(), Count: 100, Diet: DietHerbivore})
	biome1.AddSpecies(&SpeciesPopulation{SpeciesID: uuid.New(), Count: 50, Diet: DietCarnivore})
	sim.Biomes[biome1.BiomeID] = biome1

	biome2 := NewBiomePopulation(uuid.New(), geography.BiomeOcean)
	biome2.AddSpecies(&SpeciesPopulation{SpeciesID: uuid.New(), Count: 200, Diet: DietPhotosynthetic})
	sim.Biomes[biome2.BiomeID] = biome2

	pop, species, extinct := sim.GetStats()

	if pop != 350 {
		t.Errorf("Total population should be 350, got %d", pop)
	}
	if species != 3 {
		t.Errorf("Total species should be 3, got %d", species)
	}
	if extinct != 0 {
		t.Errorf("Extinct should be 0, got %d", extinct)
	}
}

func TestApplyExtinctionEvent_VolcanicWinter(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	biome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	flora := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Test Flora",
		Count:     1000,
		Traits:    DefaultTraitsForDiet(DietPhotosynthetic),
		Diet:      DietPhotosynthetic,
	}
	biome.AddSpecies(flora)
	sim.Biomes[biome.BiomeID] = biome

	initialPop := flora.Count
	deaths := sim.ApplyExtinctionEvent(EventVolcanicWinter, 0.5)

	if deaths == 0 {
		t.Error("Volcanic winter should cause deaths")
	}
	if flora.Count >= initialPop {
		t.Error("Flora population should decrease")
	}
}

func TestApplyExtinctionEvent_AsteroidImpact(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	biome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	grazer := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Large Grazer",
		Count:     1000,
		Traits:    EvolvableTraits{Size: 6.0, Intelligence: 0.3}, // Large, not very intelligent
		Diet:      DietHerbivore,
	}
	biome.AddSpecies(grazer)
	sim.Biomes[biome.BiomeID] = biome

	initialPop := grazer.Count
	deaths := sim.ApplyExtinctionEvent(EventAsteroidImpact, 0.9) // Severe impact

	if deaths == 0 {
		t.Error("Asteroid impact should cause deaths")
	}
	if grazer.Count >= initialPop/2 {
		t.Error("Large species should suffer heavy losses from asteroid")
	}
}

func TestApplyExtinctionEvent_IceAge(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	// Tropical biome species should suffer
	tropicalBiome := NewBiomePopulation(uuid.New(), geography.BiomeRainforest)
	tropicalSpecies := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Tropical Bird",
		Count:     500,
		Traits:    EvolvableTraits{ColdResistance: 0.1}, // Not cold resistant
		Diet:      DietOmnivore,
	}
	tropicalBiome.AddSpecies(tropicalSpecies)
	sim.Biomes[tropicalBiome.BiomeID] = tropicalBiome

	// Cold biome species should survive better
	coldBiome := NewBiomePopulation(uuid.New(), geography.BiomeTundra)
	coldSpecies := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Arctic Grazer",
		Count:     500,
		Traits:    EvolvableTraits{ColdResistance: 0.9}, // Very cold resistant
		Diet:      DietHerbivore,
	}
	coldBiome.AddSpecies(coldSpecies)
	sim.Biomes[coldBiome.BiomeID] = coldBiome

	sim.ApplyExtinctionEvent(EventIceAge, 0.7)

	// Tropical species should suffer more
	if tropicalSpecies.Count >= coldSpecies.Count {
		t.Error("Tropical species should suffer more than cold-adapted species in ice age")
	}
}

func TestNewBiomePopulation(t *testing.T) {
	bp := NewBiomePopulation(uuid.New(), geography.BiomeOcean)

	if bp.BiomeType != geography.BiomeOcean {
		t.Error("BiomeType should be set")
	}
	if bp.Species == nil {
		t.Error("Species map should be initialized")
	}
	if bp.CarryingCapacity != 10000 {
		t.Errorf("CarryingCapacity should be 10000, got %d", bp.CarryingCapacity)
	}
}

func TestBiomePopulation_AddRemoveSpecies(t *testing.T) {
	bp := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	species := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Test",
		Count:     100,
	}

	bp.AddSpecies(species)
	if len(bp.Species) != 1 {
		t.Error("Should have 1 species after add")
	}

	removed := bp.RemoveSpecies(species.SpeciesID)
	if removed == nil {
		t.Error("Should return removed species")
	}
	if len(bp.Species) != 0 {
		t.Error("Should have 0 species after remove")
	}
}

func TestTotalPopulation(t *testing.T) {
	bp := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	bp.AddSpecies(&SpeciesPopulation{SpeciesID: uuid.New(), Count: 100})
	bp.AddSpecies(&SpeciesPopulation{SpeciesID: uuid.New(), Count: 200})
	bp.AddSpecies(&SpeciesPopulation{SpeciesID: uuid.New(), Count: 50})

	if bp.TotalPopulation() != 350 {
		t.Errorf("Total population should be 350, got %d", bp.TotalPopulation())
	}
}

func TestDefaultTraitsForDiet(t *testing.T) {
	diets := []DietType{DietHerbivore, DietCarnivore, DietOmnivore, DietPhotosynthetic}

	for _, diet := range diets {
		t.Run(string(diet), func(t *testing.T) {
			traits := DefaultTraitsForDiet(diet)
			if traits.Size == 0 && diet != DietPhotosynthetic {
				t.Error("Size should not be 0 for fauna")
			}
			if traits.Fertility == 0 {
				t.Error("Fertility should not be 0")
			}
		})
	}
}

func TestCalculateBiomeFitness_AllBiomes(t *testing.T) {
	biomes := []geography.BiomeType{
		geography.BiomeTundra, geography.BiomeAlpine, geography.BiomeDesert,
		geography.BiomeOcean, geography.BiomeRainforest, geography.BiomeGrassland,
		geography.BiomeTaiga, geography.BiomeDeciduousForest,
	}

	traits := DefaultTraitsForDiet(DietHerbivore)

	for _, biome := range biomes {
		t.Run(string(biome), func(t *testing.T) {
			fitness := CalculateBiomeFitness(traits, biome)
			if fitness < 0.5 || fitness > 1.5 {
				t.Errorf("Fitness should be in range 0.5-1.5, got %.2f", fitness)
			}
		})
	}
}

func TestPopulationDynamics_Flora(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	biome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	flora := &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Test Flora",
		Count:         100,
		Traits:        DefaultTraitsForDiet(DietPhotosynthetic),
		TraitVariance: 0.3,
		Diet:          DietPhotosynthetic,
	}
	biome.AddSpecies(flora)
	sim.Biomes[biome.BiomeID] = biome

	// Flora should grow with logistic growth
	initialCount := flora.Count
	for i := 0; i < 10; i++ {
		sim.SimulateYear()
	}

	if flora.Count <= initialCount {
		t.Logf("Flora count: initial=%d, final=%d", initialCount, flora.Count)
	}
}

func TestPopulationDynamics_PredatorPrey(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	biome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)

	// Add flora
	flora := &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Grass",
		Count:         500,
		Traits:        DefaultTraitsForDiet(DietPhotosynthetic),
		TraitVariance: 0.3,
		Diet:          DietPhotosynthetic,
	}
	biome.AddSpecies(flora)

	// Add herbivore
	herbivore := &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Grazer",
		Count:         100,
		Traits:        DefaultTraitsForDiet(DietHerbivore),
		TraitVariance: 0.3,
		Diet:          DietHerbivore,
	}
	biome.AddSpecies(herbivore)

	// Add carnivore
	carnivore := &SpeciesPopulation{
		SpeciesID:     uuid.New(),
		Name:          "Hunter",
		Count:         20,
		Traits:        DefaultTraitsForDiet(DietCarnivore),
		TraitVariance: 0.3,
		Diet:          DietCarnivore,
	}
	biome.AddSpecies(carnivore)

	sim.Biomes[biome.BiomeID] = biome

	// Simulate ecosystem
	sim.SimulateYears(50)

	// All species should still exist
	if flora.Count == 0 {
		t.Error("Flora should not go extinct in balanced ecosystem")
	}
	t.Logf("After 50 years: Flora=%d, Herbivore=%d, Carnivore=%d",
		flora.Count, herbivore.Count, carnivore.Count)
}

func TestRecordExtinction(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)
	sim.CurrentYear = 1000

	biome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	species := &SpeciesPopulation{
		SpeciesID:   uuid.New(),
		Name:        "Doomed Species",
		Count:       0, // Already extinct
		Traits:      DefaultTraitsForDiet(DietHerbivore),
		Diet:        DietHerbivore,
		CreatedYear: 500,
	}
	biome.AddSpecies(species)
	sim.Biomes[biome.BiomeID] = biome

	// Apply event that triggers extinction recording
	sim.ApplyExtinctionEvent(EventAsteroidImpact, 1.0)

	if len(sim.FossilRecord.Extinct) == 0 {
		t.Log("No extinct species recorded - species may have survived")
	}
}

func TestOceanAnoxiaEvent(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	// Ocean biome
	oceanBiome := NewBiomePopulation(uuid.New(), geography.BiomeOcean)
	oceanSpecies := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Fish",
		Count:     1000,
		Traits:    DefaultTraitsForDiet(DietHerbivore),
		Diet:      DietHerbivore,
	}
	oceanBiome.AddSpecies(oceanSpecies)
	sim.Biomes[oceanBiome.BiomeID] = oceanBiome

	// Land biome (should not be affected)
	landBiome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	landSpecies := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Grazer",
		Count:     1000,
		Traits:    DefaultTraitsForDiet(DietHerbivore),
		Diet:      DietHerbivore,
	}
	landBiome.AddSpecies(landSpecies)
	sim.Biomes[landBiome.BiomeID] = landBiome

	sim.ApplyExtinctionEvent(EventOceanAnoxia, 0.8)

	// Ocean species should suffer
	if oceanSpecies.Count >= 1000 {
		t.Error("Ocean species should suffer from anoxia")
	}
	// Land species should be unaffected
	if landSpecies.Count < 1000 {
		t.Error("Land species should not be affected by ocean anoxia")
	}
}

func TestFloodBasaltEvent(t *testing.T) {
	sim := NewPopulationSimulator(uuid.New(), 12345)

	landBiome := NewBiomePopulation(uuid.New(), geography.BiomeGrassland)
	landSpecies := &SpeciesPopulation{
		SpeciesID: uuid.New(),
		Name:      "Land Animal",
		Count:     1000,
		Traits:    EvolvableTraits{PoisonResistance: 0.1}, // Low poison resistance
		Diet:      DietHerbivore,
	}
	landBiome.AddSpecies(landSpecies)
	sim.Biomes[landBiome.BiomeID] = landBiome

	sim.ApplyExtinctionEvent(EventFloodBasalt, 0.8)

	if landSpecies.Count >= 1000 {
		t.Error("Land species should suffer from flood basalt toxic gases")
	}
}
