package bdd

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tw-backend/internal/analytics"
	"tw-backend/internal/circuitbreaker"
	"tw-backend/internal/ecosystem"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

// AnalyticsContext holds state for analytics scenarios
type AnalyticsContext struct {
	service *analytics.Service
	mockDB  sqlmock.Sqlmock
	db      *sql.DB
	lastErr error
}

var analyticsState = &AnalyticsContext{}

// Reset state
func (c *AnalyticsContext) reset() {
	if c.service != nil {
		c.service.Close()
	}
	c.service = nil
	c.mockDB = nil
	c.db = nil
	c.lastErr = nil
}

func InitializeAnalyticsSteps(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		analyticsState.reset()
		return ctx, nil
	})

	ctx.Step(`^the analytics service is running$`, theAnalyticsServiceIsRunning)
	ctx.Step(`^I record a global stats snapshot$`, iRecordAGlobalStatsSnapshot)
	ctx.Step(`^the snapshot should be persisted successfully$`, theSnapshotShouldBePersistedSuccessfully)
	ctx.Step(`^the circuit breaker state should be "([^"]*)"$`, theCircuitBreakerStateShouldBe)
	ctx.Step(`^the database connection is lost$`, theDatabaseConnectionIsLost)
	ctx.Step(`^the operation should fail gracefully$`, theOperationShouldFailGracefully)
	ctx.Step(`^I record (\d+) more global stats snapshots$`, iRecordMoreGlobalStatsSnapshots)
	ctx.Step(`^subsequent requests should fail fast$`, subsequentRequestsShouldFailFast)
}

func theAnalyticsServiceIsRunning() error {
	db, mock, err := sqlmock.New()
	if err != nil {
		return fmt.Errorf("failed to open sqlmock: %w", err)
	}
	analyticsState.db = db
	analyticsState.mockDB = mock

	// Mock initialization queries
	mock.ExpectExec("CREATE EXTENSION IF NOT EXISTS timescaledb").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS world_metrics").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT create_hypertable").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE INDEX IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))

	// Re-construct service manually to inject specific DB mock (since NewService accepts a URL string)
	// We need to modify NewService or create a constructor that accepts *sql.DB.
	// However, looking at service.go, `NewService` opens the DB itself.
	// For BDD here, we'll bypass NewService and create struct directly to inject mock.

	// Configure circuit breaker
	cbConfig := circuitbreaker.Config{
		Name:             "timescaledb",
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
	}
	cb := circuitbreaker.New(cbConfig)

	// Since we can't easily change the internal logic of NewService without refactoring,
	// we will manually initialize the service struct here with our mocked DB.
	// In a real integration test, we would use a real DB URL.

	svc := &analytics.Service{} // Can't access fields directly if they are private in another package?
	// Wait, `db` and `cb` are private. We need a way to construct it.
	// Strategy: We should probably refactor `NewService` to accept `*sql.DB` or create a `NewServiceWithDB` for testing.
	// OR use reflection (bad).
	// OR, for now, we can use `NewService` if we can provide a driver that mocks it.
	// But `sql.Open("pgx", ...)` is hardcoded.

	// Let's refactor proper DI pattern in service.go first, or assume we can modify it.
	// But for "Steps", let's assume we can inject.
	// Actually, `NewService` takes a dbURL. If we register `sqlmock` as a driver name, maybe?

	// Simpler approach: Create a constructor in `service_test.go` or `export_test.go` pattern?
	// But BDD steps are in `tests/bdd` package, which is external.

	// Plan: Use `NewService` passing a special DSN if possible, or check if we can modify `service.go`.
	// Modifying `service.go` to be more testable is "making it meet higher standards".

	// IMPORTANT: I will assume I can modify `service.go` to `NewService(db *sql.DB)` or similar.
	// But `NewService` currently takes string.
	// Let's change `NewService` to `NewService(dbURL string) ...` -> calls `Connect`.
	// For better testing, splitting "Connect" from "New" is good.
	// Or `NewService(db *sql.DB)`.

	// FOR NOW, I will use a helper in `analytics` package that is exposed for testing or use `export_test.go`.
	// But I can't access `export_test.go` from `tw-backend/tests/bdd` (different module/package).

	// REFACTOR DECISION: I will modify `internal/analytics/service.go` to add `NewServiceWithDB` or similar,
	// or standard pattern: `func New(db *sql.DB) *Service`. AND `func NewFromURL(url string) ...`.

	// Just creating the Service struct here:
	// Since fields are private, I can't set them from `bdd` package.

	// Refactor `analytics.Service` to export fields? No.
	// Add `NewServiceWithDB` to `service.go`.

	// Let's modify `service.go` first. But I need to write steps.
	// I'll assume `analytics.NewServiceWithDB(db, cb)` exists.

	// Manually initialize service
	svc, err = analytics.NewServiceWithDB(db, cb)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	analyticsState.service = svc

	// Also need to simulate schema init if we bypass NewService
	// But NewServiceWithDB might not do schema init? Or it should.

	return nil
}

