package interaction

import (
	"tw-backend/internal/npc/personality"
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

func TestGenerateGreeting_MediumAffection(t *testing.T) {
	greeting := GenerateGreeting(40) // Between 20-60
	found := false
	for _, g := range GreetingTemplates["medium_affection"] {
		if g == greeting {
			found = true
			break
		}
	}
	if !found {
		t.Error("Greeting did not match medium affection templates")
	}
}

func TestGenerateGreeting_LowAffection(t *testing.T) {
	greeting := GenerateGreeting(10) // < 20
	found := false
	for _, g := range GreetingTemplates["low_affection"] {
		if g == greeting {
			found = true
			break
		}
	}
	if !found {
		t.Error("Greeting did not match low affection templates")
	}
}

func TestProcessTurn_GreetingStage(t *testing.T) {
	conv := &Conversation{
		CurrentStage: StageGreeting,
	}

	response := ProcessTurn(conv, "hello")

	if conv.CurrentStage != StageTopic {
		t.Errorf("Expected stage to advance to Topic, got %s", conv.CurrentStage)
	}
	if response == "" {
		t.Error("Response should not be empty")
	}
}

func TestProcessTurn_TopicStage(t *testing.T) {
	conv := &Conversation{
		CurrentStage: StageTopic,
	}

	response := ProcessTurn(conv, "about the weather")

	if conv.CurrentStage != StageResponse {
		t.Errorf("Expected stage to advance to Response, got %s", conv.CurrentStage)
	}
	if response == "" {
		t.Error("Response should not be empty")
	}
}

func TestProcessTurn_ResponseStage(t *testing.T) {
	conv := &Conversation{
		CurrentStage: StageResponse,
	}

	response := ProcessTurn(conv, "I agree")

	if conv.CurrentStage != StageEnd {
		t.Errorf("Expected stage to advance to End, got %s", conv.CurrentStage)
	}
	if response == "" {
		t.Error("Response should not be empty")
	}
}

func TestProcessTurn_EndStage(t *testing.T) {
	conv := &Conversation{
		CurrentStage: StageEnd,
	}

	response := ProcessTurn(conv, "goodbye")

	// Should return conversation ended
	if response == "" {
		t.Error("Response should not be empty")
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
