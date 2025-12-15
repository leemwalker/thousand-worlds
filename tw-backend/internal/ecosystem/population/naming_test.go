package population

import (
	"testing"

	"tw-backend/internal/worldgen/geography"
)

func TestCoveringType(t *testing.T) {
	tests := []struct {
		name     string
		covering CoveringType
		expected string
	}{
		{"Fur covering", CoveringFur, "fur"},
		{"Scales covering", CoveringScales, "scales"},
		{"Feathers covering", CoveringFeathers, "feathers"},
		{"Shell covering", CoveringShell, "shell"},
		{"Skin covering", CoveringSkin, "skin"},
		{"Bark covering", CoveringBark, "bark"},
		{"None covering", CoveringNone, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.covering) != tt.expected {
				t.Errorf("CoveringType = %s, expected %s", tt.covering, tt.expected)
			}
		})
	}
}

func TestFloraGrowthType(t *testing.T) {
	tests := []struct {
		name     string
		growth   FloraGrowthType
		expected string
	}{
		{"Evergreen", FloraEvergreen, "evergreen"},
		{"Deciduous", FloraDeciduous, "deciduous"},
		{"Annual", FloraAnnual, "annual"},
		{"Perennial", FloraPerennial, "perennial"},
		{"Succulent", FloraSucculent, "succulent"},
		{"Aquatic", FloraAquatic, "aquatic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.growth) != tt.expected {
				t.Errorf("FloraGrowthType = %s, expected %s", tt.growth, tt.expected)
			}
		})
	}
}

func TestGenerateSpeciesName_Fauna(t *testing.T) {
	tests := []struct {
		name     string
		traits   EvolvableTraits
		diet     DietType
		biome    geography.BiomeType
		contains []string // Words that should appear in the name
	}{
		{
			name: "Large furry herbivore",
			traits: EvolvableTraits{
				Size: 8.0, Speed: 3.0, Strength: 4.0,
				Social: 0.9, Covering: CoveringFur,
			},
			diet:     DietHerbivore,
			biome:    geography.BiomeGrassland,
			contains: []string{"Giant", "Woolly", "Auroch"},
		},
		{
			name: "Small fast scaled carnivore",
			traits: EvolvableTraits{
				Size: 0.5, Speed: 9.0, Strength: 2.0,
				Covering: CoveringScales,
			},
			diet:     DietCarnivore,
			biome:    geography.BiomeDesert,
			contains: []string{"Small", "Swift", "Stalker"},
		},
		{
			name: "Feathered predator",
			traits: EvolvableTraits{
				Size: 2.0, Speed: 7.0, Strength: 3.0,
				Covering: CoveringFeathers,
			},
			diet:     DietCarnivore,
			biome:    geography.BiomeTaiga,
			contains: []string{"Feathered", "Stalker"},
		},
		{
			name: "Pack hunter with intelligence",
			traits: EvolvableTraits{
				Size: 3.0, Speed: 6.0, Strength: 5.0,
				Social: 0.8, Intelligence: 0.7,
				Covering: CoveringFur,
			},
			diet:     DietCarnivore,
			biome:    geography.BiomeGrassland,
			contains: []string{"Large", "Woolly", "Packwolf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSpeciesName(tt.traits, tt.diet, tt.biome)
			for _, word := range tt.contains {
				if !containsWord(result, word) {
					t.Errorf("GenerateSpeciesName() = %q, expected to contain %q", result, word)
				}
			}
		})
	}
}

func TestGenerateSpeciesName_Flora(t *testing.T) {
	tests := []struct {
		name     string
		traits   EvolvableTraits
		biome    geography.BiomeType
		contains []string
	}{
		{
			name: "Tall forest tree",
			traits: EvolvableTraits{
				Size: 5.0, FloraGrowth: FloraDeciduous,
				Covering: CoveringBark,
			},
			biome:    geography.BiomeDeciduousForest,
			contains: []string{"Broadleaf"},
		},
		{
			name: "Small aquatic plant",
			traits: EvolvableTraits{
				Size: 0.1, FloraGrowth: FloraAquatic,
				Covering: CoveringNone,
			},
			biome:    geography.BiomeOcean,
			contains: []string{"Plankton"},
		},
		{
			name: "Desert succulent",
			traits: EvolvableTraits{
				Size: 1.0, FloraGrowth: FloraSucculent,
				Covering: CoveringNone,
			},
			biome:    geography.BiomeDesert,
			contains: []string{"Succulent"},
		},
		{
			name: "Evergreen conifer",
			traits: EvolvableTraits{
				Size: 4.0, FloraGrowth: FloraEvergreen,
				Covering: CoveringBark,
			},
			biome:    geography.BiomeTaiga,
			contains: []string{"Evergreen"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSpeciesName(tt.traits, DietPhotosynthetic, tt.biome)
			for _, word := range tt.contains {
				if !containsWord(result, word) {
					t.Errorf("GenerateSpeciesName() = %q, expected to contain %q", result, word)
				}
			}
		})
	}
}

