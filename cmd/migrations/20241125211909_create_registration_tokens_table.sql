-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS registration_tokens (
    user_id INT PRIMARY KEY REFERENCES users(id),
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_registration_tokens_token ON registration_tokens(token);
CREATE INDEX idx_registration_tokens_expires_at ON registration_tokens(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_registration_tokens_expires_at;
DROP INDEX IF EXISTS idx_registration_tokens_token;
DROP TABLE IF EXISTS registration_tokens;
-- +goose StatementEnd
