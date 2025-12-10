package worldentity

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, entity *WorldEntity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*WorldEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WorldEntity), args.Error(1)
}

func (m *MockRepository) GetByWorldID(ctx context.Context, worldID uuid.UUID) ([]*WorldEntity, error) {
	args := m.Called(ctx, worldID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*WorldEntity), args.Error(1)
}

func (m *MockRepository) GetByWorldAndType(ctx context.Context, worldID uuid.UUID, entityType EntityType) ([]*WorldEntity, error) {
	args := m.Called(ctx, worldID, entityType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*WorldEntity), args.Error(1)
}

func (m *MockRepository) GetAtPosition(ctx context.Context, worldID uuid.UUID, x, y, radius float64) ([]*WorldEntity, error) {
	args := m.Called(ctx, worldID, x, y, radius)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*WorldEntity), args.Error(1)
}

func (m *MockRepository) GetByName(ctx context.Context, worldID uuid.UUID, name string) (*WorldEntity, error) {
	args := m.Called(ctx, worldID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WorldEntity), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, entity *WorldEntity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Test CheckCollision with blocking entity
func TestCheckCollision_BlockedByEntity(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	worldID := uuid.New()
	ctx := context.Background()

	// Statue at (5, 5) with collision radius 0.8
	statue := &WorldEntity{
		ID:        uuid.New(),
		WorldID:   worldID,
		Name:      "statue",
		X:         5.0,
		Y:         5.0,
		Collision: true,
		Metadata:  map[string]interface{}{"collision_radius": 0.8},
	}

	mockRepo.On("GetByWorldID", ctx, worldID).Return([]*WorldEntity{statue}, nil)

	// Attempt to move to (5, 5) - directly on statue
	blocked, entity, err := service.CheckCollision(ctx, worldID, 5.0, 5.0)

	assert.NoError(t, err)
	assert.True(t, blocked)
	assert.Equal(t, "statue", entity.Name)
}

// Test CheckCollision with no blocking entity
func TestCheckCollision_NoCollision(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	worldID := uuid.New()
	ctx := context.Background()

	// Statue at (5, 5) with collision radius 0.8
	statue := &WorldEntity{
		ID:        uuid.New(),
		WorldID:   worldID,
		Name:      "statue",
		X:         5.0,
		Y:         5.0,
		Collision: true,
		Metadata:  map[string]interface{}{"collision_radius": 0.8},
	}

	mockRepo.On("GetByWorldID", ctx, worldID).Return([]*WorldEntity{statue}, nil)

	// Attempt to move to (3, 3) - far from statue
	blocked, entity, err := service.CheckCollision(ctx, worldID, 3.0, 3.0)

	assert.NoError(t, err)
	assert.False(t, blocked)
	assert.Nil(t, entity)
}

// Test CheckCollision with non-collidable entity
func TestCheckCollision_NonCollidableEntity(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	worldID := uuid.New()
	ctx := context.Background()

	// Item at (5, 5) without collision
	item := &WorldEntity{
		ID:        uuid.New(),
		WorldID:   worldID,
		Name:      "coin",
		X:         5.0,
		Y:         5.0,
		Collision: false,
	}

	mockRepo.On("GetByWorldID", ctx, worldID).Return([]*WorldEntity{item}, nil)

	// Attempt to move to (5, 5) - should pass through
	blocked, entity, err := service.CheckCollision(ctx, worldID, 5.0, 5.0)

	assert.NoError(t, err)
	assert.False(t, blocked)
	assert.Nil(t, entity)
}

// Test CanInteract with locked entity
func TestCanInteract_LockedEntity(t *testing.T) {
	service := NewService(nil) // No repo needed for this test

	statue := &WorldEntity{
		ID:         uuid.New(),
		Name:       "statue",
		EntityType: EntityTypeStatic,
		Locked:     true,
	}

	tests := []struct {
		action      string
		expectAllow bool
		expectMsg   string
	}{
		{"get", false, "You cannot move the statue."},
		{"take", false, "You cannot move the statue."},
		{"push", false, "You cannot move the statue."},
		{"move", false, "You cannot move the statue."},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			allowed, msg := service.CanInteract(statue, tt.action)
			assert.Equal(t, tt.expectAllow, allowed)
			assert.Equal(t, tt.expectMsg, msg)
		})
	}
}

// Test CanInteract with unlocked non-item entity
func TestCanInteract_UnlockedNonItem(t *testing.T) {
	service := NewService(nil)

	plant := &WorldEntity{
		ID:         uuid.New(),
		Name:       "tree",
		EntityType: EntityTypePlant,
		Locked:     false,
	}

	// Can't pick up a plant
	allowed, msg := service.CanInteract(plant, "get")
	assert.False(t, allowed)
	assert.Equal(t, "You cannot pick up the tree.", msg)
}

