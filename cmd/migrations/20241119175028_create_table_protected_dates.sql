-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS protected_dates (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    schedule_id INTEGER NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    facility_id INTEGER NOT NULL REFERENCES facilities(id) ON DELETE RESTRICT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    available BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT unique_schedule_date UNIQUE(schedule_id, date)
);

CREATE INDEX idx_protected_dates_schedule_id ON protected_dates(schedule_id);
CREATE INDEX idx_protected_dates_date ON protected_dates(date);
CREATE INDEX idx_protected_dates_available ON protected_dates(available);
CREATE INDEX idx_facility_date ON protected_dates(facility_id, date);
CREATE INDEX idx_user_date ON protected_dates(user_id, date);

CREATE TRIGGER update_protected_dates_updated_at
    BEFORE UPDATE ON protected_dates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_protected_dates_updated_at ON protected_dates;
DROP INDEX IF EXISTS idx_protected_dates_schedule_id;
DROP INDEX IF EXISTS idx_protected_dates_date;
DROP INDEX IF EXISTS idx_protected_dates_available;
DROP INDEX IF EXISTS idx_facility_date;
DROP INDEX IF EXISTS idx_user_date;
DROP TABLE IF EXISTS protected_dates;
-- +goose StatementEnd