-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS facilities(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    code CHAR(4) NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS facilities;
-- +goose StatementEnd
