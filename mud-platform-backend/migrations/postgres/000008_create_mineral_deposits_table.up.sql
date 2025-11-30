CREATE TABLE IF NOT EXISTS mineral_deposits (
    deposit_id UUID PRIMARY KEY,
    mineral_type VARCHAR(50) NOT NULL,
    formation_type VARCHAR(50) NOT NULL,
    location_x FLOAT NOT NULL,
    location_y FLOAT NOT NULL,
    depth FLOAT NOT NULL,
    quantity INT NOT NULL,
    concentration FLOAT NOT NULL,
    vein_size VARCHAR(20) NOT NULL,
    geological_age FLOAT NOT NULL,
    vein_shape VARCHAR(20) NOT NULL,
    vein_orientation_x FLOAT NOT NULL,
    vein_orientation_y FLOAT NOT NULL,
    vein_length FLOAT NOT NULL,
    vein_width FLOAT NOT NULL,
    surface_visible BOOLEAN NOT NULL,
    required_depth FLOAT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mineral_location ON mineral_deposits (location_x, location_y);
CREATE INDEX idx_mineral_type ON mineral_deposits (mineral_type);

CREATE TABLE IF NOT EXISTS mineral_depletion (
    deposit_id UUID PRIMARY KEY REFERENCES mineral_deposits(deposit_id),
    original_quantity INT NOT NULL,
    current_quantity INT NOT NULL,
    first_extracted TIMESTAMP WITH TIME ZONE,
    depleted_at TIMESTAMP WITH TIME ZONE,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
