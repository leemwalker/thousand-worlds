package interview

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LLMClient defines the interface for generating text
type LLMClient interface {
	Generate(prompt string) (string, error)
}

// InterviewService manages the interview process
type InterviewService struct {
	client    LLMClient
	repo      RepositoryInterface
	extractor *ExtractionService
	sessions  map[uuid.UUID]*InterviewSession // In-memory fallback for tests
}

// NewService creates a new service
func NewService(client LLMClient) *InterviewService {
	return &InterviewService{
		client:    client,
		repo:      nil, // For backward compatibility with tests
		extractor: NewExtractionService(client),
		sessions:  make(map[uuid.UUID]*InterviewSession),
	}
}

// NewServiceWithRepository creates a new service with repository support
func NewServiceWithRepository(client LLMClient, repo RepositoryInterface) *InterviewService {
	return &InterviewService{
		client:    client,
		repo:      repo,
		extractor: NewExtractionService(client),
		sessions:  make(map[uuid.UUID]*InterviewSession), // Keep for fallback
	}
}

// StartInterview initializes a new session and returns the first question
func (s *InterviewService) StartInterview(playerID uuid.UUID) (*InterviewSession, string, error) {
	sessionID := uuid.New()
	session := &InterviewSession{
		ID:       sessionID,
		PlayerID: playerID,
		State: InterviewState{
			CurrentCategory:   AllTopics[0].Category,
			CurrentTopicIndex: 0,
			Answers:           make(map[string]string),
			IsComplete:        false,
		},
		History:   make([]ConversationTurn, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database if repository is available
	if s.repo != nil {
		if err := s.repo.SaveInterview(session); err != nil {
			return nil, "", fmt.Errorf("failed to save interview: %w", err)
		}
	} else {
		// Fallback to in-memory for tests
		s.sessions[sessionID] = session
	}

	// Generate first question
	firstTopic := AllTopics[0]
	prompt := BuildInterviewPrompt(session.State, firstTopic, nil)
	question, err := s.client.Generate(prompt)
	if err != nil {
		return nil, "", err
	}

	return session, question, nil
}

// ProcessResponse handles the player's answer and generates the next question
func (s *InterviewService) ProcessResponse(sessionID uuid.UUID, response string) (string, bool, error) {
	// Load session from database if repository is available
	var session *InterviewSession
	var err error

	if s.repo != nil {
		session, err = s.repo.GetInterview(sessionID)
		if err != nil {
			return "", false, errors.New("session not found")
		}
	} else {
		// Fallback to in-memory for tests
		var ok bool
		session, ok = s.sessions[sessionID]
		if !ok {
			return "", false, errors.New("session not found")
		}
	}

	if session.State.IsComplete {
		return "The interview is already complete.", true, nil
	}

	// 1. Save answer for current topic
	currentTopic := AllTopics[session.State.CurrentTopicIndex]
	session.State.Answers[currentTopic.Name] = response

	// 2. Update history
	session.History = append(session.History, ConversationTurn{
		Answer: response,
	})

	// 3. Advance topic
	session.State.CurrentTopicIndex++
	if session.State.CurrentTopicIndex >= len(AllTopics) {
		session.State.IsComplete = true

		// Update in database
		if s.repo != nil {
			if err := s.repo.UpdateInterview(session); err != nil {
				return "", false, fmt.Errorf("failed to update interview: %w", err)
			}
		}

		return "Thank you! I have gathered all the information needed to build your world.", true, nil
	}

	// 4. Update category
	nextTopic := AllTopics[session.State.CurrentTopicIndex]
	session.State.CurrentCategory = nextTopic.Category
	session.UpdatedAt = time.Now()

	// 5. Save updated interview to database
	if s.repo != nil {
		if err := s.repo.UpdateInterview(session); err != nil {
			return "", false, fmt.Errorf("failed to update interview: %w", err)
		}
	}

	// 6. Generate next question
	prompt := BuildInterviewPrompt(session.State, nextTopic, session.History)
	question, err := s.client.Generate(prompt)
	if err != nil {
		return "", false, err
	}

	return question, false, nil
}

// ResumeInterview loads and resumes an existing interview session
func (s *InterviewService) ResumeInterview(sessionID uuid.UUID) (*InterviewSession, string, error) {
	if s.repo == nil {
		return nil, "", errors.New("repository not available")
	}

	session, err := s.repo.GetInterview(sessionID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load interview: %w", err)
	}

	if session.State.IsComplete {
		return session, "This interview is already complete.", nil
	}

	// Generate next question based on current state
	currentTopic := AllTopics[session.State.CurrentTopicIndex]
	prompt := BuildInterviewPrompt(session.State, currentTopic, session.History)
	question, err := s.client.Generate(prompt)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate question: %w", err)
	}

	return session, question, nil
}

// GetActiveInterview retrieves the player's active (incomplete) interview
func (s *InterviewService) GetActiveInterview(playerID uuid.UUID) (*InterviewSession, error) {
	if s.repo == nil {
		return nil, errors.New("repository not available")
	}

	return s.repo.GetActiveInterviewByPlayer(playerID)
}

// EditAnswer allows editing a previous answer in the interview
func (s *InterviewService) EditAnswer(sessionID uuid.UUID, topicName string, newAnswer string) error {
	if s.repo == nil {
		return errors.New("repository not available")
	}

	session, err := s.repo.GetInterview(sessionID)
	if err != nil {
		return fmt.Errorf("failed to load interview: %w", err)
	}

	// Verify topic exists
	topicExists := false
	for _, topic := range AllTopics {
		if topic.Name == topicName {
			topicExists = true
			break
		}
	}

	if !topicExists {
		return fmt.Errorf("invalid topic: %s", topicName)
	}

	// Update answer
	session.State.Answers[topicName] = newAnswer
	session.UpdatedAt = time.Now()

	// Save to database
	return s.repo.UpdateInterview(session)
}

// CompleteInterview extracts configuration and validates it
func (s *InterviewService) CompleteInterview(sessionID uuid.UUID) (*WorldConfiguration, error) {
	if s.repo == nil {
		return nil, errors.New("repository not available")
	}

	session, err := s.repo.GetInterview(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load interview: %w", err)
	}

	if !session.State.IsComplete {
		return nil, errors.New("interview is not complete")
	}

	// Extract configuration
	config, err := s.extractor.ExtractConfiguration(session, session.PlayerID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract configuration: %w", err)
	}

	// Validate configuration
	validationErrors := ValidateConfiguration(config)
	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("configuration validation failed: %d errors", len(validationErrors))
	}

	// Save configuration to database
	if err := s.repo.SaveConfiguration(config); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	return config, nil
}

// GetProgress returns the interview progress percentage
func (s *InterviewService) GetProgress(sessionID uuid.UUID) (float64, error) {
	if s.repo == nil {
		return 0, errors.New("repository not available")
	}

	session, err := s.repo.GetInterview(sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to load interview: %w", err)
	}

	total := len(AllTopics)
	answered := session.State.CurrentTopicIndex

	if session.State.IsComplete {
		return 1.0, nil
	}

	return float64(answered) / float64(total), nil
}
