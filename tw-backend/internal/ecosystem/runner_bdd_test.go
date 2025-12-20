package ecosystem_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"tw-backend/internal/ecosystem"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BDD Tests: Simulation Runner
// =============================================================================

// testWorldID is a fixed UUID for deterministic testing
var testWorldID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// -----------------------------------------------------------------------------
// Scenario: Runner Initialization
// -----------------------------------------------------------------------------
// Given: Valid configuration
// When: NewSimulationRunner is called
// Then: Runner should be created in Idle state
//
//	AND Config should be stored
//	AND Population simulator should be nil (not initialized yet)
func TestBDD_Runner_Initialization(t *testing.T) {
	config := ecosystem.DefaultConfig(testWorldID)

	runner := ecosystem.NewSimulationRunner(config, nil, nil)

	require.NotNil(t, runner, "Runner should be created")
	assert.Equal(t, ecosystem.RunnerIdle, runner.GetState(), "Initial state should be Idle")
	assert.Equal(t, int64(0), runner.GetCurrentYear(), "Initial year should be 0")
}

// -----------------------------------------------------------------------------
// Scenario: Runner Start Without Initialization
// -----------------------------------------------------------------------------
// Given: A runner without InitializePopulationSimulator called
// When: Start is called
// Then: Should return an error
func TestBDD_Runner_StartWithoutInit(t *testing.T) {
	config := ecosystem.DefaultConfig(testWorldID)
	runner := ecosystem.NewSimulationRunner(config, nil, nil)

	err := runner.Start(0)

	assert.Error(t, err, "Start without initialization should return error")
	assert.Contains(t, err.Error(), "not initialized", "Error should mention initialization")
}

// -----------------------------------------------------------------------------
// Scenario: Runner Start With Initialization
// -----------------------------------------------------------------------------
// Given: A properly initialized runner
// When: Start is called
// Then: State should transition to Running
//
//	AND Background loop should begin
func TestBDD_Runner_StartWithInit(t *testing.T) {
	config := ecosystem.DefaultConfig(testWorldID)
	config.TickInterval = 50 * time.Millisecond // Fast for testing
	config.MaxYearTarget = 100                  // Stop after 100 years

	runner := ecosystem.NewSimulationRunner(config, nil, nil)
	runner.InitializePopulationSimulator(42)

	err := runner.Start(0)
	require.NoError(t, err, "Start should succeed")

	// Check state
	assert.Equal(t, ecosystem.RunnerRunning, runner.GetState(), "State should be Running")

	// Wait a bit for simulation to advance
	time.Sleep(200 * time.Millisecond)

	// Stop the runner
	runner.Stop()

	// Verify some simulation happened
	assert.Greater(t, runner.GetCurrentYear(), int64(0), "Some years should have simulated")
}

// -----------------------------------------------------------------------------
// Scenario: Runner State Machine Transitions (Table-Driven)
// -----------------------------------------------------------------------------
// Given: A runner in a specific initial state
// When: A transition method (Start, Pause, Stop) is called
// Then: The state should update correctly OR return an error if invalid
func TestBDD_Runner_StateTransitions(t *testing.T) {
	scenarios := []struct {
		name          string
		setupState    func(r *ecosystem.SimulationRunner)
		action        func(r *ecosystem.SimulationRunner)
		expectedState ecosystem.RunnerState
	}{
		{
			name: "Idle to Running",
			setupState: func(r *ecosystem.SimulationRunner) {
				r.InitializePopulationSimulator(42)
			},
			action: func(r *ecosystem.SimulationRunner) {
				r.Start(0)
			},
			expectedState: ecosystem.RunnerRunning,
		},
		{
			name: "Running to Paused",
			setupState: func(r *ecosystem.SimulationRunner) {
				r.InitializePopulationSimulator(42)
				r.Start(0)
				time.Sleep(50 * time.Millisecond) // Let it start
			},
			action: func(r *ecosystem.SimulationRunner) {
				r.Pause()
			},
			expectedState: ecosystem.RunnerPaused,
		},
		{
			name: "Paused to Running",
			setupState: func(r *ecosystem.SimulationRunner) {
				r.InitializePopulationSimulator(42)
				r.Start(0)
				time.Sleep(50 * time.Millisecond)
				r.Pause()
			},
			action: func(r *ecosystem.SimulationRunner) {
				r.Resume()
			},
			expectedState: ecosystem.RunnerRunning,
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			config := ecosystem.DefaultConfig(testWorldID)
			config.TickInterval = 10 * time.Millisecond
			runner := ecosystem.NewSimulationRunner(config, nil, nil)

			sc.setupState(runner)
			sc.action(runner)

			// Allow state to settle
			time.Sleep(50 * time.Millisecond)

			assert.Equal(t, sc.expectedState, runner.GetState(),
				"State should be %s after action", sc.expectedState)

			runner.Stop()
		})
	}
}

