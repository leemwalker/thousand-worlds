package desire

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetBestAction_Hunger(t *testing.T) {
	npcID := uuid.New()

	// High hunger (>=70) -> Seek Food
	need := &Need{Name: NeedHunger, Value: 75}
	action := GetBestAction(need, npcID)
	assert.Equal(t, "Seek Food", action.Name)
	assert.Equal(t, "seek_food", action.Type)

	// Low hunger (<70) -> Eat Snack
	need.Value = 50
	action = GetBestAction(need, npcID)
	assert.Equal(t, "Eat Snack", action.Name)
	assert.Equal(t, "eat", action.Type)
}

func TestGetBestAction_Thirst(t *testing.T) {
	npcID := uuid.New()

	// High thirst (>=60) -> Seek Water
	need := &Need{Name: NeedThirst, Value: 65}
	action := GetBestAction(need, npcID)
	assert.Equal(t, "Seek Water", action.Name)
	assert.Equal(t, "seek_water", action.Type)

	// Low thirst (<60) -> Drink
	need.Value = 40
	action = GetBestAction(need, npcID)
	assert.Equal(t, "Drink", action.Name)
	assert.Equal(t, "drink", action.Type)
}

func TestGetBestAction_Sleep(t *testing.T) {
	npcID := uuid.New()

	// High sleep need (>=75) -> Seek Bed
	need := &Need{Name: NeedSleep, Value: 80}
	action := GetBestAction(need, npcID)
	assert.Equal(t, "Seek Bed", action.Name)
	assert.Equal(t, "seek_bed", action.Type)

	// Low sleep need (<75) -> Rest
	need.Value = 50
	action = GetBestAction(need, npcID)
	assert.Equal(t, "Rest", action.Name)
	assert.Equal(t, "rest", action.Type)
}

func TestGetBestAction_Safety(t *testing.T) {
	npcID := uuid.New()

	// High danger (>=50) -> Flee
	need := &Need{Name: NeedSafety, Value: 60}
	action := GetBestAction(need, npcID)
	assert.Equal(t, "Flee", action.Name)
	assert.Equal(t, "flee", action.Type)

	// Low danger (<50) -> Seek Shelter
	need.Value = 30
	action = GetBestAction(need, npcID)
	assert.Equal(t, "Seek Shelter", action.Name)
	assert.Equal(t, "seek_shelter", action.Type)
}

func TestGetBestAction_Social(t *testing.T) {
	npcID := uuid.New()

	// Companionship
	need := &Need{Name: NeedCompanionship, Value: 70}
	action := GetBestAction(need, npcID)
	assert.Equal(t, "Seek Company", action.Name)
	assert.Equal(t, "seek_company", action.Type)

	// Conversation
	need = &Need{Name: NeedConversation, Value: 55}
	action = GetBestAction(need, npcID)
	assert.Equal(t, "Initiate Dialogue", action.Name)
	assert.Equal(t, "talk", action.Type)

	// Affection
	need = &Need{Name: NeedAffection, Value: 75}
	action = GetBestAction(need, npcID)
	assert.Equal(t, "Give Gift", action.Name)
	assert.Equal(t, "give_gift", action.Type)
}

func TestGetBestAction_Achievement(t *testing.T) {
	npcID := uuid.New()

	// Task Completion
	need := &Need{Name: NeedTaskCompletion, Value: 60}
	action := GetBestAction(need, npcID)
	assert.Equal(t, "Work on Task", action.Name)
	assert.Equal(t, "work", action.Type)

	// Skill Improvement
	need = &Need{Name: NeedSkillImprovement, Value: 50}
	action = GetBestAction(need, npcID)
	assert.Equal(t, "Practice Skill", action.Name)
	assert.Equal(t, "practice", action.Type)

	// Resource Acquisition
	need = &Need{Name: NeedResourceAcquisition, Value: 45}
	action = GetBestAction(need, npcID)
	assert.Equal(t, "Gather Resources", action.Name)
	assert.Equal(t, "gather", action.Type)
}

func TestGetBestAction_Pleasure(t *testing.T) {
	npcID := uuid.New()

	// Curiosity
	need := &Need{Name: NeedCuriosity, Value: 55}
	action := GetBestAction(need, npcID)
	assert.Equal(t, "Explore", action.Name)
	assert.Equal(t, "explore", action.Type)

	// Hedonism
	need = &Need{Name: NeedHedonism, Value: 40}
	action = GetBestAction(need, npcID)
	assert.Equal(t, "Seek Entertainment", action.Name)
	assert.Equal(t, "seek_fun", action.Type)

	// Creativity
	need = &Need{Name: NeedCreativity, Value: 65}
	action = GetBestAction(need, npcID)
	assert.Equal(t, "Create Art", action.Name)
	assert.Equal(t, "create", action.Type)
}

func TestGetBestAction_ThresholdBoundaries(t *testing.T) {
	npcID := uuid.New()

	// Test exactly at threshold
	need := &Need{Name: NeedHunger, Value: 70}
	action := GetBestAction(need, npcID)
	assert.Equal(t, "Seek Food", action.Name)

	// Test one below threshold
	need.Value = 69
	action = GetBestAction(need, npcID)
	assert.Equal(t, "Eat Snack", action.Name)
}

func TestGetBestAction_PrioritySet(t *testing.T) {
	npcID := uuid.New()

	need := &Need{Name: NeedHunger, Value: 85}
	action := GetBestAction(need, npcID)

	assert.Equal(t, 85.0, action.Priority)
	assert.Equal(t, NeedHunger, action.Source)
}
