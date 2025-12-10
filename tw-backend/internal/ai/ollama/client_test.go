package ollama

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseResponse(t *testing.T) {
	raw := "  \"Hello there!\"  "
	parsed, err := ParseResponse(raw)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if parsed != "Hello there!" {
		t.Errorf("Expected 'Hello there!', got '%s'", parsed)
	}

	// Test Meta-text
	rawMeta := "As an AI, I cannot do that."
	_, err = ParseResponse(rawMeta)
	if err == nil {
		t.Error("Expected error for meta-text, got nil")
	}
}

func TestOllamaClient_Generate(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("Expected path /api/generate, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"response": "Test response", "done": true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")
	resp, err := client.Generate("Test prompt")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp != "Test response" {
		t.Errorf("Expected 'Test response', got '%s'", resp)
	}
}