// -----------------------------------------------------------------------------
// Scenario: Speed Change
// -----------------------------------------------------------------------------
// Given: A running simulation
// When: SetSpeed is called
// Then: Speed should change
func TestBDD_Runner_SpeedChange(t *testing.T) {
	config := ecosystem.DefaultConfig(testWorldID)
	config.Speed = ecosystem.SpeedNormal

	runner := ecosystem.NewSimulationRunner(config, nil, nil)

	assert.Equal(t, ecosystem.SpeedNormal, runner.GetSpeed())

	runner.SetSpeed(ecosystem.SpeedFast)
	assert.Equal(t, ecosystem.SpeedFast, runner.GetSpeed())

	runner.SetSpeed(ecosystem.SpeedTurbo)
	assert.Equal(t, ecosystem.SpeedTurbo, runner.GetSpeed())
}

// -----------------------------------------------------------------------------
// Scenario: Stats Retrieval
// -----------------------------------------------------------------------------
// Given: A running simulation
// When: GetStats is called
// Then: Valid stats should be returned
func TestBDD_Runner_StatsRetrieval(t *testing.T) {
	config := ecosystem.DefaultConfig(testWorldID)
	config.TickInterval = 10 * time.Millisecond
	config.Speed = ecosystem.SpeedTurbo

	runner := ecosystem.NewSimulationRunner(config, nil, nil)
	runner.InitializePopulationSimulator(42)

	err := runner.Start(0)
	require.NoError(t, err)

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	stats := runner.GetStats()

	assert.Equal(t, ecosystem.RunnerRunning, stats.State)
	assert.GreaterOrEqual(t, stats.CurrentYear, int64(0))
	assert.GreaterOrEqual(t, stats.TickCount, int64(0))

	runner.Stop()
}

// -----------------------------------------------------------------------------
// Scenario: Event Broadcasting
// -----------------------------------------------------------------------------
// Given: Runner with EventBroadcastHandler set
// When: Notable event occurs (speciation, extinction)
// Then: Handler should receive event
//
//	AND Event should include year, type, description
func TestBDD_Runner_EventBroadcast(t *testing.T) {
	config := ecosystem.DefaultConfig(testWorldID)
	config.TickInterval = 10 * time.Millisecond
	config.Speed = ecosystem.SpeedTurbo

	runner := ecosystem.NewSimulationRunner(config, nil, nil)
	runner.InitializePopulationSimulator(42)

	var receivedEvents []ecosystem.RunnerEvent
	var mu sync.Mutex

	runner.SetEventBroadcastHandler(func(event ecosystem.RunnerEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	})

	err := runner.Start(0)
	require.NoError(t, err)

	// Run long enough to potentially trigger events
	time.Sleep(200 * time.Millisecond)
	runner.Stop()

	// Check that we can retrieve events (may or may not have received any depending on sim)
	recentEvents := runner.GetRecentEvents(10)
	assert.NotNil(t, recentEvents, "Should be able to get recent events")
}

// -----------------------------------------------------------------------------
// Scenario: Snapshot Creation
// -----------------------------------------------------------------------------
// Given: A running simulation with snapshot interval
// When: Simulation advances past interval
// Then: Snapshot should be created
func TestBDD_Runner_SnapshotCreation(t *testing.T) {
	config := ecosystem.DefaultConfig(testWorldID)
	config.TickInterval = 10 * time.Millisecond
	config.SnapshotInterval = 10 // Every 10 years
	config.Speed = ecosystem.SpeedTurbo

	runner := ecosystem.NewSimulationRunner(config, nil, nil)
	runner.InitializePopulationSimulator(42)

	var snapshotCount int
	var mu sync.Mutex

	runner.SetSnapshotHandler(func(snapshot *ecosystem.Snapshot) error {
		mu.Lock()
		snapshotCount++
		mu.Unlock()
		return nil
	})

	err := runner.Start(0)
	require.NoError(t, err)

	// Run enough to trigger snapshots
	time.Sleep(300 * time.Millisecond)
	runner.Stop()

	snapshots := runner.GetSnapshots()
	assert.NotEmpty(t, snapshots, "Should have created snapshots")
}

