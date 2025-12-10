package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisAdapter_PublishSubscribe(t *testing.T) {
	// This test requires a running Redis instance
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available")
	}

	// Create two adapters (simulating two instances)
	instance1 := NewRedisAdapter(client, "instance-1")
	instance2 := NewRedisAdapter(client, "instance-2")

	defer instance1.Close()
	defer instance2.Close()

	channel := "test:broadcast"

	// Subscribe instance2
	err := instance2.Subscribe(ctx, channel)
	require.NoError(t, err)

	// Register handler on instance2
	received := make(chan *BroadcastMessage, 1)
	instance2.RegisterHandler("test_message", func(msg *BroadcastMessage) {
		received <- msg
	})

	// Publish from instance1
	testData := map[string]string{"test": "data"}
	msg := &BroadcastMessage{
		Type:      "test_message",
		Data:      testData,
		TargetIDs: []uuid.UUID{uuid.New()},
	}

	err = instance1.Publish(ctx, channel, msg)
	require.NoError(t, err)

	// Wait for message
	select {
	case receivedMsg := <-received:
		assert.Equal(t, "test_message", receivedMsg.Type)
		assert.Equal(t, "instance-1", receivedMsg.SourceID)
		assert.NotEmpty(t, receivedMsg.TargetIDs)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestRedisAdapter_SelfMessagesIgnored(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available")
	}

	instance := NewRedisAdapter(client, "instance-1")
	defer instance.Close()

	channel := "test:self"

	err := instance.Subscribe(ctx, channel)
	require.NoError(t, err)

	// Register handler
	received := make(chan *BroadcastMessage, 1)
	instance.RegisterHandler("self_test", func(msg *BroadcastMessage) {
		received <- msg
	})

	// Publish from same instance
	msg := &BroadcastMessage{
		Type: "self_test",
		Data: "test",
	}

	err = instance.Publish(ctx, channel, msg)
	require.NoError(t, err)

	// Should NOT receive own message
	select {
	case <-received:
		t.Fatal("Should not receive message from self")
	case <-time.After(500 * time.Millisecond):
		// Expected - no message received
	}
}
