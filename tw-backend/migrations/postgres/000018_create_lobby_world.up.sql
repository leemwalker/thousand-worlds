-- Insert the Lobby "world" record
-- The Lobby is a special virtual space where players gather before entering actual worlds
-- First ensure the system user exists, then create the lobby

-- Fix: Add username column if missing (since it was missing in 000013)
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='username') THEN 
        ALTER TABLE users ADD COLUMN username VARCHAR(255);
        CREATE UNIQUE INDEX idx_users_username ON users(username);
    END IF; 
END $$;

INSERT INTO users (user_id, email, password_hash, username)
VALUES ('00000000-0000-0000-0000-000000000001', 'system@mud.com', 'system_hash_placeholder', 'system')
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO worlds (id, name, shape, metadata, owner_id) VALUES
('00000000-0000-0000-0000-000000000000', 'Lobby', 'virtual', '{"description": "The Grand Lobby - a virtual gathering space between worlds"}', '00000000-0000-0000-0000-000000000001')
ON CONFLICT (id) DO NOTHING;
