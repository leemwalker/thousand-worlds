package inventory

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"tw-backend/internal/game/services/entity"
)

// MockRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) AddItem(ctx context.Context, charID uuid.UUID, item Item, quantity int) error {
	args := m.Called(ctx, charID, item, quantity)
	return args.Error(0)
}

func (m *MockRepository) RemoveItem(ctx context.Context, charID uuid.UUID, itemID uuid.UUID, quantity int) error {
	args := m.Called(ctx, charID, itemID, quantity)
	return args.Error(0)
}

func (m *MockRepository) GetInventory(ctx context.Context, charID uuid.UUID) ([]InventoryItem, error) {
	args := m.Called(ctx, charID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]InventoryItem), args.Error(1)
}

func TestInventoryService_AddItem(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	entSvc := entity.NewService()
	svc := NewService(entSvc, mockRepo)

	ctx := context.Background()
	charID := uuid.New()
	item := Item{ID: uuid.New(), Name: "Sword", Description: "Sharp"}

	// Expectation
	mockRepo.On("AddItem", ctx, charID, item, 1).Return(nil)

	// Test
	err := svc.AddItem(ctx, charID, item)

	// Verify
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestInventoryService_GetInventory(t *testing.T) {
	mockRepo := new(MockRepository)
	entSvc := entity.NewService()
	svc := NewService(entSvc, mockRepo)

	ctx := context.Background()
	charID := uuid.New()
	expectedItems := []InventoryItem{
		{ID: uuid.New(), CharacterID: charID, Name: "Shield"},
	}

	mockRepo.On("GetInventory", ctx, charID).Return(expectedItems, nil)

	items, err := svc.GetInventory(ctx, charID)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Shield", items[0].Name)
}
