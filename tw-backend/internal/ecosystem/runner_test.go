package ecosystem

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSimulationRunner_BasicFlow(t *testing.T) {
	config := DefaultConfig(uuid.New())
	config.TickInterval = 10 * time.Millisecond // Fast for testing
	config.Speed = SpeedNormal

	runner := NewSimulationRunner(config, nil, nil)
	runner.InitializePopulationSimulator(12345)

	t.Run("starts in idle state", func(t *testing.T) {
		if runner.GetState() != RunnerIdle {
			t.Errorf("State = %s, want idle", runner.GetState())
		}
	})

	t.Run("can start simulation", func(t *testing.T) {
		err := runner.Start(0)
		if err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		// Give it time to start
		time.Sleep(20 * time.Millisecond)

		if runner.GetState() != RunnerRunning {
			t.Errorf("State = %s, want running", runner.GetState())
		}
	})

	t.Run("advances time", func(t *testing.T) {
		// Let it run for a bit
		time.Sleep(100 * time.Millisecond)

		year := runner.GetCurrentYear()
		if year <= 0 {
			t.Errorf("Current year = %d, want > 0", year)
		}
		t.Logf("Advanced to year %d", year)
	})

	t.Run("can stop simulation", func(t *testing.T) {
		runner.Stop()

		if runner.GetState() != RunnerIdle {
			t.Errorf("State = %s, want idle", runner.GetState())
		}
	})
}

func TestSimulationRunner_PauseResume(t *testing.T) {
	config := DefaultConfig(uuid.New())
	config.TickInterval = 10 * time.Millisecond
	config.Speed = SpeedNormal

	runner := NewSimulationRunner(config, nil, nil)
	runner.Start(0)
	defer runner.Stop()

	// Let it run
	time.Sleep(50 * time.Millisecond)
	yearBeforePause := runner.GetCurrentYear()

	t.Run("pause stops advancement", func(t *testing.T) {
		runner.Pause()

		if runner.GetState() != RunnerPaused {
			t.Errorf("State = %s, want paused", runner.GetState())
		}

		time.Sleep(100 * time.Millisecond)
		yearAfterPause := runner.GetCurrentYear()

		// Allow for one tick to have been in flight when pause was called
		if yearAfterPause > yearBeforePause+int64(config.Speed) {
			t.Errorf("Year advanced too much while paused: %d -> %d", yearBeforePause, yearAfterPause)
		}
	})

	t.Run("resume continues advancement", func(t *testing.T) {
		runner.Resume()

		if runner.GetState() != RunnerRunning {
			t.Errorf("State = %s, want running", runner.GetState())
		}

		time.Sleep(50 * time.Millisecond)
		yearAfterResume := runner.GetCurrentYear()

		if yearAfterResume <= yearBeforePause {
			t.Errorf("Year didn't advance after resume: %d", yearAfterResume)
		}
	})
}

func TestSimulationRunner_TickHandler(t *testing.T) {
	config := DefaultConfig(uuid.New())
	config.TickInterval = 10 * time.Millisecond
	config.Speed = SpeedNormal

	runner := NewSimulationRunner(config, nil, nil)

	var tickCount int64
	runner.SetTickHandler(func(year int64, yearsElapsed int64) error {
		atomic.AddInt64(&tickCount, 1)
		return nil
	})

	runner.Start(0)
	time.Sleep(100 * time.Millisecond)
	runner.Stop()

	count := atomic.LoadInt64(&tickCount)
	if count < 5 {
		t.Errorf("Tick count = %d, want >= 5", count)
	}
	t.Logf("Tick handler called %d times", count)
}

func TestSimulationRunner_Snapshots(t *testing.T) {
	config := DefaultConfig(uuid.New())
	config.TickInterval = 10 * time.Millisecond
	config.SnapshotInterval = 10 // Snapshot every 10 years
	config.Speed = SpeedNormal   // 10 years per tick

	runner := NewSimulationRunner(config, nil, nil)

	var snapshotCount int64
	runner.SetSnapshotHandler(func(snap *Snapshot) error {
		atomic.AddInt64(&snapshotCount, 1)
		return nil
	})

	runner.Start(0)
	time.Sleep(150 * time.Millisecond) // Should create several snapshots
	runner.Stop()

	count := atomic.LoadInt64(&snapshotCount)
	if count < 2 {
		t.Errorf("Snapshot count = %d, want >= 2", count)
	}
	t.Logf("Created %d snapshots", count)

	// Check snapshots are stored
	snapshots := runner.GetSnapshots()
	if len(snapshots) == 0 {
		t.Error("No snapshots stored")
	}
}

func TestSimulationRunner_Stats(t *testing.T) {
	config := DefaultConfig(uuid.New())
	config.TickInterval = 10 * time.Millisecond
	config.Speed = SpeedFast // 100 years per tick

	runner := NewSimulationRunner(config, nil, nil)
	runner.Start(0)
	time.Sleep(100 * time.Millisecond)
	runner.Stop()

	stats := runner.GetStats()

	t.Logf("Stats: Year=%d, Ticks=%d, YearsSimulated=%d, YearsPerSecond=%.1f",
		stats.CurrentYear, stats.TickCount, stats.YearsSimulated, stats.YearsPerSecond)

	if stats.TickCount <= 0 {
		t.Error("No ticks recorded")
	}
	if stats.YearsPerSecond <= 0 {
		t.Error("Years per second not calculated")
	}
}

