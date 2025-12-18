package interaction

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"tw-backend/internal/world/interview"
)

// Service manages interactions between characters and NPCs/World
type Service struct {
	interviewService *interview.InterviewService
}

func NewService(interviewService *interview.InterviewService) *Service {
	return &Service{
		interviewService: interviewService,
	}
}

// Response represents a dialogue response
type Response struct {
	Text    string   `json:"text"`
	Options []string `json:"options"`
	NPCName string   `json:"npc_name"`
	NPCID   string   `json:"npc_id"`
}

// ProcessDialogue handles a player talking to an NPC
func (s *Service) ProcessDialogue(ctx context.Context, charID uuid.UUID, targetName string, message string) (*Response, error) {
	// P0 Implementation: Simple echo or hardcoded response
	// Future: Integrate with InterviewService for LLM, or Scripted Dialogue

	// Check if target is "System" or similar for special handling?

	// For P0, we just mock a response or use the interview service if appropriate context matches?
	// The InterviewService is for world creation interviews.
	// We might need a separate DialogueService or expand InterviewService.
	// For this task "Integrate NPC Dialogue", let's assume valid mock integration first.

	return &Response{
		Text:    fmt.Sprintf("Hello! I heard you say: '%s'", message),
		Options: []string{"Ask about quest", "Goodbye"},
		NPCName: targetName,
		NPCID:   "npc-" + strings.ToLower(targetName),
	}, nil
}
