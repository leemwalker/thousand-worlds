// Package ecosystem provides the background simulation runner.
// This manages asynchronous world simulation with snapshot-based state updates.
package ecosystem

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RunnerState represents the current state of the simulation runner
type RunnerState string

const (
	RunnerIdle     RunnerState = "idle"     // Not running
	RunnerRunning  RunnerState = "running"  // Actively simulating
	RunnerPaused   RunnerState = "paused"   // Paused for turning point
	RunnerStopping RunnerState = "stopping" // Shutdown in progress
	RunnerError    RunnerState = "error"    // Error occurred
)

// SimulationSpeed controls how fast the simulation runs
type SimulationSpeed int

const (
	SpeedPaused SimulationSpeed = 0
	SpeedSlow   SimulationSpeed = 1    // 1 year per tick
	SpeedNormal SimulationSpeed = 10   // 10 years per tick
	SpeedFast   SimulationSpeed = 100  // 100 years per tick
	SpeedTurbo  SimulationSpeed = 1000 // 1000 years per tick
)

// SimulationConfig holds configuration for the runner
type SimulationConfig struct {
	WorldID          uuid.UUID       `json:"world_id"`
	TickInterval     time.Duration   `json:"tick_interval"`     // Real-world time between ticks
	SnapshotInterval int64           `json:"snapshot_interval"` // Simulation years between snapshots
	Speed            SimulationSpeed `json:"speed"`
	MaxYearTarget    int64           `json:"max_year_target"`  // Stop at this year (0 = no limit)
	PauseOnTurning   bool            `json:"pause_on_turning"` // Pause when turning point triggers
}

// DefaultConfig returns a default simulation configuration
func DefaultConfig(worldID uuid.UUID) SimulationConfig {
	return SimulationConfig{
		WorldID:          worldID,
		TickInterval:     100 * time.Millisecond, // 10 ticks per second
		SnapshotInterval: 10,                     // Snapshot every 10 years
		Speed:            SpeedNormal,
		MaxYearTarget:    0,
		PauseOnTurning:   true,
	}
}

// Snapshot represents a point-in-time simulation state
type Snapshot struct {
	WorldID       uuid.UUID `json:"world_id"`
	Year          int64     `json:"year"`
	CreatedAt     time.Time `json:"created_at"`
	TotalSpecies  int       `json:"total_species"`
	ExtantSpecies int       `json:"extant_species"`
	SapientCount  int       `json:"sapient_count"`
	EventsSummary string    `json:"events_summary"`
	Checksum      string    `json:"checksum"` // For validation
}

// RunnerEvent represents something notable that happened during simulation
type RunnerEvent struct {
	Year        int64      `json:"year"`
	Type        string     `json:"type"` // "speciation", "extinction", "sapience", etc.
	Description string     `json:"description"`
	SpeciesID   *uuid.UUID `json:"species_id,omitempty"`
	Importance  int        `json:"importance"` // 1-10
}

// TickHandler is called for each simulation tick
type TickHandler func(year int64, yearsElapsed int64) error

// SnapshotHandler is called when a snapshot is created
type SnapshotHandler func(snapshot *Snapshot) error

// TurningPointHandler is called when a turning point occurs
type TurningPointHandler func(tp *TurningPoint) error

// SimulationRunner manages background world simulation
type SimulationRunner struct {
	config           SimulationConfig
	state            RunnerState
	currentYear      int64
	lastSnapshotYear int64

	// Handlers
	tickHandler         TickHandler
	snapshotHandler     SnapshotHandler
	turningPointHandler TurningPointHandler

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	wg     sync.WaitGroup

	// State managers
	turningPointManager *TurningPointManager

	// History
	recentEvents []RunnerEvent
	snapshots    []*Snapshot

	// Stats
	tickCount      int64
	yearsSimulated int64
	startTime      time.Time
	lastTickTime   time.Time
}

// NewSimulationRunner creates a new simulation runner
func NewSimulationRunner(config SimulationConfig) *SimulationRunner {
	ctx, cancel := context.WithCancel(context.Background())

	return &SimulationRunner{
		config:              config,
		state:               RunnerIdle,
		ctx:                 ctx,
		cancel:              cancel,
		turningPointManager: NewTurningPointManager(config.WorldID),
		recentEvents:        make([]RunnerEvent, 0),
		snapshots:           make([]*Snapshot, 0),
	}
}

