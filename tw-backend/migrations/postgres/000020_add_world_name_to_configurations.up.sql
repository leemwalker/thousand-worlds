ALTER TABLE world_configurations 
ADD COLUMN world_name VARCHAR(100);

CREATE UNIQUE INDEX idx_world_config_name_unique 
ON world_configurations(LOWER(world_name));
