// Package ecosystem provides the background simulation runner.
// This manages asynchronous world simulation with snapshot-based state updates.
package ecosystem

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"tw-backend/internal/ecosystem/pathogen"
	"tw-backend/internal/ecosystem/population"
	"tw-backend/internal/ecosystem/sapience"

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

// EventBroadcastHandler is called when an important event happens (for Watchers)
type EventBroadcastHandler func(event RunnerEvent)

// SimulationRunner manages background world simulation
type SimulationRunner struct {
	config           SimulationConfig
	state            RunnerState
	currentYear      int64
	lastSnapshotYear int64
	lastSaveYear     int64

	// Core V2 Simulation Engine
	popSim *population.PopulationSimulator

	// Subsystem Integrations (Phase 4)
	diseaseSystem    *pathogen.DiseaseSystem
	sapienceDetector *sapience.SapienceDetector
	geology          *WorldGeology // Uses existing WorldGeology from this package
	snapshotRepo     *SimulationSnapshotRepository
	stateRepo        *RunnerStateRepository

	// Handlers
	tickHandler           TickHandler
	snapshotHandler       SnapshotHandler
	turningPointHandler   TurningPointHandler
	eventBroadcastHandler EventBroadcastHandler

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
func NewSimulationRunner(config SimulationConfig, snapshotRepo *SimulationSnapshotRepository, stateRepo *RunnerStateRepository) *SimulationRunner {
	ctx, cancel := context.WithCancel(context.Background())

	return &SimulationRunner{
		config:              config,
		state:               RunnerIdle,
		ctx:                 ctx,
		cancel:              cancel,
		snapshotRepo:        snapshotRepo,
		stateRepo:           stateRepo,
		turningPointManager: NewTurningPointManager(config.WorldID),
		recentEvents:        make([]RunnerEvent, 0),
		snapshots:           make([]*Snapshot, 0),
	}
}

// InitializePopulationSimulator sets up the internal V2 engine
// MUST be called before Start()
func (sr *SimulationRunner) InitializePopulationSimulator(seed int64) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	// If we have a repo, try to load existing state
	if sr.snapshotRepo != nil {
		fmt.Printf("Attempting to load snapshot for world %s\n", sr.config.WorldID)
		sim, err := sr.snapshotRepo.LoadSnapshot(sr.ctx, sr.config.WorldID)
		if err == nil && sim != nil {
			fmt.Printf("Loaded existing simulation state for world %s (Year %d)\n", sr.config.WorldID, sim.CurrentYear)
			// Re-initialize non-serialized systems
			sim.InitializeGeographicSystems(sr.config.WorldID, seed)
			sr.popSim = sim
			sr.currentYear = sim.CurrentYear
			// Initialize subsystems (not persisted separately)
			sr.initializeSubsystems(seed)
			return
		} else if err != nil {
			fmt.Printf("Error loading snapshot: %v\n", err)
		}
	}

	// Create fresh if no saved state
	fmt.Printf("Creating fresh population simulator for world %s\n", sr.config.WorldID)
	sr.popSim = population.NewPopulationSimulator(sr.config.WorldID, seed)
	sr.popSim.InitializeGeographicSystems(sr.config.WorldID, seed)
	sr.currentYear = 0

	// Initialize subsystems
	sr.initializeSubsystems(seed)
}

// initializeSubsystems sets up disease, sapience, and geology systems
func (sr *SimulationRunner) initializeSubsystems(seed int64) {
	// Initialize Disease System
	sr.diseaseSystem = pathogen.NewDiseaseSystem(sr.config.WorldID, seed)

	// Initialize Sapience Detector (magic disabled by default)
	sr.sapienceDetector = sapience.NewSapienceDetector(sr.config.WorldID, false)

	// Geology uses existing WorldGeology from this package
	// (typically initialized separately or via worldgen)
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

// SetEventBroadcastHandler sets the handler for broadcasting events to watchers
func (sr *SimulationRunner) SetEventBroadcastHandler(handler EventBroadcastHandler) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.eventBroadcastHandler = handler
}