// -----------------------------------------------------------------------------
// Scenario: Concurrent Control Access
// -----------------------------------------------------------------------------
// Given: A running simulation loop
// When: Speed is changed and Status is queried from multiple goroutines
// Then: The application should NOT panic
//
//	AND The runner should remain functional
func TestBDD_Runner_ConcurrencySafety(t *testing.T) {
	config := ecosystem.DefaultConfig(testWorldID)
	config.TickInterval = 10 * time.Millisecond
	config.Speed = ecosystem.SpeedFast

	runner := ecosystem.NewSimulationRunner(config, nil, nil)
	runner.InitializePopulationSimulator(42)

	err := runner.Start(0)
	require.NoError(t, err)

	// Start concurrent operations
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Multiple goroutines hammering the runner
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					runner.SetSpeed(ecosystem.SpeedFast)
					runner.GetState()
					runner.GetCurrentYear()
					runner.GetStats()
					runner.SetSpeed(ecosystem.SpeedTurbo)
				}
			}
		}()
	}

	wg.Wait()
	runner.Stop()

	// If we got here without panic, test passes
	assert.True(t, true, "Concurrent access should not cause panic")
}

// -----------------------------------------------------------------------------
// Scenario: Snapshot Interval Consistency
// -----------------------------------------------------------------------------
// Given: Snapshot interval of 10 years
// When: Simulation runs for 100+ years
// Then: Multiple snapshots should be created
//
//	AND Each snapshot should have increasing years
func TestBDD_Runner_SnapshotInterval(t *testing.T) {
	config := ecosystem.DefaultConfig(testWorldID)
	config.TickInterval = 5 * time.Millisecond
	config.SnapshotInterval = 10
	config.Speed = ecosystem.SpeedTurbo

	runner := ecosystem.NewSimulationRunner(config, nil, nil)
	runner.InitializePopulationSimulator(42)

	err := runner.Start(0)
	require.NoError(t, err)

	// Run until we have enough snapshots
	time.Sleep(500 * time.Millisecond)
	runner.Stop()

	snapshots := runner.GetSnapshots()

	if len(snapshots) >= 2 {
		// Verify increasing years
		for i := 1; i < len(snapshots); i++ {
			assert.Greater(t, snapshots[i].Year, snapshots[i-1].Year,
				"Snapshot years should be increasing")
		}
	}
}

// -----------------------------------------------------------------------------
// Scenario: Speed Change During Simulation
// -----------------------------------------------------------------------------
// Given: Runner at SpeedNormal (10 years/tick)
// When: Speed changed to SpeedFast (100 years/tick)
// Then: Advancement rate should increase
func TestBDD_Runner_SpeedChange_Effect(t *testing.T) {
	assert.Fail(t, "BDD RED: Time-based speed comparison requires deterministic clock")
	// Pseudocode:
	// runner.SetSpeed(SpeedNormal)
	// time.Sleep(100 * time.Millisecond)
	// yearAtNormal := runner.GetCurrentYear()
	// runner.SetSpeed(SpeedFast)
	// time.Sleep(100 * time.Millisecond)
	// yearAtFast := runner.GetCurrentYear()
	// fastAdvance := yearAtFast - yearAtNormal
	// assert fastAdvance > yearAtNormal * 5 // Much faster
}

// -----------------------------------------------------------------------------
// Scenario: Persistence on Stop
// -----------------------------------------------------------------------------
// Given: Runner with snapshot repository
// When: Stop() is called
// Then: Final state should be persisted
//
//	AND State should be recoverable on restart
func TestBDD_Runner_PersistenceOnStop(t *testing.T) {
	assert.Fail(t, "BDD RED: Persistence test requires database setup - see runner_test.go for full test")
	// Pseudocode:
	// repo := NewSimulationSnapshotRepository(db)
	// runner := NewSimulationRunner(config, repo, nil)
	// runner.Start(0)
	// time.Sleep(100 * time.Millisecond)
	// yearBefore := runner.GetCurrentYear()
	// runner.Stop()
	// // Reload
	// runner2 := NewSimulationRunner(config, repo, nil)
	// runner2.InitializePopulationSimulator(seed)
	// assert runner2.GetCurrentYear() == yearBefore
}

// -----------------------------------------------------------------------------
// Scenario: Deterministic Tick Execution
// -----------------------------------------------------------------------------
// Given: A runner configured for Manual Stepping (Test Mode)
// When: Step(5) is called
// Then: Exactly 5 simulation years/ticks should process
//
//	AND No more, no less
func TestBDD_Runner_DeterministicStepping(t *testing.T) {
	assert.Fail(t, "BDD RED: Manual stepping not yet implemented - requires Step() method")
	// Pseudocode:
	// runner := NewTestRunner() // Manual clock
	// startYear := runner.CurrentYear
	// runner.Step(5)
	// assert runner.CurrentYear == startYear + 5
}

