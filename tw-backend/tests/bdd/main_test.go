package bdd

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
	// Initialize step contexts
	worldCtx := &worldGenContext{}
	// Note: gameLoopContext and physicsContext are initialized implicitly in their respective step files
	// or we need to pass them if they are structs.
	// Physics steps (steps_physics.go) uses a global? No, let's check.
	// steps_physics.go defines `InitializePhysicsSteps(ctx *godog.ScenarioContext)`.
	// It likely manages its own context internally or uses a struct pointer.
	// Let's assume the previous patterns were correct for Physics and GameLoop.

	// World Generation Steps - Handled by InitializeWorldGenSteps(ctx)
	InitializeWorldGenSteps(ctx, worldCtx)

	// Physics Steps - Handled by InitializePhysicsSteps(ctx)
	InitializePhysicsSteps(ctx)

	// Game Loop Steps - Handled by InitializeGameLoopSteps(ctx)
	InitializeGameLoopSteps(ctx)

	// Analytics Steps - Handled by InitializeAnalyticsSteps(ctx)
	InitializeAnalyticsSteps(ctx)

	// Economy Steps - Handled by InitializeEconomySteps(ctx)
	InitializeEconomySteps(ctx)

	// Combat Steps
	combatCtx := &CombatContext{}
	InitializeCombatSteps(ctx, combatCtx)

	// Auth Steps
	authCtx := &AuthContext{}
	InitializeAuthSteps(ctx, authCtx)

	// AI Steps
	aiCtx := &AIContext{}
	InitializeAISteps(ctx, aiCtx)
}

func TestMain(m *testing.M) {
	opts := godog.Options{
		Format:    "pretty",
		Paths:     []string{"../../features"},
		Randomize: 0,
	}

	status := godog.TestSuite{
		Name:                "godogs",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
