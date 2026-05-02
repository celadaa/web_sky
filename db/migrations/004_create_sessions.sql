-- Migración 004: tabla `sessions` (cookies de sesión persistidas).

CREATE TABLE sessions (
    token       TEXT        PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,

    CONSTRAINT sessions_expires_after_created CHECK (expires_at > created_at)
);

CREATE INDEX idx_sessions_user_id    ON sessions (user_id);
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);

COMMENT ON TABLE sessions IS 'Tokens de sesión (cookie HttpOnly).';
