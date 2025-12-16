CREATE TABLE simulation_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    year BIGINT NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    severity FLOAT NOT NULL DEFAULT 0.5,
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sim_events_world_year ON simulation_events(world_id, year);
CREATE INDEX idx_sim_events_type ON simulation_events(world_id, event_type);
