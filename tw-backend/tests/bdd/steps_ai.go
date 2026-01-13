package bdd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"

	"tw-backend/internal/ai"
	"tw-backend/internal/repository"
)

// MockEventBus
type MockEventBus struct {
	mock.Mock
	subscriptions map[string]nats.MsgHandler
}

func (m *MockEventBus) Subscribe(subject string, cb nats.MsgHandler) (*nats.Subscription, error) {
	args := m.Called(subject, cb)
	if m.subscriptions == nil {
		m.subscriptions = make(map[string]nats.MsgHandler)
	}
	m.subscriptions[subject] = cb
	return nil, args.Error(1)
}

func (m *MockEventBus) Publish(subject string, data []byte) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

// Simulate receiving a message
func (m *MockEventBus) SimulateMessage(subject string, data []byte) {
	if handler, ok := m.subscriptions[subject]; ok {
		// Wildcard handling simplistic for test
		handler(&nats.Msg{Subject: subject, Data: data})
	} else {
		// Try wildcard match if needed, e.g. "ai.response.decision.*"
		// handling exact match for now
	}
}

// MockMemoryRepo
type MockMemoryRepo struct {
	mock.Mock
}

func (m *MockMemoryRepo) GetMemoriesByWorldID(ctx context.Context, worldID string) ([]repository.Memory, error) {
	args := m.Called(ctx, worldID)
	return args.Get(0).([]repository.Memory), args.Error(1)
}

type AIContext struct {
	desireEngine *ai.DesireEngine
	scheduler    *ai.Scheduler
	eventBus     *MockEventBus
	memoryRepo   *MockMemoryRepo

	lastEntityID string
	lastWorldID  string
	lastPrompt   string
	published    map[string][]string // subject -> messages
}

func InitializeAISteps(ctx *godog.ScenarioContext, s *AIContext) {
	ctx.Step(`^the desire engine is running$`, s.theDesireEngineIsRunning)
	ctx.Step(`^I have an NPC "([^"]*)" in "([^"]*)" with context "([^"]*)"$`, s.iHaveAnNPCInWithContext)
	ctx.Step(`^the system requests a decision for "([^"]*)"$`, s.theSystemRequestsADecisionFor)
	ctx.Step(`^the engine should publish an AI request with prompt containing "([^"]*)"$`, s.theEngineShouldPublishAnAIRequestWithPromptContaining)
	ctx.Step(`^the AI service responds with "([^"]*)"$`, s.theAIServiceRespondsWith)
	ctx.Step(`^the engine should publish a spatial action "([^"]*)"$`, s.theEngineShouldPublishASpatialAction)

	ctx.Step(`^an AI scheduler with (\d+) buckets$`, s.anAISchedulerWithBuckets)
	ctx.Step(`^I register (\d+) entities$`, s.iRegisterEntities)
	ctx.Step(`^they should be distributed evenly across buckets$`, s.theyShouldBeDistributedEvenlyAcrossBuckets)
	ctx.Step(`^tick (\d+) should process (\d+) entities$`, s.tickShouldProcessEntities)
	ctx.Step(`^tick (\d+) should process the same entities as tick (\d+)$`, s.tickShouldProcessTheSameEntitiesAsTick)
}

func (s *AIContext) theDesireEngineIsRunning() error {
	s.eventBus = &MockEventBus{}
	s.memoryRepo = &MockMemoryRepo{}

	// Setup expectations that might be called during ListenForDecisions
	s.eventBus.On("Subscribe", "npc.command.decide_action", mock.Anything).Return(nil, nil)
	s.eventBus.On("Subscribe", "ai.response.decision.*", mock.Anything).Return(nil, nil)

	s.desireEngine = ai.NewDesireEngine(s.eventBus, s.memoryRepo)
	err := s.desireEngine.ListenForDecisions()

	s.published = make(map[string][]string)

	// Capture publishes
	s.eventBus.On("Publish", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		subj := args.String(0)
		data := args.Get(1).([]byte)
		s.published[subj] = append(s.published[subj], string(data))
	}).Return(nil)

	return err
}

func (s *AIContext) iHaveAnNPCInWithContext(npc, world, contextStr string) error {
	s.lastEntityID = npc
	s.lastWorldID = world

	// Setup Mock Repo response
	memories := []repository.Memory{
		{NPCID: npc, Content: contextStr, ImportanceScore: 1.0},
	}
	s.memoryRepo.On("GetMemoriesByWorldID", mock.Anything, world).Return(memories, nil)
	return nil
}

