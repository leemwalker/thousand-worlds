package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)

	// Clean up entities table
	_, err = pool.Exec(ctx, "TRUNCATE TABLE entities")
	require.NoError(t, err)

	return pool
}

func TestSpatialRepository_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresSpatialRepository(db)
	ctx := context.Background()

	// Use test continent world
	worldID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	entityID := uuid.New()
	x, y, z := 100.5, 200.1, 5.0 // meters in test continent

	err := repo.CreateEntity(ctx, worldID, entityID, x, y, z)
	require.NoError(t, err)

	entity, err := repo.GetEntity(ctx, entityID)
	require.NoError(t, err)
	assert.Equal(t, entityID, entity.ID)
	assert.Equal(t, worldID, entity.WorldID)
	assert.InDelta(t, x, entity.X, 0.1)
	assert.InDelta(t, y, entity.Y, 0.1)
	assert.InDelta(t, z, entity.Z, 0.1)
}

func TestSpatialRepository_UpdateLocation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresSpatialRepository(db)
	ctx := context.Background()

	worldID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	entityID := uuid.New()
	x, y, z := 10.0, 10.0, 0.0

	err := repo.CreateEntity(ctx, worldID, entityID, x, y, z)
	require.NoError(t, err)

	newX, newY, newZ := 15.0, 15.0, 1.0
	err = repo.UpdateEntityLocation(ctx, entityID, newX, newY, newZ)
	require.NoError(t, err)

	entity, err := repo.GetEntity(ctx, entityID)
	require.NoError(t, err)
	assert.InDelta(t, newX, entity.X, 0.1)
	assert.InDelta(t, newY, entity.Y, 0.1)
	assert.InDelta(t, newZ, entity.Z, 0.1)
}

func TestSpatialRepository_GetEntitiesNearby(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresSpatialRepository(db)
	ctx := context.Background()

	worldID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	// Create center entity
	centerID := uuid.New()
	err := repo.CreateEntity(ctx, worldID, centerID, 0, 0, 0)
	require.NoError(t, err)

	// Create nearby entity (~5m away in Cartesian space)
	nearbyID := uuid.New()
	err = repo.CreateEntity(ctx, worldID, nearbyID, 5, 0, 0)
	require.NoError(t, err)

	// Create far entity (~200m away)
	farID := uuid.New()
	err = repo.CreateEntity(ctx, worldID, farID, 200, 0, 0)
	require.NoError(t, err)

	// Search within 10m radius
	entities, err := repo.GetEntitiesNearby(ctx, worldID, 0, 0, 0, 10)
	require.NoError(t, err)

	// Should find center and nearby, but not far
	require.Len(t, entities, 2, "Should find 2 entities within 10m")

	foundNearby := false
	for _, e := range entities {
		if e.ID == nearbyID {
			foundNearby = true
			break
		}
	}
	assert.True(t, foundNearby, "Should find nearby entity")
}

func TestSpatialRepository_CalculateDistance(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresSpatialRepository(db)
	ctx := context.Background()

	worldID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	e1ID := uuid.New()
	e2ID := uuid.New()

	// 3-4-5 triangle in meters
	err := repo.CreateEntity(ctx, worldID, e1ID, 0, 0, 0)
	require.NoError(t, err)

	err = repo.CreateEntity(ctx, worldID, e2ID, 3, 4, 0)
	require.NoError(t, err)

	dist, err := repo.CalculateDistance(ctx, e1ID, e2ID)
	require.NoError(t, err)
	// Should be exactly 5m (3-4-5 triangle)
	assert.InDelta(t, 5.0, dist, 0.1)
}

func TestSpatialRepository_GetEntitiesInBounds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresSpatialRepository(db)
	ctx := context.Background()

	worldID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	// Create entities in different locations
	entity1 := uuid.New()
	err := repo.CreateEntity(ctx, worldID, entity1, 10, 10, 0) // Inside
	require.NoError(t, err)

	entity2 := uuid.New()
	err = repo.CreateEntity(ctx, worldID, entity2, 15, 15, 0) // Inside
	require.NoError(t, err)

	entity3 := uuid.New()
	err = repo.CreateEntity(ctx, worldID, entity3, 100, 100, 0) // Outside
	require.NoError(t, err)

	// Query bounding box (0,0) to (20,20)
	entities, err := repo.GetEntitiesInBounds(ctx, worldID, 0, 0, 20, 20)
	require.NoError(t, err)

	// Should find entity1 and entity2, but not entity3
	require.Len(t, entities, 2, "Should find 2 entities in bounds")
}
