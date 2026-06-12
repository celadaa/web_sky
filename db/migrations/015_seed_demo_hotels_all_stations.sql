-- Migración 015: seed demo — 3 hoteles por CADA estación.
-- ----------------------------------------------------------------------------
-- ⚠️ DATOS FICTICIOS DE DEMOSTRACIÓN — NO SON HOTELES REALES ⚠️
--
-- Sustituye el seed mínimo de la migración 014 (3 hoteles en total) por un
-- catálogo demo uniforme: cada estación recibe 3 alojamientos ficticios
-- (hotel a pie de pistas, apartamentos y hostal económico) con precios
-- derivados del forfait de la estación para que haya variedad realista.
--
-- Idempotente: slugs únicos por estación (sufijo = id de estación) con
-- ON CONFLICT (slug) DO NOTHING. Las imágenes son genéricas de Unsplash
-- (dominio permitido por la CSP). Borrar/editar desde /admin/hoteles al
-- cargar alojamientos reales.

-- Retiramos los 3 hoteles demo de la migración 014 para no duplicar
-- (eran seeds nuestros marcados [DEMO]; sus imágenes caen por CASCADE).
DELETE FROM hotels WHERE slug IN (
    'hotel-demo-pie-de-pistas',
    'apartahotel-demo-valle-nevado',
    'hostal-demo-montana-blanca'
);

-- 1/3 — Hotel a pie de pistas (gama alta)
INSERT INTO hotels (
    station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url
)
SELECT s.id,
       'Hotel Demo ' || s.name,
       'hotel-demo-' || s.id,
       '[DEMO] Alojamiento ficticio de demostración. Hotel de montaña a pie de pistas de ' || s.name ||
       ' con spa, desayuno buffet y guardaesquís. Sustituir por datos reales desde el panel de administración.',
       'Carretera de la Estación s/n',
       split_part(s.location, ',', 1),
       'España',
       0.3 + (s.id % 4) * 0.2,
       GREATEST(90, ROUND(s.price_child * 2.8)),
       'EUR',
       4.2 + (s.id % 7) * 0.1,
       80 + (s.id * 37) % 320,
       'wifi,spa,parking,desayuno,guardaesquís,restaurante',
       '', '',
       'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'
FROM stations s
ON CONFLICT (slug) DO NOTHING;

-- 2/3 — Apartamentos familiares (gama media)
INSERT INTO hotels (
    station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url
)
SELECT s.id,
       'Apartamentos Demo ' || s.name,
       'apartamentos-demo-' || s.id,
       '[DEMO] Alojamiento ficticio de demostración. Apartamentos familiares con cocina equipada y vistas a ' || s.name ||
       ', a pocos minutos en coche de los remontes.',
       'Avenida del Valle 12',
       split_part(s.location, ',', 1),
       'España',
       2.0 + (s.id % 5) * 0.8,
       GREATEST(60, ROUND(s.price_child * 1.9)),
       'EUR',
       3.9 + (s.id % 9) * 0.1,
       40 + (s.id * 53) % 260,
       'wifi,cocina,parking,lavandería,admite mascotas',
       '', '',
       'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'
FROM stations s
ON CONFLICT (slug) DO NOTHING;

-- 3/3 — Hostal económico (gama básica)
INSERT INTO hotels (
    station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url
)
SELECT s.id,
       'Hostal Demo ' || s.name,
       'hostal-demo-' || s.id,
       '[DEMO] Alojamiento ficticio de demostración. Hostal económico y acogedor en el pueblo más cercano a ' || s.name ||
       ', con cafetería y alquiler de material en la puerta.',
       'Plaza Mayor 3',
       split_part(s.location, ',', 1),
       'España',
       5.0 + (s.id % 6) * 1.1,
       GREATEST(38, ROUND(s.price_child * 1.2)),
       'EUR',
       3.5 + (s.id % 10) * 0.1,
       15 + (s.id * 29) % 180,
       'wifi,cafetería,consigna',
       '', '',
       'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop'
FROM stations s
ON CONFLICT (slug) DO NOTHING;

-- Galería demo (3 fotos genéricas) para todos los hoteles demo sin galería.
INSERT INTO hotel_images (hotel_id, image_url, alt_text, source_name, source_url, sort_order, is_primary)
SELECT h.id, v.url, '[DEMO] ' || v.alt || ' — ' || h.name, 'Unsplash (demo)', 'https://unsplash.com', v.ord, v.ord = 0
FROM hotels h
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1600&auto=format&fit=crop', 'Vista exterior', 0),
    ('https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1600&auto=format&fit=crop',   'Habitación', 1),
    ('https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1600&auto=format&fit=crop','Paisaje del entorno', 2)
) AS v(url, alt, ord)
WHERE h.slug LIKE '%demo%'
  AND NOT EXISTS (SELECT 1 FROM hotel_images hi WHERE hi.hotel_id = h.id);
