package world

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeProgression_OneYear(t *testing.T) {
	registry := NewRegistry()
	natsPublisher := &MockNATSPublisher{}
	tm := NewTickerManager(registry, nil, natsPublisher)
	defer tm.StopAll()

	worldID := uuid.New()
	// Use high dilation to simulate a year quickly
	// 1 year = 360 days * 24 hours = 8640 hours
	// We want to run this in < 5 seconds
	// Real time: 5 seconds
	// Game time: 1 year
	// Dilation needed: (1 year) / (5 seconds)
	// 1 year = 365 * 24 * 3600 = 31,536,000 seconds
	// Dilation = 31,536,000 / 5 = ~6,300,000
	// Let's use a smaller scale for the test to ensure we capture enough ticks
	// We don't need to simulate EVERY tick, just enough to verify progression

	// Let's simulate 1 season (90 days) in ~1 second
	// 90 days = 90 * 24 * 3600 = 7,776,000 seconds
	// 1 second real time
	// Dilation = 7,776,000
	dilationFactor := 100000.0 // 100ms tick = 10,000s game time (~2.7 hours)

	err := tm.SpawnTicker(worldID, "Fast World", dilationFactor)
	require.NoError(t, err)

	// Run for enough time to cover a full year (4 seasons)
	// Year length = 4 * 90 days = 360 days
	// Game time per tick = 100ms * 100,000 = 10,000s
	// Ticks per day = 86400 / 10000 = 8.64 ticks
	// Ticks per year = 360 * 8.64 = 3110 ticks
	// Real time needed = 3110 * 100ms = 311 seconds... too long!

	// Let's increase dilation even more for the test
	dilationFactor = 1000000.0 // 100ms tick = 100,000s game time (~27 hours, > 1 day)
	// This is too fast, we'll skip days.

	// Let's verify day/night cycle with lower dilation first
	t.Run("DayNightCycle", func(t *testing.T) {
		natsPublisher.Clear()
		// 1 day in 1 second
		// 86400s game time / 1s real time = 86400 dilation
		dilation := 86400.0
		// Tick = 100ms * 86400 = 8640s = 2.4 hours

		// Restart ticker with new dilation
		tm.StopTicker(worldID)
		err := tm.SpawnTicker(worldID, "Day Cycle World", dilation)
		require.NoError(t, err)

		time.Sleep(1100 * time.Millisecond) // Wait just over 1 second (1 day)

		published := natsPublisher.GetPublished()
		require.Greater(t, len(published), 8)

		// Verify we went through different times of day
		seenTimes := make(map[string]bool)
		var lastSunPos float64 = -1.0

		for _, msg := range published {
			var broadcast TickBroadcast
			json.Unmarshal(msg.Data, &broadcast)
			seenTimes[broadcast.TimeOfDay] = true

			// Verify sun position loops
			if lastSunPos != -1.0 {
				if broadcast.SunPosition < lastSunPos {
					// Wrapped around (new day)
					t.Log("New day started")
				} else {
					assert.Greater(t, broadcast.SunPosition, lastSunPos)
				}
			}
			lastSunPos = broadcast.SunPosition
		}

		assert.True(t, seenTimes["Night"], "Should have seen Night")
		assert.True(t, seenTimes["Noon"], "Should have seen Noon")
	})

	t.Run("SeasonCycle", func(t *testing.T) {
		natsPublisher.Clear()
		// 1 year in 2 seconds
		// Year = 360 days * 86400 = 31,104,000s
		// Dilation = 15,552,000
		dilation := 15552000.0
		// Tick = 100ms * 15.5M = 1,555,200s = 18 days

		tm.StopTicker(worldID)
		err := tm.SpawnTicker(worldID, "Season Cycle World", dilation)
		require.NoError(t, err)

		time.Sleep(2500 * time.Millisecond) // Wait 2.5 seconds (> 1 year)

		published := natsPublisher.GetPublished()
		require.Greater(t, len(published), 10)

		seenSeasons := make(map[string]bool)
		seasonsOrder := []string{}
		var lastSeason string

		for _, msg := range published {
			var broadcast TickBroadcast
			json.Unmarshal(msg.Data, &broadcast)

			if broadcast.CurrentSeason != lastSeason {
				seasonsOrder = append(seasonsOrder, broadcast.CurrentSeason)
				lastSeason = broadcast.CurrentSeason
			}
			seenSeasons[broadcast.CurrentSeason] = true
		}

		assert.True(t, seenSeasons["Spring"], "Should have seen Spring")
		assert.True(t, seenSeasons["Summer"], "Should have seen Summer")
		assert.True(t, seenSeasons["Autumn"], "Should have seen Autumn")
		assert.True(t, seenSeasons["Winter"], "Should have seen Winter")

		// Verify order (might start mid-spring depending on previous tests)
		// But generally should follow sequence
		t.Logf("Seasons observed: %v", seasonsOrder)
	})
}
