-- Migración 005: tabla `favorites` (relación N:M usuarios ↔ estaciones).

CREATE TABLE favorites (
    user_id    BIGINT      NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    station_id BIGINT      NOT NULL REFERENCES stations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, station_id)
);

CREATE INDEX idx_favorites_station ON favorites (station_id);

COMMENT ON TABLE favorites IS 'Estaciones marcadas como favoritas por cada usuario.';
