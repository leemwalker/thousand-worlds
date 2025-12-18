CREATE TABLE IF NOT EXISTS world_simulation_snapshot (
    world_id UUID PRIMARY KEY REFERENCES worlds(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_simulation_snapshot_world_id ON world_simulation_snapshot(world_id);
