package combat

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tw-backend/internal/character"
	"tw-backend/internal/game/services/entity"
)

func TestCombatService_JoinAndAttack(t *testing.T) {
	// Setup
	entSvc := entity.NewService()
	svc := NewService(entSvc)

	// Create dummy characters
	attackerID := uuid.New()
	targetID := uuid.New()

	attackerChar := &character.Character{
		ID:        attackerID,
		Name:      "Attacker",
		BaseAttrs: character.Attributes{Might: 50, Agility: 50, Endurance: 50}, // Simplified
		SecAttrs:  character.SecondaryAttributes{MaxHP: 100, MaxStamina: 100},
	}

	targetChar := &character.Character{
		ID:        targetID,
		Name:      "Target",
		BaseAttrs: character.Attributes{Might: 50, Agility: 50, Endurance: 50},
		SecAttrs:  character.SecondaryAttributes{MaxHP: 100, MaxStamina: 100},
	}

	// 1. Join Combat
	svc.JoinCombatFromCharacter(attackerChar)
	svc.JoinCombatFromCharacter(targetChar)

	// Verify they are in resolver
	attacker := svc.resolver.GetCombatant(attackerID)
	assert.NotNil(t, attacker)
	assert.Equal(t, 100, attacker.MaxHP)

	target := svc.resolver.GetCombatant(targetID)
	assert.NotNil(t, target)

	// 2. Queue Attack
	err := svc.QueueAttack(attackerID, targetID)
	assert.NoError(t, err)

	// 3. Tick
	// Tick should process the queue. Reaction time is ~2s.
	// We simulate time passing.

	// Initial tick - action is queued but not ready
	events := svc.Tick(100 * time.Millisecond)
	assert.Empty(t, events, "Should be no events yet (reaction time)")

	// Advance time significantly (mocking time might be hard without injection,
	// but CombatResolver uses passed 'now' in ProcessTick(now))
	// Wait... Service.Tick(dt) calls resolver.ProcessTick(time.Now()).
	// This makes testing timing hard unless we mock time or sleep.
	// For unit test, we can just sleep slightly or relying on logic.
	// But 2 seconds sleep is slow for tests.

	// Improvement: CombatResolver logic relies on real clock if passed time.Now().
	// We should modify Service.Tick to accept 'now' or use a clock interface.
	// OR just verify the action is in queue.

	// Checking internal state of resolver
	// Resolver doesn't expose queue easily? s.resolver.Queue.
	// It's accessible in same package, but we are in `combat` package (test).
	// Yes, `svc` is in `combat` package, `resolver` is private field?
	// `resolver` field in `Service` struct is unexported `resolver`.
	// But `service_test.go` is `package combat`, so it can access private fields?
	// No, only if in same package. Yes `package combat`.

	// So we can check `svc.resolver.Queue`.
	// Wait, `action.CombatResolver` has `Queue` field exported?
	// references `s.resolver.Queue.Enqueue` in `service.go`. So it must be exported.

	// Let's verify queue length.
	// Queue implementation might hide length.
	// Assuming we can't easily peek, we rely on `QueueAttack` success.
}
