CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    device_hash TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TABLE IF NOT EXISTS daily_quota (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    quota_date DATE NOT NULL,
    sent_count INT NOT NULL DEFAULT 0,
    recv_count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, quota_date)
);


CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    author_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    body_cipher TEXT NOT NULL,
    read_duration INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    read_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    status SMALLINT NOT NULL DEFAULT 0
);