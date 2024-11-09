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
INSERT INTO schedules (rdos, anchor, user_id) VALUES
    (ARRAY[6, 7], '2024-01-01', (SELECT id FROM users WHERE email = 'elrond@rivendell.med')),
    (ARRAY[5, 6], '2024-01-02', (SELECT id FROM users WHERE email = 'aragorn@minas-tirith.med')),
    (ARRAY[4, 5], '2024-01-03', (SELECT id FROM users WHERE email = 'galadriel@lothlorien.med')),
    (ARRAY[3, 4], '2024-01-04', (SELECT id FROM users WHERE email = 'thorin@erebor.med')),
    (ARRAY[2, 3], '2024-01-05', (SELECT id FROM users WHERE email = 'bilbo@shire.med')),
    (ARRAY[1, 2], '2024-01-06', (SELECT id FROM users WHERE email = 'frodo@shire.med')),
    (ARRAY[6, 7], '2024-01-07', (SELECT id FROM users WHERE email = 'arwen@rivendell.med')),
    (ARRAY[5, 6], '2024-01-08', (SELECT id FROM users WHERE email = 'faramir@minas-tirith.med')),
    (ARRAY[4, 5], '2024-01-09', (SELECT id FROM users WHERE email = 'celeborn@lothlorien.med')),
    (ARRAY[3, 4], '2024-01-10', (SELECT id FROM users WHERE email = 'balin@erebor.med'))
ON CONFLICT (user_id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Clean up in reverse order of dependencies
-- First remove schedules (due to foreign key to users)
DELETE FROM schedules
WHERE user_id IN (
    SELECT id FROM users 
    WHERE email IN (
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

-- Then remove users (due to foreign key to facilities)
DELETE FROM users 
WHERE email IN (
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
);

-- Finally remove facilities
DELETE FROM facilities 
WHERE code IN (
    'RVDL',
    'MTGH',
    'LOTH',
    'ERBR',
    'SHIR'
);
-- +goose StatementEnd
