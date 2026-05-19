-- Migración 013: datos de ejemplo para `lodgings`.
-- ----------------------------------------------------------------------------
-- Por cada estación del catálogo (`stations`) genera 3 alojamientos:
--
--   1. Hotel céntrico de gama media-alta, a pie de pistas o muy cerca.
--   2. Apartamento turístico a 5-10 minutos en coche, más barato.
--   3. Hostal/albergue económico para presupuesto ajustado.
--
-- Precios y distancias varían determinísticamente con el id de la estación
-- para que cada estación tenga una variedad realista sin tener que
-- enumerar 80 líneas. Las imágenes son URLs de Unsplash ya autorizadas
-- por la CSP (img-src https://images.unsplash.com).
--
-- Idempotente: si la migración se aplica dos veces, el WHERE NOT EXISTS
-- evita duplicar (se considera duplicado un mismo nombre para la misma
-- estación).

-- Hotel céntrico — alta gama
INSERT INTO lodgings (station_id, name, kind, image_url, distance_km, rating, price_eur, zone, description)
SELECT
    s.id,
    'Hotel ' || s.name || ' Snow Lodge',
    'hotel',
    'https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=1200&auto=format&fit=crop',
    ROUND( (0.2 + (s.id % 5) * 0.15)::numeric , 2),                    -- 0.2 a 0.8 km
    ROUND( (4.2 + ((s.id * 7) % 9) * 0.07)::numeric , 1),              -- 4.2 a 4.8
    ROUND( (110 + (s.id % 8) * 15)::numeric , 2),                      -- 110 a 215 €/noche
    'A pie de pistas',
    'Hotel boutique con spa, vistas a la montaña, desayuno bufé y acceso directo a remontes. Ideal para esquiadores que quieren comodidad sin moverse.'
FROM stations s
WHERE NOT EXISTS (
    SELECT 1 FROM lodgings l
    WHERE l.station_id = s.id AND l.name = 'Hotel ' || s.name || ' Snow Lodge'
);

-- Apartamento turístico — gama media
INSERT INTO lodgings (station_id, name, kind, image_url, distance_km, rating, price_eur, zone, description)
SELECT
    s.id,
    'Apartamentos ' || s.name || ' Pistas',
    'apartamento',
    'https://images.unsplash.com/photo-1502672260266-1c1ef2d93688?q=80&w=1200&auto=format&fit=crop',
    ROUND( (1.5 + (s.id % 4) * 0.5)::numeric , 2),                     -- 1.5 a 3.0 km
    ROUND( (3.9 + ((s.id * 11) % 10) * 0.07)::numeric , 1),            -- 3.9 a 4.6
    ROUND( (70 + (s.id % 6) * 10)::numeric , 2),                       -- 70 a 120 €/noche
    'A 5 min en coche',
    'Apartamento totalmente equipado para 4-6 personas con cocina, salón y plaza de garaje. Perfecto para familias o grupos.'
FROM stations s
WHERE NOT EXISTS (
    SELECT 1 FROM lodgings l
    WHERE l.station_id = s.id AND l.name = 'Apartamentos ' || s.name || ' Pistas'
);

-- Hostal/albergue económico
INSERT INTO lodgings (station_id, name, kind, image_url, distance_km, rating, price_eur, zone, description)
SELECT
    s.id,
    CASE WHEN s.id % 2 = 0 THEN 'Albergue ' || s.name ELSE 'Hostal ' || s.name || ' Backpackers' END,
    CASE WHEN s.id % 2 = 0 THEN 'albergue' ELSE 'hostal' END,
    'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop',
    ROUND( (3.0 + (s.id % 5) * 0.4)::numeric , 2),                     -- 3.0 a 4.6 km
    ROUND( (3.7 + ((s.id * 13) % 8) * 0.07)::numeric , 1),             -- 3.7 a 4.2
    ROUND( (35 + (s.id % 5) * 8)::numeric , 2),                        -- 35 a 67 €/noche
    'Pueblo cercano',
    'Habitaciones compartidas o privadas con baño común, zona social con chimenea y desayuno opcional. La opción más asequible para esquiadores.'
FROM stations s
WHERE NOT EXISTS (
    SELECT 1 FROM lodgings l
    WHERE l.station_id = s.id
      AND l.name IN ('Albergue ' || s.name, 'Hostal ' || s.name || ' Backpackers')
);
