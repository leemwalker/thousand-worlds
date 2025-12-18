package inventory

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockRepository
type MockRepository struct {
	items map[uuid.UUID][]InventoryItem
}

func (m *MockRepository) AddItem(ctx context.Context, charID uuid.UUID, itemID uuid.UUID, quantity int, metadata map[string]interface{}) error {
	if m.items == nil {
		m.items = make(map[uuid.UUID][]InventoryItem)
	}
	m.items[charID] = append(m.items[charID], InventoryItem{
		CharacterID: charID,
		ItemID:      itemID,
		Quantity:    quantity,
		Metadata:    metadata,
		Name:        metadata["name"].(string), // rudimentary support for test
	})
	return nil
}

func (m *MockRepository) RemoveItem(ctx context.Context, charID uuid.UUID, itemID uuid.UUID, quantity int) error {
	return nil // Mock implementation
}

func (m *MockRepository) GetInventory(ctx context.Context, charID uuid.UUID) ([]InventoryItem, error) {
	return m.items[charID], nil
}

func TestService_AddItem(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := NewService(nil, mockRepo)

	ctx := context.Background()
	charID := uuid.New()
	itemID := uuid.New()
	metadata := map[string]interface{}{"name": "Test Item"}

	err := svc.AddItem(ctx, charID, itemID, 1, metadata)
	assert.NoError(t, err)

	items, _ := mockRepo.GetInventory(ctx, charID)
	assert.Len(t, items, 1)
	assert.Equal(t, itemID, items[0].ItemID)
}

func TestInventoryService_GetInventory(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := NewService(nil, mockRepo)

	ctx := context.Background()
	charID := uuid.New()

	// Pre-populate mock repo
	mockRepo.items = map[uuid.UUID][]InventoryItem{
		charID: {
			{ID: uuid.New(), CharacterID: charID, Name: "Shield"},
		},
	}

	items, err := svc.GetInventory(ctx, charID)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Shield", items[0].Name)
}
