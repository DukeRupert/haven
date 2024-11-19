-- +goose Up
-- +goose StatementBegin
-- Add new columns
ALTER TABLE protected_dates 
    ADD COLUMN facility_id INTEGER NOT NULL,
    ADD COLUMN user_id INTEGER NOT NULL;

-- Add foreign key constraints
ALTER TABLE protected_dates
    ADD CONSTRAINT fk_protected_dates_facility
        FOREIGN KEY (facility_id)
        REFERENCES facilities(id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT fk_protected_dates_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE;

-- Add indexes for performance
CREATE INDEX idx_facility_date ON protected_dates(facility_id, date);
CREATE INDEX idx_user_date ON protected_dates(user_id, date);

-- Populate new columns from existing relationships
UPDATE protected_dates pd
SET 
    user_id = s.user_id,
    facility_id = u.facility_id
FROM schedules s
JOIN users u ON s.user_id = u.id
WHERE pd.schedule_id = s.id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_facility_date;
DROP INDEX IF EXISTS idx_user_date;

ALTER TABLE protected_dates
    DROP CONSTRAINT IF EXISTS fk_protected_dates_facility,
    DROP CONSTRAINT IF EXISTS fk_protected_dates_user,
    DROP COLUMN IF EXISTS facility_id,
    DROP COLUMN IF EXISTS user_id;
-- +goose StatementEnd