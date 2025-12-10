package resources

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateCluster(t *testing.T) {
	primaryNode := &ResourceNode{
		NodeID:    uuid.New(),
		Name:      "Oak Wood",
		Type:      ResourceVegetation,
		LocationX: 1000.0,
		LocationY: 1000.0,
		Quantity:  100,
	}

	// Run multiple times to ensure we get clusters occasionally
	clusterCreated := false
	totalNodes := 0

	for i := 0; i < 20; i++ {
		cluster := CreateCluster(primaryNode, int64(i))
		if len(cluster) > 0 {
			clusterCreated = true
			totalNodes += len(cluster)

			// Verify cluster properties
			for _, node := range cluster {
				// Should be same type/name
				assert.Equal(t, primaryNode.Name, node.Name)
				assert.Equal(t, primaryNode.Type, node.Type)

				// Should be new ID
				assert.NotEqual(t, primaryNode.NodeID, node.NodeID)

				// Should be within 200m radius
				dx := node.LocationX - primaryNode.LocationX
				dy := node.LocationY - primaryNode.LocationY
				dist := dx*dx + dy*dy
				assert.LessOrEqual(t, dist, 200.0*200.0)
				assert.GreaterOrEqual(t, dist, 50.0*50.0) // Min 50m
			}
		}
	}

	assert.True(t, clusterCreated, "Should create at least one cluster in 20 attempts")
}

func TestCreateRichDeposit(t *testing.T) {
	template := ResourceTemplate{
		Name:        "Wild Berries",
		Type:        ResourceVegetation,
		Rarity:      RarityCommon,
		MaxQuantity: 100,
		RegenRate:   10.0,
	}

	// Run multiple times to find a rich deposit
	richCreated := false

	for i := 0; i < 100; i++ {
		node := CreateRichDeposit(template, int64(i))
		if node != nil {
			richCreated = true

			// Verify rich properties
			assert.Equal(t, template.MaxQuantity*3, node.Quantity)
			assert.Equal(t, template.MaxQuantity*3, node.MaxQuantity)
			assert.Equal(t, template.RegenRate*2, node.RegenRate)
			assert.Equal(t, RarityUncommon, node.Rarity) // Upgraded rarity
		}
	}

	assert.True(t, richCreated, "Should create at least one rich deposit in 100 attempts")
}

func TestClusterBiomeAffinity(t *testing.T) {
	primaryNode := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          "Cactus",
		Type:          ResourceVegetation,
		BiomeAffinity: []string{"Desert"},
	}

	cluster := CreateCluster(primaryNode, 12345)

	if len(cluster) > 0 {
		for _, node := range cluster {
			assert.Equal(t, primaryNode.BiomeAffinity, node.BiomeAffinity)
		}
	}
}