// Start begins the simulation
func (sr *SimulationRunner) Start(startYear int64) error {
	sr.mu.Lock()
	if sr.state == RunnerRunning {
		sr.mu.Unlock()
		return nil // Already running
	}

	if sr.popSim == nil {
		sr.mu.Unlock()
		return fmt.Errorf("population simulator not initialized")
	}

	sr.state = RunnerRunning
	// If starting fresher than current state, use that, otherwise continue from where we are
	if startYear > sr.currentYear {
		sr.currentYear = startYear
	}
	sr.lastSnapshotYear = sr.currentYear
	sr.lastSaveYear = sr.currentYear
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

	// Final save
	sr.persistState()

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
	// Save on pause
	go sr.persistState()
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

// GetSpeed returns the current simulation speed
func (sr *SimulationRunner) GetSpeed() SimulationSpeed {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.config.Speed
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

// UpdateConfig updates the simulation configuration
func (sr *SimulationRunner) UpdateConfig(config SimulationConfig) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.config = config
	return nil
}

// Step advances the simulation by a specific number of ticks manually
// Useful for testing and deterministic execution
func (sr *SimulationRunner) Step(ticks int) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	// Use current speed, or default to Normal if paused/zero
	speed := sr.config.Speed
	if speed == SpeedPaused {
		speed = SpeedNormal
	}

	for i := 0; i < ticks; i++ {
		if err := sr.tickLocked(int64(speed)); err != nil {
			return err
		}
	}
	return nil
}

// runLoop is the main simulation loop
func (sr *SimulationRunner) runLoop() {
	defer sr.wg.Done()

	// Panic Recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic in runLoop: %v\n", r)
			sr.mu.Lock()
			sr.state = RunnerError
			sr.mu.Unlock()
		}
	}()

	ticker := time.NewTicker(sr.config.TickInterval)
	defer ticker.Stop()

	// Auto-save ticker (every 30 seconds)
	saveTicker := time.NewTicker(30 * time.Second)
	defer saveTicker.Stop()

	for {
		select {
		case <-sr.ctx.Done():
			return
		case <-saveTicker.C:
			sr.persistState()
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
				fmt.Printf("Simulation error: %v\n", err)
				sr.mu.Lock()
				sr.state = RunnerError
				sr.mu.Unlock()
				return
			}
		}
	}
}

// tick advances the simulation by the specified years (thread-safe wrapper)
func (sr *SimulationRunner) tick(yearsToAdvance int64) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.tickLocked(yearsToAdvance)
}

// tickLocked performs the actual simulation step (assumes lock held)
func (sr *SimulationRunner) tickLocked(yearsToAdvance int64) error {
	// Run V2 Simulation Step(s)
	// We run years one by one to ensure proper granularity of events
	for i := int64(0); i < yearsToAdvance; i++ {
		sr.popSim.SimulateYear()

		// Periodic Evolution (every 1000 years)
		if sr.popSim.CurrentYear%1000 == 0 {
			sr.popSim.ApplyEvolution()
			sr.popSim.ApplyCoEvolution()
			sr.popSim.ApplyGeneticDrift()
			sr.popSim.ApplySexualSelection()

			// Sapience Detection (every 1000 years)
			sr.updateSapienceDetection()
		}

		// Speciation & Migration (every 10000 years)
		if sr.popSim.CurrentYear%10000 == 0 {
			sr.popSim.UpdateOxygenLevel()
			sr.popSim.ApplyOxygenEffects()
			if newSpecies := sr.popSim.CheckSpeciation(); newSpecies > 0 {
				sr.broadcastEvent(RunnerEvent{
					Year:        sr.popSim.CurrentYear,
					Type:        "speciation",
					Description: fmt.Sprintf("%d new species evolved", newSpecies),
					Importance:  7,
				})
			}
			sr.popSim.ApplyMigrationCycle()

			// Disease System Update (every 10000 years)
			sr.updateDiseaseSystem()
		}

		// Geology Updates (every 100000 years)
		if sr.popSim.CurrentYear%100000 == 0 {
			sr.updateGeology(100000)
		}

		// Check for events logged by the simulator this year
		if len(sr.popSim.Events) > 0 {
			for _, evtMsg := range sr.popSim.Events {
				sr.broadcastEvent(RunnerEvent{
					Year:        sr.popSim.CurrentYear,
					Type:        "sim_event",
					Description: evtMsg,
					Importance:  5,
				})
			}
		}
	}

	// Update local state
	sr.currentYear = sr.popSim.CurrentYear
	sr.yearsSimulated += yearsToAdvance
	sr.tickCount++
	sr.lastTickTime = time.Now()

	// External Tick Handler (optional, for legacy hooks)
	if sr.tickHandler != nil {
		if err := sr.tickHandler(sr.currentYear, yearsToAdvance); err != nil {
			return err
		}
	}

	// Check for snapshot
	if sr.currentYear-sr.lastSnapshotYear >= sr.config.SnapshotInterval {
		sr.createSnapshot()
	}

	// Accumulate Divine Energy over time (Turning Point system)
	if sr.turningPointManager != nil {
		sr.turningPointManager.AccumulateEnergy(sr.currentYear)
	}

	// Check for turning points
	// Use slightly randomized check frequency to avoid performance spikes
	if sr.turningPointManager != nil && sr.currentYear%100000 == 0 && sr.currentYear > 0 {
		// Use stats from PopSim
		pop, species, extinct := sr.popSim.GetStats()
		_ = pop // unused

		tp := sr.turningPointManager.CheckForTurningPoint(
			sr.currentYear,
			int(species),
			int(extinct), // TODO: track recent extinctions only?
			nil,          // newSapientSpecies
			"",           // significantEvent
		)
		if tp != nil && sr.config.PauseOnTurning {
			sr.state = RunnerPaused
			sr.broadcastEvent(RunnerEvent{
				Year:        sr.currentYear,
				Type:        "turning_point",
				Description: fmt.Sprintf("Turning Point Reached: %s", tp.Title),
				Importance:  10,
			})
			// Call turning point handler
			if sr.turningPointHandler != nil {
				sr.mu.Unlock()
				_ = sr.turningPointHandler(tp)
				sr.mu.Lock()
			}
			// Force save on turning point
			go sr.persistState()
		}
	}

	// Check for max year
	if sr.config.MaxYearTarget > 0 && sr.currentYear >= sr.config.MaxYearTarget {
		sr.state = RunnerPaused
		sr.broadcastEvent(RunnerEvent{
			Year:        sr.currentYear,
			Type:        "system",
			Description: "Simulation reached target year",
			Importance:  1,
		})
	}

	return nil
}

