
-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    s record;
    check_date date;
    counter integer;
    user_id integer;
    facility_id integer;
BEGIN
    FOR s IN 
        SELECT 
            schedules.*, 
            users.id as uid,
            users.facility_id as fid
        FROM schedules 
        JOIN users ON schedules.user_id = users.id 
    LOOP
        user_id := s.uid;
        facility_id := s.fid;
        
        -- First weekday
        check_date := s.start_date;
        counter := 0;
        WHILE check_date < (s.start_date + interval '1 year') LOOP
            IF EXTRACT(DOW FROM check_date) = s.first_weekday THEN
                counter := counter + 1;
                IF counter % 3 = 0 THEN
                    INSERT INTO protected_dates (schedule_id, date, available, user_id, facility_id)
                    VALUES (s.id, check_date, false, user_id, facility_id)
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
                    INSERT INTO protected_dates (schedule_id, date, available, user_id, facility_id)
                    VALUES (s.id, check_date, false, user_id, facility_id)
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