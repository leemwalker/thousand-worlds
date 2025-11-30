DROP INDEX IF EXISTS idx_characters_position;
DROP INDEX IF EXISTS idx_characters_world_id;
DROP INDEX IF EXISTS idx_characters_user_id;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_users_email;

DROP FUNCTION IF EXISTS cleanup_expired_sessions();

DROP TABLE IF EXISTS characters;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
