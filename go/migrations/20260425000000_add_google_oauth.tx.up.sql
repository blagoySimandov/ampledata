ALTER TYPE source_type ADD VALUE IF NOT EXISTS 'google_sheets';

CREATE TABLE user_oauth_tokens (
    user_id VARCHAR NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR NOT NULL,
    access_token TEXT NOT NULL DEFAULT '',
    refresh_token TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, provider)
);

CREATE TABLE oauth_states (
    state VARCHAR NOT NULL PRIMARY KEY,
    user_id VARCHAR NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_oauth_states_expires_at ON oauth_states (expires_at);
