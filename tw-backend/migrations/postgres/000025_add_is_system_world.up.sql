ALTER TABLE worlds ADD COLUMN is_system_world BOOLEAN DEFAULT FALSE;
UPDATE worlds SET is_system_world = TRUE WHERE name = 'Lobby';
