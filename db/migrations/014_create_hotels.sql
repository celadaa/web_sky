-- Migración 014: tablas `hotels` y `hotel_images` (módulo de hoteles).
-- ----------------------------------------------------------------------------
-- Modeliza los alojamientos asociados a cada estación que se muestran en
-- /hoteles, /hoteles/{slug} y en el paso 3 del planificador de estancia.
--
-- La reserva NO se procesa internamente: cada hotel puede llevar un enlace
-- externo (booking_url / official_url) donde el usuario completa la reserva.
-- Las imágenes se gestionan manualmente (URLs oficiales autorizadas o
-- ficheros propios en /static). NUNCA se descargan de terceros de forma
-- automática.

CREATE TABLE hotels (
    id                      BIGSERIAL PRIMARY KEY,
    station_id              BIGINT      NOT NULL REFERENCES stations(id) ON DELETE CASCADE,

    -- Identificación
    name                    TEXT        NOT NULL,
    slug                    TEXT        NOT NULL UNIQUE,
    description             TEXT        NOT NULL DEFAULT '',

    -- Localización
    address                 TEXT        NOT NULL DEFAULT '',
    city                    TEXT        NOT NULL DEFAULT '',
    country                 TEXT        NOT NULL DEFAULT 'España',
    distance_to_station_km  NUMERIC(6,2) NOT NULL DEFAULT 0,

    -- Precio orientativo (por noche) y valoración
    price_from              NUMERIC(8,2) NOT NULL DEFAULT 0,
    currency                TEXT        NOT NULL DEFAULT 'EUR',
    rating                  NUMERIC(2,1) NOT NULL DEFAULT 0,
    review_count            INTEGER     NOT NULL DEFAULT 0,

    -- Servicios, separados por coma ("wifi,spa,parking"). Se parsean en Go.
    amenities               TEXT        NOT NULL DEFAULT '',

    -- Enlaces externos (validados en el service: solo https).
    booking_url             TEXT        NOT NULL DEFAULT '',
    official_url            TEXT        NOT NULL DEFAULT '',

    -- Imagen principal para las cards (la galería vive en hotel_images).
    main_image_url          TEXT        NOT NULL DEFAULT '',

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT hotels_distance_nn CHECK (distance_to_station_km >= 0),
    CONSTRAINT hotels_price_nn    CHECK (price_from >= 0),
    CONSTRAINT hotels_rating_rng  CHECK (rating >= 0 AND rating <= 5),
    CONSTRAINT hotels_reviews_nn  CHECK (review_count >= 0)
);

CREATE INDEX idx_hotels_station ON hotels (station_id);
CREATE INDEX idx_hotels_slug    ON hotels (slug);

CREATE TRIGGER trg_hotels_updated_at
BEFORE UPDATE ON hotels
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE  hotels             IS 'Alojamientos asociados a estaciones. La reserva se completa en la web externa del hotel.';
COMMENT ON COLUMN hotels.amenities   IS 'Servicios separados por coma; se parsean en el modelo Go.';
COMMENT ON COLUMN hotels.price_from  IS 'Precio orientativo por noche, no es precio de venta interno.';

-- Galería: un hotel puede tener muchas imágenes.
CREATE TABLE hotel_images (
    id          BIGSERIAL PRIMARY KEY,
    hotel_id    BIGINT      NOT NULL REFERENCES hotels(id) ON DELETE CASCADE,
    image_url   TEXT        NOT NULL,
    alt_text    TEXT        NOT NULL DEFAULT '',
    source_name TEXT        NOT NULL DEFAULT '',
    source_url  TEXT        NOT NULL DEFAULT '',
    sort_order  INTEGER     NOT NULL DEFAULT 0,
    is_primary  BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hotel_images_hotel ON hotel_images (hotel_id, sort_order);

COMMENT ON TABLE hotel_images IS 'Galería de fotos por hotel. URLs autorizadas manualmente o ficheros propios en /static.';

-- ----------------------------------------------------------------------------
-- DATOS SEED / DEMO — ⚠️ NO SON HOTELES REALES ⚠️
-- ----------------------------------------------------------------------------
-- Hoteles ficticios de demostración para poder probar el módulo. Las fotos
-- son imágenes genéricas de montaña de Unsplash (dominio ya permitido por la
-- CSP). Sustituir o borrar cuando se carguen alojamientos reales desde el
-- panel /admin/hoteles. Idempotente gracias a ON CONFLICT (slug) DO NOTHING.

INSERT INTO hotels (
    station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url
)
SELECT s.id,
       'Hotel Demo Pie de Pistas', 'hotel-demo-pie-de-pistas',
       '[DEMO] Alojamiento ficticio de demostración. Hotel de montaña a pie de pistas con spa, desayuno buffet y guardaesquís. Sustituir por datos reales desde el panel de administración.',
       'Carretera de la Estación s/n', 'Naut Aran', 'España',
       0.30, 145.00, 'EUR', 4.6, 312,
       'wifi,spa,parking,desayuno,guardaesquís,restaurante',
       '', '',
       'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'
FROM stations s WHERE s.name = 'Baqueira Beret'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO hotels (
    station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url
)
SELECT s.id,
       'Apartahotel Demo Valle Nevado', 'apartahotel-demo-valle-nevado',
       '[DEMO] Alojamiento ficticio de demostración. Apartamentos familiares con cocina equipada y vistas al valle, a 5 minutos en coche de los remontes.',
       'Avenida del Valle 12', 'Vielha', 'España',
       4.50, 89.00, 'EUR', 4.2, 187,
       'wifi,cocina,parking,lavandería,admite mascotas',
       '', '',
       'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'
FROM stations s WHERE s.name = 'Baqueira Beret'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO hotels (
    station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url
)
SELECT s.id,
       'Hostal Demo Montaña Blanca', 'hostal-demo-montana-blanca',
       '[DEMO] Alojamiento ficticio de demostración. Hostal económico y acogedor en el centro del pueblo, con cafetería y alquiler de material en la puerta.',
       'Plaza Mayor 3', 'Monachil', 'España',
       8.00, 55.00, 'EUR', 3.9, 96,
       'wifi,cafetería,consigna',
       '', '',
       'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop'
FROM stations s WHERE s.name = 'Sierra Nevada'
ON CONFLICT (slug) DO NOTHING;

-- Galería demo para el primer hotel (imágenes genéricas Unsplash permitidas por CSP).
INSERT INTO hotel_images (hotel_id, image_url, alt_text, source_name, source_url, sort_order, is_primary)
SELECT h.id, v.url, v.alt, 'Unsplash (demo)', 'https://unsplash.com', v.ord, v.prim
FROM hotels h
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1600&auto=format&fit=crop', '[DEMO] Fachada del hotel junto a las pistas', 0, TRUE),
    ('https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1600&auto=format&fit=crop',   '[DEMO] Habitación con vistas a la montaña', 1, FALSE),
    ('https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1600&auto=format&fit=crop','[DEMO] Paisaje nevado del entorno', 2, FALSE)
) AS v(url, alt, ord, prim)
WHERE h.slug = 'hotel-demo-pie-de-pistas'
  AND NOT EXISTS (SELECT 1 FROM hotel_images hi WHERE hi.hotel_id = h.id);
