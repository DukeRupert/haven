-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS schedules (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    first_day INTEGER NOT NULL CHECK (first_day >= 0 AND first_day <= 6),
    second_day INTEGER NOT NULL CHECK (second_day >= 0 AND second_day <= 6),
    start_date DATE NOT NULL
);

CREATE TABLE IF NOT EXISTS protected_dates (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    schedule_id INTEGER NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    available BOOLEAN NOT NULL DEFAULT FALSE
);

-- Create indexes for commonly queried fields
CREATE INDEX idx_schedules_user_id ON schedules(user_id);
CREATE INDEX idx_protected_dates_schedule_id ON protected_dates(schedule_id);
CREATE INDEX idx_protected_dates_date ON protected_dates(date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS protected_dates;
DROP TABLE IF EXISTS schedules;

DROP INDEX IF EXISTS idx_schedules_user_id;
DROP INDEX IF EXISTS idx_protected_dates_schedule_id;
DROP INDEX IF EXISTS idx_protected_dates_date;
-- +goose StatementEnd