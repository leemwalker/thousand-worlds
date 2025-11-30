package action

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
)

// BenchmarkProcessTick_100Combatants tests processing 100 concurrent combatants
// Target: < 100ms per tick
func BenchmarkProcessTick_100Combatants(b *testing.B) {
	resolver := NewCombatResolver()
	now := time.Now()

	// Create 100 combatants
	combatantIDs := make([]uuid.UUID, 100)
	for i := 0; i < 100; i++ {
		id := uuid.New()
		combatantIDs[i] = id

		combatant := &Combatant{
			EntityID:       id,
			CurrentStamina: 1000,
			MaxStamina:     1000,
			CurrentHP:      100,
			MaxHP:          100,
			Agility:        rand.Intn(100),
			LastActionTime: now.Add(-1 * time.Second),
			CombatState:    StateInCombat,
			StatusEffects:  []StatusEffect{},
		}
		resolver.AddCombatant(combatant)

		// Queue one action per combatant
		action := &CombatAction{
			ActionID:   uuid.New(),
			ActorID:    id,
			TargetID:   combatantIDs[rand.Intn(100)],
			ActionType: ActionAttack,
			QueuedAt:   now.Add(time.Duration(rand.Intn(1000)) * time.Millisecond),
			ExecuteAt:  now.Add(time.Duration(rand.Intn(2000)) * time.Millisecond),
			Resolved:   false,
		}
		resolver.Queue.Enqueue(action)
	}

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Process a tick at a future time when all actions are ready
		resolver.ProcessTick(now.Add(3 * time.Second))

		// Re-queue actions for next iteration
		if i < b.N-1 {
			for _, id := range combatantIDs {
				action := &CombatAction{
					ActionID:   uuid.New(),
					ActorID:    id,
					TargetID:   combatantIDs[rand.Intn(100)],
					ActionType: ActionAttack,
					QueuedAt:   now.Add(time.Duration(rand.Intn(1000)) * time.Millisecond),
					ExecuteAt:  now.Add(time.Duration(rand.Intn(2000)) * time.Millisecond),
					Resolved:   false,
				}
				resolver.Queue.Enqueue(action)
			}
		}
	}
}

// BenchmarkQueueOperations tests queue enqueue/dequeue performance
func BenchmarkQueueOperations(b *testing.B) {
	q := NewCombatQueue()
	now := time.Now()

	actions := make([]*CombatAction, 1000)
	for i := 0; i < 1000; i++ {
		actions[i] = &CombatAction{
			ActionID:   uuid.New(),
			ActorID:    uuid.New(),
			TargetID:   uuid.New(),
			ActionType: ActionAttack,
			QueuedAt:   now,
			ExecuteAt:  now.Add(time.Duration(rand.Intn(5000)) * time.Millisecond),
			Resolved:   false,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Enqueue all
		for _, action := range actions {
			q.Enqueue(action)
		}

		// Dequeue all
		for j := 0; j < 1000; j++ {
			q.Dequeue()
		}
	}
}

// BenchmarkReactionTimeCalculation tests reaction time calculation performance
func BenchmarkReactionTimeCalculation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateReactionTime(ActionAttack, AttackNormal, 50, 1.0)
	}
}

// BenchmarkCanQueueAction tests validation performance
func BenchmarkCanQueueAction(b *testing.B) {
	now := time.Now()
	combatant := &Combatant{
		EntityID:       uuid.New(),
		CurrentStamina: 100,
		MaxStamina:     100,
		CurrentHP:      100,
		MaxHP:          100,
		Agility:        50,
		LastActionTime: now.Add(-1 * time.Second),
		CurrentAction:  nil,
		CombatState:    StateInCombat,
		StatusEffects:  []StatusEffect{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CanQueueAction(combatant, ActionAttack, AttackNormal, now)
	}
}

// TestPerformance_100CombatantsPerTick verifies <100ms per tick with 100 combatants
func TestPerformance_100CombatantsPerTick(t *testing.T) {
	resolver := NewCombatResolver()
	now := time.Now()

	// Create 100 combatants
	for i := 0; i < 100; i++ {
		id := uuid.New()

		combatant := &Combatant{
			EntityID:       id,
			CurrentStamina: 1000,
			MaxStamina:     1000,
			CurrentHP:      100,
			MaxHP:          100,
			Agility:        rand.Intn(100),
			LastActionTime: now.Add(-1 * time.Second),
			CombatState:    StateInCombat,
			StatusEffects:  []StatusEffect{},
		}
		resolver.AddCombatant(combatant)

		// Queue one action per combatant, all ready to execute
		action := &CombatAction{
			ActionID:   uuid.New(),
			ActorID:    id,
			TargetID:   uuid.New(),
			ActionType: ActionAttack,
			QueuedAt:   now.Add(-1 * time.Second),
			ExecuteAt:  now.Add(-100 * time.Millisecond), // All ready
			Resolved:   false,
		}
		resolver.Queue.Enqueue(action)
	}

	// Measure tick processing time
	start := time.Now()
	resolved := resolver.ProcessTick(now)
	elapsed := time.Since(start)

	// Verify all actions were processed
	if len(resolved) != 100 {
		t.Errorf("Expected 100 actions resolved, got %d", len(resolved))
	}

	// Verify performance: should be < 100ms
	maxDuration := 100 * time.Millisecond
	if elapsed > maxDuration {
		t.Errorf("Tick took %v, expected < %v with 100 combatants", elapsed, maxDuration)
	}

	t.Logf("Processed 100 combatants in %v", elapsed)
}
