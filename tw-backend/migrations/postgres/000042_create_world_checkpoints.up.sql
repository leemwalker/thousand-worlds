CREATE TABLE world_checkpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    year BIGINT NOT NULL,
    checkpoint_type VARCHAR(20) NOT NULL, -- 'full' or 'delta'
    state_data BYTEA NOT NULL, -- Compressed gob data
    species_count INT NOT NULL,
    population_sum BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (world_id, year)
);

CREATE INDEX idx_checkpoints_world ON world_checkpoints(world_id);
