package dialogue

import (
	"mud-platform-backend/internal/npc/desire"

	"github.com/google/uuid"
)

func (s *DialogueService) fetchNPCState(id uuid.UUID) (*NPCState, error) {
	// This would typically call the repo
	// For MVP, we assume the repo returns the structs we need
	// We might need to map from repo types to our NPCState if they differ
	// But assuming direct mapping for now

	// Mocking the fetch since we don't have the full repo implementation details in this file
	// In real code, this calls s.npcRepo.GetNPC(id), etc.

	// Let's assume s.npcRepo has these methods as defined in interface
	char, err := s.npcRepo.GetNPC(id)
	if err != nil {
		return nil, err
	}
	p, err := s.npcRepo.GetPersonality(id)
	if err != nil {
		return nil, err
	}
	m, err := s.npcRepo.GetMood(id)
	if err != nil {
		return nil, err
	}

	return &NPCState{
		Name:        char.Name,
		Personality: p,
		Mood:        m,
		Attributes:  char.BaseAttrs,
	}, nil
}

func (s *DialogueService) determineIntent(profile *desire.DesireProfile) (string, string) {
	// Find highest need
	var topNeed *desire.Need
	highestVal := -1.0

	for _, n := range profile.Needs {
		if n.Value > highestVal {
			highestVal = n.Value
			topNeed = n
		}
	}

	if topNeed == nil {
		return IntentNeutral, "calm and observant"
	}

	switch topNeed.Name {
	case desire.NeedHunger:
		if highestVal > 70 {
			return IntentSeekingFood, "desperately hungry and hoping to find food soon"
		}
	case desire.NeedCompanionship:
		if highestVal > 60 {
			return IntentSeekingConnection, "feeling lonely and want to connect with someone"
		}
	case desire.NeedSafety:
		if highestVal > 50 {
			return IntentSeekingSafety, "nervous and looking for reassurance"
		}
	case desire.NeedTaskCompletion:
		if highestVal > 60 {
			return IntentFocusedOnGoal, "focused on your current goal and slightly distracted"
		}
	}

	return IntentNeutral, "calm and observant"
}
