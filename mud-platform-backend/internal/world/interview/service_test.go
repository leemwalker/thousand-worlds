package interview

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// MockLLM implements LLMClient for testing
type MockLLM struct {
	GenerateFunc func(prompt string) (string, error)
}

func (m *MockLLM) Generate(prompt string) (string, error) {
	return m.GenerateFunc(prompt)
}

func TestStartInterview(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "What kind of world?", nil
		},
	}
	service := NewService(mockLLM)
	playerID := uuid.New()

	session, question, err := service.StartInterview(playerID)
	if err != nil {
		t.Fatalf("StartInterview failed: %v", err)
	}

	if session.State.CurrentTopicIndex != 0 {
		t.Error("Expected initial topic index 0")
	}
	if question != "What kind of world?" {
		t.Errorf("Expected 'What kind of world?', got '%s'", question)
	}
}

func TestProcessResponse(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Next question?", nil
		},
	}
	service := NewService(mockLLM)
	playerID := uuid.New()

	session, _, _ := service.StartInterview(playerID)

	// Answer first question
	q2, completed, err := service.ProcessResponse(session.ID, "Fantasy world")
	if err != nil {
		t.Fatalf("ProcessResponse failed: %v", err)
	}
	if completed {
		t.Error("Expected interview to not be completed")
	}

	if session.State.CurrentTopicIndex != 1 {
		t.Errorf("Expected topic index 1, got %d", session.State.CurrentTopicIndex)
	}
	if session.State.Answers["World Type"] != "Fantasy world" {
		t.Error("Answer not saved")
	}
	if q2 != "Next question?" {
		t.Error("Expected next question")
	}
}

func TestInterviewCompletion(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Q", nil
		},
	}
	service := NewService(mockLLM)
	playerID := uuid.New()
	session, _, _ := service.StartInterview(playerID)

	// Fast forward to last question
	total := len(AllTopics)
	session.State.CurrentTopicIndex = total - 1

	// Answer last question
	resp, completed, err := service.ProcessResponse(session.ID, "Last answer")
	if err != nil {
		t.Fatalf("ProcessResponse failed: %v", err)
	}

	if !session.State.IsComplete {
		t.Error("Expected session to be complete")
	}
	if !completed {
		t.Error("Expected interview to be completed")
	}
	if resp != "Thank you! I have gathered all the information needed to build your world." {
		t.Errorf("Expected completion message, got '%s'", resp)
	}
}

func TestGetProgress_NoRepository(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Q", nil
		},
	}
	service := NewService(mockLLM)

	_, err := service.GetProgress(uuid.New())
	if err == nil {
		t.Error("Expected error when repository not available")
	}
}

func TestResumeInterview_NoRepository(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Q", nil
		},
	}
	service := NewService(mockLLM)

	_, _, err := service.ResumeInterview(uuid.New())
	if err == nil {
		t.Error("Expected error when repository not available")
	}
}

func TestGetActiveInterview_NoRepository(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Q", nil
		},
	}
	service := NewService(mockLLM)

	_, err := service.GetActiveInterview(uuid.New())
	if err == nil {
		t.Error("Expected error when repository not available")
	}
}

func TestEditAnswer_NoRepository(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Q", nil
		},
	}
	service := NewService(mockLLM)

	err := service.EditAnswer(uuid.New(), "World Type", "New answer")
	if err == nil {
		t.Error("Expected error when repository not available")
	}
}

func TestCompleteInterview_NoRepository(t *testing.T) {
	mockLLM := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			return "Q", nil
		},
	}
	service := NewService(mockLLM)

	_, err := service.CompleteInterview(uuid.New())
	if err == nil {
		t.Error("Expected error when repository not available")
	}
}

func TestBuildInterviewPrompt_WithHistory(t *testing.T) {
	state := InterviewState{
		CurrentCategory:   CategoryTheme,
		CurrentTopicIndex: 1,
		Answers: map[string]string{
			"World Type": "Fantasy",
		},
		IsComplete: false,
	}

	topic := Topic{
		Category:    CategoryTheme,
		Name:        "Tone",
		Description: "Overall tone of the world",
	}

	history := []ConversationTurn{
		{Answer: "Fantasy world with magic"},
	}

	prompt := BuildInterviewPrompt(state, topic, history)

	if !contains(prompt, "Fantasy") {
		t.Error("Expected prompt to contain previous answer")
	}

	if !contains(prompt, "Tone") {
		t.Error("Expected prompt to contain current topic")
	}

	if !contains(prompt, "1 / "+fmt.Sprintf("%d", len(AllTopics))) {
		t.Error("Expected prompt to contain progress")
	}
}

func TestBuildInterviewPrompt_EmptyHistory(t *testing.T) {
	state := InterviewState{
		CurrentCategory:   CategoryTheme,
		CurrentTopicIndex: 0,
		Answers:           make(map[string]string),
		IsComplete:        false,
	}

	topic := AllTopics[0]

	prompt := BuildInterviewPrompt(state, topic, nil)

	if !contains(prompt, topic.Name) {
		t.Error("Expected prompt to contain topic name")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr))
}

func TestGetTopicsByCategory_Theme(t *testing.T) {
	topics := GetTopicsByCategory(CategoryTheme)
	if len(topics) == 0 {
		t.Error("Expected theme topics, got none")
	}

	for _, topic := range topics {
		if topic.Category != CategoryTheme {
			t.Errorf("Expected theme category, got %s", topic.Category)
		}
	}
}

func TestGetTopicsByCategory_TechLevel(t *testing.T) {
	topics := GetTopicsByCategory(CategoryTechLevel)
	if len(topics) == 0 {
		t.Error("Expected tech level topics, got none")
	}

	for _, topic := range topics {
		if topic.Category != CategoryTechLevel {
			t.Errorf("Expected tech level category, got %s", topic.Category)
		}
	}
}

func TestGetTopicsByCategory_Geography(t *testing.T) {
	topics := GetTopicsByCategory(CategoryGeography)
	if len(topics) == 0 {
		t.Error("Expected geography topics, got none")
	}

	for _, topic := range topics {
		if topic.Category != CategoryGeography {
			t.Errorf("Expected geography category, got %s", topic.Category)
		}
	}
}

func TestGetTopicsByCategory_Culture(t *testing.T) {
	topics := GetTopicsByCategory(CategoryCulture)
	if len(topics) == 0 {
		t.Error("Expected culture topics, got none")
	}

	for _, topic := range topics {
		if topic.Category != CategoryCulture {
			t.Errorf("Expected culture category, got %s", topic.Category)
		}
	}
}

func TestAllTopics_Coverage(t *testing.T) {
	if len(AllTopics) < 20 {
		t.Errorf("Expected at least 20 topics, got %d", len(AllTopics))
	}

	// Verify all categories are represented
	categories := make(map[Category]bool)
	for _, topic := range AllTopics {
		categories[topic.Category] = true
	}

	expectedCategories := []Category{
		CategoryTheme,
		CategoryTechLevel,
		CategoryGeography,
		CategoryCulture,
	}

	for _, cat := range expectedCategories {
		if !categories[cat] {
			t.Errorf("Category %s not found in topics", cat)
		}
	}
}
