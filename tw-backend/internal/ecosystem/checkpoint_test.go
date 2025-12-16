package ecosystem

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestNewCheckpointManager(t *testing.T) {
	worldID := uuid.New()

	t.Run("creates manager with defaults", func(t *testing.T) {
		cfg := CheckpointManagerConfig{
			WorldID: worldID,
		}

		mgr := NewCheckpointManager(cfg)

		if mgr.worldID != worldID {
			t.Errorf("World ID mismatch: got %v, want %v", mgr.worldID, worldID)
		}

		if mgr.fullCheckpointFreq != 1_000_000 {
			t.Errorf("Default frequency should be 1M years, got %d", mgr.fullCheckpointFreq)
		}
	})

	t.Run("uses custom frequency", func(t *testing.T) {
		cfg := CheckpointManagerConfig{
			WorldID:            worldID,
			FullCheckpointFreq: 500_000,
		}

		mgr := NewCheckpointManager(cfg)

		if mgr.fullCheckpointFreq != 500_000 {
			t.Errorf("Frequency should be 500k years, got %d", mgr.fullCheckpointFreq)
		}
	})
}

func TestCheckpointManager_ShouldCreateFullCheckpoint(t *testing.T) {
	worldID := uuid.New()
	cfg := CheckpointManagerConfig{
		WorldID:            worldID,
		FullCheckpointFreq: 1_000_000,
	}
	mgr := NewCheckpointManager(cfg)

	tests := []struct {
		year     int64
		expected bool
	}{
		{0, false},        // Initial state, lastFullYear is 0
		{500_000, false},  // Not yet at frequency
		{1_000_000, true}, // At frequency
		{1_500_000, true}, // Past frequency
		{2_000_000, true}, // Double frequency
	}

	for _, tt := range tests {
		result := mgr.ShouldCreateFullCheckpoint(tt.year)
		if result != tt.expected {
			t.Errorf("ShouldCreateFullCheckpoint(%d) = %v, want %v", tt.year, result, tt.expected)
		}
	}
}

func TestCheckpointManager_Compression(t *testing.T) {
	worldID := uuid.New()
	cfg := CheckpointManagerConfig{WorldID: worldID}
	mgr := NewCheckpointManager(cfg)

	t.Run("compress and decompress", func(t *testing.T) {
		original := []byte("This is test data for compression. Repeated: AAAAAAAAAAAAAAAA")

		compressed, err := mgr.compress(original)
		if err != nil {
			t.Fatalf("Compression failed: %v", err)
		}

		// Compressed should be smaller (for repetitive data)
		if len(compressed) >= len(original) {
			t.Logf("Warning: compressed size %d >= original size %d (may be expected for small data)", len(compressed), len(original))
		}

		decompressed, err := mgr.decompress(compressed)
		if err != nil {
			t.Fatalf("Decompression failed: %v", err)
		}

		if string(decompressed) != string(original) {
			t.Errorf("Decompressed data doesn't match original")
		}
	})
}

func TestCheckpointManager_Serialization(t *testing.T) {
	worldID := uuid.New()
	speciesID := uuid.New()
	regionID := uuid.New()

	cfg := CheckpointManagerConfig{WorldID: worldID}
	mgr := NewCheckpointManager(cfg)

	t.Run("serialize and deserialize state", func(t *testing.T) {
		original := &SimulationState{
			Version: 1,
			Year:    5_000_000,
			WorldID: worldID,
			Species: []SpeciesState{
				{
					ID:          speciesID,
					Name:        "Test Species",
					OriginYear:  1_000_000,
					GeneticCode: make([]float32, 200),
					IsExtinct:   false,
				},
			},
			Populations: []PopulationState{
				{
					SpeciesID: speciesID,
					RegionID:  regionID,
					Count:     10000,
					Juveniles: 2000,
				},
			},
			ClimateState: &ClimateState{
				GlobalTemperature: 15.0,
				OxygenLevel:       0.21,
				CO2Level:          0.0004,
			},
		}

		// Set some genetic values
		original.Species[0].GeneticCode[0] = 0.5
		original.Species[0].GeneticCode[99] = 0.8

		serialized, err := mgr.serializeState(original)
		if err != nil {
			t.Fatalf("Serialization failed: %v", err)
		}

		restored, err := mgr.deserializeState(serialized)
		if err != nil {
			t.Fatalf("Deserialization failed: %v", err)
		}

		// Verify restored state
		if restored.Year != original.Year {
			t.Errorf("Year mismatch: got %d, want %d", restored.Year, original.Year)
		}

		if len(restored.Species) != 1 {
			t.Errorf("Species count mismatch: got %d, want 1", len(restored.Species))
		}

		if restored.Species[0].Name != "Test Species" {
			t.Errorf("Species name mismatch: got %s", restored.Species[0].Name)
		}

		if restored.Species[0].GeneticCode[0] != 0.5 {
			t.Errorf("Genetic code[0] mismatch: got %f, want 0.5", restored.Species[0].GeneticCode[0])
		}

		if restored.ClimateState.OxygenLevel != 0.21 {
			t.Errorf("OxygenLevel mismatch: got %f, want 0.21", restored.ClimateState.OxygenLevel)
		}
	})
}

