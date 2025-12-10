package eventstore

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EventStore defines the methods for storing and retrieving events.
type EventStore interface {
	AppendEvent(ctx context.Context, event Event) error
	GetEventsByAggregate(ctx context.Context, aggregateID string, fromVersion int64) ([]Event, error)
	GetEventsByType(ctx context.Context, eventType EventType, fromTimestamp, toTimestamp time.Time) ([]Event, error)
	GetAllEvents(ctx context.Context, fromTimestamp time.Time, limit int) ([]Event, error)
}

// PostgresEventStore implements EventStore using PostgreSQL.
type PostgresEventStore struct {
	pool *pgxpool.Pool
}

// NewPostgresEventStore creates a new PostgresEventStore.
func NewPostgresEventStore(pool *pgxpool.Pool) *PostgresEventStore {
	return &PostgresEventStore{pool: pool}
}

func (s *PostgresEventStore) AppendEvent(ctx context.Context, event Event) error {
	query := `
		INSERT INTO events (id, event_type, aggregate_id, aggregate_type, version, timestamp, payload, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := s.pool.Exec(ctx, query,
		event.ID,
		event.EventType,
		event.AggregateID,
		event.AggregateType,
		event.Version,
		event.Timestamp,
		event.Payload,
		event.Metadata,
	)
	return err
}

func (s *PostgresEventStore) GetEventsByAggregate(ctx context.Context, aggregateID string, fromVersion int64) ([]Event, error) {
	query := `
		SELECT id, event_type, aggregate_id, aggregate_type, version, timestamp, payload, metadata
		FROM events
		WHERE aggregate_id = $1 AND version >= $2
		ORDER BY version ASC
	`
	rows, err := s.pool.Query(ctx, query, aggregateID, fromVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(
			&e.ID,
			&e.EventType,
			&e.AggregateID,
			&e.AggregateType,
			&e.Version,
			&e.Timestamp,
			&e.Payload,
			&e.Metadata,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *PostgresEventStore) GetEventsByType(ctx context.Context, eventType EventType, fromTimestamp, toTimestamp time.Time) ([]Event, error) {
	query := `
		SELECT id, event_type, aggregate_id, aggregate_type, version, timestamp, payload, metadata
		FROM events
		WHERE event_type = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC
	`
	rows, err := s.pool.Query(ctx, query, eventType, fromTimestamp, toTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(
			&e.ID,
			&e.EventType,
			&e.AggregateID,
			&e.AggregateType,
			&e.Version,
			&e.Timestamp,
			&e.Payload,
			&e.Metadata,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *PostgresEventStore) GetAllEvents(ctx context.Context, fromTimestamp time.Time, limit int) ([]Event, error) {
	query := `
		SELECT id, event_type, aggregate_id, aggregate_type, version, timestamp, payload, metadata
		FROM events
		WHERE timestamp >= $1
		ORDER BY timestamp ASC
		LIMIT $2
	`
	rows, err := s.pool.Query(ctx, query, fromTimestamp, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(
			&e.ID,
			&e.EventType,
			&e.AggregateID,
			&e.AggregateType,
			&e.Version,
			&e.Timestamp,
			&e.Payload,
			&e.Metadata,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