// SetTickHandler sets the handler called for each tick
func (sr *SimulationRunner) SetTickHandler(handler TickHandler) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.tickHandler = handler
}

// SetSnapshotHandler sets the handler called for snapshots
func (sr *SimulationRunner) SetSnapshotHandler(handler SnapshotHandler) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.snapshotHandler = handler
}

// SetTurningPointHandler sets the handler called for turning points
func (sr *SimulationRunner) SetTurningPointHandler(handler TurningPointHandler) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.turningPointHandler = handler
}

// Start begins the background simulation
func (sr *SimulationRunner) Start(startYear int64) error {
	sr.mu.Lock()
	if sr.state == RunnerRunning {
		sr.mu.Unlock()
		return nil // Already running
	}

	sr.state = RunnerRunning
	sr.currentYear = startYear
	sr.lastSnapshotYear = startYear
	sr.startTime = time.Now()
	sr.lastTickTime = time.Now()
	sr.mu.Unlock()

	sr.wg.Add(1)
	go sr.runLoop()

	return nil
}

// Stop gracefully stops the simulation
func (sr *SimulationRunner) Stop() {
	sr.mu.Lock()
	if sr.state != RunnerRunning && sr.state != RunnerPaused {
		sr.mu.Unlock()
		return
	}
	sr.state = RunnerStopping
	sr.mu.Unlock()

	sr.cancel()
	sr.wg.Wait()

	sr.mu.Lock()
	sr.state = RunnerIdle
	sr.mu.Unlock()
}

// Pause pauses the simulation (can resume)
func (sr *SimulationRunner) Pause() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if sr.state == RunnerRunning {
		sr.state = RunnerPaused
	}
}

// Resume resumes a paused simulation
func (sr *SimulationRunner) Resume() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if sr.state == RunnerPaused {
		sr.state = RunnerRunning
	}
}

// SetSpeed changes the simulation speed
func (sr *SimulationRunner) SetSpeed(speed SimulationSpeed) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.config.Speed = speed
}

// GetState returns the current simulation state
func (sr *SimulationRunner) GetState() RunnerState {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.state
}

// GetCurrentYear returns the current simulation year
func (sr *SimulationRunner) GetCurrentYear() int64 {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.currentYear
}

// GetStats returns simulation statistics
func (sr *SimulationRunner) GetStats() SimulationStats {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	elapsed := time.Since(sr.startTime)
	yearsPerSecond := 0.0
	if elapsed.Seconds() > 0 {
		yearsPerSecond = float64(sr.yearsSimulated) / elapsed.Seconds()
	}

	return SimulationStats{
		State:           sr.state,
		CurrentYear:     sr.currentYear,
		TickCount:       sr.tickCount,
		YearsSimulated:  sr.yearsSimulated,
		RealTimeElapsed: elapsed,
		YearsPerSecond:  yearsPerSecond,
		SnapshotCount:   len(sr.snapshots),
	}
}

// SimulationStats contains runtime statistics
type SimulationStats struct {
	State           RunnerState   `json:"state"`
	CurrentYear     int64         `json:"current_year"`
	TickCount       int64         `json:"tick_count"`
	YearsSimulated  int64         `json:"years_simulated"`
	RealTimeElapsed time.Duration `json:"real_time_elapsed"`
	YearsPerSecond  float64       `json:"years_per_second"`
	SnapshotCount   int           `json:"snapshot_count"`
}

// runLoop is the main simulation loop
func (sr *SimulationRunner) runLoop() {
	defer sr.wg.Done()

	ticker := time.NewTicker(sr.config.TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sr.ctx.Done():
			return
		case <-ticker.C:
			sr.mu.RLock()
			state := sr.state
			speed := sr.config.Speed
			sr.mu.RUnlock()

			// Skip if paused or stopping
			if state == RunnerPaused || state == RunnerStopping {
				continue
			}
			if speed == SpeedPaused {
				continue
			}

			// Perform tick
			if err := sr.tick(int64(speed)); err != nil {
				sr.mu.Lock()
				sr.state = RunnerError
				sr.mu.Unlock()
				return
			}
		}
	}
}

