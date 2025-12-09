package interview

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"mud-platform-backend/internal/repository"
	"mud-platform-backend/internal/worldgen/orchestrator"

	"github.com/google/uuid"
)

// LLMClient defines the interface for generating text
type LLMClient interface {
	Generate(prompt string) (string, error)
}

// InterviewService manages the interview process
type InterviewService struct {
	client        LLMClient
	repo          Repository
	worldRepo     repository.WorldRepository
	extractor     *ExtractionService
	nameGenerator *NameGenerator
}

// NewService creates a new service
func NewService(client LLMClient, repo Repository, worldRepo repository.WorldRepository) *InterviewService {
	return &InterviewService{
		client:        client,
		repo:          repo,
		worldRepo:     worldRepo,
		extractor:     NewExtractionService(client),
		nameGenerator: NewNameGenerator(client),
	}
}

// NewServiceWithRepository is deprecated, use NewService
func NewServiceWithRepository(client LLMClient, repo Repository, worldRepo repository.WorldRepository) *InterviewService {
	return NewService(client, repo, worldRepo)
}

// StartInterview initializes a new session and returns the first question
func (s *InterviewService) StartInterview(ctx context.Context, playerID uuid.UUID) (*InterviewSession, string, error) {
	// Check if active interview exists
	existing, err := s.repo.GetInterview(ctx, playerID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check existing interview: %w", err)
	}
	if existing != nil && existing.Status == StatusInProgress {
		return s.ResumeInterview(ctx, playerID)
	}

	// Create new interview
	interview, err := s.repo.CreateInterview(ctx, playerID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create interview: %w", err)
	}

	// Construct session object for internal use
	session := &InterviewSession{
		ID:       interview.ID,
		PlayerID: playerID,
		State: InterviewState{
			CurrentCategory:   AllTopics[0].Category,
			CurrentTopicIndex: 0,
			Answers:           make(map[string]string),
			IsComplete:        false,
		},
		History:   make([]ConversationTurn, 0),
		CreatedAt: interview.CreatedAt,
		UpdatedAt: interview.UpdatedAt,
	}

	// Generate first question
	firstTopic := AllTopics[0]
	prompt := BuildInterviewPrompt(session.State, firstTopic, nil)
	question, err := s.client.Generate(prompt)
	if err != nil {
		return nil, "", err
	}

	instruction := "Welcome, creator. We shall begin crafting your world together. I will ask you questions, and you may respond using the \"reply\" command.\n\n"
	return session, instruction + question, nil
}

