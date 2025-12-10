package service

import (
	"context"
	"testing"

	"tw-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSpatialRepository is a mock implementation of repository.SpatialRepository
type MockSpatialRepository struct {
	mock.Mock
}

func (m *MockSpatialRepository) CreateEntity(ctx context.Context, worldID, entityID uuid.UUID, x, y, z float64) error {
	args := m.Called(ctx, worldID, entityID, x, y, z)
	return args.Error(0)
}

func (m *MockSpatialRepository) UpdateEntityLocation(ctx context.Context, entityID uuid.UUID, x, y, z float64) error {
	args := m.Called(ctx, entityID, x, y, z)
	return args.Error(0)
}

func (m *MockSpatialRepository) GetEntity(ctx context.Context, entityID uuid.UUID) (*repository.Entity, error) {
	args := m.Called(ctx, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Entity), args.Error(1)
}

func (m *MockSpatialRepository) GetEntitiesNearby(ctx context.Context, worldID uuid.UUID, x, y, z, radius float64) ([]repository.Entity, error) {
	args := m.Called(ctx, worldID, x, y, z, radius)
	return args.Get(0).([]repository.Entity), args.Error(1)
}

func (m *MockSpatialRepository) GetEntitiesInBounds(ctx context.Context, worldID uuid.UUID, minX, minY, maxX, maxY float64) ([]repository.Entity, error) {
	args := m.Called(ctx, worldID, minX, minY, maxX, maxY)
	return args.Get(0).([]repository.Entity), args.Error(1)
}

func (m *MockSpatialRepository) CalculateDistance(ctx context.Context, entity1ID, entity2ID uuid.UUID) (float64, error) {
	args := m.Called(ctx, entity1ID, entity2ID)
	return args.Get(0).(float64), args.Error(1)
}

func TestUpdateLocation(t *testing.T) {
	mockRepo := new(MockSpatialRepository)
	service := NewSpatialService(mockRepo)

	ctx := context.Background()
	entityID := uuid.New()
	x, y, z := 10.0, 20.0, 30.0

	// Expectation
	mockRepo.On("UpdateEntityLocation", ctx, entityID, x, y, z).Return(nil)

	// Execute
	err := service.UpdateLocation(ctx, entityID, x, y, z)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateLocation_Error(t *testing.T) {
	mockRepo := new(MockSpatialRepository)
	service := NewSpatialService(mockRepo)

	ctx := context.Background()
	entityID := uuid.New()
	x, y, z := 10.0, 20.0, 30.0

	// Expectation
	mockRepo.On("UpdateEntityLocation", ctx, entityID, x, y, z).Return(assert.AnError)

	// Execute
	err := service.UpdateLocation(ctx, entityID, x, y, z)

	// Assert
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
