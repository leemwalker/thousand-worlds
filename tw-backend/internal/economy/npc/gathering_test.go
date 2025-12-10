package npc

import (
	"context"
	"testing"

	"tw-backend/internal/economy/resources"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockResourceFinder struct {
	mock.Mock
}

func (m *MockResourceFinder) FindNearbyNodes(location uuid.UUID, radius float64) ([]*resources.ResourceNode, error) {
	args := m.Called(location, radius)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*resources.ResourceNode), args.Error(1)
}

type MockHarvester struct {
	mock.Mock
}

func (m *MockHarvester) Harvest(nodeID uuid.UUID, skillLevel int, toolQuality float64) (*resources.HarvestResult, error) {
	args := m.Called(nodeID, skillLevel, toolQuality)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*resources.HarvestResult), args.Error(1)
}

type MockInventoryManager struct {
	mock.Mock
}

func (m *MockInventoryManager) Add(itemID uuid.UUID, quantity int) error {
	args := m.Called(itemID, quantity)
	return args.Error(0)
}

func (m *MockInventoryManager) IsFull() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestGatherResources(t *testing.T) {
	// Setup
	finder := new(MockResourceFinder)
	harvester := new(MockHarvester)
	inventory := new(MockInventoryManager)

	npcID := uuid.New()
	locationID := uuid.New()

	npc := &NPC{
		ID:         npcID,
		Location:   locationID,
		Occupation: OccupationFarmer,
		Skills:     map[string]int{"farming": 50},
		Inventory:  inventory,
		Desires: &Desires{
			ResourceAcquisition: 80.0, // High need
		},
		Equipment: &Equipment{
			ToolQuality: 1.0,
		},
	}

	// Mock data
	grainNode := &resources.ResourceNode{
		NodeID:   uuid.New(),
		Name:     "Wild Grain",
		Type:     resources.ResourceVegetation,
		Quantity: 100,
	}

	oreNode := &resources.ResourceNode{
		NodeID:   uuid.New(),
		Name:     "Iron Ore",
		Type:     resources.ResourceMineral,
		Quantity: 100,
	}

	// Expectation: Find nearby nodes
	// Should filter out oreNode because farmer prefers grain
	finder.On("FindNearbyNodes", locationID, 200.0).Return([]*resources.ResourceNode{grainNode, oreNode}, nil)

	// Expectation: Harvest grain
	yield := &resources.HarvestResult{
		YieldAmount: 5,
		Success:     true,
	}
	harvester.On("Harvest", grainNode.NodeID, 50, 1.0).Return(yield, nil)

	// Expectation: Add to inventory
	inventory.On("Add", mock.Anything, 5).Return(nil)

	// Execute
	err := npc.GatherResources(context.Background(), finder, harvester)

	// Verify
	assert.NoError(t, err)
	assert.Less(t, npc.Desires.ResourceAcquisition, 80.0, "Desire should decrease")

	finder.AssertExpectations(t)
	harvester.AssertExpectations(t)
	inventory.AssertExpectations(t)
}

func TestGatherResources_LowDesire(t *testing.T) {
	npc := &NPC{
		Desires: &Desires{
			ResourceAcquisition: 20.0, // Low need
		},
	}

	err := npc.GatherResources(context.Background(), nil, nil)
	assert.NoError(t, err)
	// Should return early without calling dependencies
}

func TestGatherResources_NoSuitableResources(t *testing.T) {
	finder := new(MockResourceFinder)

	npc := &NPC{
		Location:   uuid.New(),
		Occupation: OccupationFarmer,
		Desires: &Desires{
			ResourceAcquisition: 80.0,
		},
	}

	// Only ore available, farmer doesn't want it
	oreNode := &resources.ResourceNode{
		Name: "Iron Ore",
		Type: resources.ResourceMineral,
	}

	finder.On("FindNearbyNodes", mock.Anything, mock.Anything).Return([]*resources.ResourceNode{oreNode}, nil)

	err := npc.GatherResources(context.Background(), finder, nil)
	assert.Equal(t, ErrNoSuitableResources, err)
}
