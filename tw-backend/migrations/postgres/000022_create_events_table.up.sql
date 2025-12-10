-- Create events table for event sourcing
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    aggregate_type TEXT NOT NULL,
    version BIGINT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    payload JSONB NOT NULL,
    metadata JSONB,
    
    -- Ensure version uniqueness per aggregate
    CONSTRAINT unique_aggregate_version UNIQUE (aggregate_id, version)
);

-- Index for querying events by aggregate
CREATE INDEX idx_events_aggregate_id ON events(aggregate_id, version);

-- Index for querying events by type and timestamp
CREATE INDEX idx_events_type_timestamp ON events(event_type, timestamp);

-- Index for querying all events by timestamp
CREATE INDEX idx_events_timestamp ON events(timestamp);