func TestDescribePopulation(t *testing.T) {
	tests := []struct {
		count    int64
		name     string
		expected string
	}{
		{1, "Wolf", "A lone Wolf"},
		{5, "Rabbit", "A few Rabbits"},
		{25, "Deer", "A small herd of Deer"},
		{100, "Bison", "A herd of Bison"},
		{350, "Antelope", "A large herd of Antelope"},
		{750, "Locust", "A thriving swarm of Locusts"},
		{3000, "Kelp", "A massive forest of Kelp"},
		{8000, "Plankton", "An ecosystem dominated by Plankton"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := DescribePopulation(tt.count, tt.name)
			if result != tt.expected {
				t.Errorf("DescribePopulation(%d, %q) = %q, expected %q", tt.count, tt.name, result, tt.expected)
			}
		})
	}
}

func TestCalculateBiomeFitness(t *testing.T) {
	tests := []struct {
		name   string
		traits EvolvableTraits
		biome  geography.BiomeType
		minFit float64
		maxFit float64
	}{
		{
			name:   "Cold-resistant in tundra",
			traits: EvolvableTraits{ColdResistance: 0.9, HeatResistance: 0.2},
			biome:  geography.BiomeTundra,
			minFit: 1.1,
			maxFit: 1.5,
		},
		{
			name:   "Heat-resistant in desert",
			traits: EvolvableTraits{HeatResistance: 0.9, ColdResistance: 0.2, NightVision: 0.8},
			biome:  geography.BiomeDesert,
			minFit: 1.1,
			maxFit: 1.5,
		},
		{
			name:   "Cold-adapted in desert (bad fit)",
			traits: EvolvableTraits{ColdResistance: 0.9, HeatResistance: 0.1},
			biome:  geography.BiomeDesert,
			minFit: 0.5,
			maxFit: 0.9,
		},
		{
			name:   "Fast in grassland",
			traits: EvolvableTraits{Speed: 8.0, Social: 0.8},
			biome:  geography.BiomeGrassland,
			minFit: 1.2,
			maxFit: 1.5,
		},
		{
			name:   "Camouflaged in rainforest",
			traits: EvolvableTraits{Camouflage: 0.9, Intelligence: 0.8, HeatResistance: 0.7},
			biome:  geography.BiomeRainforest,
			minFit: 1.2,
			maxFit: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBiomeFitness(tt.traits, tt.biome)
			if result < tt.minFit || result > tt.maxFit {
				t.Errorf("CalculateBiomeFitness() = %.2f, expected between %.2f and %.2f", result, tt.minFit, tt.maxFit)
			}
		})
	}
}

func TestGetCoveringForDiet(t *testing.T) {
	tests := []struct {
		diet     DietType
		biome    geography.BiomeType
		expected CoveringType
	}{
		{DietPhotosynthetic, geography.BiomeOcean, CoveringNone},
		{DietPhotosynthetic, geography.BiomeGrassland, CoveringBark},
		{DietHerbivore, geography.BiomeTundra, CoveringFur},
		{DietHerbivore, geography.BiomeDesert, CoveringScales},
		{DietCarnivore, geography.BiomeOcean, CoveringScales},
	}

	for _, tt := range tests {
		t.Run(string(tt.diet)+"_"+string(tt.biome), func(t *testing.T) {
			result := GetCoveringForDiet(tt.diet, tt.biome)
			if result != tt.expected {
				t.Errorf("GetCoveringForDiet(%s, %s) = %s, expected %s",
					tt.diet, tt.biome, result, tt.expected)
			}
		})
	}
}

func TestGetFloraGrowthForBiome(t *testing.T) {
	tests := []struct {
		biome    geography.BiomeType
		expected FloraGrowthType
	}{
		{geography.BiomeTaiga, FloraEvergreen},
		{geography.BiomeAlpine, FloraEvergreen},
		{geography.BiomeDeciduousForest, FloraDeciduous},
		{geography.BiomeDesert, FloraSucculent},
		{geography.BiomeOcean, FloraAquatic},
		{geography.BiomeGrassland, FloraPerennial},
		{geography.BiomeRainforest, FloraEvergreen},
		{geography.BiomeTundra, FloraPerennial},
	}

	for _, tt := range tests {
		t.Run(string(tt.biome), func(t *testing.T) {
			result := GetFloraGrowthForBiome(tt.biome)
			if result != tt.expected {
				t.Errorf("GetFloraGrowthForBiome(%s) = %s, expected %s",
					tt.biome, result, tt.expected)
			}
		})
	}
}

func TestGetAllEpochs(t *testing.T) {
	epochs := GetAllEpochs()
	if len(epochs) != 10 {
		t.Errorf("Expected 10 epochs, got %d", len(epochs))
	}
}

func TestGetAllGoals(t *testing.T) {
	goals := GetAllGoals()
	if len(goals) != 12 {
		t.Errorf("Expected 12 goals, got %d", len(goals))
	}
}

// Helper function to check if a string contains a word
func containsWord(s, word string) bool {
	return len(s) > 0 && len(word) > 0 &&
		(s == word ||
			len(s) >= len(word) &&
				(s[:len(word)] == word || s[len(s)-len(word):] == word ||
					indexOf(s, " "+word+" ") >= 0 || indexOf(s, " "+word) >= 0 || indexOf(s, word+" ") >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
