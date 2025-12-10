package ollama

import (
	"errors"
	"strings"
)

// ParseResponse extracts and cleans the dialogue text
func ParseResponse(raw string) (string, error) {
	sanitized := SanitizeResponse(raw)
	if err := ValidateResponse(sanitized); err != nil {
		return "", err
	}
	return sanitized, nil
}

// SanitizeResponse cleans up the output
func SanitizeResponse(raw string) string {
	// Remove quotes if they wrap the entire response
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "\"")
	trimmed = strings.TrimSuffix(trimmed, "\"")

	// Remove markdown code blocks if present (sometimes models do this)
	trimmed = strings.ReplaceAll(trimmed, "```", "")

	return strings.TrimSpace(trimmed)
}

// ValidateResponse checks if the response is valid
func ValidateResponse(text string) error {
	if len(text) == 0 {
		return errors.New("empty response")
	}
	if len(text) > 500 {
		return errors.New("response too long")
	}

	// Check for AI meta-text
	lower := strings.ToLower(text)
	if strings.Contains(lower, "as an ai") || strings.Contains(lower, "i cannot") {
		return errors.New("response contains AI meta-text")
	}

	return nil
}
