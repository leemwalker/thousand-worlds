package personality

import (
	"testing"
	"time"

	"mud-platform-backend/internal/npc/memory"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestApplyExperienceModifiers_ChildhoodTrauma(t *testing.T) {
	p := NewPersonality()
	initialNeuroticism := p.Neuroticism.Value

	mem := memory.Memory{
		ID:        uuid.New(),
		NPCID:     uuid.New(),
		Type:      memory.MemoryTypeEvent,
		Timestamp: time.Now(),
		Tags:      []string{EventTrauma},
		Content: memory.EventContent{
			Description: "Lost parents at young age",
		},
	}

	ApplyExperienceModifiers(p, []memory.Memory{mem})

	// Trauma should increase neuroticism (+20)
	assert.Greater(t, p.Neuroticism.Value, initialNeuroticism, "Trauma should increase neuroticism")
	assert.InDelta(t, initialNeuroticism+20, p.Neuroticism.Value, 0.1)
}

func TestApplyExperienceModifiers_Nurtured(t *testing.T) {
	p := NewPersonality()
	initialAgreeableness := p.Agreeableness.Value

	mem := memory.Memory{
		ID:        uuid.New(),
		NPCID:     uuid.New(),
		Type:      memory.MemoryTypeEvent,
		Timestamp: time.Now(),
		Tags:      []string{EventNurtured},
		Content: memory.EventContent{
			Description: "Raised in loving family",
		},
	}

	ApplyExperienceModifiers(p, []memory.Memory{mem})

	assert.Greater(t, p.Agreeableness.Value, initialAgreeableness, "Nurturing should increase agreeableness")
	assert.InDelta(t, initialAgreeableness+15, p.Agreeableness.Value, 0.1)
}

func TestApplyExperienceModifiers_Challenged(t *testing.T) {
	p := NewPersonality()
	initialConscientiousness := p.Conscientiousness.Value

	mem := memory.Memory{
		ID:        uuid.New(),
		NPCID:     uuid.New(),
		Type:      memory.MemoryTypeEvent,
		Timestamp: time.Now(),
		Tags:      []string{EventChallenged},
		Content: memory.EventContent{
			Description: "Had to work hard from young age",
		},
	}

	ApplyExperienceModifiers(p, []memory.Memory{mem})

	assert.Greater(t, p.Conscientiousness.Value, initialConscientiousness)
	assert.InDelta(t, initialConscientiousness+10, p.Conscientiousness.Value, 0.1)
}

func TestApplyExperienceModifiers_SocialSuccess(t *testing.T) {
	p := NewPersonality()
	initialExtraversion := p.Extraversion.Value

	mem := memory.Memory{
		ID:        uuid.New(),
		NPCID:     uuid.New(),
		Type:      memory.MemoryTypeEvent,
		Timestamp: time.Now(),
		Tags:      []string{EventSocialSuccess},
		Content: memory.EventContent{
			Description: "Made many friends, well-liked",
		},
	}

	ApplyExperienceModifiers(p, []memory.Memory{mem})

	assert.Greater(t, p.Extraversion.Value, initialExtraversion)
	assert.InDelta(t, initialExtraversion+10, p.Extraversion.Value, 0.1)
}

func TestApplyExperienceModifiers_RepeatedFailures(t *testing.T) {
	p := NewPersonality()
	initialNeuroticism := p.Neuroticism.Value
	initialConscientiousness := p.Conscientiousness.Value

	mem := memory.Memory{
		ID:        uuid.New(),
		NPCID:     uuid.New(),
		Type:      memory.MemoryTypeEvent,
		Timestamp: time.Now(),
		Tags:      []string{EventSocialFailure},
		Content: memory.EventContent{
			Description: "Failed at multiple endeavors",
		},
	}

	ApplyExperienceModifiers(p, []memory.Memory{mem})

	assert.Greater(t, p.Neuroticism.Value, initialNeuroticism, "Failures increase anxiety")
	assert.Less(t, p.Conscientiousness.Value, initialConscientiousness, "Failures reduce discipline")
}

func TestApplyExperienceModifiers_Exploration(t *testing.T) {
	p := NewPersonality()
	initialOpenness := p.Openness.Value

	mem := memory.Memory{
		ID:        uuid.New(),
		NPCID:     uuid.New(),
		Type:      memory.MemoryTypeEvent,
		Timestamp: time.Now(),
		Tags:      []string{EventExploration},
		Content: memory.EventContent{
			Description: "Traveled to distant lands",
		},
	}

	ApplyExperienceModifiers(p, []memory.Memory{mem})

	assert.Greater(t, p.Openness.Value, initialOpenness)
	assert.InDelta(t, initialOpenness+10, p.Openness.Value, 0.1)
}

func TestApplyExperienceModifiers_MultipleEvents(t *testing.T) {
	p := NewPersonality()

	memories := []memory.Memory{
		{
			ID:        uuid.New(),
			NPCID:     uuid.New(),
			Type:      memory.MemoryTypeEvent,
			Timestamp: time.Now(),
			Tags:      []string{EventNurtured},
		},
		{
			ID:        uuid.New(),
			NPCID:     uuid.New(),
			Type:      memory.MemoryTypeEvent,
			Timestamp: time.Now(),
			Tags:      []string{EventSocialSuccess},
		},
		{
			ID:        uuid.New(),
			NPCID:     uuid.New(),
			Type:      memory.MemoryTypeEvent,
			Timestamp: time.Now(),
			Tags:      []string{EventExploration},
		},
	}

	ApplyExperienceModifiers(p, memories)

	// All three traits should have increased
	assert.InDelta(t, 65.0, p.Agreeableness.Value, 1.0) // 50 + 15
	assert.InDelta(t, 60.0, p.Extraversion.Value, 1.0)  // 50 + 10
	assert.InDelta(t, 60.0, p.Openness.Value, 1.0)      // 50 + 10
}

func TestApplyExperienceModifiers_Clamping(t *testing.T) {
	p := NewPersonality()

	// Set trait near max
	p.Openness.Value = 95

	mem := memory.Memory{
		ID:        uuid.New(),
		NPCID:     uuid.New(),
		Type:      memory.MemoryTypeEvent,
		Timestamp: time.Now(),
		Tags:      []string{EventExploration},
	}

	ApplyExperienceModifiers(p, []memory.Memory{mem})

	// Should be clamped at 100
	assert.Equal(t, 100.0, p.Openness.Value, "Should clamp at 100")
}

func TestApplyExperienceModifiers_ClampingMin(t *testing.T) {
	p := NewPersonality()

	// Set trait near min
	p.Conscientiousness.Value = 5

	mem := memory.Memory{
		ID:        uuid.New(),
		NPCID:     uuid.New(),
		Type:      memory.MemoryTypeEvent,
		Timestamp: time.Now(),
		Tags:      []string{EventSocialFailure},
	}

	ApplyExperienceModifiers(p, []memory.Memory{mem})

	// Should be clamped at 0
	assert.Equal(t, 0.0, p.Conscientiousness.Value, "Should clamp at 0")
}

func TestClampTrait_Max(t *testing.T) {
	trait := &Trait{Name: TraitOpenness, Value: 120}
	clampTrait(trait)
	assert.Equal(t, 100.0, trait.Value)
}

func TestClampTrait_Min(t *testing.T) {
	trait := &Trait{Name: TraitOpenness, Value: -10}
	clampTrait(trait)
	assert.Equal(t, 0.0, trait.Value)
}

func TestClampTrait_Normal(t *testing.T) {
	trait := &Trait{Name: TraitOpenness, Value: 50}
	clampTrait(trait)
	assert.Equal(t, 50.0, trait.Value, "Should not change valid values")
}

func TestApplyExperienceModifiers_CumulativeEffects(t *testing.T) {
	p := NewPersonality()
	initialNeuroticism := p.Neuroticism.Value

	// Simulate a traumatic life with multiple negative events
	memories := []memory.Memory{
		{
			ID:        uuid.New(),
			NPCID:     uuid.New(),
			Type:      memory.MemoryTypeEvent,
			Timestamp: time.Now(),
			Tags:      []string{EventTrauma},
		},
		{
			ID:        uuid.New(),
			NPCID:     uuid.New(),
			Type:      memory.MemoryTypeEvent,
			Timestamp: time.Now(),
			Tags:      []string{EventSocialFailure},
		},
	}

	ApplyExperienceModifiers(p, memories)

	// Net effect: +20 (trauma) + 10 (failure) = +30 neuroticism
	assert.InDelta(t, initialNeuroticism+30, p.Neuroticism.Value, 0.1)
}

func TestApplyExperienceModifiers_EmptyEvents(t *testing.T) {
	p := NewPersonality()
	original := *p

	ApplyExperienceModifiers(p, []memory.Memory{})

	// Should remain unchanged
	assert.Equal(t, original.Openness.Value, p.Openness.Value)
	assert.Equal(t, original.Conscientiousness.Value, p.Conscientiousness.Value)
	assert.Equal(t, original.Extraversion.Value, p.Extraversion.Value)
	assert.Equal(t, original.Agreeableness.Value, p.Agreeableness.Value)
	assert.Equal(t, original.Neuroticism.Value, p.Neuroticism.Value)
}

func TestApplyExperienceModifiers_EventContentDetection(t *testing.T) {
	p := NewPersonality()

	// Test using EventContent.EventType instead of tags
	mem := memory.Memory{
		ID:        uuid.New(),
		NPCID:     uuid.New(),
		Type:      memory.MemoryTypeEvent,
		Timestamp: time.Now(),
		Content: memory.EventContent{
			Description: "Traumatic experience",
			EventType:   "childhood_trauma",
		},
	}

	ApplyExperienceModifiers(p, []memory.Memory{mem})

	// Should detect "trauma" in EventType and apply modifier
	assert.Greater(t, p.Neuroticism.Value, 50.0, "Should increase from trauma keyword")
}

func TestApplyExperienceModifiers_CompoundTags(t *testing.T) {
	p := NewPersonality()

	// Memory with multiple personality-forming tags
	mem := memory.Memory{
		ID:        uuid.New(),
		NPCID:     uuid.New(),
		Type:      memory.MemoryTypeEvent,
		Timestamp: time.Now(),
		Tags:      []string{EventSocialSuccess, EventExploration},
	}

	ApplyExperienceModifiers(p, []memory.Memory{mem})

	// Should apply both modifiers
	assert.InDelta(t, 60.0, p.Extraversion.Value, 1.0) // +10 from social success
	assert.InDelta(t, 60.0, p.Openness.Value, 1.0)     // +10 from exploration
}
