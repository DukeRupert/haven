-- +goose Up
-- +goose StatementBegin
INSERT INTO facilities (name, code) VALUES
    ('Rivendell Medical Center', 'RVDL'),
    ('Minas Tirith General Hospital', 'MTGH'),
    ('Lothlorien Health Center', 'LOTH'),
    ('Erebor Mountain Clinic', 'ERBR'),
    ('Shire Community Hospital', 'SHIR')
ON CONFLICT (code) DO NOTHING;

-- Then seed users (using facilities' IDs)
INSERT INTO users (first_name, last_name, initials, email, password, facility_id, role) VALUES
    -- Rivendell Staff
    ('Elrond', 'Peredhel', 'EP', 'elrond@rivendell.med', '$2a$10$lC4wfrnR9TCtzH7v2nlI8O/HjBwvMhoYBSb5OfcLzwIWs8TJwh/XG', 
        (SELECT id FROM facilities WHERE code = 'RVDL'), 'super'),
    ('Arwen', 'Evenstar', 'AE', 'arwen@rivendell.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'RVDL'), 'admin'),
    -- Minas Tirith Staff
    ('Aragorn', 'Elessar', 'AE', 'aragorn@minas-tirith.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'MTGH'), 'super'),
    ('Faramir', 'Steward', 'FS', 'faramir@minas-tirith.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'MTGH'), 'user'),
    -- Lothlorien Staff
    ('Galadriel', 'Light', 'GL', 'galadriel@lothlorien.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'LOTH'), 'admin'),
    ('Celeborn', 'Elder', 'CE', 'celeborn@lothlorien.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'LOTH'), 'user'),
    -- Erebor Staff
    ('Thorin', 'Oakenshield', 'TO', 'thorin@erebor.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'ERBR'), 'super'),
    ('Balin', 'Fundin', 'BF', 'balin@erebor.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'ERBR'), 'user'),
    -- Shire Staff
    ('Bilbo', 'Baggins', 'BB', 'bilbo@shire.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'SHIR'), 'admin'),
    ('Frodo', 'Baggins', 'FB', 'frodo@shire.med', '$2a$10$xVQk0kTlMoTvQEIKN7kzPO2aHnwAWXz9SJ9nz0x1XR4b8r2j5n5Yq',
        (SELECT id FROM facilities WHERE code = 'SHIR'), 'user')
ON CONFLICT (email) DO NOTHING;

-- Finally seed schedules
INSERT INTO schedules (user_id, first_weekday, second_weekday, start_date) 
SELECT u.id, s.first_weekday, s.second_weekday, s.start_date::date
FROM (VALUES
    ('elrond@rivendell.med', 5, 6, '2024-01-01'::date),    -- Fri, Sat
    ('aragorn@minas-tirith.med', 4, 5, '2024-01-02'::date), -- Thu, Fri
    ('galadriel@lothlorien.med', 3, 4, '2024-01-03'::date), -- Wed, Thu
    ('thorin@erebor.med', 2, 3, '2024-01-04'::date),        -- Tue, Wed
    ('bilbo@shire.med', 1, 2, '2024-01-05'::date),          -- Mon, Tue
    ('frodo@shire.med', 0, 1, '2024-01-06'::date),          -- Sun, Mon
    ('arwen@rivendell.med', 5, 6, '2024-01-07'::date),      -- Fri, Sat
    ('faramir@minas-tirith.med', 4, 5, '2024-01-08'::date), -- Thu, Fri
    ('celeborn@lothlorien.med', 3, 4, '2024-01-09'::date),  -- Wed, Thu
    ('balin@erebor.med', 2, 3, '2024-01-10'::date)          -- Tue, Wed
) AS s(email, first_weekday, second_weekday, start_date)
JOIN users u ON u.email = s.email
ON CONFLICT ON CONSTRAINT unique_user_schedule DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    -- First remove schedules if the table exists
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schedules') THEN
        DELETE FROM schedules WHERE user_id IN (
            SELECT id FROM users WHERE email IN (
                'elrond@rivendell.med',
                'arwen@rivendell.med',
                'aragorn@minas-tirith.med',
                'faramir@minas-tirith.med',
                'galadriel@lothlorien.med',
                'celeborn@lothlorien.med',
                'thorin@erebor.med',
                'balin@erebor.med',
                'bilbo@shire.med',
                'frodo@shire.med'
            )
        );
    END IF;
END $$;

-- Then remove users
DELETE FROM users WHERE facility_id IN (
    SELECT id FROM facilities WHERE code IN (
        'RVDL',
        'MTGH',
        'LOTH',
        'ERBR',
        'SHIR'
    )
);

-- Finally remove facilities
DELETE FROM facilities WHERE code IN (
    'RVDL',
    'MTGH',
    'LOTH',
    'ERBR',
    'SHIR'
);
-- +goose StatementEnd
