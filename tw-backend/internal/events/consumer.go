// Package events provides event publishing and consumption for simulation events.
package events

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	pb "tw-backend/api/proto"
)

// EventConsumerInterface defines the methods for consuming simulation events.
type EventConsumerInterface interface {
	GetEventsFromSequence(ctx context.Context, worldID string, fromSeq uint64) ([]*pb.SimulationEvent, error)
	GetLatestSequence(ctx context.Context) (uint64, error)
	GetStreamInfo(ctx context.Context) (*jetstream.StreamInfo, error)
	Close() error
}

// EventConsumer reads events from NATS JetStream for crash recovery and replay.
type EventConsumer struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	stream jetstream.Stream
}

// Verify EventConsumer implements EventConsumerInterface
var _ EventConsumerInterface = (*EventConsumer)(nil)

// NewEventConsumer creates a consumer connected to NATS JetStream.
func NewEventConsumer(natsURL string) (*EventConsumer, error) {
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

	// Get existing stream (should already exist from publisher)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := js.Stream(ctx, StreamConfig.Name)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("stream get: %w", err)
	}

	return &EventConsumer{
		nc:     nc,
		js:     js,
		stream: stream,
	}, nil
}

// GetEventsFromSequence returns all events after the given sequence number for a world.
// Used for crash recovery to replay events since last snapshot.
func (c *EventConsumer) GetEventsFromSequence(ctx context.Context, worldID string, fromSeq uint64) ([]*pb.SimulationEvent, error) {
	// Create an ephemeral ordered consumer starting from the sequence
	consumerConfig := jetstream.OrderedConsumerConfig{
		FilterSubjects:   []string{fmt.Sprintf("simulation.%s.>", worldID)},
		DeliverPolicy:    jetstream.DeliverByStartSequencePolicy,
		OptStartSeq:      fromSeq + 1, // Start AFTER the fromSeq
		MaxResetAttempts: 3,
	}

	consumer, err := c.js.OrderedConsumer(ctx, StreamConfig.Name, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("create ordered consumer: %w", err)
	}

	var events []*pb.SimulationEvent

	// Fetch messages in batches
	for {
		// Use FetchNoWait to get available messages without blocking forever
		msgs, err := consumer.FetchNoWait(100)
		if err != nil {
			return nil, fmt.Errorf("fetch messages: %w", err)
		}

		count := 0
		for msg := range msgs.Messages() {
			var event pb.SimulationEvent
			if err := proto.Unmarshal(msg.Data(), &event); err != nil {
				// Log and skip malformed messages
				continue
			}
			events = append(events, &event)
			count++
		}

		if err := msgs.Error(); err != nil {
			return nil, fmt.Errorf("message iteration error: %w", err)
		}

		// If we got fewer than batch size, we've consumed all available
		if count < 100 {
			break
		}
	}

	return events, nil
}

// GetLatestSequence returns the last sequence number in the stream.
func (c *EventConsumer) GetLatestSequence(ctx context.Context) (uint64, error) {
	info, err := c.stream.Info(ctx)
	if err != nil {
		return 0, fmt.Errorf("stream info: %w", err)
	}
	return info.State.LastSeq, nil
}

// GetStreamInfo returns information about the event stream.
func (c *EventConsumer) GetStreamInfo(ctx context.Context) (*jetstream.StreamInfo, error) {
	return c.stream.Info(ctx)
}

// Close releases resources.
func (c *EventConsumer) Close() error {
	if c.nc != nil {
		c.nc.Close()
	}
	return nil
}
