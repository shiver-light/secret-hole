-- 用户扩展字段
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS nickname TEXT,
  ADD COLUMN IF NOT EXISTS avatar_b64 TEXT,
  ADD COLUMN IF NOT EXISTS gender_color TEXT;

-- 前台心跳（App 前台且活跃）
CREATE TABLE IF NOT EXISTS user_presence (
  user_id        BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  is_foreground  BOOLEAN NOT NULL DEFAULT false,
  last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 临时好友（保留 streak 与累计时长；到期不删行，仅清时间）
CREATE TABLE IF NOT EXISTS temp_friendships (
  user_min BIGINT REFERENCES users(id) ON DELETE CASCADE,
  user_max BIGINT REFERENCES users(id) ON DELETE CASCADE,
  streak   INT NOT NULL DEFAULT 0,
  last_started_at TIMESTAMPTZ,
  active_until   TIMESTAMPTZ,
  total_active_seconds BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (user_min, user_max),
  CHECK (user_min < user_max)
);
CREATE INDEX IF NOT EXISTS idx_temp_friendships_until
  ON temp_friendships (active_until);

-- 临时好友期间的点对点消息（到期将被清理）
CREATE TABLE IF NOT EXISTS direct_messages (
  id         BIGSERIAL PRIMARY KEY,
  sender_id  BIGINT REFERENCES users(id) ON DELETE CASCADE,
  recv_id    BIGINT REFERENCES users(id) ON DELETE CASCADE,
  body       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_dm_pair_time
  ON direct_messages (sender_id, recv_id, created_at);

-- 树洞公共消息的一次回复
ALTER TABLE messages
  ADD COLUMN IF NOT EXISTS reply_sender_id BIGINT,
  ADD COLUMN IF NOT EXISTS reply_body TEXT;