func TestCheckpointManager_CreateCheckpoint(t *testing.T) {
	worldID := uuid.New()
	ctx := context.Background()

	cfg := CheckpointManagerConfig{
		WorldID:            worldID,
		FullCheckpointFreq: 1_000_000,
	}
	mgr := NewCheckpointManager(cfg)

	t.Run("creates checkpoint without store", func(t *testing.T) {
		state := &SimulationState{
			Version: 1,
			Year:    1_000_000,
			WorldID: worldID,
			Species: []SpeciesState{
				{
					ID:          uuid.New(),
					Name:        "Test",
					GeneticCode: make([]float32, 200),
				},
			},
			Populations: []PopulationState{
				{Count: 1000, Juveniles: 200},
			},
		}

		checkpoint, err := mgr.CreateCheckpoint(ctx, state)
		if err != nil {
			t.Fatalf("CreateCheckpoint failed: %v", err)
		}

		if checkpoint.Year != 1_000_000 {
			t.Errorf("Year mismatch: got %d", checkpoint.Year)
		}

		if checkpoint.Type != CheckpointFull {
			t.Errorf("Expected full checkpoint, got %s", checkpoint.Type)
		}

		if checkpoint.SpeciesCount != 1 {
			t.Errorf("Species count mismatch: got %d", checkpoint.SpeciesCount)
		}

		if checkpoint.PopulationSum != 1200 {
			t.Errorf("Population sum mismatch: got %d, want 1200", checkpoint.PopulationSum)
		}

		if len(checkpoint.StateData) == 0 {
			t.Error("State data should not be empty")
		}
	})

	t.Run("creates delta after full", func(t *testing.T) {
		// Reset manager
		mgr2 := NewCheckpointManager(cfg)

		// First checkpoint at 1M years - should be full
		state1 := &SimulationState{Year: 1_000_000, WorldID: worldID}
		cp1, _ := mgr2.CreateCheckpoint(ctx, state1)
		if cp1.Type != CheckpointFull {
			t.Errorf("First checkpoint should be full, got %s", cp1.Type)
		}

		// Second checkpoint at 1.5M years - should be delta
		state2 := &SimulationState{Year: 1_500_000, WorldID: worldID}
		cp2, _ := mgr2.CreateCheckpoint(ctx, state2)
		if cp2.Type != CheckpointDelta {
			t.Errorf("Second checkpoint should be delta, got %s", cp2.Type)
		}

		// Third checkpoint at 2M years - should be full (1M after last full)
		state3 := &SimulationState{Year: 2_000_000, WorldID: worldID}
		cp3, _ := mgr2.CreateCheckpoint(ctx, state3)
		if cp3.Type != CheckpointFull {
			t.Errorf("Third checkpoint should be full, got %s", cp3.Type)
		}
	})
}

func TestCheckpointManager_RestoreCheckpoint(t *testing.T) {
	worldID := uuid.New()
	ctx := context.Background()

	cfg := CheckpointManagerConfig{WorldID: worldID}
	mgr := NewCheckpointManager(cfg)

	original := &SimulationState{
		Version: 1,
		Year:    5_000_000,
		WorldID: worldID,
		Species: []SpeciesState{
			{ID: uuid.New(), Name: "RestoreTest", GeneticCode: make([]float32, 200)},
		},
	}

	checkpoint, err := mgr.CreateCheckpoint(ctx, original)
	if err != nil {
		t.Fatalf("CreateCheckpoint failed: %v", err)
	}

	restored, err := mgr.RestoreCheckpoint(ctx, checkpoint)
	if err != nil {
		t.Fatalf("RestoreCheckpoint failed: %v", err)
	}

	if restored.Year != original.Year {
		t.Errorf("Year mismatch: got %d, want %d", restored.Year, original.Year)
	}

	if restored.Species[0].Name != "RestoreTest" {
		t.Errorf("Species name mismatch: got %s", restored.Species[0].Name)
	}
}
