package ai

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"tw-backend/internal/ai/area"
	"tw-backend/internal/npc/memory"

	"github.com/google/uuid"
)

// MockClient implements area.LLMClient
type MockClient struct {
	Delay time.Duration
}

func (m *MockClient) Generate(prompt string) (string, error) {
	time.Sleep(m.Delay)
	return "Generated Description", nil
}

func TestLoad_ConcurrentAreaDescriptions(t *testing.T) {
	// Setup
	mockClient := &MockClient{Delay: 10 * time.Millisecond} // Fast mock
	cache := area.NewAreaCache()
	service := area.NewAreaDescriptionService(mockClient, cache)

	concurrency := 10
	duration := 2 * time.Second // Short load test for CI

	var errors int64

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(id)))

			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate requesting same few locations to test cache
					locIdx := r.Intn(5)
					data := area.ContextData{
						Location:   memory.Location{WorldID: uuid.Nil, X: float64(locIdx), Y: 0, Z: 0},
						WorldName:  "Test",
						Biome:      "Forest",
						Weather:    "Clear",
						TimeOfDay:  "Day",
						Season:     "Spring",
						Perception: 50,
					}

					// We can't easily check hit/miss from return value alone without inspecting cache or service internals
					// But we can infer from timing or just trust the functional tests.
					// Here we just want to ensure no crashes and decent throughput.

					// To verify cache hit rate, we'd need metrics exposed.
					// For this test, we'll just run it.

					_, err := service.GenerateAreaDescription(context.Background(), data)
					if err != nil {
						atomic.AddInt64(&errors, 1)
					}

					// Sleep a bit
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("Load Test Complete: %v elapsed", elapsed)
	t.Logf("Errors: %d", errors)

	if errors > 0 {
		t.Errorf("Expected 0 errors, got %d", errors)
	}
}
