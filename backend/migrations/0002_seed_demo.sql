-- +goose Up
-- +goose StatementBegin
INSERT INTO teams (id, name) VALUES (1, 'Demo FC')
    ON CONFLICT DO NOTHING;
SELECT setval(pg_get_serial_sequence('teams', 'id'), (SELECT max(id) FROM teams));

INSERT INTO players (name, number, position, team_id, device_id, status) VALUES
    ('Ivan Petrov',   7,  'MF', 1, 'ESP32-TRACKER-001', 'offline'),
    ('Sergey Orlov',  10, 'FW', 1, 'ESP32-TRACKER-002', 'offline'),
    ('Anton Belov',   4,  'DF', 1, 'ESP32-TRACKER-003', 'offline')
    ON CONFLICT (device_id) DO NOTHING;

-- password for all demo users is "password" (bcrypt hash)
INSERT INTO users (username, email, password_hash, role, team_id) VALUES
    ('coach', 'coach@demo.local', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqxU7m0l3g0kU9r6Yb0k0k3g0kU9r', 'coach', 1)
    ON CONFLICT (username) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE username = 'coach';
DELETE FROM players WHERE device_id IN ('ESP32-TRACKER-001','ESP32-TRACKER-002','ESP32-TRACKER-003');
DELETE FROM teams WHERE id = 1;
-- +goose StatementEnd
