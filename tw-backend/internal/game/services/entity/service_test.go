package entity

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityService_AddEntity(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	ent := &Entity{
		Name: "Test Entity",
		Type: EntityTypeItem,
	}

	err := svc.AddEntity(ctx, ent)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, ent.ID)

	// Verify it's there
	assert.Equal(t, ent, svc.entities[ent.ID])
}

func TestEntityService_RemoveEntity(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	ent := &Entity{Name: "To Delete"}
	_ = svc.AddEntity(ctx, ent)

	err := svc.RemoveEntity(ctx, ent.ID)
	assert.NoError(t, err)

	_, exists := svc.entities[ent.ID]
	assert.False(t, exists)

	// Remove again
	err = svc.RemoveEntity(ctx, ent.ID)
	assert.Error(t, err)
}

func TestEntityService_GetEntitiesAt(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	worldID := uuid.New()

	// Add nearby entity
	ent1 := &Entity{
		Name:    "Nearby",
		WorldID: worldID,
		X:       10, Y: 10,
	}
	_ = svc.AddEntity(ctx, ent1)

	// Add distant entity
	ent2 := &Entity{
		Name:    "Far",
		WorldID: worldID,
		X:       100, Y: 100,
	}
	_ = svc.AddEntity(ctx, ent2)

	// Add different world entity
	ent3 := &Entity{
		Name:    "Other World",
		WorldID: uuid.New(),
		X:       10, Y: 10,
	}
	_ = svc.AddEntity(ctx, ent3)

	// Search at 10,10 radius 20
	res, err := svc.GetEntitiesAt(ctx, worldID, 10, 10, 20)
	require.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "Nearby", res[0].Name)
}
