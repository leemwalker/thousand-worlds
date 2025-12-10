package dialogue

import (
	"mud-platform-backend/internal/npc/relationship"
	"testing"
)

func TestGetFallbackResponse(t *testing.T) {
	s := &DialogueService{}

	tests := []struct {
		name     string
		rel      *relationship.Relationship
		expected string
	}{
		{
			name: "High Affection",
			rel: &relationship.Relationship{
				CurrentAffinity: relationship.Affinity{Affection: 60},
			},
			expected: "...smiles warmly but seems distracted.",
		},
		{
			name: "Low Affection",
			rel: &relationship.Relationship{
				CurrentAffinity: relationship.Affinity{Affection: -30},
			},
			expected: "...grunts noncommittally.",
		},
		{
			name: "High Fear",
			rel: &relationship.Relationship{
				CurrentAffinity: relationship.Affinity{Fear: 60},
			},
			expected: "...looks away nervously.",
		},
		{
			name: "Neutral",
			rel: &relationship.Relationship{
				CurrentAffinity: relationship.Affinity{Affection: 0},
			},
			expected: "...nods silently.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.getFallbackResponse(tt.rel)
			if got != tt.expected {
				t.Errorf("Expected: %s, Got: %s", tt.expected, got)
			}
		})
	}
}
