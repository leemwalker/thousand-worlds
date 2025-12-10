package gateway

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerate_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/generate", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Send streaming response
		w.Write([]byte(`{"response": "Hello", "done": false}` + "\n"))
		w.Write([]byte(`{"response": " World", "done": true}` + "\n"))
	}))
	defer server.Close()

	// Save original env
	originalURL := os.Getenv("OLLAMA_URL")
	os.Setenv("OLLAMA_URL", server.URL)
	defer os.Setenv("OLLAMA_URL", originalURL)

	client := NewOllamaClient()
	resp, err := client.Generate("prompt", "model")

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", resp)
}

func TestGenerate_HTTPError(t *testing.T) {
	// No server running at this URL
	originalURL := os.Getenv("OLLAMA_URL")
	os.Setenv("OLLAMA_URL", "http://localhost:12345")
	defer os.Setenv("OLLAMA_URL", originalURL)

	client := NewOllamaClient()
	// Reduce timeout for fast fail
	client.httpClient.Timeout = 100 * time.Millisecond

	_, err := client.Generate("prompt", "model")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestGenerate_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	originalURL := os.Getenv("OLLAMA_URL")
	os.Setenv("OLLAMA_URL", server.URL)
	defer os.Setenv("OLLAMA_URL", originalURL)

	client := NewOllamaClient()
	_, err := client.Generate("prompt", "model")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ollama returned non-200 status: 500")
}

func TestGenerate_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"response": "Hello", "done": false}` + "\n"))
		w.Write([]byte(`{malformed json` + "\n"))
	}))
	defer server.Close()

	originalURL := os.Getenv("OLLAMA_URL")
	os.Setenv("OLLAMA_URL", server.URL)
	defer os.Setenv("OLLAMA_URL", originalURL)

	client := NewOllamaClient()
	_, err := client.Generate("prompt", "model")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode chunk")
}