// ProcessResponse handles the player's answer and generates the next question
func (s *InterviewService) ProcessResponse(ctx context.Context, playerID uuid.UUID, response string) (string, bool, error) {
	// Load session
	session, err := s.loadSession(ctx, playerID)
	if err != nil {
		return "", false, err
	}
	if session == nil {
		return "", false, errors.New("no active interview found")
	}

	if session.State.IsComplete {
		return "The interview is already complete.", true, nil
	}

	// 0. Check for "change" command
	// Format: change <topic> to <value>
	lowerResp := strings.ToLower(strings.TrimSpace(response))
	if strings.HasPrefix(lowerResp, "change ") {
		parts := strings.SplitN(response, " to ", 2)
		if len(parts) == 2 {
			// Extract topic name from "change <topic>"
			// Remove "change " prefix (7 chars) and trim
			topicPart := strings.TrimSpace(parts[0][7:])
			newValue := strings.TrimSpace(parts[1])

			// Find topic
			var foundTopic *Topic
			for _, t := range AllTopics {
				if strings.Contains(strings.ToLower(t.Name), strings.ToLower(topicPart)) {
					foundTopic = &t
					break
				}
			}

			if foundTopic != nil {
				if err := s.EditAnswer(ctx, playerID, session.ID, foundTopic.Name, newValue); err != nil {
					return "", false, fmt.Errorf("failed to edit answer: %w", err)
				}

				// Update session state in memory
				session.State.Answers[foundTopic.Name] = newValue

				// If in review mode (index >= len), show review again
				if session.State.CurrentTopicIndex >= len(AllTopics) {
					summary := s.generateReviewSummary(session)
					return fmt.Sprintf("Updated %s to '%s'.\n\n%s", foundTopic.Name, newValue, summary), false, nil
				}

				// Otherwise, repeat current question with confirmation
				currentTopic := AllTopics[session.State.CurrentTopicIndex]
				prompt := BuildInterviewPrompt(session.State, currentTopic, session.History)
				question, err := s.client.Generate(prompt)
				if err != nil {
					return "", false, err
				}
				return fmt.Sprintf("Updated %s to '%s'.\n\n%s", foundTopic.Name, newValue, question), false, nil
			}
		}
	}

	// Check for Review Mode
	if session.State.CurrentTopicIndex >= len(AllTopics) {
		return s.handleReviewResponse(ctx, session, playerID, response)
	}

	// Get current topic
	currentTopic := AllTopics[session.State.CurrentTopicIndex]

	// Special handling for World Name topic - validate before saving
	if currentTopic.Name == "World Name" {
		// Validate world name
		if strings.TrimSpace(response) == "" {
			return "Please provide a name for your world.", false, nil
		}

		// Check for invalid characters (allow letters, numbers, spaces, hyphens, apostrophes)
		validName := regexp.MustCompile(`^[a-zA-Z0-9\s\-']+$`)
		if !validName.MatchString(response) || len(response) > 100 {
			return "That world name is not valid. Please use only letters, numbers, spaces, hyphens, and apostrophes (max 100 characters).", false, nil
		}

		// Check if name is already taken
		taken, err := s.repo.IsWorldNameTaken(ctx, response)
		if err != nil {
			return "", false, fmt.Errorf("failed to check world name: %w", err)
		}

		if taken {
			// Generate alternative names
			alternatives, err := s.generateNameSuggestions(ctx, session)
			if err != nil {
				return "That world name is already taken. Please choose another name.", false, nil
			}
			return fmt.Sprintf("That world name is already taken. Please choose another name.\n\nHere are some alternative suggestions:\n%s", alternatives), false, nil
		}
	}

	// 1. Save answer for current topic
	if err := s.repo.SaveAnswer(ctx, session.ID, session.State.CurrentTopicIndex, response); err != nil {
		return "", false, fmt.Errorf("failed to save answer: %w", err)
	}

	// 2. Update history (in memory for prompt generation)
	session.History = append(session.History, ConversationTurn{
		Answer: response,
	})
	session.State.Answers[currentTopic.Name] = response

	// 3. Advance topic
	nextIndex := session.State.CurrentTopicIndex + 1

	// Update index explicitly
	if err := s.repo.UpdateQuestionIndex(ctx, session.ID, nextIndex); err != nil {
		return "", false, fmt.Errorf("failed to update question index: %w", err)
	}
	session.State.CurrentTopicIndex = nextIndex

	if nextIndex >= len(AllTopics) {
		// Determine we just finished the last question. Enter Review Mode.
		return s.generateReviewSummary(session), false, nil
	}

	// 5. Generate next question
	nextTopic := AllTopics[nextIndex]
	prompt := BuildInterviewPrompt(session.State, nextTopic, session.History)
	question, err := s.client.Generate(prompt)
	if err != nil {
		return "", false, err
	}

	return question, false, nil
}

