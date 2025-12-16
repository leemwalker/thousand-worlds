CREATE TABLE world_runner_state (
    world_id UUID PRIMARY KEY REFERENCES worlds(id) ON DELETE CASCADE,
    current_year BIGINT NOT NULL DEFAULT 0,
    speed INT NOT NULL DEFAULT 1,
    state VARCHAR(20) NOT NULL DEFAULT 'idle',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
