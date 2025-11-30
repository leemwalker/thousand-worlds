CREATE TABLE IF NOT EXISTS worlds (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    shape VARCHAR(50) NOT NULL,
    radius FLOAT,
    bounds_min_x FLOAT,
    bounds_min_y FLOAT,
    bounds_min_z FLOAT,
    bounds_max_x FLOAT,
    bounds_max_y FLOAT,
    bounds_max_z FLOAT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_worlds_created_at ON worlds(created_at);
