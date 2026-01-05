// Package events provides event publishing infrastructure for simulation events.
// Events are published to NATS JetStream for persistence, replay, and distribution.
package events

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	pb "tw-backend/api/proto"
)

// Publisher defines the interface for emitting simulation events.
// Implementations can be NATSPublisher for production or NoOpPublisher for testing.
type Publisher interface {
	// PublishTectonic emits a tectonic update event.
	// TIMING: Adaptive based on planetary heat (100K-10M years)
	PublishTectonic(ctx context.Context, event *pb.TectonicUpdate) error

	// PublishErosion emits an erosion update event.
	// TIMING: Static 10M year intervals (only when heat <= 4.0)
	PublishErosion(ctx context.Context, event *pb.ErosionUpdate) error

	// PublishPhase emits a phase transition event (Great Deluge, etc.).
	// TIMING: Event-driven when conditions are met
	PublishPhase(ctx context.Context, event *pb.PhaseTransition) error

	// PublishVolcanic emits a volcanic event.
	// TIMING: Adaptive based on planetary heat
	PublishVolcanic(ctx context.Context, event *pb.VolcanicEvent) error

	// PublishAtmosphere emits an atmosphere composition update.
	// TIMING: Era transitions only
	PublishAtmosphere(ctx context.Context, event *pb.AtmosphereUpdate) error

	// PublishSnapshot emits a full heightmap snapshot.
	// TIMING: Static 100M year intervals or on-demand
	PublishSnapshot(ctx context.Context, event *pb.HeightmapSnapshot) error

	// GetLastSequence returns the last JetStream sequence number assigned.
	// Used for crash recovery to track position in event log.
	GetLastSequence() uint64

	// Close releases resources. Safe to call multiple times.
	Close() error
}

// NATSPublisher implements Publisher using NATS JetStream for durable event storage.
type NATSPublisher struct {
	nc       *nats.Conn
	js       jetstream.JetStream
	stream   jetstream.Stream
	localSeq atomic.Int64  // Local sequence for event ordering
	lastSeq  atomic.Uint64 // Last JetStream-assigned sequence
}

// StreamConfig defines the JetStream stream configuration for simulation events.
var StreamConfig = jetstream.StreamConfig{
	Name:        "SIMULATION_EVENTS",
	Description: "Simulation events for world state evolution",
	Subjects: []string{
		"simulation.>", // simulation.{world_id}.{event_type}
	},
	Storage:    jetstream.FileStorage,
	Retention:  jetstream.LimitsPolicy,
	MaxMsgs:    1_000_000,           // 1M events per stream (covers ~2B years at 10M intervals)
	MaxAge:     30 * 24 * time.Hour, // 30 day retention
	Replicas:   1,                   // Increase for production HA
	Discard:    jetstream.DiscardOld,
	Duplicates: 5 * time.Minute, // Dedup window
}

// NewNATSPublisher creates a publisher connected to NATS JetStream.
// The stream is created if it doesn't exist.
func NewNATSPublisher(natsURL string) (*NATSPublisher, error) {
	nc, err := nats.Connect(natsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("jetstream init: %w", err)
	}

	// Create or update the stream
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := js.CreateOrUpdateStream(ctx, StreamConfig)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("stream create: %w", err)
	}

	return &NATSPublisher{
		nc:     nc,
		js:     js,
		stream: stream,
	}, nil
}

// publish sends a wrapped SimulationEvent to the specified subject.
func (p *NATSPublisher) publish(ctx context.Context, subject string, event *pb.SimulationEvent) error {
	// Set local sequence and timestamp
	event.Sequence = p.localSeq.Add(1)
	event.TimestampUnix = time.Now().Unix()

	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	ack, err := p.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}

	// Store the JetStream-assigned sequence for crash recovery
	p.lastSeq.Store(ack.Sequence)

	return nil
}

// GetLastSequence returns the last JetStream sequence number assigned.
func (p *NATSPublisher) GetLastSequence() uint64 {
	return p.lastSeq.Load()
}

func (p *NATSPublisher) PublishTectonic(ctx context.Context, ev *pb.TectonicUpdate) error {
	subject := fmt.Sprintf("simulation.%s.tectonic", ev.WorldId)
	return p.publish(ctx, subject, &pb.SimulationEvent{
		Event: &pb.SimulationEvent_Tectonic{Tectonic: ev},
	})
}

func (p *NATSPublisher) PublishErosion(ctx context.Context, ev *pb.ErosionUpdate) error {
	subject := fmt.Sprintf("simulation.%s.erosion", ev.WorldId)
	return p.publish(ctx, subject, &pb.SimulationEvent{
		Event: &pb.SimulationEvent_Erosion{Erosion: ev},
	})
}

func (p *NATSPublisher) PublishPhase(ctx context.Context, ev *pb.PhaseTransition) error {
	subject := fmt.Sprintf("simulation.%s.phase", ev.WorldId)
	return p.publish(ctx, subject, &pb.SimulationEvent{
		Event: &pb.SimulationEvent_Phase{Phase: ev},
	})
}

func (p *NATSPublisher) PublishVolcanic(ctx context.Context, ev *pb.VolcanicEvent) error {
	subject := fmt.Sprintf("simulation.%s.volcanic", ev.WorldId)
	return p.publish(ctx, subject, &pb.SimulationEvent{
		Event: &pb.SimulationEvent_Volcanic{Volcanic: ev},
	})
}

func (p *NATSPublisher) PublishAtmosphere(ctx context.Context, ev *pb.AtmosphereUpdate) error {
	subject := fmt.Sprintf("simulation.%s.atmosphere", ev.WorldId)
	return p.publish(ctx, subject, &pb.SimulationEvent{
		Event: &pb.SimulationEvent_Atmosphere{Atmosphere: ev},
	})
}

func (p *NATSPublisher) PublishSnapshot(ctx context.Context, ev *pb.HeightmapSnapshot) error {
	subject := fmt.Sprintf("simulation.%s.snapshot", ev.WorldId)
	return p.publish(ctx, subject, &pb.SimulationEvent{
		Event: &pb.SimulationEvent_Snapshot{Snapshot: ev},
	})
}

func (p *NATSPublisher) Close() error {
	if p.nc != nil {
		p.nc.Close()
	}
	return nil
}
