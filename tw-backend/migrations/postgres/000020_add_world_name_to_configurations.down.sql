DROP INDEX IF EXISTS idx_world_config_name_unique;
ALTER TABLE world_configurations DROP COLUMN IF EXISTS world_name;
