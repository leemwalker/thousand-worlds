-- Performance optimization indexes for spatial queries, user activity, and JSONB searches
-- Phase 1.1: Database and Query Optimization

-- Composite index for spatial queries (world + position)
-- Improves queries like "find all characters in world X near position Y"
-- Complexity: O(log N) vs O(N) table scan
CREATE INDEX IF NOT EXISTS idx_characters_world_position 
ON characters(world_id, position);

-- Index for user activity queries
-- Improves queries like "find users who haven't logged in for X days"
-- Complexity: O(log N) for range queries on last_login
CREATE INDEX IF NOT EXISTS idx_users_last_login 
ON users(last_login) 
WHERE last_login IS NOT NULL;

-- GIN index for JSONB metadata searches on worlds
-- Improves queries like "find worlds with specific metadata properties"
-- Complexity: O(log N) for JSONB containment queries vs O(N) sequential scan
CREATE INDEX IF NOT EXISTS idx_worlds_metadata_gin 
ON worlds USING GIN(metadata);

-- Composite index for session cleanup queries
-- Improves the cleanup_expired_sessions() function performance
-- Complexity: O(log N) for finding expired sessions
CREATE INDEX IF NOT EXISTS idx_sessions_expires_user 
ON sessions(expires_at, user_id);

-- Partial index for active sessions only
-- Reduces index size by only indexing non-expired sessions
-- Complexity: Smaller index = faster lookups
CREATE INDEX IF NOT EXISTS idx_sessions_active 
ON sessions(user_id, expires_at) 
WHERE expires_at > CURRENT_TIMESTAMP;

-- Comment explaining performance impact
COMMENT ON INDEX idx_characters_world_position IS 
'Composite index for spatial queries. Reduces nearby player lookup from O(N) to O(log N).';

COMMENT ON INDEX idx_worlds_metadata_gin IS 
'GIN index for JSONB queries. Enables efficient metadata searches with @>, ?, ?& operators.';
