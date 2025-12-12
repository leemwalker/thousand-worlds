package world

import (
	"time"

	"github.com/rs/zerolog/log"
)

const CatchupDilationFactor = 100.0

// performCatchupThenRun performs catch-up ticks and then starts normal ticker
func (tm *TickerManager) performCatchupThenRun(t *ticker, world *WorldState) {
	// Calculate missed time
	pauseDuration := time.Since(world.PausedAt)
	// Missed game time = pause duration * dilation
	missedGameTime := time.Duration(float64(pauseDuration) * t.dilationFactor)
	// Missed ticks = pause duration / tick interval (how many 100ms periods passed)
	missedTicks := int64(pauseDuration / t.tickInterval)

	log.Info().
		Str("world_id", t.worldID.String()).
		Dur("pause_duration", pauseDuration).
		Dur("missed_game_time", missedGameTime).
		Int64("catch_up_ticks", missedTicks).
		Msg("Performing catch-up")

	// Clear pause time in registry
	_ = tm.registry.UpdateWorld(t.worldID, func(w *WorldState) {
		w.PausedAt = time.Time{}
	})

	// Emit catch-up ticks at 100x speed
	catchupInterval := t.tickInterval / CatchupDilationFactor

	for i := int64(0); i < missedTicks; i++ {
		select {
		case <-t.stopCh:
			log.Info().Str("world_id", t.worldID.String()).Msg("Catch-up interrupted by stop")
			return // Stopped during catch-up
		default:
			tm.tick(t)                  // Reuse existing tick logic
			time.Sleep(catchupInterval) // 100x faster
		}
	}

	log.Info().
		Str("world_id", t.worldID.String()).
		Int64("catch_up_ticks", missedTicks).
		Msg("Catch-up complete, starting normal ticker")

	// After catch-up, start normal ticker
	tm.runTicker(t)
}
