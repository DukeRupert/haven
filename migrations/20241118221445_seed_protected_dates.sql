-- In a new migration file (something like 20240318000000_generate_protected_dates.sql)
-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    s record;
    check_date date;
    counter integer;
BEGIN
    FOR s IN SELECT * FROM schedules LOOP
        -- First weekday
        check_date := s.start_date;
        counter := 0;
        WHILE check_date < (s.start_date + interval '1 year') LOOP
            IF EXTRACT(DOW FROM check_date) = s.first_weekday THEN
                counter := counter + 1;
                IF counter % 3 = 0 THEN
                    INSERT INTO protected_dates (schedule_id, date, available)
                    VALUES (s.id, check_date, false)
                    ON CONFLICT (schedule_id, date) DO NOTHING;
                END IF;
            END IF;
            check_date := check_date + interval '1 day';
        END LOOP;

        -- Second weekday
        check_date := s.start_date;
        counter := 0;
        WHILE check_date < (s.start_date + interval '1 year') LOOP
            IF EXTRACT(DOW FROM check_date) = s.second_weekday THEN
                counter := counter + 1;
                IF counter % 3 = 0 THEN
                    INSERT INTO protected_dates (schedule_id, date, available)
                    VALUES (s.id, check_date, false)
                    ON CONFLICT (schedule_id, date) DO NOTHING;
                END IF;
            END IF;
            check_date := check_date + interval '1 day';
        END LOOP;
    END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE protected_dates;
-- +goose StatementEnd
