-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create events table
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    event_type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(255) NOT NULL,
    version BIGINT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    payload JSONB NOT NULL,
    metadata JSONB,
    UNIQUE(aggregate_id, version)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_events_aggregate_id_version ON events (aggregate_id, version);
CREATE INDEX IF NOT EXISTS idx_events_event_type ON events (event_type);
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events (timestamp);
