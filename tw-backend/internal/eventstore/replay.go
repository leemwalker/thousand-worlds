package eventstore

import (
	"context"
	"time"
)

// ReplayEngine defines methods for replaying events to reconstruct state.
type ReplayEngine interface {
	ReplayEvents(ctx context.Context, aggregateID string, fromVersion, toVersion int64) ([]Event, error)
	RewindToTimestamp(ctx context.Context, aggregateID string, timestamp time.Time) ([]Event, error)
	// FastForwardFrom would typically apply events to a state, but here we just return events for now
	// as the state application logic depends on the specific aggregate.
}

// PostgresReplayEngine implements ReplayEngine using EventStore.
type PostgresReplayEngine struct {
	store EventStore
}

func NewPostgresReplayEngine(store EventStore) *PostgresReplayEngine {
	return &PostgresReplayEngine{store: store}
}

func (r *PostgresReplayEngine) ReplayEvents(ctx context.Context, aggregateID string, fromVersion, toVersion int64) ([]Event, error) {
	// We can reuse GetEventsByAggregate but we need to filter by toVersion
	// Since GetEventsByAggregate only takes fromVersion, we might need to fetch all and filter,
	// or add a new method to EventStore.
	// For efficiency, let's add a direct query here or assume we can fetch and filter.
	// Given the requirements, let's implement the query directly here or extend EventStore.
	// Extending EventStore is cleaner but for now let's just query via the store if possible?
	// No, PostgresReplayEngine has access to store, but store interface doesn't have range query.
	// Let's just use GetEventsByAggregate and filter in memory for now, or better, cast to PostgresEventStore if we need raw access?
	// Actually, the prompt says "ReplayEvents(aggregateID, fromVersion, toVersion)".
	// Let's implement it by fetching from fromVersion and filtering.

	events, err := r.store.GetEventsByAggregate(ctx, aggregateID, fromVersion)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("DEBUG: ReplayEvents fetched %d events for agg %s from ver %d\n", len(events), aggregateID, fromVersion)

	var filtered []Event
	for _, e := range events {
		if e.Version <= toVersion {
			filtered = append(filtered, e)
		} else {
			// Since events are ordered by version, we can stop early
			break
		}
	}
	return filtered, nil
}

func (r *PostgresReplayEngine) RewindToTimestamp(ctx context.Context, aggregateID string, timestamp time.Time) ([]Event, error) {
	// Fetch all events for aggregate (from version 0)
	events, err := r.store.GetEventsByAggregate(ctx, aggregateID, 0)
	if err != nil {
		return nil, err
	}

	var filtered []Event
	for _, e := range events {
		if !e.Timestamp.After(timestamp) {
			filtered = append(filtered, e)
		} else {
			// Since we can't guarantee timestamp order strictly matches version order (though it should),
			// we shouldn't break early unless we are sure.
			// But usually version order implies timestamp order.
			// Let's just filter.
		}
	}
	return filtered, nil
}
