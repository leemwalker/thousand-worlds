package interaction

import (
	"mud-platform-backend/internal/npc/personality"
	"testing"
)

func TestGenerateGreeting(t *testing.T) {
	// High Affection
	greeting := GenerateGreeting(80)
	// Should be from high_affection list
	found := false
	for _, g := range GreetingTemplates["high_affection"] {
		if g == greeting {
			found = true
			break
		}
	}
	if !found {
		t.Error("Greeting did not match high affection templates")
	}
}

func TestGenerateResponse(t *testing.T) {
	p := personality.NewPersonality()
	p.Agreeableness.Value = 90.0 // High Agreeableness

	affection := 50
	topic := Topic{Text: "test topic"}

	text, responseType := GenerateResponse(p, affection, topic)

	if responseType != "agreeable" {
		t.Errorf("Expected agreeable response, got %s", responseType)
	}

	if text == "" {
		t.Error("Generated response text is empty")
	}
}
