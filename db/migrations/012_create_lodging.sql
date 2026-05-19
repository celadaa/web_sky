-- Migración 012: alojamientos cercanos a las estaciones + reservas.
-- ----------------------------------------------------------------------------
-- Modelo:
--   `lodgings`         → catálogo de hoteles/hostales/apartamentos asociados a
--                        una `stations` concreta. Imagen, distancia, valoración
--                        y precio orientativo por noche.
--   `lodging_bookings` → reservas confirmadas (o simuladas) por usuarios
--                        autenticados. Snapshot del nombre, tipo, precio por
--                        noche y total para que el ticket sea legible aunque
--                        el catálogo cambie en el futuro (mismo enfoque que
--                        `order_items` para forfaits).
--
-- NO toca tablas existentes. Si la app se ejecuta sin esta migración, el
-- código Go detecta la lista de alojamientos como vacía y la sección
-- "Alojamientos cercanos" simplemente no se muestra (no rompe nada).

CREATE TABLE lodgings (
    id            BIGSERIAL PRIMARY KEY,
    station_id    BIGINT       NOT NULL REFERENCES stations(id) ON DELETE CASCADE,

    name          TEXT         NOT NULL,
    kind          TEXT         NOT NULL,            -- hotel | hostal | apartamento | albergue | casa_rural
    image_url     TEXT         NOT NULL DEFAULT '',
    distance_km   NUMERIC(5,2) NOT NULL,            -- a la estación o pistas
    rating        NUMERIC(2,1) NOT NULL DEFAULT 0,  -- 0.0 a 5.0
    price_eur     NUMERIC(8,2) NOT NULL,            -- precio orientativo por noche
    zone          TEXT         NOT NULL DEFAULT '', -- p.ej. "Centro de Pradollano"
    description   TEXT         NOT NULL DEFAULT '',

    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT lodgings_kind_ok      CHECK (kind IN ('hotel','hostal','apartamento','albergue','casa_rural')),
    CONSTRAINT lodgings_distance_nn  CHECK (distance_km >= 0),
    CONSTRAINT lodgings_rating_range CHECK (rating >= 0 AND rating <= 5),
    CONSTRAINT lodgings_price_nn     CHECK (price_eur >= 0)
);

CREATE INDEX idx_lodgings_station_id ON lodgings (station_id);
CREATE INDEX idx_lodgings_price      ON lodgings (price_eur);

CREATE TRIGGER trg_lodgings_updated_at
BEFORE UPDATE ON lodgings
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE lodgings IS 'Catálogo de alojamientos cercanos a estaciones de esquí.';

-- ----------------------------------------------------------------------------

CREATE TABLE lodging_bookings (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT       NOT NULL REFERENCES users(id)     ON DELETE RESTRICT,
    lodging_id      BIGINT       NOT NULL REFERENCES lodgings(id)  ON DELETE RESTRICT,

    -- Snapshots para historial — el ticket sigue siendo legible aunque
    -- el alojamiento cambie de precio o desaparezca del catálogo.
    lodging_name    TEXT         NOT NULL,
    lodging_kind    TEXT         NOT NULL,
    station_id      BIGINT       NOT NULL REFERENCES stations(id)  ON DELETE RESTRICT,
    station_name    TEXT         NOT NULL,

    check_in        DATE         NOT NULL,
    check_out       DATE         NOT NULL,
    nights          INTEGER      NOT NULL,
    guests          INTEGER      NOT NULL,
    price_per_night NUMERIC(8,2) NOT NULL,
    total_eur       NUMERIC(10,2) NOT NULL,
    status          TEXT         NOT NULL DEFAULT 'confirmed',  -- confirmed | cancelled

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT lodging_bookings_kind_ok    CHECK (lodging_kind IN ('hotel','hostal','apartamento','albergue','casa_rural')),
    CONSTRAINT lodging_bookings_dates_ok   CHECK (check_out > check_in),
    CONSTRAINT lodging_bookings_nights_pos CHECK (nights > 0),
    CONSTRAINT lodging_bookings_guests_pos CHECK (guests > 0 AND guests <= 10),
    CONSTRAINT lodging_bookings_ppn_nn     CHECK (price_per_night >= 0),
    CONSTRAINT lodging_bookings_total_nn   CHECK (total_eur >= 0),
    CONSTRAINT lodging_bookings_status_ok  CHECK (status IN ('confirmed','cancelled'))
);

CREATE INDEX idx_lodging_bookings_user_id    ON lodging_bookings (user_id);
CREATE INDEX idx_lodging_bookings_lodging_id ON lodging_bookings (lodging_id);
CREATE INDEX idx_lodging_bookings_created_at ON lodging_bookings (created_at DESC);

COMMENT ON TABLE lodging_bookings IS 'Reservas de alojamiento confirmadas por usuarios.';
