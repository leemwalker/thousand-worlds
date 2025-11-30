-- World Interviews Table
-- Stores interview session state for world creation interviews

CREATE TABLE IF NOT EXISTS world_interviews (
    id UUID PRIMARY KEY,
    player_id UUID NOT NULL,
    current_category VARCHAR(50) NOT NULL,
    current_topic_index INT NOT NULL,
    answers JSONB NOT NULL DEFAULT '{}',
    history JSONB NOT NULL DEFAULT '[]',
    is_complete BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for finding player's active interviews
CREATE INDEX IF NOT EXISTS idx_world_interviews_player_id ON world_interviews(player_id);

-- Index for cleanup/archival queries
CREATE INDEX IF NOT EXISTS idx_world_interviews_created_at ON world_interviews(created_at);

-- Index for completed interview queries
CREATE INDEX IF NOT EXISTS idx_world_interviews_is_complete ON world_interviews(is_complete);
