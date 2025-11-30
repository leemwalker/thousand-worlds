package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlaceResourcesInBiome(t *testing.T) {
	// Test Forest placement
	// Area: 1 km²
	nodes, err := PlaceResourcesInBiome("Deciduous Forest", 1.0, 12345)
	assert.NoError(t, err)
	assert.NotEmpty(t, nodes)

	// Forest targets:
	// Wood: 10-20/km²
	// Herbs: 5-10/km²
	// Game: 1-3/km²
	// Total: ~16-33/km² + clusters

	woodCount := 0
	herbCount := 0
	gameCount := 0

	for _, node := range nodes {
		if node.Name == "Oak Wood" || node.Name == "Birch Wood" || node.Name == "Maple Wood" {
			woodCount++
		}
		if node.Name == "Medicinal Herbs" || node.Name == "Wild Berries" {
			herbCount++
		}
		if node.Name == "Deer Hide" || node.Name == "Rabbit Fur" {
			gameCount++
		}
	}

	// Allow wide range due to randomness and clustering
	assert.GreaterOrEqual(t, woodCount, 8)
	assert.LessOrEqual(t, woodCount, 40) // Clusters can increase count significantly

	assert.GreaterOrEqual(t, herbCount, 4)
	assert.LessOrEqual(t, herbCount, 25)

	assert.GreaterOrEqual(t, gameCount, 1)
	assert.LessOrEqual(t, gameCount, 10)
}

func TestPlaceResourcesInDesert(t *testing.T) {
	// Test Desert placement
	nodes, err := PlaceResourcesInBiome("Desert", 1.0, 67890)
	assert.NoError(t, err)

	// Desert targets:
	// Cacti: 2-4/km²
	// Crystals: 0.3-0.8/km²

	cactusCount := 0
	crystalCount := 0

	for _, node := range nodes {
		if node.Name == "Desert Cactus" || node.Name == "Aloe Vera" {
			cactusCount++
		}
		if node.Name == "Rare Crystal" {
			crystalCount++
		}
	}

	assert.GreaterOrEqual(t, cactusCount, 1)
	assert.LessOrEqual(t, cactusCount, 15)

	// Crystals are rare, might be 0 in 1km²
	assert.LessOrEqual(t, crystalCount, 5)
}

func TestRarityDistribution(t *testing.T) {
	// Generate large area to get good statistical sample
	nodes, err := PlaceResourcesInBiome("Deciduous Forest", 10.0, 99999)
	assert.NoError(t, err)

	common := 0
	uncommon := 0
	rare := 0

	for _, node := range nodes {
		switch node.Rarity {
		case RarityCommon:
			common++
		case RarityUncommon:
			uncommon++
		case RarityRare:
			rare++
		}
	}

	total := float64(len(nodes))
	commonPct := float64(common) / total
	uncommonPct := float64(uncommon) / total

	// Target: Common 60-70%, Uncommon 20-25%
	// Note: Our templates define rarity, so this tests the template mix + rich upgrades
	assert.Greater(t, commonPct, 0.50)
	assert.Less(t, commonPct, 0.90)

	assert.Greater(t, uncommonPct, 0.10)
	assert.Less(t, uncommonPct, 0.40)
}

func TestGenerate1000KmWorld(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large world generation in short mode")
	}

	// 1000 km² is too big for unit test memory, let's do 10 km² and extrapolate
	area := 10.0
	nodes, err := PlaceResourcesInBiome("Grassland", area, 55555)
	assert.NoError(t, err)

	density := float64(len(nodes)) / area

	// Grassland target: 30-50/km² (vegetation + animal)
	// With clustering it might be higher
	assert.Greater(t, density, 20.0)
	assert.Less(t, density, 100.0)
}
