-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS http_sessions (
    id BIGSERIAL PRIMARY KEY,
    key BYTEA NOT NULL,
    data BYTEA NOT NULL,
    created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    modified_on TIMESTAMPTZ,
    expires_on TIMESTAMPTZ
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS http_sessions_key_idx ON http_sessions (key);
CREATE INDEX IF NOT EXISTS http_sessions_expiry_idx ON http_sessions (expires_on);

-- Add comment to table
COMMENT ON TABLE http_sessions IS 'Stores web session data for user authentication';

-- Add comments to columns
COMMENT ON COLUMN http_sessions.id IS 'Unique identifier for the session record';
COMMENT ON COLUMN http_sessions.key IS 'Session key used for lookup';
COMMENT ON COLUMN http_sessions.data IS 'Encrypted session data';
COMMENT ON COLUMN http_sessions.created_on IS 'Timestamp when the session was created';
COMMENT ON COLUMN http_sessions.modified_on IS 'Timestamp when the session was last modified';
COMMENT ON COLUMN http_sessions.expires_on IS 'Timestamp when the session expires';


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS http_sessions_expiry_idx;
DROP INDEX IF EXISTS http_sessions_key_idx;
DROP TABLE IF EXISTS http_sessions;
-- +goose StatementEnd
