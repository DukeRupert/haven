-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_protected_dates()
RETURNS TRIGGER AS $$
DECLARE
    check_date date;
    counter integer;
BEGIN
    -- Delete existing protected dates for this schedule
    DELETE FROM protected_dates WHERE schedule_id = NEW.id;
    
    -- First weekday
    check_date := NEW.start_date;
    counter := 0;
    WHILE check_date < (NEW.start_date + interval '1 year') LOOP
        IF EXTRACT(DOW FROM check_date) = NEW.first_weekday THEN
            counter := counter + 1;
            IF counter % 3 = 0 THEN
                INSERT INTO protected_dates (schedule_id, date, available, user_id, facility_id)
                VALUES (NEW.id, check_date, false, NEW.user_id, 
                    (SELECT facility_id FROM users WHERE id = NEW.user_id));
            END IF;
        END IF;
        check_date := check_date + interval '1 day';
    END LOOP;

    -- Second weekday
    check_date := NEW.start_date;
    counter := 0;
    WHILE check_date < (NEW.start_date + interval '1 year') LOOP
        IF EXTRACT(DOW FROM check_date) = NEW.second_weekday THEN
            counter := counter + 1;
            IF counter % 3 = 0 THEN
                INSERT INTO protected_dates (schedule_id, date, available, user_id, facility_id)
                VALUES (NEW.id, check_date, false, NEW.user_id,
                    (SELECT facility_id FROM users WHERE id = NEW.user_id));
            END IF;
        END IF;
        check_date := check_date + interval '1 day';
    END LOOP;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER schedule_update_trigger
AFTER INSERT OR UPDATE ON schedules
FOR EACH ROW
EXECUTE FUNCTION update_protected_dates();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS schedule_update_trigger ON schedules;
DROP FUNCTION IF EXISTS update_protected_dates();
-- +goose StatementEnd