// persistState saves both the runner status and the simulation blob
func (sr *SimulationRunner) persistState() {
	if sr.stateRepo != nil {
		_ = sr.SaveState(sr.stateRepo) // Save lightweight state
	}
	if sr.snapshotRepo != nil && sr.popSim != nil {
		sr.mu.RLock()
		// Make a copy or handle concurrency?
		// For now we lock during save, which might pause sim briefly.
		// Optimized: Save logic handles serialization.
		err := sr.snapshotRepo.SaveSnapshot(context.Background(), sr.config.WorldID, sr.popSim)
		sr.mu.RUnlock()
		if err != nil {
			fmt.Printf("Failed to persist simulation state: %v\n", err)
		} else {
			// fmt.Printf("Persisted state at year %d\n", sr.currentYear)
		}
	}
}

// broadcastEvent sends an event to the handler
func (sr *SimulationRunner) broadcastEvent(event RunnerEvent) {
	sr.AddEvent(event)
	if sr.eventBroadcastHandler != nil {
		// Non-blocking send ideally
		sr.eventBroadcastHandler(event)
	}
}

// createSnapshot creates a new snapshot
func (sr *SimulationRunner) createSnapshot() {
	// Simple metadata snapshot for now
	// If popSim is not initialized, abort
	if sr.popSim == nil {
		return
	}

	pop, species, _ := sr.popSim.GetStats() // Get stats from V2 sim
	_ = pop

	snapshot := &Snapshot{
		WorldID:       sr.config.WorldID,
		Year:          sr.currentYear,
		CreatedAt:     time.Now(),
		TotalSpecies:  int(species),
		ExtantSpecies: int(species), // distinct from total if tracking history
		SapientCount:  0,            // TODO: Track sapients
	}

	sr.snapshots = append(sr.snapshots, snapshot)
	sr.lastSnapshotYear = sr.currentYear

	// Call snapshot handler (without lock)
	if sr.snapshotHandler != nil {
		sr.mu.Unlock()
		sr.snapshotHandler(snapshot)
		sr.mu.Lock()
	}
}

