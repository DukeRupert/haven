-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS schedules(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rdos INTEGER[],
    anchor DATE,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS schedules;
-- +goose StatementEnd
