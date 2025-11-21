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

	worldID := uuid.New()
	entityID := uuid.New()
	x, y, z := 10.5, 20.1, 5.0

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

	worldID := uuid.New()
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

	worldID := uuid.New()
	centerLon, centerLat := -97.7431, 30.2672

	// Create center entity
	centerID := uuid.New()
	err := repo.CreateEntity(ctx, worldID, centerID, centerLon, centerLat, 0)
	require.NoError(t, err)

	// Create nearby entity (~5m away: 0.00005 degrees ≈ 5.5m at this latitude)
	nearbyID := uuid.New()
	err = repo.CreateEntity(ctx, worldID, nearbyID, centerLon+0.00005, centerLat, 0)
	require.NoError(t, err)

	// Create far entity (~200m away)
	farID := uuid.New()
	err = repo.CreateEntity(ctx, worldID, farID, centerLon+0.002, centerLat, 0)
	require.NoError(t, err)

	// Search within 10m radius
	entities, err := repo.GetEntitiesNearby(ctx, worldID, centerLon, centerLat, 0, 10)
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

	worldID := uuid.New()
	e1ID := uuid.New()
	e2ID := uuid.New()

	lon1, lat1 := -97.7431, 30.2672
	// Move ~100m east: 0.001 degrees ≈ 111m at this latitude
	lon2, lat2 := -97.7421, 30.2672

	err := repo.CreateEntity(ctx, worldID, e1ID, lon1, lat1, 0)
	require.NoError(t, err)

	err = repo.CreateEntity(ctx, worldID, e2ID, lon2, lat2, 0)
	require.NoError(t, err)

	dist, err := repo.CalculateDistance(ctx, e1ID, e2ID)
	require.NoError(t, err)
	// Distance should be ~100m (allowing some tolerance for spherical calculation)
	assert.InDelta(t, 100.0, dist, 15.0, "Distance should be approximately 100m")
}

func TestSpatialRepository_GetEntitiesInBounds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresSpatialRepository(db)
	ctx := context.Background()

	worldID := uuid.New()

	// Create entities in different locations around Austin, TX
	entity1 := uuid.New()
	err := repo.CreateEntity(ctx, worldID, entity1, -97.7431, 30.2672, 0) // Center
	require.NoError(t, err)

	entity2 := uuid.New()
	err = repo.CreateEntity(ctx, worldID, entity2, -97.7400, 30.2700, 0) // Inside bounds
	require.NoError(t, err)

	entity3 := uuid.New()
	err = repo.CreateEntity(ctx, worldID, entity3, -97.8000, 30.3000, 0) // Outside bounds
	require.NoError(t, err)

	// Query bounding box
	minLon, minLat := -97.7500, 30.2600
	maxLon, maxLat := -97.7300, 30.2800
	entities, err := repo.GetEntitiesInBounds(ctx, worldID, minLon, minLat, maxLon, maxLat)
	require.NoError(t, err)

	// Should find entity1 and entity2, but not entity3
	require.Len(t, entities, 2, "Should find 2 entities in bounds")
}
