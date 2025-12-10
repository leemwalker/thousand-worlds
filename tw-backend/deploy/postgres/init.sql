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

-- Create worlds table
CREATE TABLE IF NOT EXISTS worlds (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    shape VARCHAR(20) NOT NULL,  -- 'sphere', 'cube', 'infinite'
    radius NUMERIC(12,2),  -- meters (required for sphere)
    bounds_min_x NUMERIC(12,2),  -- for cube worlds
    bounds_max_x NUMERIC(12,2),
    bounds_min_y NUMERIC(12,2),
    bounds_max_y NUMERIC(12,2),
    bounds_min_z NUMERIC(12,2),
    bounds_max_z NUMERIC(12,2),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create entities table (Cartesian coordinates)
CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY,
    world_id UUID NOT NULL REFERENCES worlds(id),
    position GEOMETRY(POINTZ, 0) NOT NULL,
    metadata JSONB
);

-- Create spatial index
CREATE INDEX IF NOT EXISTS idx_entities_position ON entities USING GIST (position);
CREATE INDEX IF NOT EXISTS idx_entities_world_position ON entities (world_id, position);

