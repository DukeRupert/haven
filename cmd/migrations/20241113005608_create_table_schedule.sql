-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS schedules (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    first_weekday INTEGER NOT NULL CHECK (first_weekday >= 0 AND first_weekday <= 6),
    second_weekday INTEGER NOT NULL CHECK (second_weekday >= 0 AND second_weekday <= 6),
    start_date DATE NOT NULL,
    CONSTRAINT unique_user_schedule UNIQUE(user_id)
);

CREATE INDEX idx_schedules_user_id ON schedules(user_id);

CREATE TRIGGER update_schedules_updated_at
    BEFORE UPDATE ON schedules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_schedules_updated_at ON schedules;
DROP INDEX IF EXISTS idx_schedules_user_id;
DROP TABLE IF EXISTS schedules;
-- +goose StatementEnd
