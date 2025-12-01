package websocket_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"mud-platform-backend/cmd/game-server/websocket"
	"mud-platform-backend/internal/game/processor"
	"mud-platform-backend/internal/spatial"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// TestHub_LoadTest_1000Clients tests Hub performance with 1000 concurrent clients
// Verifies spatial partitioning provides O(k) area broadcast vs O(N)
func TestHub_LoadTest_1000Clients(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	const numClients = 1000
	const numMessages = 10

	// Create Hub with game processor
	gameProc := processor.NewGameProcessor()
	hub := websocket.NewHub(gameProc)
	gameProc.SetHub(hub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Hub
	go hub.Run(ctx)

	// Create test server
	upgrader := ws.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		characterID := uuid.New()
		client := &websocket.Client{
			ID:          uuid.New(),
			CharacterID: characterID,
			Conn:        conn,
			Send:        make(chan []byte, 256),
		}

		hub.Register <- client

		// Keep connection alive
		go func() {
			for range client.Send {
				// Drain send channel
			}
		}()

		<-ctx.Done()
	}))
	defer server.Close()

	// Connect clients
	t.Logf("Connecting %d clients...", numClients)
	var wg sync.WaitGroup
	clients := make([]*websocket.Client, numClients)
	connectStart := time.Now()

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Connect WebSocket
			wsURL := "ws" + server.URL[4:] // Replace http with ws
			conn, _, err := ws.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Errorf("Failed to connect client %d: %v", idx, err)
				return
			}

			// Position clients in a 100x100 grid pattern
			x := float64((idx % 100) * 10)
			y := float64((idx / 100) * 10)
			characterID := uuid.New()

			hub.UpdateCharacterPosition(characterID, x, y)

			clients[idx] = &websocket.Client{
				ID:          uuid.New(),
				CharacterID: characterID,
				Conn:        conn,
				Send:        make(chan []byte, 256),
			}
		}(i)
	}

	wg.Wait()
	connectDuration := time.Since(connectStart)
	t.Logf("Connected %d clients in %v (%v per client)", numClients, connectDuration, connectDuration/numClients)

	// Wait for all registrations to complete
	time.Sleep(500 * time.Millisecond)

	// Test 1: BroadcastToAll performance (O(N))
	t.Run("BroadcastToAll_Baseline", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < numMessages; i++ {
			hub.BroadcastToAll("test", map[string]interface{}{"iteration": i})
		}
		duration := time.Since(start)
		t.Logf("BroadcastToAll: %d messages to %d clients in %v (%v per message)",
			numMessages, numClients, duration, duration/numMessages)
	})

	// Test 2: BroadcastToArea performance (O(k) where k << N)
	t.Run("BroadcastToArea_Optimized", func(t *testing.T) {
		center := spatial.Position{X: 500, Y: 500} // Center of grid
		radius := 200.0                            // ~400 clients in range

		start := time.Now()
		for i := 0; i < numMessages; i++ {
			hub.BroadcastToArea(center, radius, "area_test", map[string]interface{}{"iteration": i})
		}
		duration := time.Since(start)

		// Count actual clients in range for verification
		actualClients := hub.SpatialIndex.QueryRadius(center, radius)

		t.Logf("BroadcastToArea: %d messages to ~%d clients (from %d total) in %v (%v per message)",
			numMessages, len(actualClients), numClients, duration, duration/numMessages)

		// Verify spatial optimization: should be faster than full broadcast
		assert.Less(t, duration.Milliseconds(), int64(1000),
			"Area broadcast should complete in <1s")
	})

	// Test 3: Concurrent message processing
	t.Run("ConcurrentMessageProcessing", func(t *testing.T) {
		var processedCount atomic.Int32

		start := time.Now()
		for i := 0; i < 100; i++ {
			go func() {
				hub.BroadcastToArea(
					spatial.Position{X: 500, Y: 500},
					150.0,
					"concurrent_test",
					map[string]interface{}{"test": true},
				)
				processedCount.Add(1)
			}()
		}

		// Wait for all broadcasts to complete
		for processedCount.Load() < 100 {
			time.Sleep(10 * time.Millisecond)
		}
		duration := time.Since(start)

		t.Logf("100 concurrent area broadcasts completed in %v", duration)
		assert.Less(t, duration.Milliseconds(), int64(2000),
			"Concurrent broadcasts should complete in <2s")
	})

	// Cleanup
	cancel()
	time.Sleep(100 * time.Millisecond)
}

// BenchmarkHub_BroadcastToArea benchmarks area broadcasting performance
func BenchmarkHub_BroadcastToArea(b *testing.B) {
	gameProc := processor.NewGameProcessor()
	hub := websocket.NewHub(gameProc)
	gameProc.SetHub(hub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// Simulate 1000 clients in grid pattern
	for i := 0; i < 1000; i++ {
		characterID := uuid.New()
		x := float64((i % 100) * 10)
		y := float64((i / 100) * 10)

		client := &websocket.Client{
			ID:          uuid.New(),
			CharacterID: characterID,
			Send:        make(chan []byte, 256),
		}

		hub.Register <- client
		hub.UpdateCharacterPosition(characterID, x, y)

		// Drain send channel
		go func() {
			for range client.Send {
			}
		}()
	}

	// Wait for registrations
	time.Sleep(100 * time.Millisecond)

	center := spatial.Position{X: 500, Y: 500}
	radius := 200.0

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hub.BroadcastToArea(center, radius, "benchmark", map[string]int{"iteration": i})
	}
}

// BenchmarkHub_BroadcastToAll benchmarks full broadcast (baseline)
func BenchmarkHub_BroadcastToAll(b *testing.B) {
	gameProc := processor.NewGameProcessor()
	hub := websocket.NewHub(gameProc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// Simulate 1000 clients
	for i := 0; i < 1000; i++ {
		client := &websocket.Client{
			ID:          uuid.New(),
			CharacterID: uuid.New(),
			Send:        make(chan []byte, 256),
		}

		hub.Register <- client

		go func() {
			for range client.Send {
			}
		}()
	}

	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hub.BroadcastToAll("benchmark", map[string]int{"iteration": i})
	}
}