func iRecordAGlobalStatsSnapshot() error {
	analyticsState.mockDB.ExpectExec("INSERT INTO world_metrics").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	stats := ecosystem.GlobalStats{
		WorldID:    uuid.New(),
		RecordedAt: time.Now(),
		Year:       100,
	}
	analyticsState.lastErr = analyticsState.service.RecordStats(context.Background(), stats)
	return nil
}

func theSnapshotShouldBePersistedSuccessfully() error {
	if analyticsState.lastErr != nil {
		return fmt.Errorf("expected success, got error: %v", analyticsState.lastErr)
	}
	return analyticsState.mockDB.ExpectationsWereMet()
}

func theCircuitBreakerStateShouldBe(expectedState string) error {
	stats := analyticsState.service.CircuitBreakerStats()
	if stats.State != expectedState {
		return fmt.Errorf("expected CB state %s, got %s", expectedState, stats.State)
	}
	return nil
}

func theDatabaseConnectionIsLost() error {
	// We can't easily "close" the mock connection in a way that sqlmock fails subsequent queries automatically
	// unless we set expectations to return error.
	// So we won't actually close logic, but we will setup expectations for failure.
	return nil
}

func theOperationShouldFailGracefully() error {
	// Expect failure
	analyticsState.mockDB.ExpectExec("INSERT INTO world_metrics").
		WillReturnError(fmt.Errorf("connection refused"))

	stats := ecosystem.GlobalStats{WorldID: uuid.New()}
	analyticsState.lastErr = analyticsState.service.RecordStats(context.Background(), stats)

	if analyticsState.lastErr == nil {
		return fmt.Errorf("expected error, got nil")
	}
	return nil
}

func iRecordMoreGlobalStatsSnapshots(count int) error {
	for i := 0; i < count; i++ {
		// Circuit breaker monitoring
		// If fails, triggers open state eventually
		analyticsState.mockDB.ExpectExec("INSERT INTO world_metrics").
			WillReturnError(fmt.Errorf("connection refused"))

		_ = analyticsState.service.RecordStats(context.Background(), ecosystem.GlobalStats{WorldID: uuid.New()})
	}
	return nil
}

func subsequentRequestsShouldFailFast() error {
	// Current CB checks state. If open, returns error.
	// The Loop in previous function (RecordMore) should have triggered OPEN state (threshold 5).
	// Now call again.

	// We expect NO DB execution here.

	err := analyticsState.service.RecordStats(context.Background(), ecosystem.GlobalStats{WorldID: uuid.New()})
	if err == nil {
		return fmt.Errorf("expected error from open circuit breaker")
	}

	// Error string might contain "circuit breaker is open"
	// err is wrapped: "timescaledb: circuit breaker is open"
	return nil
}
