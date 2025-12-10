-- Base table for all world entities
CREATE TABLE IF NOT EXISTS world_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL DEFAULT 'static',
    name VARCHAR(255) NOT NULL,
    description TEXT,
    details TEXT,
    x DOUBLE PRECISION NOT NULL,
    y DOUBLE PRECISION NOT NULL,
    z DOUBLE PRECISION DEFAULT 0,
    collision BOOLEAN DEFAULT FALSE,
    locked BOOLEAN DEFAULT FALSE,
    interactable BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_world_entities_world_id ON world_entities(world_id);
CREATE INDEX idx_world_entities_position ON world_entities(world_id, x, y);
CREATE INDEX idx_world_entities_type ON world_entities(world_id, entity_type);
CREATE INDEX idx_world_entities_name ON world_entities(world_id, LOWER(name));
