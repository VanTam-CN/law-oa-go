CREATE TABLE IF NOT EXISTS auth_token_sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    device_id VARCHAR(128) NOT NULL DEFAULT '',
    ip VARCHAR(45) NOT NULL DEFAULT '',
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    access_token_uuid VARCHAR(64) NOT NULL,
    refresh_token_uuid VARCHAR(64) NOT NULL,
    access_token_expires TIMESTAMPTZ NOT NULL,
    refresh_token_expires TIMESTAMPTZ NOT NULL,
    access_revoked_at TIMESTAMPTZ NULL,
    refresh_revoked_at TIMESTAMPTZ NULL,
    revoked_at TIMESTAMPTZ NULL,
    device_revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_auth_token_sessions_access_uuid UNIQUE (access_token_uuid),
    CONSTRAINT uq_auth_token_sessions_refresh_uuid UNIQUE (refresh_token_uuid),
    CONSTRAINT chk_auth_token_session_user CHECK (user_id > 0)
);

CREATE INDEX IF NOT EXISTS idx_auth_token_sessions_user
    ON auth_token_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_auth_token_sessions_device
    ON auth_token_sessions (device_id);
CREATE INDEX IF NOT EXISTS idx_auth_token_sessions_access_revoked
    ON auth_token_sessions (access_revoked_at);
CREATE INDEX IF NOT EXISTS idx_auth_token_sessions_refresh_revoked
    ON auth_token_sessions (refresh_revoked_at);
CREATE INDEX IF NOT EXISTS idx_auth_token_sessions_revoked
    ON auth_token_sessions (revoked_at);
CREATE INDEX IF NOT EXISTS idx_auth_token_sessions_device_revoked
    ON auth_token_sessions (device_revoked_at);
