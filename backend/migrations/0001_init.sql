-- +goose Up
-- +goose StatementBegin
CREATE TABLE teams (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE players (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    number     INT,
    position   VARCHAR(50),
    team_id    INT REFERENCES teams(id) ON DELETE SET NULL,
    device_id  VARCHAR(255) UNIQUE NOT NULL,
    status     VARCHAR(50)  NOT NULL DEFAULT 'offline',
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id               SERIAL PRIMARY KEY,
    player_id        INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    start_time       TIMESTAMPTZ NOT NULL,
    end_time         TIMESTAMPTZ,
    duration_minutes INT,
    session_type     VARCHAR(50) NOT NULL DEFAULT 'training',
    status           VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sessions_player ON sessions(player_id, start_time DESC);
CREATE UNIQUE INDEX idx_sessions_one_active
    ON sessions(player_id) WHERE status = 'active';

CREATE TABLE raw_data (
    id           BIGSERIAL PRIMARY KEY,
    session_id   INT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    player_id    INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    timestamp    TIMESTAMPTZ NOT NULL,
    latitude     DOUBLE PRECISION,
    longitude    DOUBLE PRECISION,
    altitude     DOUBLE PRECISION,
    gps_speed    DOUBLE PRECISION,
    gps_accuracy DOUBLE PRECISION,
    accel_x      DOUBLE PRECISION,
    accel_y      DOUBLE PRECISION,
    accel_z      DOUBLE PRECISION,
    gyro_x       DOUBLE PRECISION,
    gyro_y       DOUBLE PRECISION,
    gyro_z       DOUBLE PRECISION,
    heart_rate   INT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_raw_session_time ON raw_data(session_id, timestamp);

CREATE TABLE metrics (
    id                 SERIAL PRIMARY KEY,
    session_id         INT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    player_id          INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    speed_max          DOUBLE PRECISION NOT NULL DEFAULT 0,
    speed_avg          DOUBLE PRECISION NOT NULL DEFAULT 0,
    speed_min          DOUBLE PRECISION NOT NULL DEFAULT 0,
    distance_total     DOUBLE PRECISION NOT NULL DEFAULT 0,
    sprint_count       INT NOT NULL DEFAULT 0,
    sprint_avg_length  DOUBLE PRECISION NOT NULL DEFAULT 0,
    sprint_max_length  DOUBLE PRECISION NOT NULL DEFAULT 0,
    acceleration_count INT NOT NULL DEFAULT 0,
    player_load        DOUBLE PRECISION NOT NULL DEFAULT 0,
    low_intensity_time  INT NOT NULL DEFAULT 0,
    med_intensity_time  INT NOT NULL DEFAULT 0,
    high_intensity_time INT NOT NULL DEFAULT 0,
    jump_height_max    DOUBLE PRECISION NOT NULL DEFAULT 0,
    jump_count         INT NOT NULL DEFAULT 0,
    hr_max             INT,
    hr_avg             INT,
    hr_min             INT,
    duration_minutes   INT NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (session_id)
);

CREATE TABLE heatmap_data (
    id              SERIAL PRIMARY KEY,
    session_id      INT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    player_id       INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    grid_x          INT NOT NULL,
    grid_y          INT NOT NULL,
    time_seconds    INT NOT NULL DEFAULT 0,
    avg_speed       DOUBLE PRECISION NOT NULL DEFAULT 0,
    intensity_level VARCHAR(20) NOT NULL DEFAULT 'low',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (session_id, grid_x, grid_y)
);
CREATE INDEX idx_heatmap_session ON heatmap_data(session_id, grid_x, grid_y);

CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    username      VARCHAR(255) UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(50)  NOT NULL DEFAULT 'coach',
    team_id       INT REFERENCES teams(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS heatmap_data;
DROP TABLE IF EXISTS metrics;
DROP TABLE IF EXISTS raw_data;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
