-- Add username column to users table
ALTER TABLE users ADD COLUMN username VARCHAR(50);

-- Backfill existing users with username derived from email (part before @)
UPDATE users SET username = SPLIT_PART(email, '@', 1) WHERE username IS NULL;

-- Make username NOT NULL and UNIQUE
ALTER TABLE users ALTER COLUMN username SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);

-- Create index on username
CREATE INDEX idx_users_username ON users(username);
