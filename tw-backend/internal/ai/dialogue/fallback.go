package dialogue

import "tw-backend/internal/npc/relationship"

func (s *DialogueService) getFallbackResponse(rel *relationship.Relationship) string {
	if rel.CurrentAffinity.Affection > 50 {
		return "...smiles warmly but seems distracted."
	}
	if rel.CurrentAffinity.Affection < -20 {
		return "...grunts noncommittally."
	}
	if rel.CurrentAffinity.Fear > 50 {
		return "...looks away nervously."
	}
	return "...nods silently."
}