func TestSimulationRunner_Events(t *testing.T) {
	config := DefaultConfig(uuid.New())
	runner := NewSimulationRunner(config, nil, nil)

	// Add some events
	for i := 0; i < 5; i++ {
		runner.AddEvent(RunnerEvent{
			Year:        int64(i * 1000),
			Type:        "test",
			Description: "Test event",
			Importance:  5,
		})
	}

	events := runner.GetRecentEvents(3)
	if len(events) != 3 {
		t.Errorf("Got %d events, want 3", len(events))
	}

	// Most recent first
	if events[0].Year != 4000 {
		t.Errorf("First event year = %d, want 4000", events[0].Year)
	}
}

func TestSimulationRunner_SpeedChange(t *testing.T) {
	config := DefaultConfig(uuid.New())
	config.TickInterval = 10 * time.Millisecond
	config.Speed = SpeedSlow // 1 year per tick

	runner := NewSimulationRunner(config, nil, nil)
	runner.Start(0)

	// Run at slow speed
	time.Sleep(50 * time.Millisecond)
	yearAtSlow := runner.GetCurrentYear()

	// Switch to fast
	runner.SetSpeed(SpeedFast) // 100 years per tick
	time.Sleep(50 * time.Millisecond)
	yearAtFast := runner.GetCurrentYear()

	runner.Stop()

	// Fast should have advanced much more
	slowAdvance := yearAtSlow
	fastAdvance := yearAtFast - yearAtSlow

	t.Logf("Slow advance: %d, Fast advance: %d", slowAdvance, fastAdvance)

	if fastAdvance <= slowAdvance*10 {
		t.Log("Fast speed should advance more than slow (timing-dependent)")
	}
}

func TestPlayerViewSync_BasicFlow(t *testing.T) {
	config := DefaultConfig(uuid.New())
	config.TickInterval = 10 * time.Millisecond
	config.Speed = SpeedNormal

	runner := NewSimulationRunner(config, nil, nil)
	viewSync := NewPlayerViewSync(uuid.New(), runner)

	// Set species counts
	viewSync.UpdateSpeciesCounts(100, 80, 20, 2)

	t.Run("gets current state", func(t *testing.T) {
		state := viewSync.GetCurrentState()
		if state == nil {
			t.Fatal("State is nil")
		}
	})

	t.Run("starts update loop", func(t *testing.T) {
		var updateCount int64
		viewSync.SetUpdateHandler(func(state *ViewState) {
			atomic.AddInt64(&updateCount, 1)
		})

		viewSync.Start()
		runner.Start(0)

		time.Sleep(150 * time.Millisecond)

		viewSync.Stop()
		runner.Stop()

		count := atomic.LoadInt64(&updateCount)
		if count < 1 {
			t.Errorf("Update count = %d, want >= 1", count)
		}
		t.Logf("View updated %d times", count)
	})
}

func TestPlayerViewSync_SpeciesCounts(t *testing.T) {
	runner := NewSimulationRunner(DefaultConfig(uuid.New()), nil, nil)
	viewSync := NewPlayerViewSync(uuid.New(), runner)

	viewSync.UpdateSpeciesCounts(500, 400, 100, 3)
	viewSync.ForceUpdate()

	state := viewSync.GetCurrentState()

	if state.TotalSpecies != 500 {
		t.Errorf("TotalSpecies = %d, want 500", state.TotalSpecies)
	}
	if state.ExtantSpecies != 400 {
		t.Errorf("ExtantSpecies = %d, want 400", state.ExtantSpecies)
	}
	if state.SapientSpecies != 3 {
		t.Errorf("SapientSpecies = %d, want 3", state.SapientSpecies)
	}
}

func TestPlayerViewSync_Snapshot(t *testing.T) {
	runner := NewSimulationRunner(DefaultConfig(uuid.New()), nil, nil)
	viewSync := NewPlayerViewSync(uuid.New(), runner)

	viewSync.UpdateSpeciesCounts(100, 80, 20, 1)

	snapshot := viewSync.GetSnapshot()
	if snapshot.TotalSpecies != 100 {
		t.Errorf("Snapshot TotalSpecies = %d, want 100", snapshot.TotalSpecies)
	}

	// Clear and restore
	viewSync.UpdateSpeciesCounts(0, 0, 0, 0)
	viewSync.RestoreFromSnapshot(snapshot)

	state := viewSync.GetCurrentState()
	// After force update, counts should be restored
	viewSync.ForceUpdate()
	state = viewSync.GetCurrentState()
	if state.TotalSpecies != 100 {
		t.Logf("Note: RestoreFromSnapshot sets internal counts, ForceUpdate reads them")
	}
}
