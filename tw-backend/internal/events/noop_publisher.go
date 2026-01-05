package events

import (
	"context"

	pb "tw-backend/api/proto"
)

// NoOpPublisher discards all events. Use for testing or when event infrastructure is disabled.
type NoOpPublisher struct{}

// Verify NoOpPublisher implements Publisher at compile time
var _ Publisher = (*NoOpPublisher)(nil)

func (NoOpPublisher) PublishTectonic(context.Context, *pb.TectonicUpdate) error     { return nil }
func (NoOpPublisher) PublishErosion(context.Context, *pb.ErosionUpdate) error       { return nil }
func (NoOpPublisher) PublishPhase(context.Context, *pb.PhaseTransition) error       { return nil }
func (NoOpPublisher) PublishVolcanic(context.Context, *pb.VolcanicEvent) error      { return nil }
func (NoOpPublisher) PublishAtmosphere(context.Context, *pb.AtmosphereUpdate) error { return nil }
func (NoOpPublisher) PublishSnapshot(context.Context, *pb.HeightmapSnapshot) error  { return nil }
func (NoOpPublisher) GetLastSequence() uint64                                       { return 0 }
func (NoOpPublisher) Close() error                                                  { return nil }

// NewNoOpPublisher creates a publisher that discards all events.
func NewNoOpPublisher() Publisher {
	return NoOpPublisher{}
}
