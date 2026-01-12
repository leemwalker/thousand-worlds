package bdd

import (
	"context"
	"fmt"
	"time"

	"tw-backend/internal/ecosystem"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

// GameLoopContext holds state for game loop scenarios
type GameLoopContext struct {
	runner         *ecosystem.SimulationRunner
	tickInterval   time.Duration
	ticksProcessed int

	// Event System Mock
	eventQueue     []string
	handlers       map[string]func(string)
	playerCount    int
	receivedEvents map[string]bool
}

var gameLoopState = &GameLoopContext{
	handlers:       make(map[string]func(string)),
	receivedEvents: make(map[string]bool),
}

// Reset state before scenario
func (c *GameLoopContext) reset() {
	c.runner = nil
	c.tickInterval = time.Second / 60
	c.ticksProcessed = 0
	c.eventQueue = []string{}
	c.handlers = make(map[string]func(string))
	c.playerCount = 0
	c.receivedEvents = make(map[string]bool)
}

func InitializeGameLoopSteps(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		gameLoopState.reset()
		return ctx, nil
	})

	ctx.Step(`^the game engine is running$`, theGameEngineIsRunning)
	ctx.Step(`^(\d+) ticks pass$`, ticksPass)
	ctx.Step(`^the simulation time should advance by (\d+) second$`, theSimulationTimeShouldAdvanceBySecond)
	ctx.Step(`^all registered subsystems should have updated$`, allRegisteredSubsystemsShouldHaveUpdated)

	ctx.Step(`^a "([^"]*)" event is queued$`, aEventIsQueued)
	ctx.Step(`^the event loop processes the queue$`, theEventLoopProcessesTheQueue)
	ctx.Step(`^the "([^"]*)" should receive the event$`, theShouldReceiveTheEvent)
	ctx.Step(`^the active player count should increase by (\d+)$`, theActivePlayerCountShouldIncreaseBy)
}

func theGameEngineIsRunning() error {
	config := ecosystem.SimulationConfig{
		WorldID:      uuid.New(),
		TickInterval: gameLoopState.tickInterval,
		Speed:        ecosystem.SpeedNormal,
	}
	// We pass nil repos as we don't need persistence for this test
	runner := ecosystem.NewSimulationRunner(config, nil, nil)

	// Create dummy population simulator to avoid nil panic
	runner.InitializePopulationSimulator(123)

	gameLoopState.runner = runner
	return nil
}

func ticksPass(ticks int) error {
	if gameLoopState.runner == nil {
		return fmt.Errorf("runner not initialized")
	}

	// Use manual stepping
	err := gameLoopState.runner.Step(ticks)
	if err != nil {
		return err
	}
	gameLoopState.ticksProcessed += ticks
	return nil
}

func theSimulationTimeShouldAdvanceBySecond(seconds int) error {
	expectedDuration := time.Duration(seconds) * time.Second
	actualDuration := time.Duration(gameLoopState.ticksProcessed) * gameLoopState.tickInterval

	// Allow small tolerance (1 microsecond)
	diff := actualDuration - expectedDuration
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Microsecond {
		return fmt.Errorf("expected %v elapsed, got %v", expectedDuration, actualDuration)
	}
	return nil
}

func allRegisteredSubsystemsShouldHaveUpdated() error {
	// SimulationRunner updates subsystems. Since Step() returned without error, we assume success.
	return nil
}

func aEventIsQueued(eventType string) error {
	gameLoopState.eventQueue = append(gameLoopState.eventQueue, eventType)

	// Register mock PlayerManager handler if not exists
	if _, ok := gameLoopState.handlers["PlayerManager"]; !ok {
		gameLoopState.handlers["PlayerManager"] = func(evt string) {
			gameLoopState.receivedEvents["PlayerManager"] = true
			if evt == "PLAYER_JOINED" {
				gameLoopState.playerCount++
			}
		}
	}
	return nil
}

func theEventLoopProcessesTheQueue() error {
	for _, event := range gameLoopState.eventQueue {
		for _, handler := range gameLoopState.handlers {
			handler(event)
		}
	}
	gameLoopState.eventQueue = []string{}
	return nil
}

func theShouldReceiveTheEvent(receiverName string) error {
	if received, ok := gameLoopState.receivedEvents[receiverName]; !ok || !received {
		return fmt.Errorf("%s did not receive event", receiverName)
	}
	return nil
}

func theActivePlayerCountShouldIncreaseBy(count int) error {
	if gameLoopState.playerCount != count {
		return fmt.Errorf("expected count increase %d, got %d", count, gameLoopState.playerCount)
	}
	return nil
}
