package dialogue

import (
	"testing"
	"tw-backend/internal/character"
	"tw-backend/internal/npc/memory"
	"tw-backend/internal/npc/personality"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateConversationMemory(t *testing.T) {
	mockMemRepo := new(MockMemRepo)
	s := &DialogueService{
		memoryRepo: mockMemRepo,
	}

	npcID := uuid.New()
	speakerID := uuid.New()
	input := "Hello"
	output := "Hi there"
	emotion := "happy"
	weight := 0.8

	mockMemRepo.On("CreateMemory", mock.MatchedBy(func(m memory.Memory) bool {
		return m.NPCID == npcID &&
			m.Type == memory.MemoryTypeConversation &&
			m.EmotionalWeight == weight &&
			m.Content.(memory.ConversationContent).Dialogue[0].Text == input &&
			m.Content.(memory.ConversationContent).Dialogue[1].Text == output &&
			m.Content.(memory.ConversationContent).Dialogue[1].Emotion == emotion
	})).Return(nil).Once()

	s.createConversationMemory(npcID, speakerID, input, output, emotion, weight)

	mockMemRepo.AssertExpectations(t)
}

func TestFetchNPCState(t *testing.T) {
	mockNPCRepo := new(MockNPCRepo)
	s := &DialogueService{
		npcRepo: mockNPCRepo,
	}

	npcID := uuid.New()
	mockNPCRepo.On("GetNPC", npcID).Return(&character.Character{Name: "Bob"}, nil)
	mockNPCRepo.On("GetPersonality", npcID).Return(personality.NewPersonality(), nil)
	mockNPCRepo.On("GetMood", npcID).Return(personality.NewMood("Neutral", 0.5), nil)

	state, err := s.fetchNPCState(npcID)
	assert.NoError(t, err)
	assert.Equal(t, "Bob", state.Name)
	assert.Equal(t, "Neutral", state.Mood.Type)

	mockNPCRepo.AssertExpectations(t)
}
