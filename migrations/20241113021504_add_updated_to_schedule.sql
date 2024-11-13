-- +goose Up
-- +goose StatementBegin
ALTER TABLE schedules 
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Update existing rows to have updated_at match created_at
UPDATE schedules SET updated_at = created_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE schedules 
DROP COLUMN updated_at;
-- +goose StatementEnd