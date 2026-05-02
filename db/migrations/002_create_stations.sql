-- Migración 002: tabla `stations` (estaciones de esquí).
-- ----------------------------------------------------------------------------
-- Modeliza el catálogo de estaciones que se muestra en /estaciones,
-- /estacion/{id} y /forfaits. Los precios viven aquí porque cada
-- estación tiene su propio forfait con tres tarifas (adulto/niño/senior).

CREATE TABLE stations (
    id              BIGSERIAL PRIMARY KEY,

    -- Identificación
    name            TEXT        NOT NULL UNIQUE,
    location        TEXT        NOT NULL,

    -- Información meteo / operativa (snapshot, no histórico)
    distance_km     NUMERIC(6,2) NOT NULL DEFAULT 0,
    temperature_c   INTEGER     NOT NULL DEFAULT 0,
    snow_base_cm    INTEGER     NOT NULL DEFAULT 0,
    snow_new_cm     INTEGER     NOT NULL DEFAULT 0,
    slopes_open     INTEGER     NOT NULL DEFAULT 0,
    slopes_total    INTEGER     NOT NULL DEFAULT 0,
    lifts_open      INTEGER     NOT NULL DEFAULT 0,
    lifts_total     INTEGER     NOT NULL DEFAULT 0,
    last_snowfall   TEXT        NOT NULL DEFAULT '',
    altitude        TEXT        NOT NULL DEFAULT '',
    ski_km          NUMERIC(6,2) NOT NULL DEFAULT 0,
    difficulty      TEXT        NOT NULL DEFAULT '',
    phone           TEXT        NOT NULL DEFAULT '',
    image_url       TEXT        NOT NULL DEFAULT '',
    description     TEXT        NOT NULL DEFAULT '',

    -- Forfait (precios por día)
    price_adult     NUMERIC(8,2) NOT NULL DEFAULT 0,
    price_child     NUMERIC(8,2) NOT NULL DEFAULT 0,
    price_senior    NUMERIC(8,2) NOT NULL DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints semánticos
    CONSTRAINT stations_distance_nn  CHECK (distance_km   >= 0),
    CONSTRAINT stations_snow_base_nn CHECK (snow_base_cm  >= 0),
    CONSTRAINT stations_snow_new_nn  CHECK (snow_new_cm   >= 0),
    CONSTRAINT stations_slopes_open  CHECK (slopes_open   >= 0 AND slopes_open  <= slopes_total),
    CONSTRAINT stations_slopes_total CHECK (slopes_total  >= 0),
    CONSTRAINT stations_lifts_open   CHECK (lifts_open    >= 0 AND lifts_open   <= lifts_total),
    CONSTRAINT stations_lifts_total  CHECK (lifts_total   >= 0),
    CONSTRAINT stations_ski_km_nn    CHECK (ski_km        >= 0),
    CONSTRAINT stations_p_adult_nn   CHECK (price_adult   >= 0),
    CONSTRAINT stations_p_child_nn   CHECK (price_child   >= 0),
    CONSTRAINT stations_p_senior_nn  CHECK (price_senior  >= 0)
);

CREATE INDEX idx_stations_distance ON stations (distance_km);

CREATE TRIGGER trg_stations_updated_at
BEFORE UPDATE ON stations
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE  stations             IS 'Catálogo de estaciones de esquí.';
COMMENT ON COLUMN stations.distance_km IS 'Distancia (km) desde la ubicación de referencia (Madrid en la demo).';
