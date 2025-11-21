package world

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_RegisterWorld(t *testing.T) {
	registry := NewRegistry()

	world := &WorldState{
		ID:             uuid.New(),
		Name:           "Test World",
		Status:         StatusPaused,
		DilationFactor: 1.0,
		CreatedAt:      time.Now(),
	}

	err := registry.RegisterWorld(world)
	require.NoError(t, err, "Should register world successfully")

	retrieved, err := registry.GetWorld(world.ID)
	require.NoError(t, err)
	assert.Equal(t, world.ID, retrieved.ID)
	assert.Equal(t, world.Name, retrieved.Name)
}

func TestRegistry_RegisterWorld_Duplicate(t *testing.T) {
	registry := NewRegistry()

	worldID := uuid.New()
	world := &WorldState{
		ID:        worldID,
		Name:      "Test World",
		Status:    StatusPaused,
		CreatedAt: time.Now(),
	}

	err := registry.RegisterWorld(world)
	require.NoError(t, err)

	// Try to register again
	duplicate := &WorldState{
		ID:        worldID,
		Name:      "Duplicate",
		Status:    StatusPaused,
		CreatedAt: time.Now(),
	}

	err = registry.RegisterWorld(duplicate)
	assert.Error(t, err, "Should return error for duplicate world ID")
}

func TestRegistry_GetWorld_NotFound(t *testing.T) {
	registry := NewRegistry()

	nonExistentID := uuid.New()
	_, err := registry.GetWorld(nonExistentID)
	assert.Error(t, err, "Should return error for non-existent world")
}

func TestRegistry_UpdateWorld(t *testing.T) {
	registry := NewRegistry()

	world := &WorldState{
		ID:        uuid.New(),
		Name:      "Test World",
		Status:    StatusPaused,
		TickCount: 0,
		CreatedAt: time.Now(),
	}

	err := registry.RegisterWorld(world)
	require.NoError(t, err)

	// Update the world
	err = registry.UpdateWorld(world.ID, func(w *WorldState) {
		w.TickCount = 10
		w.Status = StatusRunning
		w.GameTime = 1 * time.Second
	})
	require.NoError(t, err)

	// Verify update
	updated, err := registry.GetWorld(world.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(10), updated.TickCount)
	assert.Equal(t, StatusRunning, updated.Status)
	assert.Equal(t, 1*time.Second, updated.GameTime)
}

func TestRegistry_UpdateWorld_NotFound(t *testing.T) {
	registry := NewRegistry()

	nonExistentID := uuid.New()
	err := registry.UpdateWorld(nonExistentID, func(w *WorldState) {
		w.TickCount = 10
	})
	assert.Error(t, err, "Should return error when updating non-existent world")
}

func TestRegistry_ListWorlds(t *testing.T) {
	registry := NewRegistry()

	world1 := &WorldState{
		ID:        uuid.New(),
		Name:      "World 1",
		Status:    StatusPaused,
		CreatedAt: time.Now(),
	}
	world2 := &WorldState{
		ID:        uuid.New(),
		Name:      "World 2",
		Status:    StatusRunning,
		CreatedAt: time.Now(),
	}

	err := registry.RegisterWorld(world1)
	require.NoError(t, err)
	err = registry.RegisterWorld(world2)
	require.NoError(t, err)

	worlds := registry.ListWorlds()
	assert.Len(t, worlds, 2, "Should return all registered worlds")

	// Check both worlds are in the list
	worldIDs := make(map[uuid.UUID]bool)
	for _, w := range worlds {
		worldIDs[w.ID] = true
	}
	assert.True(t, worldIDs[world1.ID])
	assert.True(t, worldIDs[world2.ID])
}

func TestRegistry_RemoveWorld(t *testing.T) {
	registry := NewRegistry()

	world := &WorldState{
		ID:        uuid.New(),
		Name:      "Test World",
		Status:    StatusPaused,
		CreatedAt: time.Now(),
	}

	err := registry.RegisterWorld(world)
	require.NoError(t, err)

	err = registry.RemoveWorld(world.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = registry.GetWorld(world.ID)
	assert.Error(t, err, "Should not find removed world")
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	const numGoroutines = 100
	const numWorlds = 10

	// Create some initial worlds
	worldIDs := make([]uuid.UUID, numWorlds)
	for i := 0; i < numWorlds; i++ {
		worldIDs[i] = uuid.New()
		world := &WorldState{
			ID:        worldIDs[i],
			Name:      fmt.Sprintf("World %d", i),
			Status:    StatusPaused,
			TickCount: 0,
			CreatedAt: time.Now(),
		}
		err := registry.RegisterWorld(world)
		require.NoError(t, err)
	}

	// Concurrently read, update, and list
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Randomly read, update, or list
			switch idx % 3 {
			case 0:
				// Read
				worldID := worldIDs[idx%numWorlds]
				_, err := registry.GetWorld(worldID)
				if err != nil {
					errors <- err
				}
			case 1:
				// Update
				worldID := worldIDs[idx%numWorlds]
				err := registry.UpdateWorld(worldID, func(w *WorldState) {
					w.TickCount++
				})
				if err != nil {
					errors <- err
				}
			case 2:
				// List
				_ = registry.ListWorlds()
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	// Verify all worlds still exist
	worlds := registry.ListWorlds()
	assert.Len(t, worlds, numWorlds, "All worlds should still exist after concurrent access")
}
