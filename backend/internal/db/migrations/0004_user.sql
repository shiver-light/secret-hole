-- 用户表（若已存在就 ALTER：这里给完整定义）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables
                   WHERE table_name = 'users' AND table_schema = 'public') THEN
        CREATE TABLE users (
            id            BIGSERIAL PRIMARY KEY,
            name          TEXT NOT NULL UNIQUE,
            avatar_base64 TEXT,
            gender_color  TEXT,
            password_hash TEXT NOT NULL,
            created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
        );
    ELSE
        -- 如果表已存在但缺少列，就按需添加
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                       WHERE table_name = 'users' AND column_name = 'avatar_base64') THEN
            ALTER TABLE users ADD COLUMN avatar_base64 TEXT;
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                       WHERE table_name = 'users' AND column_name = 'gender_color') THEN
            ALTER TABLE users ADD COLUMN gender_color TEXT;
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                       WHERE table_name = 'users' AND column_name = 'password_hash') THEN
            ALTER TABLE users ADD COLUMN password_hash TEXT NOT NULL DEFAULT '';
        END IF;
    END IF;
END
$$;

-- 登录token 表（短期会话/长期都可；可选加过期时间）
CREATE TABLE IF NOT EXISTS auth_tokens (
  token           TEXT PRIMARY KEY,
  user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_auth_tokens_user ON auth_tokens(user_id);