// handleReviewResponse handles interaction during review phase
func (s *InterviewService) handleReviewResponse(ctx context.Context, session *InterviewSession, playerID uuid.UUID, response string) (string, bool, error) {
	lowerResp := strings.ToLower(strings.TrimSpace(response))
	if lowerResp == "yes" || lowerResp == "create" || lowerResp == "confirm" || lowerResp == "looks good" {
		// Proceed to completion
		if err := s.repo.UpdateInterviewStatus(ctx, session.ID, StatusCompleted); err != nil {
			return "", false, fmt.Errorf("failed to complete interview: %w", err)
		}

		// Extract and save configuration
		config, err := s.extractor.ExtractConfiguration(session, playerID)
		if err != nil {
			return "", false, fmt.Errorf("failed to extract configuration: %w", err)
		}

		if err := s.repo.SaveConfiguration(ctx, config); err != nil {
			return "", false, fmt.Errorf("failed to save configuration: %w", err)
		}

		// Create the world
		radius := 1000.0
		world := &repository.World{
			ID:        uuid.New(),
			Name:      config.WorldName,
			OwnerID:   playerID,
			Shape:     repository.WorldShapeSphere,
			Radius:    &radius,
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
		}

		// Add world metadata
		world.Metadata["theme"] = config.Theme
		world.Metadata["description"] = fmt.Sprintf("A %s world with %s tone.", config.Theme, config.Tone)

		fmt.Printf("[DEBUG] Attempting to create world: ID=%s, Name='%s', OwnerID=%s\n", world.ID, world.Name, world.OwnerID)
		if err := s.worldRepo.CreateWorld(ctx, world); err != nil {
			fmt.Printf("[DEBUG] Failed to create world: %v\n", err)
			return "", false, fmt.Errorf("failed to create world: %w", err)
		}
		fmt.Printf("[DEBUG] Successfully created world %s\n", world.ID)

		// Generate procedural content for the world
		fmt.Printf("[DEBUG] Generating procedural content for world %s\n", world.ID)
		generator := orchestrator.NewGeneratorService()
		generated, err := generator.GenerateWorld(ctx, world.ID, config)
		if err != nil {
			// Log error but don't fail - world record exists
			fmt.Printf("[WARN] Failed to generate world content: %v\n", err)
		} else {
			fmt.Printf("[DEBUG] World generation completed in %v\n", generated.Metadata.GenerationTime)
			fmt.Printf("[DEBUG] Generated: %d plates, %d biomes, sea level: %.2f\n",
				len(generated.Geography.Plates),
				len(generated.Geography.Biomes),
				generated.Metadata.SeaLevel)

			// Store generation metadata in world metadata
			world.Metadata["generated"] = true
			world.Metadata["generation_seed"] = generated.Metadata.Seed
			world.Metadata["generation_time"] = generated.Metadata.GenerationTime.String()
			world.Metadata["sea_level"] = generated.Metadata.SeaLevel
			world.Metadata["land_ratio"] = generated.Metadata.LandRatio
			world.Metadata["dimensions"] = map[string]int{
				"width":  generated.Metadata.DimensionsX,
				"height": generated.Metadata.DimensionsY,
			}

			// Update world with generation metadata
			if err := s.worldRepo.UpdateWorld(ctx, world); err != nil {
				fmt.Printf("[WARN] Failed to update world metadata: %v\n", err)
			}

			// TODO: Persist geography, minerals, species to permanent storage
			// For now, generation is ephemeral and will need to be regenerated on world load
			// Future PR will add database tables for heightmap, minerals, species persistence
		}

		// Link configuration to the new world ID
		config.WorldID = &world.ID
		if err := s.repo.SaveConfiguration(ctx, config); err != nil {
			// Just log error, don't fail as world is created
			fmt.Printf("[ERROR] failed to link configuration to world: %v\n", err)
		}

		return fmt.Sprintf("Thank you! I have gathered all the information, and your world has been created. Your world is being forged. You may now enter it using 'enter %s'.", world.Name), true, nil
	}

	// Assume user wants to change something but didn't use "change" command correctly, or just chatting
	return "Please type 'reply yes' to create your world, or use 'reply change <topic> to <value>' to modify an answer.\n\n" + s.generateReviewSummary(session), false, nil
}

// generateReviewSummary creates a summary of the current world configuration for review
func (s *InterviewService) generateReviewSummary(session *InterviewSession) string {
	summary := "Here is the vision for your world:\n\n"
	for _, t := range AllTopics {
		ans, ok := session.State.Answers[t.Name]
		if !ok {
			ans = "(Not answered)"
		}
		summary += fmt.Sprintf("- **%s**: %s\n", t.Name, ans)
	}
	summary += "\nIs this correct? Type 'reply yes' to create, or 'reply change <topic> to <value>' to modify."
	return summary
}

