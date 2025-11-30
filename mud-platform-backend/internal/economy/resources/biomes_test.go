package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetForestResources(t *testing.T) {
	templates := GetResourceTemplatesForBiome("Deciduous Forest")

	assert.NotEmpty(t, templates)

	// Check for expected forest resources
	hasWood := false
	hasHerbs := false
	hasGame := false

	for _, tmpl := range templates {
		if tmpl.Type == ResourceVegetation && tmpl.Name == "Oak Wood" {
			hasWood = true
			assert.Equal(t, RarityCommon, tmpl.Rarity)
			assert.Equal(t, "gathering", tmpl.RequiredSkill)
		}
		if tmpl.Type == ResourceVegetation && tmpl.Name == "Medicinal Herbs" {
			hasHerbs = true
		}
		if tmpl.Type == ResourceAnimal && tmpl.Name == "Deer Hide" {
			hasGame = true
		}
	}

	assert.True(t, hasWood, "Forest should have wood resources")
	assert.True(t, hasHerbs, "Forest should have herb resources")
	assert.True(t, hasGame, "Forest should have game animal resources")
}

func TestGetGrasslandResources(t *testing.T) {
	templates := GetResourceTemplatesForBiome("Grassland")

	assert.NotEmpty(t, templates)

	hasGrain := false
	hasFiber := false
	hasGrazing := false

	for _, tmpl := range templates {
		if tmpl.Name == "Wild Grain" {
			hasGrain = true
			assert.Equal(t, ResourceVegetation, tmpl.Type)
		}
		if tmpl.Name == "Cotton Fiber" {
			hasFiber = true
		}
		if tmpl.Name == "Bison Hide" {
			hasGrazing = true
			assert.Equal(t, ResourceAnimal, tmpl.Type)
		}
	}

	assert.True(t, hasGrain)
	assert.True(t, hasFiber)
	assert.True(t, hasGrazing)
}

func TestGetDesertResources(t *testing.T) {
	templates := GetResourceTemplatesForBiome("Desert")

	assert.NotEmpty(t, templates)

	hasCactus := false
	hasCrystal := false

	for _, tmpl := range templates {
		if tmpl.Name == "Desert Cactus" {
			hasCactus = true
			assert.Equal(t, RarityUncommon, tmpl.Rarity)
		}
		if tmpl.Name == "Rare Crystal" {
			hasCrystal = true
			assert.Equal(t, ResourceSpecial, tmpl.Type)
			assert.Equal(t, RarityRare, tmpl.Rarity)
		}
	}

	assert.True(t, hasCactus)
	assert.True(t, hasCrystal)
}

func TestGetOceanResources(t *testing.T) {
	templates := GetResourceTemplatesForBiome("Ocean")

	assert.NotEmpty(t, templates)

	hasFish := false
	hasPearls := false
	hasKelp := false

	for _, tmpl := range templates {
		if tmpl.Name == "Fish" {
			hasFish = true
			assert.Equal(t, ResourceAnimal, tmpl.Type)
			assert.Equal(t, RarityCommon, tmpl.Rarity)
		}
		if tmpl.Name == "Pearl" {
			hasPearls = true
			assert.Equal(t, RarityRare, tmpl.Rarity)
		}
		if tmpl.Name == "Kelp" {
			hasKelp = true
		}
	}

	assert.True(t, hasFish)
	assert.True(t, hasPearls)
	assert.True(t, hasKelp)
}

func TestGetTundraResources(t *testing.T) {
	templates := GetResourceTemplatesForBiome("Tundra")

	assert.NotEmpty(t, templates)

	hasFur := false
	hasIceCrystal := false

	for _, tmpl := range templates {
		if tmpl.Name == "Arctic Fox Fur" {
			hasFur = true
			assert.Equal(t, ResourceAnimal, tmpl.Type)
		}
		if tmpl.Name == "Ice Crystal" {
			hasIceCrystal = true
			assert.Equal(t, ResourceSpecial, tmpl.Type)
		}
	}

	assert.True(t, hasFur)
	assert.True(t, hasIceCrystal)
}

func TestResourceTemplateProperties(t *testing.T) {
	templates := GetResourceTemplatesForBiome("Deciduous Forest")

	for _, tmpl := range templates {
		// All templates must have valid properties
		assert.NotEmpty(t, tmpl.Name)
		assert.NotEmpty(t, tmpl.Type)
		assert.NotEmpty(t, tmpl.Rarity)
		assert.Greater(t, tmpl.MaxQuantity, 0)
		assert.GreaterOrEqual(t, tmpl.RegenRate, 0.0)
		assert.GreaterOrEqual(t, tmpl.CooldownHours, 0)
		assert.NotEmpty(t, tmpl.RequiredSkill)
		assert.GreaterOrEqual(t, tmpl.MinSkillLevel, 0)
		assert.LessOrEqual(t, tmpl.MinSkillLevel, 100)
	}
}

func TestVegetationRegenRates(t *testing.T) {
	templates := GetResourceTemplatesForBiome("Deciduous Forest")

	for _, tmpl := range templates {
		if tmpl.Type == ResourceVegetation {
			// Fast-growing vegetation: 10-20 units/day
			// Slow-growing trees: 1-3 units/day
			if tmpl.Name == "Oak Wood" || tmpl.Name == "Birch Wood" {
				assert.GreaterOrEqual(t, tmpl.RegenRate, 1.0)
				assert.LessOrEqual(t, tmpl.RegenRate, 3.0)
			} else {
				// Herbs and fast vegetation
				assert.GreaterOrEqual(t, tmpl.RegenRate, 0.0)
				assert.LessOrEqual(t, tmpl.RegenRate, 20.0)
			}
		}
	}
}

func TestAnimalRegenRates(t *testing.T) {
	templates := GetResourceTemplatesForBiome("Grassland")

	for _, tmpl := range templates {
		if tmpl.Type == ResourceAnimal {
			// Animal resources: 5-10 units/day
			assert.GreaterOrEqual(t, tmpl.RegenRate, 5.0)
			assert.LessOrEqual(t, tmpl.RegenRate, 10.0)
		}
	}
}

func TestBiomeAffinityRainforest(t *testing.T) {
	templates := GetResourceTemplatesForBiome("Rainforest")

	assert.NotEmpty(t, templates)

	// Rainforests should have dense vegetation
	hasExoticWood := false
	hasMedicinalPlants := false

	for _, tmpl := range templates {
		if tmpl.Name == "Mahogany Wood" {
			hasExoticWood = true
		}
		if tmpl.Name == "Rare Medicinal Plant" {
			hasMedicinalPlants = true
		}
	}

	assert.True(t, hasExoticWood || hasMedicinalPlants, "Rainforest should have unique vegetation")
}
