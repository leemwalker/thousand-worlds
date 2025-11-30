package dialogue

import (
	"context"
	"fmt"

	"mud-platform-backend/internal/ai/cache"
	"mud-platform-backend/internal/ai/ollama"
	"mud-platform-backend/internal/ai/prompt"

	"github.com/google/uuid"
)

// NewDialogueService creates a new service instance
func NewDialogueService(
	npcRepo NPCRepository,
	memRepo MemoryRepository,
	relRepo RelationshipRepository,
	desireRepo DesireRepository,
	pb *prompt.PromptBuilder,
	client *ollama.OllamaClient,
	cache *cache.DialogueCache,
) *DialogueService {
	return &DialogueService{
		npcRepo:          npcRepo,
		memoryRepo:       memRepo,
		relationshipRepo: relRepo,
		desireRepo:       desireRepo,
		promptBuilder:    pb,
		ollamaClient:     client,
		dialogueCache:    cache,
	}
}

// GenerateDialogue processes a dialogue request
func (s *DialogueService) GenerateDialogue(ctx context.Context, npcID, speakerID uuid.UUID, input string) (*DialogueResponse, error) {
	// 1. Fetch Context
	npcState, err := s.fetchNPCState(npcID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch npc state: %w", err)
	}

	rel, err := s.relationshipRepo.GetRelationship(npcID, speakerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relationship: %w", err)
	}

	desires, err := s.desireRepo.GetDesireProfile(npcID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch desires: %w", err)
	}

	memories, err := s.memoryRepo.GetMemories(npcID, 5) // Get last 5 relevant
	if err != nil {
		return nil, fmt.Errorf("failed to fetch memories: %w", err)
	}

	// 2. Determine Intent
	intent, intentDesc := s.determineIntent(desires)

	// 3. Detect Drift (if applicable)
	// Assuming we can check if inhabited via some flag or just always check drift metrics
	// For now, let's try to get drift metrics
	drift, _ := s.relationshipRepo.GetDriftMetrics(npcID)

	// 4. Check Cache
	// contextHash := cache.GenerateContextHash(...) // We need to expose this or reimplement
	// For now, skip cache implementation details to focus on flow

	// 5. Build Prompt
	s.promptBuilder.WithNPC(npcState.Name, 30, "Human", "Villager", npcState.Personality, npcState.Mood, intentDesc, 0.0, npcState.Attributes) // Age/Species/Job hardcoded for now or need to be in NPCState
	s.promptBuilder.WithSpeaker("Speaker", rel)                                                                                                // Speaker name?
	s.promptBuilder.WithConversation(intent, input)
	s.promptBuilder.WithMemories(memories)

	if drift != nil {
		// Fetch profiles
		base, err := s.relationshipRepo.GetBehavioralProfile(npcID)
		if err != nil {
			// Log error or ignore drift? Ignore for now
		} else {
			curr, err := s.relationshipRepo.GetCurrentBehavior(npcID)
			if err == nil {
				s.promptBuilder.WithDrift(drift, base, curr)
			}
		}
	}

	promptStr, err := s.promptBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// 6. Call AI
	rawResp, err := s.ollamaClient.Generate(promptStr)
	usedFallback := false
	var finalText string

	if err != nil {
		// Fallback
		finalText = s.getFallbackResponse(rel)
		usedFallback = true
	} else {
		parsed, err := ollama.ParseResponse(rawResp)
		if err != nil {
			finalText = s.getFallbackResponse(rel)
			usedFallback = true
		} else {
			finalText = parsed
		}
	}

	// 7. Analyze Response
	emotion, weight := s.inferEmotionalReaction(finalText)

	// 8. Update State (Async?)
	// We should probably do this async or return it to be handled
	// For now, synchronous
	if !usedFallback {
		go func() {
			s.createConversationMemory(npcID, speakerID, input, finalText, emotion, weight)
			s.updateRelationship(npcID, speakerID, finalText, rel)
		}()
	}

	return &DialogueResponse{
		Text:              finalText,
		EmotionalReaction: emotion,
		EmotionalWeight:   weight,
		UsedFallback:      usedFallback,
	}, nil
}
