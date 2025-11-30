package resources

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateResourceNode(node *ResourceNode) error {
	args := m.Called(node)
	return args.Error(0)
}

func (m *MockRepository) GetResourceNode(nodeID uuid.UUID) (*ResourceNode, error) {
	args := m.Called(nodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ResourceNode), args.Error(1)
}

func (m *MockRepository) GetAllResourceNodes() ([]*ResourceNode, error) {
	args := m.Called()
	return args.Get(0).([]*ResourceNode), args.Error(1)
}

func (m *MockRepository) UpdateResourceNode(node *ResourceNode) error {
	args := m.Called(node)
	return args.Error(0)
}

// Implement other interface methods with stubs
func (m *MockRepository) GetResourceNodesByType(resourceType ResourceType) ([]*ResourceNode, error) {
	return nil, nil
}
func (m *MockRepository) GetResourceNodesByBiome(biomeType string) ([]*ResourceNode, error) {
	return nil, nil
}
func (m *MockRepository) GetResourceNodesInRadius(x, y, radius float64) ([]*ResourceNode, error) {
	return nil, nil
}
func (m *MockRepository) DeleteResourceNode(nodeID uuid.UUID) error      { return nil }
func (m *MockRepository) GetMineralDeposits() ([]*MineralDeposit, error) { return nil, nil }
func (m *MockRepository) GetMineralDepositByID(depositID uuid.UUID) (*MineralDeposit, error) {
	return nil, nil
}

func TestRegenerateVegetation(t *testing.T) {
	lastHarvested := time.Now().Add(-24 * time.Hour) // 24 hours ago

	node := &ResourceNode{
		NodeID:        uuid.New(),
		Name:          "Medicinal Herbs",
		Type:          ResourceVegetation,
		Quantity:      50,
		MaxQuantity:   100,
		RegenRate:     10.0, // 10 units per day
		RegenCooldown: 12 * time.Hour,
		LastHarvested: &lastHarvested,
	}

	repo := new(MockRepository)
	repo.On("GetAllResourceNodes").Return([]*ResourceNode{node}, nil)
	repo.On("UpdateResourceNode", mock.MatchedBy(func(n *ResourceNode) bool {
		return n.Quantity == 60 // 50 + 10 units
	})).Return(nil)

	// Simulate 24 hours passing
	err := RegenerateResources(24*time.Hour, repo)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestRegenerateCooldownNotExpired(t *testing.T) {
	lastHarvested := time.Now().Add(-1 * time.Hour) // 1 hour ago

	node := &ResourceNode{
		NodeID:        uuid.New(),
		Type:          ResourceVegetation,
		Quantity:      50,
		MaxQuantity:   100,
		RegenRate:     10.0,
		RegenCooldown: 24 * time.Hour, // 24 hour cooldown
		LastHarvested: &lastHarvested,
	}

	repo := new(MockRepository)
	repo.On("GetAllResourceNodes").Return([]*ResourceNode{node}, nil)
	// Update should NOT be called

	err := RegenerateResources(1*time.Hour, repo)

	assert.NoError(t, err)
	repo.AssertNotCalled(t, "UpdateResourceNode")
}

func TestSkipMineralRegen(t *testing.T) {
	node := &ResourceNode{
		NodeID:      uuid.New(),
		Type:        ResourceMineral,
		Quantity:    50,
		MaxQuantity: 100,
		RegenRate:   0.0,
	}

	repo := new(MockRepository)
	repo.On("GetAllResourceNodes").Return([]*ResourceNode{node}, nil)
	// Update should NOT be called for minerals

	err := RegenerateResources(24*time.Hour, repo)

	assert.NoError(t, err)
	repo.AssertNotCalled(t, "UpdateResourceNode")
}

func TestRegenerateCapAtMax(t *testing.T) {
	lastHarvested := time.Now().Add(-24 * time.Hour)

	node := &ResourceNode{
		NodeID:        uuid.New(),
		Type:          ResourceVegetation,
		Quantity:      95,
		MaxQuantity:   100,
		RegenRate:     10.0, // +10 units
		RegenCooldown: 0,
		LastHarvested: &lastHarvested,
	}

	repo := new(MockRepository)
	repo.On("GetAllResourceNodes").Return([]*ResourceNode{node}, nil)
	repo.On("UpdateResourceNode", mock.MatchedBy(func(n *ResourceNode) bool {
		return n.Quantity == 100 // Capped at 100
	})).Return(nil)

	err := RegenerateResources(24*time.Hour, repo)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// Mock species population function for testing

func TestAnimalPopulationCap(t *testing.T) {
	speciesID := uuid.New()
	lastHarvested := time.Now().Add(-24 * time.Hour)

	node := &ResourceNode{
		NodeID:        uuid.New(),
		Type:          ResourceAnimal,
		Quantity:      10,
		MaxQuantity:   100,
		RegenRate:     50.0, // High regen
		SpeciesID:     &speciesID,
		LastHarvested: &lastHarvested,
	}

	// Mock population of 20 animals
	// Max resources should be 20 * 2 = 40
	oldFunc := GetSpeciesPopulationFunc
	GetSpeciesPopulationFunc = func(id uuid.UUID) int {
		return 20
	}
	defer func() { GetSpeciesPopulationFunc = oldFunc }()

	repo := new(MockRepository)
	repo.On("GetAllResourceNodes").Return([]*ResourceNode{node}, nil)
	repo.On("UpdateResourceNode", mock.MatchedBy(func(n *ResourceNode) bool {
		return n.Quantity == 40 // Capped by population
	})).Return(nil)

	err := RegenerateResources(24*time.Hour, repo)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
