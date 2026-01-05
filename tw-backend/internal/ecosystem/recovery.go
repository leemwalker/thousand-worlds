package ecosystem

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	pb "tw-backend/api/proto"
	"tw-backend/internal/events"
	"tw-backend/internal/repository"
	"tw-backend/internal/storage"
	"tw-backend/internal/worldgen/geography"
)

// RecoveryService handles crash recovery by loading snapshots and replaying events.
type RecoveryService struct {
	snapshotStore storage.SnapshotStoreInterface
	eventConsumer events.EventConsumerInterface
	saveRepo      repository.SaveRepository
}

// NewRecoveryService creates a new recovery service.
func NewRecoveryService(
	snapshotStore storage.SnapshotStoreInterface,
	eventConsumer events.EventConsumerInterface,
	saveRepo repository.SaveRepository,
) *RecoveryService {
	return &RecoveryService{
		snapshotStore: snapshotStore,
		eventConsumer: eventConsumer,
		saveRepo:      saveRepo,
	}
}

// RecoveryResult contains the outcome of a world recovery operation.
type RecoveryResult struct {
	SnapshotYear     int64                      // Year from the loaded snapshot
	EventsReplayed   int                        // Number of events replayed after snapshot
	CurrentYear      int64                      // Final year after replay
	RecoveryDuration time.Duration              // Time taken for recovery
	Heightmap        *geography.SphereHeightmap // Recovered heightmap (nil if recovery failed)
}

// RecoverWorld loads the latest snapshot and replays events to restore world state.
// Returns nil, ErrSaveNotFound if no save exists for this world.
func (r *RecoveryService) RecoverWorld(ctx context.Context, worldID uuid.UUID) (*RecoveryResult, error) {
	startTime := time.Now()

	// 1. Get the latest save metadata
	save, err := r.saveRepo.GetLatestSave(ctx, worldID)
	if err != nil {
		return nil, fmt.Errorf("get latest save: %w", err)
	}

	result := &RecoveryResult{
		SnapshotYear: save.Year,
		CurrentYear:  save.Year,
	}

	// 2. Load the heightmap snapshot from MinIO
	snapshotData, err := r.snapshotStore.Download(ctx, save.SnapshotKey)
	if err != nil {
		return nil, fmt.Errorf("download snapshot %s: %w", save.SnapshotKey, err)
	}

	// 3. Deserialize the heightmap (auto-detects resolution from header)
	result.Heightmap, err = events.DeserializeHeightmapAuto(snapshotData)
	if err != nil {
		return nil, fmt.Errorf("deserialize heightmap: %w", err)
	}

	// 4. Replay events from NATS starting after the snapshot's sequence
	if r.eventConsumer != nil && save.EventSequence > 0 {
		evts, err := r.eventConsumer.GetEventsFromSequence(ctx, worldID.String(), save.EventSequence)
		if err != nil {
			// Log but don't fail - we have the snapshot at least
			fmt.Printf("[Recovery] Warning: failed to replay events: %v\n", err)
		} else {
			result.EventsReplayed = len(evts)

			// Apply events to heightmap
			for _, evt := range evts {
				r.applyEvent(result.Heightmap, evt)
				// Track the latest year from events
				if tectonic := evt.GetTectonic(); tectonic != nil {
					result.CurrentYear = tectonic.Year
				} else if erosion := evt.GetErosion(); erosion != nil {
					result.CurrentYear = erosion.Year
				}
			}
		}
	}

	result.RecoveryDuration = time.Since(startTime)
	return result, nil
}

// applyEvent applies a simulation event to the heightmap.
// This reconstructs state changes from the event log.
// Note: For now, we skip event replay since the current proto doesn't include
// per-cell deltas in TectonicUpdate. Full reconstruction relies on snapshots.
func (r *RecoveryService) applyEvent(hm *geography.SphereHeightmap, evt *pb.SimulationEvent) {
	if hm == nil || evt == nil {
		return
	}

	// The current proto doesn't include elevation deltas in TectonicUpdate or ErosionUpdate
	// that we could apply directly. Future enhancement could add:
	// - ElevationDeltas map[int64]float64 to TectonicUpdate
	// - SedimentDeltas map[int64]float64 to ErosionUpdate
	//
	// For now, crash recovery relies primarily on snapshots, with events
	// used mainly for tracking the current year.
	//
	// TODO: Extend proto with delta fields for true event sourcing
}

// CreateRecoverySave creates a save point for crash recovery.
// Called automatically during simulation at snapshot intervals.
func (r *RecoveryService) CreateRecoverySave(ctx context.Context, worldID uuid.UUID, year int64, snapshotKey string, eventSequence uint64) error {
	save := &repository.WorldSave{
		WorldID:       worldID,
		SnapshotKey:   snapshotKey,
		EventSequence: eventSequence,
		Year:          year,
		Metadata: map[string]any{
			"type": "auto_recovery",
		},
	}

	if err := r.saveRepo.CreateSave(ctx, save); err != nil {
		return fmt.Errorf("create save: %w", err)
	}

	// Clean up old saves, keeping the last 10
	deleted, err := r.saveRepo.DeleteOldSaves(ctx, worldID, 10)
	if err != nil {
		fmt.Printf("[Recovery] Warning: failed to clean old saves: %v\n", err)
	} else if deleted > 0 {
		fmt.Printf("[Recovery] Cleaned up %d old saves\n", deleted)
	}

	return nil
}