// Test CanInteract with unlocked item
func TestCanInteract_UnlockedItem(t *testing.T) {
	service := NewService(nil)

	item := &WorldEntity{
		ID:         uuid.New(),
		Name:       "gem",
		EntityType: EntityTypeItem,
		Locked:     false,
	}

	allowed, msg := service.CanInteract(item, "get")
	assert.True(t, allowed)
	assert.Empty(t, msg)
}

// Test GetEntitiesAt filters correctly
func TestGetEntitiesAt(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	worldID := uuid.New()
	ctx := context.Background()

	entities := []*WorldEntity{
		{ID: uuid.New(), WorldID: worldID, Name: "statue", X: 5.0, Y: 5.0},
		{ID: uuid.New(), WorldID: worldID, Name: "fountain", X: 8.0, Y: 8.0},
		{ID: uuid.New(), WorldID: worldID, Name: "bench", X: 2.0, Y: 2.0},
	}

	mockRepo.On("GetByWorldID", ctx, worldID).Return(entities, nil)

	// Search from (5, 4) with radius 2 - should only find statue
	result, err := service.GetEntitiesAt(ctx, worldID, 5.0, 4.0, 2.0)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "statue", result[0].Name)
}

// Test CollisionRadius with custom metadata
func TestCollisionRadius_CustomMetadata(t *testing.T) {
	entity := &WorldEntity{
		Metadata: map[string]interface{}{"collision_radius": 1.5},
	}
	assert.Equal(t, 1.5, entity.CollisionRadius())
}

// Test CollisionRadius default
func TestCollisionRadius_Default(t *testing.T) {
	entity := &WorldEntity{}
	assert.Equal(t, 0.5, entity.CollisionRadius())
}

// Test GetGlyph with custom metadata
func TestGetGlyph_CustomMetadata(t *testing.T) {
	entity := &WorldEntity{
		Metadata: map[string]interface{}{"glyph": "🗿"},
	}
	assert.Equal(t, "🗿", entity.GetGlyph())
}

// Test GetGlyph defaults by type
func TestGetGlyph_DefaultsByType(t *testing.T) {
	tests := []struct {
		entityType EntityType
		expected   string
	}{
		{EntityTypeStatic, "◼"},
		{EntityTypeNPC, "👤"},
		{EntityTypeCreature, "🐾"},
		{EntityTypeItem, "📦"},
		{EntityTypePlant, "🌿"},
		{EntityTypeStructure, "🏠"},
		{EntityTypeResource, "⛏"},
	}

	for _, tt := range tests {
		t.Run(string(tt.entityType), func(t *testing.T) {
			entity := &WorldEntity{EntityType: tt.entityType}
			assert.Equal(t, tt.expected, entity.GetGlyph())
		})
	}
}

// Test caching behavior
func TestCaching_LoadsOnceFromDB(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	worldID := uuid.New()
	ctx := context.Background()

	entities := []*WorldEntity{
		{ID: uuid.New(), WorldID: worldID, Name: "statue"},
	}

	// Expect only one DB call
	mockRepo.On("GetByWorldID", ctx, worldID).Return(entities, nil).Once()

	// First call loads from DB
	result1, err := service.GetEntitiesInWorld(ctx, worldID)
	assert.NoError(t, err)
	assert.Len(t, result1, 1)

	// Second call uses cache
	result2, err := service.GetEntitiesInWorld(ctx, worldID)
	assert.NoError(t, err)
	assert.Len(t, result2, 1)

	mockRepo.AssertExpectations(t)
}

// Test cache invalidation
func TestCaching_InvalidatesOnCreate(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	worldID := uuid.New()
	ctx := context.Background()

	entities := []*WorldEntity{
		{ID: uuid.New(), WorldID: worldID, Name: "statue"},
	}

	mockRepo.On("GetByWorldID", ctx, worldID).Return(entities, nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)

	// Load cache
	_, _ = service.GetEntitiesInWorld(ctx, worldID)

	// Create new entity - should invalidate cache
	newEntity := &WorldEntity{WorldID: worldID, Name: "fountain"}
	_ = service.Create(ctx, newEntity)

	// Next call should hit DB again
	_, _ = service.GetEntitiesInWorld(ctx, worldID)

	// Should have been called twice
	mockRepo.AssertNumberOfCalls(t, "GetByWorldID", 2)
}
