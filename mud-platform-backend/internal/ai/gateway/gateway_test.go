package gateway_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mud-platform-backend/internal/ai/gateway"
)

func TestAIGatewayIntegration(t *testing.T) {
	// Skip if integration tests are not enabled or dependencies are missing
	if os.Getenv("TEST_INTEGRATION") != "true" {
		t.Skip("Skipping integration test. Set TEST_INTEGRATION=true to run.")
	}

	// Connect to NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	nc, err := nats.Connect(natsURL)
	require.NoError(t, err, "Failed to connect to NATS")
	defer nc.Close()

	// Initialize Ollama Client (assumes Ollama is running)
	client := gateway.NewOllamaClient()

	// Start Worker
	gateway.StartWorker(nc, client)

	// Start Listener
	err = gateway.StartListener(nc)
	require.NoError(t, err, "Failed to start listener")

	// Subscribe to response
	responseChan := make(chan gateway.AIResponse, 1)
	sub, err := nc.Subscribe("ai.response.test-id", func(msg *nats.Msg) {
		var resp gateway.AIResponse
		err := json.Unmarshal(msg.Data, &resp)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
			return
		}
		responseChan <- resp
	})
	require.NoError(t, err, "Failed to subscribe to response")
	defer sub.Unsubscribe()

	// Publish Request
	req := gateway.AIRequest{
		ID:     "test-id",
		Prompt: "Say hello in one word.",
		Model:  "tinyllama",
	}
	reqData, err := json.Marshal(req)
	require.NoError(t, err, "Failed to marshal request")

	err = nc.Publish("ai.request.test", reqData)
	require.NoError(t, err, "Failed to publish request")

	// Wait for response
	select {
	case resp := <-responseChan:
		assert.Equal(t, "test-id", resp.ID)
		assert.NotEmpty(t, resp.Response)
		assert.Empty(t, resp.Error)
		t.Logf("Received AI Response: %s", resp.Response)
	case <-time.After(125 * time.Second): // Slightly longer than client timeout
		t.Fatal("Timed out waiting for AI response")
	}
}
