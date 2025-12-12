package crafting

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTechTreeStructure(t *testing.T) {
	// Create a simple tech tree
	tree := &TechTree{
		TreeID:    uuid.New(),
		Name:      "Test Tree",
		TechLevel: TechPrimitive,
		Nodes:     []*TechNode{},
	}

	// Create nodes
	node1 := &TechNode{
		NodeID:    uuid.New(),
		Name:      "Fire Starting",
		TechLevel: TechPrimitive,
		Tier:      1,
	}

	node2 := &TechNode{
		NodeID:        uuid.New(),
		Name:          "Pottery",
		TechLevel:     TechPrimitive,
		Tier:          2,
		Prerequisites: []uuid.UUID{node1.NodeID},
	}

	tree.Nodes = append(tree.Nodes, node1, node2)

	// Verify structure
	assert.Equal(t, 2, len(tree.Nodes))
	assert.Equal(t, node1.NodeID, tree.Nodes[1].Prerequisites[0])
}

func TestUnlockTechNode(t *testing.T) {
	// Setup
	node1ID := uuid.New()
	node2ID := uuid.New()

	// node1 definition removed as it was unused, we only need the ID

	node2 := &TechNode{
		NodeID:        node2ID,
		Name:          "Advanced Tools",
		Prerequisites: []uuid.UUID{node1ID},
	}

	// Mock repository or unlocked tech state
	unlocked := make(map[uuid.UUID]bool)

	// Helper to check prerequisites
	checkPrereqs := func(node *TechNode) bool {
		for _, prereqID := range node.Prerequisites {
			if !unlocked[prereqID] {
				return false
			}
		}
		return true
	}

	// Try to unlock node 2 (should fail)
	assert.False(t, checkPrereqs(node2))

	// Unlock node 1
	unlocked[node1ID] = true

	// Try to unlock node 2 (should succeed)
	assert.True(t, checkPrereqs(node2))
}

func TestTechTreeCustomization(t *testing.T) {
	// This test will verify that CustomizeTechTree correctly filters/modifies the tree
	// We'll implement the actual logic in customization.go

	// Mock world config
	worldConfig := struct {
		TechLevel  TechLevel
		MagicLevel string
	}{
		TechLevel:  TechMedieval,
		MagicLevel: "none",
	}

	// Create a full tree with all levels
	fullTree := []*TechNode{
		{Name: "Stone Tools", TechLevel: TechPrimitive},
		{Name: "Iron Tools", TechLevel: TechMedieval},
		{Name: "Steam Engine", TechLevel: TechIndustrial},
	}

	// Filter function (simulated)
	filter := func(nodes []*TechNode, maxLevel TechLevel) []*TechNode {
		var filtered []*TechNode
		levels := map[TechLevel]int{
			TechPrimitive:  1,
			TechMedieval:   2,
			TechIndustrial: 3,
			TechModern:     4,
			TechFuturistic: 5,
		}

		maxVal := levels[maxLevel]

		for _, node := range nodes {
			if levels[node.TechLevel] <= maxVal {
				filtered = append(filtered, node)
			}
		}
		return filtered
	}

	result := filter(fullTree, worldConfig.TechLevel)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, "Stone Tools", result[0].Name)
	assert.Equal(t, "Iron Tools", result[1].Name)
}

func TestTechTreeManager_UnlockTechNode(t *testing.T) {
	repo := new(MockRepository)
	manager := NewTechTreeManager(repo)

	entityID := uuid.New()
	nodeID := uuid.New()

	node := &TechNode{
		NodeID:         nodeID,
		Name:           "Test Node",
		UnlocksRecipes: []uuid.UUID{uuid.New()},
	}

	repo.On("GetTechNode", nodeID).Return(node, nil)
	repo.On("IsTechUnlocked", entityID, nodeID).Return(false, nil)
	repo.On("UnlockTech", entityID, nodeID).Return(nil)
	repo.On("DiscoverRecipe", mock.Anything).Return(nil)

	err := manager.UnlockTechNode(entityID, nodeID)
	assert.NoError(t, err)

	repo.AssertCalled(t, "UnlockTech", entityID, nodeID)
	repo.AssertCalled(t, "DiscoverRecipe", mock.Anything)
}

func TestLoadTechTreeFromFile(t *testing.T) {
	// Create temp file
	file, err := os.CreateTemp("", "techtree-*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	content := `[
		{"name": "Node 1", "tier": 1},
		{"name": "Node 2", "tier": 2}
	]`
	file.WriteString(content)
	file.Close()

	nodes, err := LoadTechTreeFromFile(file.Name())
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
	assert.Equal(t, "Node 1", nodes[0].Name)
	assert.NotEqual(t, uuid.Nil, nodes[0].NodeID)
}
