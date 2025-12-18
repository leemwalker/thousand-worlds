package combat

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"tw-backend/internal/character"
	"tw-backend/internal/combat/action"
	"tw-backend/internal/game/services/entity"
)

// CombatEvent represents an event occurring during combat resolution
type CombatEvent struct {
	Type      string                 `json:"type"` // attack_start, damage, death
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Service manages the combat system
type Service struct {
	resolver      *action.CombatResolver
	entityService *entity.Service
}

// NewService creates a new combat service
func NewService(entityService *entity.Service) *Service {
	return &Service{
		resolver:      action.NewCombatResolver(),
		entityService: entityService,
	}
}

// JoinCombat adds an entity to the combat session
func (s *Service) JoinCombat(combatant *action.Combatant) {
	s.resolver.AddCombatant(combatant)
}

// JoinCombatFromCharacter creates a combatant from a character and joins combat
func (s *Service) JoinCombatFromCharacter(char *character.Character) {
	combatant := &action.Combatant{
		EntityID:       char.ID,
		MaxHP:          char.SecAttrs.MaxHP,
		CurrentHP:      char.SecAttrs.MaxHP, // TODO: Load from persistence
		MaxStamina:     char.SecAttrs.MaxStamina,
		CurrentStamina: char.SecAttrs.MaxStamina, // TODO: Load from persistence
		Agility:        char.BaseAttrs.Agility,
		CombatState:    action.StateIdle,
	}
	s.JoinCombat(combatant)
}

// QueueAttack queues an attack action
func (s *Service) QueueAttack(attackerID, targetID uuid.UUID) error {
	// Calculate reaction time based on agility (placeholder logic)
	// Base 2 seconds, reduced by agility
	attacker := s.resolver.GetCombatant(attackerID)
	if attacker == nil {
		// Initialize combatant if not found?
		// For now, assume they must have joined.
		return fmt.Errorf("attacker not found in combat")
	}

	// Reaction time: 2000ms - (Agility * 10ms), min 500ms
	agilityMod := time.Duration(attacker.Agility*10) * time.Millisecond
	reactionTime := 2*time.Second - agilityMod
	if reactionTime < 500*time.Millisecond {
		reactionTime = 500 * time.Millisecond
	}

	queueAction := action.NewCombatAction(attackerID, targetID, action.ActionAttack, reactionTime)
	s.resolver.Queue.Enqueue(queueAction)

	return nil
}

// Tick processes one tick of the combat simulation
func (s *Service) Tick(dt time.Duration) []CombatEvent {
	now := time.Now()
	resolved := s.resolver.ProcessTick(now)

	var events []CombatEvent

	for _, act := range resolved {
		// Logic to apply damage would go here (Phase 7.2)
		// For now, we generate events indicating what happened

		evt := CombatEvent{
			Type:      "combat_action",
			Timestamp: now,
			Data: map[string]interface{}{
				"action_id": act.ActionID,
				"actor_id":  act.ActorID,
				"target_id": act.TargetID,
				"type":      string(act.ActionType),
				"resolved":  true,
			},
		}
		events = append(events, evt)

		log.Printf("[COMBAT] Action resolved: %s -> %s (%s)", act.ActorID, act.TargetID, act.ActionType)
	}

	return events
}
