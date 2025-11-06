CREATE INDEX IF NOT EXISTS idx_messages_unread ON messages (status) WHERE status = 0;
CREATE INDEX IF NOT EXISTS idx_messages_expiry ON messages (expires_at) WHERE expires_at IS NOT NULL;