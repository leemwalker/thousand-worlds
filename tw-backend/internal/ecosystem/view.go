// Package ecosystem provides player view synchronization for smooth simulation viewing.
// This enables players to observe simulation progress with interpolated states.
package ecosystem

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// ViewState represents what the player sees at a given moment
type ViewState struct {
	WorldID          uuid.UUID `json:"world_id"`
	Year             int64     `json:"year"`
	InterpolatedYear float64   `json:"interpolated_year"` // For smooth animation

	// Species summary
	TotalSpecies   int `json:"total_species"`
	ExtantSpecies  int `json:"extant_species"`
	ExtinctSpecies int `json:"extinct_species"`
	SapientSpecies int `json:"sapient_species"`

	// Recent activity
	RecentEvents []RunnerEvent `json:"recent_events"`

	// Simulation status
	SimulationState RunnerState `json:"simulation_state"`
	SimulationSpeed int         `json:"simulation_speed"`
	YearsPerSecond  float64     `json:"years_per_second"`

	// Turning point info
	HasPendingTurningPoint bool          `json:"has_pending_turning_point"`
	PendingTurningPoint    *TurningPoint `json:"pending_turning_point,omitempty"`

	// Time info
	RealTimeElapsed time.Duration `json:"real_time_elapsed"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// ViewConfig configures the view synchronization
type ViewConfig struct {
	UpdateInterval  time.Duration `json:"update_interval"`   // How often to push updates
	EventBufferSize int           `json:"event_buffer_size"` // How many events to keep
	InterpolationOn bool          `json:"interpolation_on"`  // Enable smooth year display
}

// DefaultViewConfig returns default view configuration
func DefaultViewConfig() ViewConfig {
	return ViewConfig{
		UpdateInterval:  100 * time.Millisecond, // 10 updates per second
		EventBufferSize: 25,
		InterpolationOn: true,
	}
}

// ViewUpdateHandler is called when the view state changes
type ViewUpdateHandler func(state *ViewState)

// PlayerViewSync manages what the player sees during simulation
type PlayerViewSync struct {
	worldID uuid.UUID
	config  ViewConfig
	runner  *SimulationRunner

	// Current view state
	currentState   *ViewState
	lastSimYear    int64
	lastUpdateTime time.Time

	// Species counts (updated by simulation)
	totalSpecies   int
	extantSpecies  int
	extinctSpecies int
	sapientSpecies int

	// Handlers
	updateHandler ViewUpdateHandler

	// Control
	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
}

// NewPlayerViewSync creates a new view synchronizer
func NewPlayerViewSync(worldID uuid.UUID, runner *SimulationRunner) *PlayerViewSync {
	return &PlayerViewSync{
		worldID:      worldID,
		config:       DefaultViewConfig(),
		runner:       runner,
		currentState: &ViewState{WorldID: worldID},
		stopChan:     make(chan struct{}),
	}
}

// SetConfig sets the view configuration
func (pvs *PlayerViewSync) SetConfig(config ViewConfig) {
	pvs.mu.Lock()
	defer pvs.mu.Unlock()
	pvs.config = config
}

// SetUpdateHandler sets the handler called when view updates
func (pvs *PlayerViewSync) SetUpdateHandler(handler ViewUpdateHandler) {
	pvs.mu.Lock()
	defer pvs.mu.Unlock()
	pvs.updateHandler = handler
}

// UpdateSpeciesCounts updates the species counts (called by simulation)
func (pvs *PlayerViewSync) UpdateSpeciesCounts(total, extant, extinct, sapient int) {
	pvs.mu.Lock()
	defer pvs.mu.Unlock()
	pvs.totalSpecies = total
	pvs.extantSpecies = extant
	pvs.extinctSpecies = extinct
	pvs.sapientSpecies = sapient
}

// Start begins the view update loop
func (pvs *PlayerViewSync) Start() {
	pvs.mu.Lock()
	if pvs.running {
		pvs.mu.Unlock()
		return
	}
	pvs.running = true
	pvs.stopChan = make(chan struct{})
	pvs.mu.Unlock()

	go pvs.updateLoop()
}

// Stop stops the view update loop
func (pvs *PlayerViewSync) Stop() {
	pvs.mu.Lock()
	if !pvs.running {
		pvs.mu.Unlock()
		return
	}
	pvs.running = false
	pvs.mu.Unlock()

	close(pvs.stopChan)
}

// GetCurrentState returns the current view state
func (pvs *PlayerViewSync) GetCurrentState() *ViewState {
	pvs.mu.RLock()
	defer pvs.mu.RUnlock()

	// Return a copy
	state := *pvs.currentState
	return &state
}

// updateLoop is the main view update loop
func (pvs *PlayerViewSync) updateLoop() {
	ticker := time.NewTicker(pvs.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pvs.stopChan:
			return
		case <-ticker.C:
			pvs.update()
		}
	}
}

// update refreshes the view state
func (pvs *PlayerViewSync) update() {
	pvs.mu.Lock()
	defer pvs.mu.Unlock()

	now := time.Now()
	stats := pvs.runner.GetStats()

	// Calculate interpolated year for smooth display
	currentYear := stats.CurrentYear
	interpolatedYear := float64(currentYear)

	if pvs.config.InterpolationOn && pvs.lastUpdateTime.Unix() > 0 {
		// Interpolate between last update and now based on speed
		elapsed := now.Sub(pvs.lastUpdateTime).Seconds()
		yearsPassedRecently := float64(currentYear - pvs.lastSimYear)
		if yearsPassedRecently > 0 {
			// Smooth transition
			progress := elapsed / pvs.config.UpdateInterval.Seconds()
			if progress > 1 {
				progress = 1
			}
			interpolatedYear = float64(pvs.lastSimYear) + yearsPassedRecently*progress
		}
	}

	// Get recent events
	recentEvents := pvs.runner.GetRecentEvents(pvs.config.EventBufferSize)

	// Check for turning point
	tpm := pvs.runner.GetTurningPointManager()
	pendingTP := tpm.GetPendingTurningPoint()

	// Build view state
	pvs.currentState = &ViewState{
		WorldID:                pvs.worldID,
		Year:                   currentYear,
		InterpolatedYear:       interpolatedYear,
		TotalSpecies:           pvs.totalSpecies,
		ExtantSpecies:          pvs.extantSpecies,
		ExtinctSpecies:         pvs.extinctSpecies,
		SapientSpecies:         pvs.sapientSpecies,
		RecentEvents:           recentEvents,
		SimulationState:        stats.State,
		SimulationSpeed:        int(pvs.runner.config.Speed),
		YearsPerSecond:         stats.YearsPerSecond,
		HasPendingTurningPoint: pendingTP != nil,
		PendingTurningPoint:    pendingTP,
		RealTimeElapsed:        stats.RealTimeElapsed,
		LastUpdated:            now,
	}

	pvs.lastSimYear = currentYear
	pvs.lastUpdateTime = now

	// Call update handler (without lock)
	if pvs.updateHandler != nil {
		state := *pvs.currentState
		pvs.mu.Unlock()
		pvs.updateHandler(&state)
		pvs.mu.Lock()
	}
}

// ForceUpdate forces an immediate view update
func (pvs *PlayerViewSync) ForceUpdate() *ViewState {
	pvs.update()
	return pvs.GetCurrentState()
}

// GetSnapshot returns a complete view snapshot for saving/loading
func (pvs *PlayerViewSync) GetSnapshot() *ViewSnapshot {
	pvs.mu.RLock()
	defer pvs.mu.RUnlock()

	return &ViewSnapshot{
		WorldID:        pvs.worldID,
		Year:           pvs.currentState.Year,
		TotalSpecies:   pvs.totalSpecies,
		ExtantSpecies:  pvs.extantSpecies,
		ExtinctSpecies: pvs.extinctSpecies,
		SapientSpecies: pvs.sapientSpecies,
		CreatedAt:      time.Now(),
	}
}

// ViewSnapshot is a serializable snapshot of the view state
type ViewSnapshot struct {
	WorldID        uuid.UUID `json:"world_id"`
	Year           int64     `json:"year"`
	TotalSpecies   int       `json:"total_species"`
	ExtantSpecies  int       `json:"extant_species"`
	ExtinctSpecies int       `json:"extinct_species"`
	SapientSpecies int       `json:"sapient_species"`
	CreatedAt      time.Time `json:"created_at"`
}

// RestoreFromSnapshot restores view state from a snapshot
func (pvs *PlayerViewSync) RestoreFromSnapshot(snapshot *ViewSnapshot) {
	pvs.mu.Lock()
	defer pvs.mu.Unlock()

	pvs.totalSpecies = snapshot.TotalSpecies
	pvs.extantSpecies = snapshot.ExtantSpecies
	pvs.extinctSpecies = snapshot.ExtinctSpecies
	pvs.sapientSpecies = snapshot.SapientSpecies
	pvs.lastSimYear = snapshot.Year
}

// SeekToYear attempts to jump the view to a specific year
// Returns true if successful (snapshot exists near that year)
func (pvs *PlayerViewSync) SeekToYear(targetYear int64) bool {
	snapshot := pvs.runner.GetSnapshotAtYear(targetYear)
	if snapshot == nil {
		return false
	}

	pvs.mu.Lock()
	defer pvs.mu.Unlock()

	pvs.lastSimYear = snapshot.Year
	pvs.currentState.Year = snapshot.Year
	pvs.currentState.InterpolatedYear = float64(snapshot.Year)

	return true
}

// GetAvailableYearRange returns the range of years with snapshots
func (pvs *PlayerViewSync) GetAvailableYearRange() (minYear, maxYear int64) {
	snapshots := pvs.runner.GetSnapshots()
	if len(snapshots) == 0 {
		return 0, 0
	}

	minYear = snapshots[0].Year
	maxYear = snapshots[0].Year

	for _, snap := range snapshots {
		if snap.Year < minYear {
			minYear = snap.Year
		}
		if snap.Year > maxYear {
			maxYear = snap.Year
		}
	}

	return minYear, maxYear
}