// AddEvent records a simulation event
func (sr *SimulationRunner) AddEvent(event RunnerEvent) {
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

// Helper to randomly pick an element (used by tests mostly)
func pickRandom[T any](rng *rand.Rand, slice []T) T {
	return slice[rng.Intn(len(slice))]
}

// =============================================================================
// Phase 4: Subsystem Integration Methods
// =============================================================================

// updateDiseaseSystem simulates pathogen dynamics and outbreaks
func (sr *SimulationRunner) updateDiseaseSystem() {
	if sr.diseaseSystem == nil || sr.popSim == nil {
		return
	}

	// Build species info map for disease system
	speciesInfo := make(map[uuid.UUID]pathogen.SpeciesInfo)
	for biomeID, biome := range sr.popSim.Biomes {
		for speciesID, species := range biome.Species {
			// Aggregate species info across biomes
			info, exists := speciesInfo[speciesID]
			if !exists {
				info = pathogen.SpeciesInfo{
					Population:        species.Count,
					DiseaseResistance: float32(species.Traits.DiseaseResistance),
					DietType:          string(species.Diet),
					Density:           0.5, // Default density
				}
			} else {
				info.Population += species.Count
			}
			speciesInfo[speciesID] = info
			_ = biomeID // used in loop
		}
	}

	// Update disease system
	sr.diseaseSystem.Update(sr.popSim.CurrentYear, speciesInfo)

	// Check for pandemics and broadcast events
	pandemics := sr.diseaseSystem.GetPandemics()
	for _, outbreak := range pandemics {
		sr.broadcastEvent(RunnerEvent{
			Year:        sr.popSim.CurrentYear,
			Type:        "pandemic",
			Description: fmt.Sprintf("Pandemic outbreak affecting species"),
			SpeciesID:   &outbreak.SpeciesID,
			Importance:  9,
		})
	}
}

// updateSapienceDetection checks for emerging sapience in species
func (sr *SimulationRunner) updateSapienceDetection() {
	if sr.sapienceDetector == nil || sr.popSim == nil {
		return
	}

	// Evaluate all species for sapience potential
	for _, biome := range sr.popSim.Biomes {
		for speciesID, species := range biome.Species {
			// Convert population traits to sapience traits (scale 0-10)
			traits := sapience.SpeciesTraits{
				Intelligence:  species.Traits.Intelligence * 10, // Convert 0-1 to 0-10
				Social:        species.Traits.Social * 10,
				ToolUse:       species.Traits.Intelligence * 5, // Approximate tool use from intelligence
				Communication: species.Traits.Social * 8,       // Approximate from social
				MagicAffinity: 0,                               // No magic by default
				Population:    species.Count,
				Generation:    species.Generation,
			}

			candidate := sr.sapienceDetector.Evaluate(
				speciesID,
				species.Name,
				traits,
				sr.popSim.CurrentYear,
			)

			// Broadcast sapience events
			if candidate != nil && candidate.Level == sapience.SapienceSapient {
				sr.broadcastEvent(RunnerEvent{
					Year:        sr.popSim.CurrentYear,
					Type:        "sapience",
					Description: fmt.Sprintf("Species '%s' has achieved sapience!", species.Name),
					SpeciesID:   &speciesID,
					Importance:  10,
				})
			} else if candidate != nil && candidate.Level == sapience.SapienceProtoSapient {
				sr.broadcastEvent(RunnerEvent{
					Year:        sr.popSim.CurrentYear,
					Type:        "proto_sapience",
					Description: fmt.Sprintf("Species '%s' shows proto-sapient behavior", species.Name),
					SpeciesID:   &speciesID,
					Importance:  7,
				})
			}
		}
	}
}

// updateGeology simulates geological processes over time
func (sr *SimulationRunner) updateGeology(yearsElapsed int64) {
	if sr.geology == nil {
		return
	}

	// Simulate geological processes
	sr.geology.SimulateGeology(yearsElapsed, 0.0) // No temperature modifier

	// Check for significant geological events
	// (The WorldGeology system handles internal events)
}

// SetGeology allows external injection of geology system
func (sr *SimulationRunner) SetGeology(geology *WorldGeology) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.geology = geology
}

// GetDiseaseSystem returns the disease system for external access
func (sr *SimulationRunner) GetDiseaseSystem() *pathogen.DiseaseSystem {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.diseaseSystem
}

// GetSapienceDetector returns the sapience detector for external access
func (sr *SimulationRunner) GetSapienceDetector() *sapience.SapienceDetector {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.sapienceDetector
}