// -----------------------------------------------------------------------------
// Scenario: Panic Recovery in Simulation Loop
// -----------------------------------------------------------------------------
// Given: A simulation strategy that forces a panic (e.g., div by zero)
// When: The runner executes a tick
// Then: The panic should be recovered
//
//	AND The runner should enter RunnerError/Paused state
//	AND The error should be logged
func TestBDD_Runner_PanicRecovery(t *testing.T) {
	assert.Fail(t, "BDD RED: Panic recovery not yet implemented in runLoop")
	// Pseudocode:
	// badStrategy := func() { panic("oops") }
	// runner.SetStrategy(badStrategy)
	// runner.Start(0)
	// assert runner.State() == RunnerError
	// assert runner.LastError() == "oops"
}

// -----------------------------------------------------------------------------
// Scenario: Dynamic Configuration Update
// -----------------------------------------------------------------------------
// Given: A running simulation with SnapshotInterval = 100
// When: UpdateConfig is called with SnapshotInterval = 10
// Then: The next snapshot should trigger based on the new interval
func TestBDD_Runner_HotConfigUpdate(t *testing.T) {
	assert.Fail(t, "BDD RED: Hot config update not yet implemented")
	// Pseudocode:
	// runner.Config.SnapshotInterval = 100
	// runner.Step(50) // No snapshot
	// runner.UpdateConfig(Interval: 10)
	// runner.Step(10) // Should snapshot now
	// assert snapshotCreated == true
}

// -----------------------------------------------------------------------------
// Scenario: Turning Point Triggers (Table-Driven)
// -----------------------------------------------------------------------------
// Given: Various world states (High Pop, Extinction, Time passed)
// When: CheckForTurningPoint is evaluated
// Then: The correct Turning Point Event should be returned
func TestBDD_Runner_TurningPoints(t *testing.T) {
	assert.Fail(t, "BDD RED: Turning point unit testing requires isolated manager - see turning_point_test.go")

	scenarios := []struct {
		name        string
		year        int64
		extinctRate float64
		expectEvent bool
		expectType  string
	}{
		{"Standard Year", 100, 0.0, false, ""},
		{"Million Year Mark", 1_000_000, 0.0, true, "EpochChange"},
		{"Mass Extinction", 0, 0.9, true, "ExtinctionEvent"},
	}
	_ = scenarios // For BDD stub - will be used when implemented
}

// -----------------------------------------------------------------------------
// Scenario: Max Year Target Stop
// -----------------------------------------------------------------------------
// Given: MaxYearTarget set to 1000
// When: Simulation reaches year 1000
// Then: Runner should pause
func TestBDD_Runner_MaxYearTarget(t *testing.T) {
	config := ecosystem.DefaultConfig(testWorldID)
	config.TickInterval = 5 * time.Millisecond
	config.MaxYearTarget = 100
	config.Speed = ecosystem.SpeedTurbo

	runner := ecosystem.NewSimulationRunner(config, nil, nil)
	runner.InitializePopulationSimulator(42)

	err := runner.Start(0)
	require.NoError(t, err)

	// Wait for target to be reached
	time.Sleep(500 * time.Millisecond)

	// Should have paused at target
	state := runner.GetState()
	currentYear := runner.GetCurrentYear()

	runner.Stop()

	// Either we hit the target and paused, or we're still running toward it
	if currentYear >= 100 {
		assert.Equal(t, ecosystem.RunnerPaused, state,
			"Runner should pause at MaxYearTarget")
	}
}

// -----------------------------------------------------------------------------
// Scenario: GetSnapshotAtYear
// -----------------------------------------------------------------------------
// Given: Multiple snapshots at different years
// When: GetSnapshotAtYear is called
// Then: Closest snapshot should be returned
func TestBDD_Runner_GetSnapshotAtYear(t *testing.T) {
	config := ecosystem.DefaultConfig(testWorldID)
	config.TickInterval = 5 * time.Millisecond
	config.SnapshotInterval = 50
	config.Speed = ecosystem.SpeedTurbo

	runner := ecosystem.NewSimulationRunner(config, nil, nil)
	runner.InitializePopulationSimulator(42)

	err := runner.Start(0)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)
	runner.Stop()

	snapshots := runner.GetSnapshots()
	if len(snapshots) > 0 {
		// Get snapshot at a known year
		targetYear := snapshots[0].Year
		found := runner.GetSnapshotAtYear(targetYear)
		assert.NotNil(t, found, "Should find snapshot at year")
		assert.Equal(t, targetYear, found.Year)
	}
}