// tick advances the simulation by the specified years
func (sr *SimulationRunner) tick(yearsToAdvance int64) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	// Call tick handler
	if sr.tickHandler != nil {
		if err := sr.tickHandler(sr.currentYear, yearsToAdvance); err != nil {
			return err
		}
	}

	// Advance time
	sr.currentYear += yearsToAdvance
	sr.yearsSimulated += yearsToAdvance
	sr.tickCount++
	sr.lastTickTime = time.Now()

	// Check for snapshot
	if sr.currentYear-sr.lastSnapshotYear >= sr.config.SnapshotInterval {
		sr.createSnapshot()
	}

	// Accumulate Divine Energy over time
	if sr.turningPointManager != nil {
		sr.turningPointManager.AccumulateEnergy(sr.currentYear)
	}

	// Check for turning points every 100,000 years
	if sr.turningPointManager != nil && sr.currentYear%100000 == 0 && sr.currentYear > 0 {
		// Get relevant stats for turning point check (simplified for now)
		tp := sr.turningPointManager.CheckForTurningPoint(
			sr.currentYear,
			0,   // totalSpecies - would need to be passed in via tick handler
			0,   // recentExtinctions
			nil, // newSapientSpecies
			"",  // significantEvent
		)
		if tp != nil && sr.config.PauseOnTurning {
			sr.state = RunnerPaused
			// Call turning point handler (unlocked to allow resolution)
			if sr.turningPointHandler != nil {
				sr.mu.Unlock()
				_ = sr.turningPointHandler(tp)
				sr.mu.Lock()
			}
		}
	}

	// Check for max year
	if sr.config.MaxYearTarget > 0 && sr.currentYear >= sr.config.MaxYearTarget {
		sr.state = RunnerPaused
	}

	return nil
}

// createSnapshot creates a new snapshot
func (sr *SimulationRunner) createSnapshot() {
	snapshot := &Snapshot{
		WorldID:   sr.config.WorldID,
		Year:      sr.currentYear,
		CreatedAt: time.Now(),
	}

	sr.snapshots = append(sr.snapshots, snapshot)
	sr.lastSnapshotYear = sr.currentYear

	// Call snapshot handler (without lock)
	if sr.snapshotHandler != nil {
		// Unlock temporarily for handler
		sr.mu.Unlock()
		sr.snapshotHandler(snapshot)
		sr.mu.Lock()
	}
}

// AddEvent records a simulation event
func (sr *SimulationRunner) AddEvent(event RunnerEvent) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.recentEvents = append(sr.recentEvents, event)

	// Keep only recent events (last 100)
	if len(sr.recentEvents) > 100 {
		sr.recentEvents = sr.recentEvents[len(sr.recentEvents)-100:]
	}
}

// GetRecentEvents returns recent simulation events
func (sr *SimulationRunner) GetRecentEvents(count int) []RunnerEvent {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	if count > len(sr.recentEvents) {
		count = len(sr.recentEvents)
	}

	// Return most recent first
	result := make([]RunnerEvent, count)
	for i := 0; i < count; i++ {
		result[i] = sr.recentEvents[len(sr.recentEvents)-1-i]
	}

	return result
}

// GetSnapshots returns all snapshots
func (sr *SimulationRunner) GetSnapshots() []*Snapshot {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	result := make([]*Snapshot, len(sr.snapshots))
	copy(result, sr.snapshots)
	return result
}

// GetSnapshotAtYear returns the snapshot closest to the given year
func (sr *SimulationRunner) GetSnapshotAtYear(year int64) *Snapshot {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var closest *Snapshot
	minDiff := int64(^uint64(0) >> 1) // Max int64

	for _, snap := range sr.snapshots {
		diff := year - snap.Year
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			closest = snap
		}
	}

	return closest
}

// TriggerTurningPoint manually triggers a turning point
func (sr *SimulationRunner) TriggerTurningPoint(title, description string) *TurningPoint {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	tp := &TurningPoint{
		ID:          uuid.New(),
		WorldID:     sr.config.WorldID,
		Year:        sr.currentYear,
		Trigger:     TriggerPlayerRequest,
		Title:       title,
		Description: description,
	}

	sr.turningPointManager.TurningPoints[tp.ID] = tp
	sr.turningPointManager.PendingTurningPoint = &tp.ID

	if sr.config.PauseOnTurning {
		sr.state = RunnerPaused
	}

	// Call turning point handler
	if sr.turningPointHandler != nil {
		sr.mu.Unlock()
		sr.turningPointHandler(tp)
		sr.mu.Lock()
	}

	return tp
}

// GetTurningPointManager returns the turning point manager
func (sr *SimulationRunner) GetTurningPointManager() *TurningPointManager {
	return sr.turningPointManager
}
