package interaction

import (
	"testing"
	"time"

	"mud-platform-backend/internal/npc/memory"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSelectTopic_HighEmotionalWeight(t *testing.T) {
	// Memory with high emotional weight  should score high
	mem1 := memory.Memory{
		ID:              uuid.New(),
		NPCID:           uuid.New(),
		Type:            memory.MemoryTypeEvent,
		Timestamp:       time.Now().Add(-1 * time.Hour), // Recent
		EmotionalWeight: 0.9,
		Content: memory.EventContent{
			Description: "witnessed a dramatic event",
		},
	}

	mem2 := memory.Memory{
		ID:              uuid.New(),
		NPCID:           uuid.New(),
		Type:            memory.MemoryTypeObservation,
		Timestamp:       time.Now().Add(-1 * time.Hour),
		EmotionalWeight: 0.2, // Low emotion
		Content: memory.ObservationContent{
			Event: "saw a bird",
		},
	}

	topic := SelectTopic([]memory.Memory{mem1, mem2}, nil)

	// Should select mem1 (high emotion)
	assert.Greater(t, topic.Score, 0.0)
	assert.Contains(t, topic.Text, "dramatic")
}

func TestSelectTopic_SharedExperience(t *testing.T) {
	indivMem := memory.Memory{
		ID:              uuid.New(),
		NPCID:           uuid.New(),
		Type:            memory.MemoryTypeEvent,
		Timestamp:       time.Now().Add(-2 * time.Hour),
		EmotionalWeight: 0.6,
		Content: memory.EventContent{
			Description: "went shopping alone",
		},
	}

	sharedMem := memory.Memory{
		ID:              uuid.New(),
		NPCID:           uuid.New(),
		Type:            memory.MemoryTypeEvent,
		Timestamp:       time.Now().Add(-2 * time.Hour),
		EmotionalWeight: 0.6, // Same emotion
		Content: memory.EventContent{
			Description: "we survived the attack together",
		},
	}

	topic := SelectTopic([]memory.Memory{indivMem}, []memory.Memory{sharedMem})

	// Shared should win due to +0.3 bonus
	assert.Contains(t, topic.Text, "attack")
}

func TestSelectTopic_Recency(t *testing.T) {
	recentMem := memory.Memory{
		ID:              uuid.New(),
		NPCID:           uuid.New(),
		Type:            memory.MemoryTypeEvent,
		Timestamp:       time.Now().Add(-1 * time.Hour), // 1 hour ago
		EmotionalWeight: 0.5,
		Content: memory.EventContent{
			Description: "just happened",
		},
	}

	oldMem := memory.Memory{
		ID:              uuid.New(),
		NPCID:           uuid.New(),
		Type:            memory.MemoryTypeEvent,
		Timestamp:       time.Now().Add(-20 * time.Hour), // 20 hours ago
		EmotionalWeight: 0.5,                             // Same emotion
		Content: memory.EventContent{
			Description: "old event",
		},
	}

	topic := SelectTopic([]memory.Memory{recentMem, oldMem}, nil)

	// Recent should score higher
	assert.Contains(t, topic.Text, "just")
}

func TestSelectTopic_Fallback(t *testing.T) {
	// No memories provided
	topic := SelectTopic([]memory.Memory{}, nil)

	// Should fallback to small talk
	assert.Equal(t, "the weather", topic.Text)
	assert.Equal(t, 0.0, topic.Score)
}

func TestSelectTopic_ScoreFormula(t *testing.T) {
	// Test exact scoring formula: (emotion × 0.4) + (recency × 0.3) + (shared × 0.3)

	// Recent memory with 0.8 emotion, 1.0 recency (0 hours old), not shared
	mem := memory.Memory{
		ID:              uuid.New(),
		NPCID:           uuid.New(),
		Type:            memory.MemoryTypeEvent,
		Timestamp:       time.Now(),
		EmotionalWeight: 0.8,
		Content: memory.EventContent{
			Description: "test event",
		},
	}

	topic := SelectTopic([]memory.Memory{mem}, nil)

	// Expected: (0.8 * 0.4) + (1.0 * 0.3) + (0 * 0.3) = 0.32 + 0.3 = 0.62
	assert.InDelta(t, 0.62, topic.Score, 0.01)
}

func TestSelectTopic_SharedBonus(t *testing.T) {
	// Same memory, one shared, one not
	timestamp := time.Now()

	mem := memory.Memory{
		ID:              uuid.New(),
		NPCID:           uuid.New(),
		Type:            memory.MemoryTypeEvent,
		Timestamp:       timestamp,
		EmotionalWeight: 0.6,
		Content: memory.EventContent{
			Description: "event",
		},
	}

	// Not shared: (0.6 * 0.4) + (1.0 * 0.3) + (0 * 0.3) = 0.54
	topicIndiv := SelectTopic([]memory.Memory{mem}, nil)

	// Shared: (0.6 * 0.4) + (1.0 * 0.3) + (1.0 * 0.3) = 0.84
	topicShared := SelectTopic(nil, []memory.Memory{mem})

	assert.InDelta(t, 0.54, topicIndiv.Score, 0.01)
	assert.InDelta(t, 0.84, topicShared.Score, 0.01)
	assert.Greater(t, topicShared.Score, topicIndiv.Score)
}

func TestSelectTopic_VeryOldMemory(t *testing.T) {
	// Memory over 24 hours old should have 0 recency
	oldMem := memory.Memory{
		ID:              uuid.New(),
		NPCID:           uuid.New(),
		Type:            memory.MemoryTypeEvent,
		Timestamp:       time.Now().Add(-30 * time.Hour),
		EmotionalWeight: 1.0, // Max emotion
		Content: memory.EventContent{
			Description: "very old",
		},
	}

	topic := SelectTopic([]memory.Memory{oldMem}, nil)

	// Expected: (1.0 * 0.4) + (0 * 0.3) + (0 * 0.3) = 0.4
	assert.InDelta(t, 0.4, topic.Score, 0.01)
}

func TestGenerateTopicStatement(t *testing.T) {
	topic := Topic{
		Text:  "the dragon attack",
		Score: 0.8,
	}

	statement := GenerateTopicStatement(topic)

	// Should use template and include topic
	assert.Contains(t, statement, "dragon attack")
}

func TestFormatTopic_EventContent(t *testing.T) {
	mem := memory.Memory{
		Content: memory.EventContent{
			Description: "an exciting adventure",
		},
	}

	text := formatTopic(mem)
	assert.Equal(t, "an exciting adventure", text)
}

func TestFormatTopic_OtherContent(t *testing.T) {
	mem := memory.Memory{
		Content: memory.ObservationContent{
			Event: "observed something",
		},
	}

	text := formatTopic(mem)
	assert.Equal(t, "something I saw", text)
}
