package area

import (
	"context"
	"tw-backend/internal/ai/ollama"
	"tw-backend/internal/npc/memory"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGenerateKey(t *testing.T) {
	id := uuid.New()
	k1 := GenerateKey(id, 10, 10, 0, "Sunny", "Noon", "Summer", 20) // Bucket 0
	k2 := GenerateKey(id, 10, 10, 0, "Sunny", "Noon", "Summer", 25) // Bucket 0
	k3 := GenerateKey(id, 10, 10, 0, "Sunny", "Noon", "Summer", 30) // Bucket 1

	if k1 != k2 {
		t.Error("Keys should be same for same bucket")
	}
	if k1 == k3 {
		t.Error("Keys should differ for different buckets")
	}
}

func TestAreaDescriptionService_Generate(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"response": "A beautiful forest.", "done": true}`))
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model")
	cache := NewAreaCache()
	service := NewAreaDescriptionService(client, cache)

	data := ContextData{
		Location:   memory.Location{WorldID: uuid.New(), X: 0, Y: 0, Z: 0},
		WorldName:  "TestWorld",
		Biome:      "Forest",
		Weather:    "Clear",
		TimeOfDay:  "Day",
		Season:     "Spring",
		Perception: 50,
	}

	// 1. First Call (Cache Miss)
	desc, err := service.GenerateAreaDescription(context.Background(), data)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if desc != "A beautiful forest." {
		t.Errorf("Expected 'A beautiful forest.', got '%s'", desc)
	}

	// 2. Second Call (Cache Hit)
	// We can verify this by shutting down server or checking cache directly
	// Let's check cache
	key := GenerateKey(data.Location.WorldID, 0, 0, 0, "Clear", "Day", "Spring", 50)
	cached, ok := cache.Get(key)
	if !ok || cached != "A beautiful forest." {
		t.Error("Cache should contain description")
	}
}

func TestBuildPrompt(t *testing.T) {
	service := &AreaDescriptionService{}
	data := ContextData{
		Location:    memory.Location{WorldID: uuid.New(), X: 10, Y: 20, Z: 5},
		WorldName:   "Azeroth",
		Biome:       "Forest",
		Terrain:     "Hilly",
		Weather:     "Rain",
		Temperature: 15.5,
		TimeOfDay:   "Dusk",
		Season:      "Autumn",
		Entities:    []string{"Wolf", "Bear"},
		Structures:  []string{"Cabin"},
		Perception:  80,
	}

	prompt := service.buildPrompt(data)

	checks := []string{
		"COORDINATES: 10.0, 20.0, 5.0 in Azeroth",
		"BIOME: Forest",
		"WEATHER: Rain (15.5°C)",
		"Wolf",
		"Cabin",
		"OBSERVER PERCEPTION: 80/100",
		"High (76-100): Rich, nuanced detail",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("Prompt missing: %s", check)
		}
	}
}
