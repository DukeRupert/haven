-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN registration_completed BOOLEAN NOT NULL DEFAULT false;

-- Update existing users to have registration completed
UPDATE users 
SET registration_completed = CASE 
    WHEN password IS NOT NULL AND password != '' THEN true 
    ELSE false 
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP COLUMN registration_completed;
-- +goose StatementEnd
