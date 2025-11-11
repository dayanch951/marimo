-- Drop indexes
DROP INDEX IF EXISTS idx_refresh_tokens_revoked;
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_refresh_tokens_token;

-- Drop refresh_tokens table
DROP TABLE IF EXISTS refresh_tokens;
