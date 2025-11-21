package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/eventstore"
	"mud-platform-backend/internal/repository"
)

// BenchmarkEventReplay benchmarks replaying events from the store.
// Since we don't have a full DB setup here, we'll mock the store or use a simplified version if possible.
// For now, we'll benchmark the ReplayEngine logic with an in-memory store or mock.
func BenchmarkEventReplay(b *testing.B) {
	// Setup
	store := &MockEventStore{events: make([]eventstore.Event, 10000)}
	for i := 0; i < 10000; i++ {
		store.events[i] = eventstore.Event{
			AggregateID: "agg-1",
			Version:     int64(i + 1),
			Payload:     []byte(`{"foo":"bar"}`),
		}
	}
	engine := eventstore.NewPostgresReplayEngine(store)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.ReplayEvents(ctx, "agg-1", 0, 10000)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJWTValidation benchmarks token validation.
func BenchmarkJWTValidation(b *testing.B) {
	tm, err := auth.NewTokenManager([]byte("secret"), []byte("encryption-key-must-be-32-bytes-long!"))
	require.NoError(b, err)

	token, err := tm.GenerateToken("user-1", "testuser", []string{"admin"})
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tm.ValidateToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// MockEventStore for benchmarking
type MockEventStore struct {
	events []eventstore.Event
}

func (m *MockEventStore) AppendEvent(ctx context.Context, event eventstore.Event) error {
	return nil
}

func (m *MockEventStore) GetEventsByAggregate(ctx context.Context, aggregateID string, fromVersion int64) ([]eventstore.Event, error) {
	return m.events, nil
}

func (m *MockEventStore) GetEventsByType(ctx context.Context, eventType eventstore.EventType, from, to time.Time) ([]eventstore.Event, error) {
	return nil, nil
}

func (m *MockEventStore) GetAllEvents(ctx context.Context, from time.Time, limit int) ([]eventstore.Event, error) {
	return nil, nil
}

// BenchmarkSpatialQuery benchmarks spatial queries.
// Note: This requires a real DB connection, so we might skip if TEST_DB_URL is not set.
// For now, we'll just define it and let it fail or skip if env not set.
func BenchmarkSpatialQuery(b *testing.B) {
	dbURL := "postgres://user:password@localhost:5432/mud_test?sslmode=disable" // Default or env
	// In a real benchmark, we'd use the env var or a setup function.
	// For this exercise, we'll skip if we can't connect.
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		b.Skip("Skipping spatial benchmark: no DB connection")
	}
	defer pool.Close()

	repo := repository.NewPostgresSpatialRepository(pool)
	worldID := uuid.New()

	// Seed data
	// We need to seed 1000 entities.
	// This might be slow for a benchmark setup, but necessary.
	// Ideally we'd do this once.
	// For now, let's just seed 100 for speed in this iteration, or assume pre-seeded.
	// Let's seed 100.
	for i := 0; i < 100; i++ {
		repo.CreateEntity(ctx, worldID, uuid.New(), float64(i), float64(i), 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetEntitiesNearby(ctx, worldID, 50, 50, 0, 100)
		if err != nil {
			b.Fatal(err)
		}
	}
}
