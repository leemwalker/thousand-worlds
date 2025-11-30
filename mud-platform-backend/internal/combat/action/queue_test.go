package action

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCombatQueue_Ordering(t *testing.T) {
	q := NewCombatQueue()
	actorID := uuid.New()
	targetID := uuid.New()

	// Create actions with different reaction times
	// Action 1: 1000ms
	a1 := NewCombatAction(actorID, targetID, ActionAttack, 1000*time.Millisecond)
	// Action 2: 500ms (Should be first)
	a2 := NewCombatAction(actorID, targetID, ActionDefend, 500*time.Millisecond)
	// Action 3: 1500ms (Should be last)
	a3 := NewCombatAction(actorID, targetID, ActionFlee, 1500*time.Millisecond)

	// Enqueue in mixed order
	q.Enqueue(a1)
	q.Enqueue(a3)
	q.Enqueue(a2)

	// Dequeue and verify order
	first := q.Dequeue()
	if first != a2 {
		t.Errorf("Expected a2 (Defend, 500ms) first, got %v", first.ActionType)
	}

	second := q.Dequeue()
	if second != a1 {
		t.Errorf("Expected a1 (Attack, 1000ms) second, got %v", second.ActionType)
	}

	third := q.Dequeue()
	if third != a3 {
		t.Errorf("Expected a3 (Flee, 1500ms) third, got %v", third.ActionType)
	}
}

func TestCombatQueue_Empty(t *testing.T) {
	q := NewCombatQueue()
	if q.Dequeue() != nil {
		t.Error("Expected nil from empty queue")
	}
	if q.Peek() != nil {
		t.Error("Expected nil peek from empty queue")
	}
}

// TestCombatQueue_MultiCombatantScenario tests the specific scenario from requirements:
// - Player A queues Normal Attack at T=0 (Agility 60, reaction time 820ms)
// - NPC B queues Quick Attack at T=100ms (Agility 40, reaction time 704ms)
// - NPC C queues Heavy Attack at T=50ms (Agility 70, reaction time 1185ms)
// Expected execution order:
// 1. T=804ms: NPC B's Quick Attack (queued at 100ms + 704ms)
// 2. T=820ms: Player A's Normal Attack (queued at 0ms + 820ms)
// 3. T=1235ms: NPC C's Heavy Attack (queued at 50ms + 1185ms)
func TestCombatQueue_MultiCombatantScenario(t *testing.T) {
	q := NewCombatQueue()
	baseTime := time.Now()

	playerA := uuid.New()
	npcB := uuid.New()
	npcC := uuid.New()

	// Player A: Normal Attack at T=0, Agility 60
	// Reaction time = 1000ms * (1 - (60/100)*0.3) = 1000 * 0.82 = 820ms
	rtPlayerA := CalculateReactionTime(ActionAttack, AttackNormal, 60, 1.0)
	actionA := &CombatAction{
		ActionID:     uuid.New(),
		ActorID:      playerA,
		TargetID:     npcB,
		ActionType:   ActionAttack,
		ReactionTime: rtPlayerA,
		QueuedAt:     baseTime,
		ExecuteAt:    baseTime.Add(rtPlayerA),
		Resolved:     false,
	}

	// NPC B: Quick Attack at T=100ms, Agility 40
	// Reaction time = 800ms * (1 - (40/100)*0.3) = 800 * 0.88 = 704ms
	rtNpcB := CalculateReactionTime(ActionAttack, AttackQuick, 40, 1.0)
	queueTimeB := baseTime.Add(100 * time.Millisecond)
	actionB := &CombatAction{
		ActionID:     uuid.New(),
		ActorID:      npcB,
		TargetID:     playerA,
		ActionType:   ActionAttack,
		ReactionTime: rtNpcB,
		QueuedAt:     queueTimeB,
		ExecuteAt:    queueTimeB.Add(rtNpcB),
		Resolved:     false,
	}

	// NPC C: Heavy Attack at T=50ms, Agility 70
	// Reaction time = 1500ms * (1 - (70/100)*0.3) = 1500 * 0.79 = 1185ms
	rtNpcC := CalculateReactionTime(ActionAttack, AttackHeavy, 70, 1.0)
	queueTimeC := baseTime.Add(50 * time.Millisecond)
	actionC := &CombatAction{
		ActionID:     uuid.New(),
		ActorID:      npcC,
		TargetID:     playerA,
		ActionType:   ActionAttack,
		ReactionTime: rtNpcC,
		QueuedAt:     queueTimeC,
		ExecuteAt:    queueTimeC.Add(rtNpcC),
		Resolved:     false,
	}

	// Enqueue actions
	q.Enqueue(actionA)
	q.Enqueue(actionB)
	q.Enqueue(actionC)

	// Dequeue and verify execution order
	first := q.Dequeue()
	if first.ActorID != npcB {
		t.Errorf("Expected NPC B first (executes at ~804ms), got %v", first.ActorID)
	}
	expectedFirstExec := queueTimeB.Add(rtNpcB)
	if !first.ExecuteAt.Equal(expectedFirstExec) {
		t.Errorf("Expected first action at %v, got %v", expectedFirstExec, first.ExecuteAt)
	}

	second := q.Dequeue()
	if second.ActorID != playerA {
		t.Errorf("Expected Player A second (executes at ~820ms), got %v", second.ActorID)
	}
	expectedSecondExec := baseTime.Add(rtPlayerA)
	if !second.ExecuteAt.Equal(expectedSecondExec) {
		t.Errorf("Expected second action at %v, got %v", expectedSecondExec, second.ExecuteAt)
	}

	third := q.Dequeue()
	if third.ActorID != npcC {
		t.Errorf("Expected NPC C third (executes at ~1235ms), got %v", third.ActorID)
	}
	expectedThirdExec := queueTimeC.Add(rtNpcC)
	if !third.ExecuteAt.Equal(expectedThirdExec) {
		t.Errorf("Expected third action at %v, got %v", expectedThirdExec, third.ExecuteAt)
	}

	// Verify order by execution time
	if !first.ExecuteAt.Before(second.ExecuteAt) {
		t.Error("First action should execute before second")
	}
	if !second.ExecuteAt.Before(third.ExecuteAt) {
		t.Error("Second action should execute before third")
	}
}
