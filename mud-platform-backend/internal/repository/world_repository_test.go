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

func ptr[T any](v T) *T {
	return &v
}

func TestWorldRepository_CreateAndGet(t *testing.T) {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewPostgresWorldRepository(pool)

	// Create a sphere world
	sphereWorld := &World{
		ID:     uuid.New(),
		Name:   "Test Planet",
		Shape:  WorldShapeSphere,
		Radius: ptr(1000.0),
		Metadata: map[string]interface{}{
			"description": "A test planet",
		},
	}

	err = repo.CreateWorld(ctx, sphereWorld)
	require.NoError(t, err)

	// Get the world back
	retrieved, err := repo.GetWorld(ctx, sphereWorld.ID)
	require.NoError(t, err)

	assert.Equal(t, sphereWorld.ID, retrieved.ID)
	assert.Equal(t, sphereWorld.Name, retrieved.Name)
	assert.Equal(t, sphereWorld.Shape, retrieved.Shape)
	assert.NotNil(t, retrieved.Radius)
	assert.InDelta(t, *sphereWorld.Radius, *retrieved.Radius, 0.01)
}

func TestWorldRepository_CreateCubeWorld(t *testing.T) {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewPostgresWorldRepository(pool)

	// Create a cube world
	cubeWorld := &World{
		ID:        uuid.New(),
		Name:      "Test Building",
		Shape:     WorldShapeCube,
		BoundsMin: &Vector3{X: 0, Y: 0, Z: 0},
		BoundsMax: &Vector3{X: 100, Y: 100, Z: 20},
		Metadata:  map[string]interface{}{},
	}

	err = repo.CreateWorld(ctx, cubeWorld)
	require.NoError(t, err)

	// Get the world back
	retrieved, err := repo.GetWorld(ctx, cubeWorld.ID)
	require.NoError(t, err)

	assert.Equal(t, WorldShapeCube, retrieved.Shape)
	assert.Nil(t, retrieved.Radius)
	assert.NotNil(t, retrieved.BoundsMin)
	assert.NotNil(t, retrieved.BoundsMax)
	assert.InDelta(t, 100.0, retrieved.BoundsMax.X, 0.01)
}

func TestWorldRepository_ListWorlds(t *testing.T) {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewPostgresWorldRepository(pool)

	worlds, err := repo.ListWorlds(ctx)
	require.NoError(t, err)

	// Should have at least our seed worlds
	assert.GreaterOrEqual(t, len(worlds), 3)

	// Check that seed worlds exist
	foundEarth := false
	for _, w := range worlds {
		if w.Name == "Earth" {
			foundEarth = true
			assert.Equal(t, WorldShapeSphere, w.Shape)
			break
		}
	}
	assert.True(t, foundEarth, "Should find Earth world")
}

func TestWorldRepository_UpdateWorld(t *testing.T) {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewPostgresWorldRepository(pool)

	// Create a world
	world := &World{
		ID:       uuid.New(),
		Name:     "Original Name",
		Shape:    WorldShapeSphere,
		Radius:   ptr(500.0),
		Metadata: map[string]interface{}{},
	}

	err = repo.CreateWorld(ctx, world)
	require.NoError(t, err)

	// Update it
	world.Name = "Updated Name"
	world.Radius = ptr(600.0)

	err = repo.UpdateWorld(ctx, world)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetWorld(ctx, world.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.InDelta(t, 600.0, *retrieved.Radius, 0.01)
}

func TestWorldRepository_DeleteWorld(t *testing.T) {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewPostgresWorldRepository(pool)

	// Create a world
	world := &World{
		ID:       uuid.New(),
		Name:     "To Delete",
		Shape:    WorldShapeSphere,
		Radius:   ptr(100.0),
		Metadata: map[string]interface{}{},
	}

	err = repo.CreateWorld(ctx, world)
	require.NoError(t, err)

	// Delete it
	err = repo.DeleteWorld(ctx, world.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = repo.GetWorld(ctx, world.ID)
	assert.Error(t, err) // Should error because it doesn't exist
}
