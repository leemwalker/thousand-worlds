package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// BroadcastMessage represents a message to broadcast across instances
type BroadcastMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	SourceID  string      `json:"source_id"`  // Instance ID that sent the message
	TargetIDs []uuid.UUID `json:"target_ids"` // Character IDs to broadcast to
}

// RedisAdapter handles pub/sub for cross-instance communication
// Enables horizontal scaling by broadcasting messages across game server instances
type RedisAdapter struct {
	client     *redis.Client
	instanceID string
	pubsub     *redis.PubSub
	handlers   map[string]func(msg *BroadcastMessage)
}

// NewRedisAdapter creates a new Redis pub/sub adapter
func NewRedisAdapter(client *redis.Client, instanceID string) *RedisAdapter {
	return &RedisAdapter{
		client:     client,
		instanceID: instanceID,
		handlers:   make(map[string]func(msg *BroadcastMessage)),
	}
}

// Subscribe subscribes to a Redis channel and starts listening
// Typically subscribe to "game:broadcast" for cross-instance messages
func (r *RedisAdapter) Subscribe(ctx context.Context, channel string) error {
	r.pubsub = r.client.Subscribe(ctx, channel)

	// Wait for confirmation
	_, err := r.pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe to channel %s: %w", channel, err)
	}

	log.Printf("Subscribed to Redis channel: %s (instance: %s)", channel, r.instanceID)

	// Start message processing goroutine
	go r.processMessages(ctx)

	return nil
}

// Publish publishes a broadcast message to all instances
// Complexity: O(1) Redis PUBLISH
func (r *RedisAdapter) Publish(ctx context.Context, channel string, msg *BroadcastMessage) error {
	msg.SourceID = r.instanceID

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return r.client.Publish(ctx, channel, data).Err()
}

// RegisterHandler registers a handler for a specific message type
func (r *RedisAdapter) RegisterHandler(msgType string, handler func(msg *BroadcastMessage)) {
	r.handlers[msgType] = handler
}

// processMessages processes incoming Redis pub/sub messages
func (r *RedisAdapter) processMessages(ctx context.Context) {
	ch := r.pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case redisMsg := <-ch:
			var msg BroadcastMessage
			if err := json.Unmarshal([]byte(redisMsg.Payload), &msg); err != nil {
				log.Printf("Error unmarshaling broadcast message: %v", err)
				continue
			}

			// Skip messages from self
			if msg.SourceID == r.instanceID {
				continue
			}

			// Route to appropriate handler
			if handler, ok := r.handlers[msg.Type]; ok {
				handler(&msg)
			} else {
				log.Printf("No handler registered for message type: %s", msg.Type)
			}
		}
	}
}

// Close closes the pub/sub connection
func (r *RedisAdapter) Close() error {
	if r.pubsub != nil {
		return r.pubsub.Close()
	}
	return nil
}

// BroadcastToCharacters is a helper to broadcast to specific characters across instances
func (r *RedisAdapter) BroadcastToCharacters(ctx context.Context, channel string, characterIDs []uuid.UUID, msgType string, data interface{}) error {
	msg := &BroadcastMessage{
		Type:      msgType,
		Data:      data,
		TargetIDs: characterIDs,
	}

	return r.Publish(ctx, channel, msg)
}
