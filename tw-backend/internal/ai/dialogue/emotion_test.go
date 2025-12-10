package dialogue

import (
	"testing"
)

func TestInferEmotionalReaction(t *testing.T) {
	s := &DialogueService{}

	tests := []struct {
		input    string
		expected string
		weight   float64
	}{
		{"This is wonderful news!", "joy", 0.7},
		{"I am furious about this.", "anger", 0.8},
		{"That is terrifying...", "fear", 0.7},
		{"It is so sad.", "sadness", 0.6},
		{"Wow!!", "excited", 0.6},
		{"What??", "confused", 0.5},
		{"Just a normal sentence.", "neutral", 0.1},
	}

	for _, tt := range tests {
		emo, weight := s.inferEmotionalReaction(tt.input)
		if emo != tt.expected {
			t.Errorf("Input: %s, Expected: %s, Got: %s", tt.input, tt.expected, emo)
		}
		if weight != tt.weight {
			t.Errorf("Input: %s, Expected Weight: %f, Got: %f", tt.input, tt.weight, weight)
		}
	}
}
