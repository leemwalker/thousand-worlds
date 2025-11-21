package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOllamaClient() *OllamaClient {
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		url = "http://ollama:11434"
	}

	timeoutStr := os.Getenv("OLLAMA_TIMEOUT")
	timeout := 30 * time.Second
	if timeoutStr != "" {
		if t, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = t
		} else {
			log.Warn().Err(err).Str("timeout", timeoutStr).Msg("Invalid OLLAMA_TIMEOUT, using default 30s")
		}
	}

	return &OllamaClient{
		baseURL: url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponseChunk struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (c *OllamaClient) Generate(prompt string, model string) (string, error) {
	if model == "" {
		model = "llama3" // Default model, can be changed
	}

	reqBody := GenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: true, // User asked to parse streaming response
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Debug().Str("model", model).Int("prompt_len", len(prompt)).Msg("Sending request to Ollama")

	resp, err := c.httpClient.Post(c.baseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request to ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var chunk GenerateResponseChunk
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to decode chunk: %w", err)
		}

		fullResponse.WriteString(chunk.Response)

		if chunk.Done {
			break
		}
	}

	return fullResponse.String(), nil
}
