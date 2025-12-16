// Package ecosystem provides checkpoint functionality for world simulation state.
// Checkpoints enable rewinding the simulation to previous states and debugging.
package ecosystem

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CheckpointType indicates whether this is a full snapshot or delta
type CheckpointType string

const (
	CheckpointFull  CheckpointType = "full"  // Complete world state
	CheckpointDelta CheckpointType = "delta" // Changes since last full checkpoint
)

// Checkpoint represents a saved simulation state
type Checkpoint struct {
	ID            uuid.UUID      `json:"id"`
	WorldID       uuid.UUID      `json:"world_id"`
	Year          int64          `json:"year"`
	Type          CheckpointType `json:"checkpoint_type"`
	StateData     []byte         `json:"-"` // Compressed state, not included in JSON
	SpeciesCount  int            `json:"species_count"`
	PopulationSum int64          `json:"population_sum"`
	CreatedAt     time.Time      `json:"created_at"`
}

// CheckpointStore interface for checkpoint persistence
type CheckpointStore interface {
	Save(ctx context.Context, checkpoint *Checkpoint) error
	Load(ctx context.Context, worldID uuid.UUID, year int64) (*Checkpoint, error)
	LoadLatest(ctx context.Context, worldID uuid.UUID) (*Checkpoint, error)
	LoadNearestBefore(ctx context.Context, worldID uuid.UUID, year int64) (*Checkpoint, error)
	List(ctx context.Context, worldID uuid.UUID) ([]Checkpoint, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// CheckpointManager handles creating and restoring simulation checkpoints
type CheckpointManager struct {
	worldID            uuid.UUID
	store              CheckpointStore
	logger             *SimulationLogger
	fullCheckpointFreq int64 // Years between full checkpoints (default 1M)
	lastFullYear       int64
	mu                 sync.Mutex
}

// CheckpointManagerConfig holds configuration for checkpoint manager
type CheckpointManagerConfig struct {
	WorldID            uuid.UUID
	Store              CheckpointStore
	Logger             *SimulationLogger
	FullCheckpointFreq int64 // Default: 1,000,000 years
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(cfg CheckpointManagerConfig) *CheckpointManager {
	freq := cfg.FullCheckpointFreq
	if freq == 0 {
		freq = 1_000_000 // 1 million years default
	}
	return &CheckpointManager{
		worldID:            cfg.WorldID,
		store:              cfg.Store,
		logger:             cfg.Logger,
		fullCheckpointFreq: freq,
		lastFullYear:       0,
	}
}

// SimulationState represents the serializable simulation state
// This should be updated as the simulation evolves
type SimulationState struct {
	Version       int               `json:"version"` // For migration compatibility
	Year          int64             `json:"year"`
	WorldID       uuid.UUID         `json:"world_id"`
	Species       []SpeciesState    `json:"species"`
	Populations   []PopulationState `json:"populations"`
	Pathogens     []PathogenState   `json:"pathogens"`
	TectonicState *TectonicState    `json:"tectonic_state,omitempty"`
	ClimateState  *ClimateState     `json:"climate_state,omitempty"`
}

// SpeciesState represents a serializable species
type SpeciesState struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	AncestorID   *uuid.UUID `json:"ancestor_id,omitempty"`
	OriginYear   int64      `json:"origin_year"`
	GeneticCode  []float32  `json:"genetic_code"` // 200 genes
	ActiveBlanks []int      `json:"active_blanks"`
	IsExtinct    bool       `json:"is_extinct"`
	ExtinctYear  int64      `json:"extinct_year,omitempty"`
}

// PopulationState represents a population in a specific region
type PopulationState struct {
	SpeciesID   uuid.UUID        `json:"species_id"`
	RegionID    uuid.UUID        `json:"region_id"`
	Count       int64            `json:"count"`
	Juveniles   int64            `json:"juveniles"`
	LastContact map[string]int64 `json:"last_contact,omitempty"` // Species ID -> year
}

// PathogenState represents a serializable pathogen
type PathogenState struct {
	ID               uuid.UUID   `json:"id"`
	Name             string      `json:"name"`
	Type             string      `json:"type"` // virus, bacteria, fungi, prion
	Virulence        float32     `json:"virulence"`
	Transmissibility float32     `json:"transmissibility"`
	HostIDs          []uuid.UUID `json:"host_ids"`
	Status           string      `json:"status"` // transient, endemic, dormant
}

// TectonicState represents the tectonic plate configuration
type TectonicState struct {
	Fragmentation float32 `json:"fragmentation"`
	// Will be expanded with hex grid data
}

// ClimateState represents the climate configuration
type ClimateState struct {
	GlobalTemperature float32 `json:"global_temperature"`
	OxygenLevel       float32 `json:"oxygen_level"`
	CO2Level          float32 `json:"co2_level"`
}

// ShouldCreateFullCheckpoint returns true if a full checkpoint is due
func (cm *CheckpointManager) ShouldCreateFullCheckpoint(year int64) bool {
	return year-cm.lastFullYear >= cm.fullCheckpointFreq
}

// CreateCheckpoint creates a checkpoint from the current simulation state
func (cm *CheckpointManager) CreateCheckpoint(ctx context.Context, state *SimulationState) (*Checkpoint, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	checkpointType := CheckpointDelta
	if cm.ShouldCreateFullCheckpoint(state.Year) {
		checkpointType = CheckpointFull
		cm.lastFullYear = state.Year
	}

	// Serialize state
	stateData, err := cm.serializeState(state)
	if err != nil {
		return nil, err
	}

	// Compress state data
	compressedData, err := cm.compress(stateData)
	if err != nil {
		return nil, err
	}

	checkpoint := &Checkpoint{
		ID:            uuid.New(),
		WorldID:       cm.worldID,
		Year:          state.Year,
		Type:          checkpointType,
		StateData:     compressedData,
		SpeciesCount:  len(state.Species),
		PopulationSum: cm.sumPopulations(state),
		CreatedAt:     time.Now(),
	}

	// Save to store
	if cm.store != nil {
		if err := cm.store.Save(ctx, checkpoint); err != nil {
			return nil, err
		}
	}

	// Log checkpoint
	if cm.logger != nil {
		cm.logger.LogCheckpoint(ctx, state.Year, string(checkpointType), checkpoint.SpeciesCount, checkpoint.PopulationSum)
	}

	return checkpoint, nil
}

// RestoreCheckpoint loads and deserializes a checkpoint
func (cm *CheckpointManager) RestoreCheckpoint(ctx context.Context, checkpoint *Checkpoint) (*SimulationState, error) {
	// Decompress state data
	stateData, err := cm.decompress(checkpoint.StateData)
	if err != nil {
		return nil, err
	}

	// Deserialize state
	state, err := cm.deserializeState(stateData)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// RestoreToYear loads the nearest checkpoint before or at the given year
func (cm *CheckpointManager) RestoreToYear(ctx context.Context, year int64) (*SimulationState, error) {
	if cm.store == nil {
		return nil, nil
	}

	checkpoint, err := cm.store.LoadNearestBefore(ctx, cm.worldID, year)
	if err != nil {
		return nil, err
	}
	if checkpoint == nil {
		return nil, nil
	}

	return cm.RestoreCheckpoint(ctx, checkpoint)
}

// serializeState serializes the simulation state using gob
func (cm *CheckpointManager) serializeState(state *SimulationState) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(state); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// deserializeState deserializes the simulation state from gob
func (cm *CheckpointManager) deserializeState(data []byte) (*SimulationState, error) {
	var state SimulationState
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&state); err != nil {
		return nil, err
	}
	return &state, nil
}

// compress compresses data using gzip
func (cm *CheckpointManager) compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decompress decompresses gzip data
func (cm *CheckpointManager) decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

// sumPopulations calculates total population across all populations
func (cm *CheckpointManager) sumPopulations(state *SimulationState) int64 {
	var total int64
	for _, pop := range state.Populations {
		total += pop.Count + pop.Juveniles
	}
	return total
}

// GetCheckpointYears returns a list of years with available checkpoints
func (cm *CheckpointManager) GetCheckpointYears(ctx context.Context) ([]int64, error) {
	if cm.store == nil {
		return nil, nil
	}

	checkpoints, err := cm.store.List(ctx, cm.worldID)
	if err != nil {
		return nil, err
	}

	years := make([]int64, len(checkpoints))
	for i, cp := range checkpoints {
		years[i] = cp.Year
	}
	return years, nil
}