// ResumeInterview loads and resumes an existing interview session
func (s *InterviewService) ResumeInterview(ctx context.Context, playerID uuid.UUID) (*InterviewSession, string, error) {
	session, err := s.loadSession(ctx, playerID)
	if err != nil {
		return nil, "", err
	}
	if session == nil {
		return nil, "", errors.New("no active interview found")
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

// generateNameSuggestions generates alternative world names based on the interview answers
func (s *InterviewService) generateNameSuggestions(ctx context.Context, session *InterviewSession) (string, error) {
	// Build a description from the interview answers
	description := ""
	if theme, ok := session.State.Answers["Core Concept"]; ok {
		description += theme + ". "
	}
	if env, ok := session.State.Answers["Environment"]; ok {
		description += env + ". "
	}

	prompt := fmt.Sprintf(`Based on the following world description, generate EXACTLY 3 creative, unique world names. 
Each name should be on a new line with no numbering or formatting.

World Description: %s

Return only the 3 names, one per line.`, description)

	response, err := s.client.Generate(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate name suggestions: %w", err)
	}

	// Clean up the response
	lines := strings.Split(strings.TrimSpace(response), "\n")
	suggestions := make([]string, 0, 3)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Remove any numbering like "1. " or "- "
		line = regexp.MustCompile(`^[\d\-\*\.\)]+\s*`).ReplaceAllString(line, "")
		if line != "" && len(suggestions) < 3 {
			suggestions = append(suggestions, line)
		}
	}

	return strings.Join(suggestions, "\n"), nil
}

// loadSession reconstructs the session from DB
func (s *InterviewService) loadSession(ctx context.Context, playerID uuid.UUID) (*InterviewSession, error) {
	interview, err := s.repo.GetInterview(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if interview == nil {
		return nil, nil
	}

	answers, err := s.repo.GetAnswers(ctx, interview.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get answers: %w", err)
	}

	// Reconstruct state
	state := InterviewState{
		CurrentTopicIndex: interview.CurrentQuestionIndex,
		Answers:           make(map[string]string),
		IsComplete:        interview.Status == StatusCompleted,
	}

	if interview.CurrentQuestionIndex < len(AllTopics) {
		state.CurrentCategory = AllTopics[interview.CurrentQuestionIndex].Category
	}

	var history []ConversationTurn
	for _, a := range answers {
		if a.QuestionIndex < len(AllTopics) {
			topic := AllTopics[a.QuestionIndex]
			state.Answers[topic.Name] = a.AnswerText
			// We don't store the question text in DB, so we can't fully reconstruct history with questions
			// But prompts usually only need answers or previous context.
			// If we need questions, we might need to regenerate them or store them.
			// For now, let's assume answers are enough or we leave question empty.
			history = append(history, ConversationTurn{
				Answer: a.AnswerText,
			})
		}
	}

	return &InterviewSession{
		ID:        interview.ID,
		PlayerID:  playerID,
		State:     state,
		History:   history,
		CreatedAt: interview.CreatedAt,
		UpdatedAt: interview.UpdatedAt,
	}, nil
}

// GetActiveInterview retrieves the player's active (incomplete) interview
func (s *InterviewService) GetActiveInterview(ctx context.Context, playerID uuid.UUID) (*InterviewSession, error) {
	return s.loadSession(ctx, playerID)
}

// GetProgress returns the interview progress percentage
func (s *InterviewService) GetProgress(ctx context.Context, playerID uuid.UUID) (float64, error) {
	interview, err := s.repo.GetInterview(ctx, playerID)
	if err != nil {
		return 0, err
	}
	if interview == nil {
		return 0, nil
	}

	if interview.Status == StatusCompleted {
		return 1.0, nil
	}

	total := len(AllTopics)
	answered := interview.CurrentQuestionIndex
	return float64(answered) / float64(total), nil
}

// CompleteInterview retrieves the configuration for a completed interview
func (s *InterviewService) CompleteInterview(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) (*WorldConfiguration, error) {
	interview, err := s.repo.GetInterview(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get interview: %w", err)
	}
	if interview == nil {
		return nil, fmt.Errorf("interview not found")
	}

	if interview.ID != sessionID {
		return nil, fmt.Errorf("session ID mismatch")
	}

	if interview.Status != StatusCompleted {
		return nil, fmt.Errorf("interview is not complete")
	}

	// Retrieve configuration
	config, err := s.repo.GetConfigurationByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	return config, nil
}

// EditAnswer updates a previous answer
func (s *InterviewService) EditAnswer(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, topicName string, newAnswer string) error {
	interview, err := s.repo.GetInterview(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get interview: %w", err)
	}
	if interview == nil || interview.ID != sessionID {
		return fmt.Errorf("interview not found or mismatch")
	}

	// Find topic index
	topicIndex := -1
	for i, t := range AllTopics {
		if t.Name == topicName {
			topicIndex = i
			break
		}
	}
	if topicIndex == -1 {
		return fmt.Errorf("topic not found: %s", topicName)
	}

	// Update answer in repo
	if err := s.repo.SaveAnswer(ctx, sessionID, topicIndex, newAnswer); err != nil {
		return fmt.Errorf("failed to save answer: %w", err)
	}

	return nil
}
