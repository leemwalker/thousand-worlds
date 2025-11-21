package gateway

// AIRequest represents a request for AI generation.
type AIRequest struct {
	ID              string `json:"id"`
	Prompt          string `json:"prompt"`
	Model           string `json:"model"`
	ResponseSubject string `json:"-"` // Subject to publish the response to
}

// AIResponse represents the response from the AI Gateway.
type AIResponse struct {
	ID       string `json:"id"`
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}
