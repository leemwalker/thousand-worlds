CREATE TABLE IF NOT EXISTS resource_nodes (
    node_id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    rarity VARCHAR(20) NOT NULL,
    location_x FLOAT NOT NULL,
    location_y FLOAT NOT NULL,
    location_z FLOAT NOT NULL,
    quantity INT NOT NULL,
    max_quantity INT NOT NULL,
    regen_rate FLOAT NOT NULL DEFAULT 0.0,
    regen_cooldown_hours INT NOT NULL DEFAULT 0,
    last_harvested TIMESTAMP WITH TIME ZONE,
    required_skill VARCHAR(50) NOT NULL,
    min_skill_level INT NOT NULL,
    
    -- Mineral-specific fields (only populated for mineral type)
    mineral_deposit_id UUID REFERENCES mineral_deposits(deposit_id),
    depth FLOAT,
    
    -- Animal-specific fields (only populated for animal type)
    species_id UUID,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Spatial index for location-based queries
CREATE INDEX idx_resource_location ON resource_nodes (location_x, location_y);

-- Type filtering
CREATE INDEX idx_resource_type ON resource_nodes (type);

-- Mineral deposit references
CREATE INDEX idx_resource_mineral ON resource_nodes (mineral_deposit_id) WHERE mineral_deposit_id IS NOT NULL;

-- Species references
CREATE INDEX idx_resource_species ON resource_nodes (species_id) WHERE species_id IS NOT NULL;

-- Rarity filtering
CREATE INDEX idx_resource_rarity ON resource_nodes (rarity);

-- Junction table for biome affinity (many-to-many)
CREATE TABLE IF NOT EXISTS resource_biome_affinity (
    node_id UUID REFERENCES resource_nodes(node_id) ON DELETE CASCADE,
    biome_type VARCHAR(50) NOT NULL,
    PRIMARY KEY (node_id, biome_type)
);

CREATE INDEX idx_biome_affinity_node ON resource_biome_affinity (node_id);
CREATE INDEX idx_biome_affinity_biome ON resource_biome_affinity (biome_type);
