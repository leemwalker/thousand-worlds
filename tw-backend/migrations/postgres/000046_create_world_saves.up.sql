CREATE TABLE IF NOT EXISTS world_saves (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    player_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
    snapshot_key TEXT NOT NULL,           -- MinIO object key
    event_sequence BIGINT NOT NULL,       -- NATS JetStream sequence for replay alignment
    year BIGINT NOT NULL,                 -- Simulation year at save
    created_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_world_saves_world_id ON world_saves(world_id);
CREATE INDEX idx_world_saves_player_id ON world_saves(player_id);
CREATE INDEX idx_world_saves_world_year ON world_saves(world_id, year DESC);
