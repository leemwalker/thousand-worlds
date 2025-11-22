package desire

import (
	"github.com/google/uuid"
)

// Need Tiers
const (
	TierSurvival    = 1
	TierSocial      = 2
	TierAchievement = 3
	TierPleasure    = 4
)

// Need Names
const (
	// Survival
	NeedHunger = "hunger"
	NeedThirst = "thirst"
	NeedSleep  = "sleep"
	NeedSafety = "safety"

	// Social
	NeedCompanionship = "companionship"
	NeedConversation  = "conversation"
	NeedAffection     = "affection"

	// Achievement
	NeedTaskCompletion      = "task_completion"
	NeedSkillImprovement    = "skill_improvement"
	NeedResourceAcquisition = "resource_acquisition"

	// Pleasure
	NeedCuriosity  = "curiosity"
	NeedHedonism   = "hedonism"
	NeedCreativity = "creativity"
)

// Need represents a single drive
type Need struct {
	Name      string  `json:"name"`
	Value     float64 `json:"value"` // 0-100
	Tier      int     `json:"tier"`
	DecayRate float64 `json:"decay_rate"` // Base increase per hour
}

// DesireProfile holds all needs for an NPC
type DesireProfile struct {
	NPCID uuid.UUID        `json:"npc_id"`
	Needs map[string]*Need `json:"needs"`
}

// NewDesireProfile creates a default profile
func NewDesireProfile(npcID uuid.UUID) *DesireProfile {
	dp := &DesireProfile{
		NPCID: npcID,
		Needs: make(map[string]*Need),
	}

	// Initialize all needs
	// Survival
	dp.Needs[NeedHunger] = &Need{Name: NeedHunger, Tier: TierSurvival, DecayRate: 1.0}
	dp.Needs[NeedThirst] = &Need{Name: NeedThirst, Tier: TierSurvival, DecayRate: 1.5}
	dp.Needs[NeedSleep] = &Need{Name: NeedSleep, Tier: TierSurvival, DecayRate: 1.0}
	dp.Needs[NeedSafety] = &Need{Name: NeedSafety, Tier: TierSurvival, DecayRate: 0.0} // Context dependent

	// Social
	dp.Needs[NeedCompanionship] = &Need{Name: NeedCompanionship, Tier: TierSocial, DecayRate: 0.5}
	dp.Needs[NeedConversation] = &Need{Name: NeedConversation, Tier: TierSocial, DecayRate: 1.0}
	dp.Needs[NeedAffection] = &Need{Name: NeedAffection, Tier: TierSocial, DecayRate: 0.2}

	// Achievement
	dp.Needs[NeedTaskCompletion] = &Need{Name: NeedTaskCompletion, Tier: TierAchievement, DecayRate: 0.0} // Task dependent
	dp.Needs[NeedSkillImprovement] = &Need{Name: NeedSkillImprovement, Tier: TierAchievement, DecayRate: 0.3}
	dp.Needs[NeedResourceAcquisition] = &Need{Name: NeedResourceAcquisition, Tier: TierAchievement, DecayRate: 0.0} // Wealth dependent

	// Pleasure
	dp.Needs[NeedCuriosity] = &Need{Name: NeedCuriosity, Tier: TierPleasure, DecayRate: 0.0}   // Context dependent
	dp.Needs[NeedHedonism] = &Need{Name: NeedHedonism, Tier: TierPleasure, DecayRate: 0.0}     // Boredom dependent
	dp.Needs[NeedCreativity] = &Need{Name: NeedCreativity, Tier: TierPleasure, DecayRate: 0.0} // Occupation dependent

	return dp
}

// Action represents a behavior to fulfill a need
type Action struct {
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	TargetID uuid.UUID `json:"target_id,omitempty"`
	Priority float64   `json:"priority"`
	Source   string    `json:"source_need"`
}
