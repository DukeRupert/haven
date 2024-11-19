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

CREATE TABLE IF NOT EXISTS protected_dates (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    schedule_id INTEGER NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    available BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT unique_schedule_date UNIQUE(schedule_id, date)
);

-- Create indexes for commonly queried fields
CREATE INDEX idx_schedules_user_id ON schedules(user_id);
CREATE INDEX idx_protected_dates_schedule_id ON protected_dates(schedule_id);
CREATE INDEX idx_protected_dates_date ON protected_dates(date);
CREATE INDEX idx_protected_dates_available ON protected_dates(available);

-- Create triggers for updated_at
CREATE TRIGGER update_schedules_updated_at
    BEFORE UPDATE ON schedules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_protected_dates_updated_at
    BEFORE UPDATE ON protected_dates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_protected_dates_updated_at ON protected_dates;
DROP TRIGGER IF EXISTS update_schedules_updated_at ON schedules;
DROP INDEX IF EXISTS idx_schedules_user_id;
DROP INDEX IF EXISTS idx_protected_dates_schedule_id;
DROP INDEX IF EXISTS idx_protected_dates_date;
DROP INDEX IF EXISTS idx_protected_dates_available;
DROP TABLE IF EXISTS protected_dates;
DROP TABLE IF EXISTS schedules;
-- +goose StatementEnd
