-- Create refresh_tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for token lookups
CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_tokens_expires ON refresh_tokens(expires_at) WHERE revoked = FALSE;