func (s *AIContext) theSystemRequestsADecisionFor(npc string) error {
	cmd := ai.DecideActionCommand{
		EntityID:   npc,
		WorldID:    s.lastWorldID,
		WorldState: "Test State",
	}
	data, _ := json.Marshal(cmd)

	// Manually trigger handler because MockEventBus needs help mapping
	// Real EventBus would call handler.
	// We need to capture the handler in Subscribe.
	// My MockEventBus simple implementation stores one handler per subject.

	s.eventBus.SimulateMessage("npc.command.decide_action", data)
	return nil
}

func (s *AIContext) theEngineShouldPublishAnAIRequestWithPromptContaining(text string) error {
	// Check s.published["ai.request.decision"]
	msgs := s.published["ai.request.decision"]
	if len(msgs) == 0 {
		return fmt.Errorf("no AI request published")
	}

	lastMsg := msgs[len(msgs)-1]
	var req ai.AIRequest
	json.Unmarshal([]byte(lastMsg), &req)

	// Check prompt
	if !strings.Contains(req.Prompt, text) { // contains helper or strings.Contains
		// Need strings package
		return fmt.Errorf("prompt doesn't contain '%s'. Got: %s", text, req.Prompt)
	}
	return nil
}

func (s *AIContext) theAIServiceRespondsWith(response string) error {
	resp := ai.AIResponse{
		Response: response,
		EntityID: s.lastEntityID,
	}
	data, _ := json.Marshal(resp)

	// Trigger handler
	// Subscribe was "ai.response.decision.*".
	// MockEventBus logic needs to handle this or we simulate on exact subject and logic finds handler?
	// My MockEventBus `SimulateMessage` looks up stored subscriptions.
	// It stored "ai.response.decision.*".
	// If I call Simulate ("ai.response.decision.123"), logic must match.
	// My MockEventBus logic above was:
	// if handler, ok := m.subscriptions[subject]; ok ... else {}
	// It doesn't handle wildcards.
	// I should invoke the handler registered for "ai.response.decision.*".

	// Quick fix: Invoke explicitly
	if handler, ok := s.eventBus.subscriptions["ai.response.decision.*"]; ok {
		handler(&nats.Msg{Subject: "ai.response.decision.123", Data: data})
	} else {
		return fmt.Errorf("handler for AI response not found")
	}
	return nil
}

func (s *AIContext) theEngineShouldPublishASpatialAction(action string) error {
	msgs := s.published["spatial.command.action"] // Service uses this subject
	if len(msgs) == 0 {
		return fmt.Errorf("no spatial action published")
	}
	if msgs[len(msgs)-1] != action {
		return fmt.Errorf("expected '%s', got '%s'", action, msgs[len(msgs)-1])
	}
	return nil
}

// Scheduler Steps

func (s *AIContext) anAISchedulerWithBuckets(buckets int) error {
	s.scheduler = ai.NewScheduler(buckets)
	return nil
}

func (s *AIContext) iRegisterEntities(count int) error {
	for i := 0; i < count; i++ {
		s.scheduler.RegisterEntity(uuid.New())
	}
	return nil
}

func (s *AIContext) theyShouldBeDistributedEvenlyAcrossBuckets() error {
	_, _, perBucket := s.scheduler.GetStats()
	for i, count := range perBucket {
		if count != 2 {
			return fmt.Errorf("bucket %d has %d entities, expected 2", i, count)
		}
	}
	return nil
}

func (s *AIContext) tickShouldProcessEntities(tick int, count int) error {
	entities := s.scheduler.GetEntitiesForTick(int64(tick))
	if len(entities) != count {
		return fmt.Errorf("tick %d processed %d entities, expected %d", tick, len(entities), count)
	}
	return nil
}

func (s *AIContext) tickShouldProcessTheSameEntitiesAsTick(tick1, tick2 int) error {
	e1 := s.scheduler.GetEntitiesForTick(int64(tick1))
	e2 := s.scheduler.GetEntitiesForTick(int64(tick2))

	if len(e1) != len(e2) {
		return fmt.Errorf("lengths differ")
	}

	// Maps for comparison
	m1 := make(map[uuid.UUID]bool)
	for _, id := range e1 {
		m1[id] = true
	}

	for _, id := range e2 {
		if !m1[id] {
			return fmt.Errorf("entity %s in tick %d but not %d", id, tick2, tick1)
		}
	}
	return nil
}
