CREATE TABLE IF NOT EXISTS character_inventory (
    id UUID PRIMARY KEY,
    character_id UUID NOT NULL REFERENCES characters(character_id) ON DELETE CASCADE,
    item_id UUID NOT NULL, -- For now just a UUID, later FK to items table
    quantity INT NOT NULL DEFAULT 1,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT quantity_positive CHECK (quantity > 0)
);

CREATE INDEX idx_character_inventory_character_id ON character_inventory(character_id);
