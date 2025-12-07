-- Rollback performance optimization indexes

DROP INDEX IF EXISTS idx_sessions_expires_user;
DROP INDEX IF EXISTS idx_worlds_metadata_gin;
DROP INDEX IF EXISTS idx_users_last_login;
DROP INDEX IF EXISTS idx_characters_world_position;
