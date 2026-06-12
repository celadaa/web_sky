-- Migración 016: catálogo de HOTELES REALES (sustituye los seeds demo).
-- ----------------------------------------------------------------------------
-- Hoteles reales verificados (junio 2026) cercanos a cada estación.
--
-- IMPORTANTE:
--  * price_from es ORIENTATIVO (tarifa baja típica de temporada); la reserva
--    y el precio real se consultan en el enlace externo.
--  * rating/review_count se dejan a 0 (la plantilla los oculta): rellenar
--    desde /admin/hoteles con datos verificados si se desea mostrarlos.
--  * Las IMÁGENES son fotos GENÉRICAS de montaña/alojamiento con licencia
--    Unsplash (dominio permitido por la CSP). NO son fotos de los hoteles:
--    las fotos reales se ven en el enlace de Booking/web oficial. Para usar
--    fotos reales hace falta autorización del hotel o una API licenciada.
--  * booking_url usa enlaces de búsqueda de Booking (no deep-links), que no
--    se rompen si el hotel cambia de ficha.
--  * Distancias a pistas aproximadas (km por carretera/remontes).
--
-- Idempotente: ON CONFLICT (slug) DO NOTHING. Las estaciones se localizan
-- con ILIKE para tolerar tildes.

-- Fuera los seeds ficticios de las migraciones 014/015 (eran demos nuestros).
DELETE FROM hotels
WHERE slug LIKE 'hotel-demo-%'
   OR slug LIKE 'apartamentos-demo-%'
   OR slug LIKE 'hostal-demo-%'
   OR slug IN ('hotel-demo-pie-de-pistas','apartahotel-demo-valle-nevado','hostal-demo-montana-blanca');

-- Imágenes genéricas (licencia Unsplash, dominio permitido por la CSP):
--  A) https://images.unsplash.com/photo-1605540436563-5bca919ae766  (montaña nevada)
--  B) https://images.unsplash.com/photo-1518602164578-cd0074062767  (paisaje invernal)
--  C) https://images.unsplash.com/photo-1551882547-ff40c63fe5fa     (interior alojamiento)

-- ─── Baqueira Beret ─────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Val de Neu', 'hotel-val-de-neu',
  'Hotel 5★ Gran Lujo en el corazón de Baqueira 1500, a unos 100 m de la telecabina. Spa, gastronomía de alto nivel y guardaesquís. Precio orientativo; foto ilustrativa de archivo.',
  'Baqueira', 'España', 0.1, 320, 'spa,wifi,restaurante,guardaesquís,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Val+de+Neu+Baqueira+Beret', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Montarto', 'hotel-montarto',
  'Clásico 4★ de Baqueira 1500 junto a las pistas (a unos 350 m del TS Bosque), con spa, gimnasio y descuentos en alquiler de material. Precio orientativo; foto ilustrativa de archivo.',
  'Baqueira', 'España', 0.3, 150, 'spa,gimnasio,wifi,restaurante,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Montarto+Baqueira+Beret', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Tuc Blanc', 'hotel-tuc-blanc',
  'Hotel 3★ a pie de pista en Baqueira 1500, a unos 100 m del telesilla Bosque: se sale esquiando. Piscina cubierta, sauna y jacuzzi. Precio orientativo; foto ilustrativa de archivo.',
  'Baqueira', 'España', 0.1, 120, 'piscina,sauna,wifi,restaurante,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Tuc+Blanc+Baqueira', 'https://www.hoteltucblancbaqueira.com/',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Baqueira%'
ON CONFLICT (slug) DO NOTHING;

-- ─── La Molina ──────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel HG La Molina', 'hotel-hg-la-molina',
  'Hotel de la cadena HG junto a las pistas de La Molina, con piscina y ambiente familiar. Precio orientativo; foto ilustrativa de archivo.',
  'La Molina', 'España', 0.3, 110, 'piscina,wifi,restaurante,parking,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=HG+La+Molina', 'https://www.hghoteles.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Solineu', 'hotel-solineu',
  'Hotel 3★ en el núcleo urbano de la estación de La Molina, con spa, sauna y baño de vapor. Precio orientativo; foto ilustrativa de archivo.',
  'La Molina', 'España', 0.5, 95, 'spa,sauna,wifi,restaurante',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Solineu+La+Molina', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Adserà', 'hotel-adsera',
  'Hotel de montaña muy cercano a las estaciones de La Molina y Masella, en plena Cerdanya. Precio orientativo; foto ilustrativa de archivo.',
  'La Molina', 'España', 1.0, 85, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Adsera+La+Molina', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'La Molina'
ON CONFLICT (slug) DO NOTHING;

-- ─── Masella (dominio Alp 2500; mismos alojamientos de la Cerdanya) ────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel HG La Molina (Alp 2500)', 'hotel-hg-la-molina-masella',
  'Hotel de la cadena HG en La Molina, dentro del dominio Alp 2500 conectado con Masella. Precio orientativo; foto ilustrativa de archivo.',
  'La Molina', 'España', 6.0, 110, 'piscina,wifi,restaurante,parking,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=HG+La+Molina', 'https://www.hghoteles.com/',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Adserà (Alp 2500)', 'hotel-adsera-masella',
  'Hotel de montaña en la Cerdanya, a un corto trayecto de los accesos de Masella por el dominio Alp 2500. Precio orientativo; foto ilustrativa de archivo.',
  'La Molina', 'España', 6.5, 85, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Adsera+La+Molina', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Solineu (Alp 2500)', 'hotel-solineu-masella',
  'Hotel 3★ con spa en La Molina, bien comunicado con Masella dentro del dominio Alp 2500. Precio orientativo; foto ilustrativa de archivo.',
  'La Molina', 'España', 7.0, 95, 'spa,sauna,wifi,restaurante',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Solineu+La+Molina', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Masella'
ON CONFLICT (slug) DO NOTHING;

-- ─── Vall de Núria ──────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Vall de Núria', 'hotel-vall-de-nuria',
  'El hotel del santuario de Núria, junto a las pistas; se llega con el tren cremallera desde Ribes de Freser. Precio orientativo; foto ilustrativa de archivo.',
  'Queralbs', 'España', 0.1, 120, 'wifi,restaurante,actividades familiares',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Vall+de+Nuria', 'https://www.valldenuria.cat/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel-Spa Resguard dels Vents', 'hotel-resguard-dels-vents',
  'Hotel-spa 4★ con vistas a la Vall de Ribes, a unos minutos del cremallera de Núria en Ribes de Freser. Precio orientativo; foto ilustrativa de archivo.',
  'Ribes de Freser', 'España', 14.0, 130, 'spa,wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Resguard+dels+Vents+Ribes+de+Freser', 'https://hotelresguard.com/',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Catalunya Ribes de Freser', 'hotel-catalunya-ribes',
  'Hotel sencillo en Ribes de Freser, cerca de la estación del cremallera que sube a Vall de Núria. Precio orientativo; foto ilustrativa de archivo.',
  'Ribes de Freser', 'España', 14.0, 70, 'wifi,restaurante',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Catalunya+Ribes+de+Freser', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Vall de N%ria'
ON CONFLICT (slug) DO NOTHING;

-- ─── Vallter 2000 ───────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel La Coma', 'hotel-la-coma-setcases',
  'Hotel 3★ en Setcases, el pueblo más cercano a Vallter 2000, recomendado por la propia estación. Precio orientativo; foto ilustrativa de archivo.',
  'Setcases', 'España', 12.0, 95, 'wifi,restaurante,piscina,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+La+Coma+Setcases', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Grèvol Spa', 'hotel-grevol-spa',
  'Hotel 4★ con spa en la Vall de Camprodon, de estilo alpino, a un corto trayecto de Vallter 2000. Precio orientativo; foto ilustrativa de archivo.',
  'Llanars', 'España', 20.0, 140, 'spa,piscina,wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Grevol+Llanars', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Hotelet del Bac', 'hotelet-del-bac',
  'Pequeño hotel con encanto en Camprodon, buena base para esquiar en Vallter 2000 y visitar la vall. Precio orientativo; foto ilustrativa de archivo.',
  'Camprodon', 'España', 18.0, 90, 'wifi,desayuno',
  'https://www.booking.com/searchresults.es.html?ss=Hotelet+del+Bac+Camprodon', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Vallter%'
ON CONFLICT (slug) DO NOTHING;

-- ─── Espot Esquí ────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Saurat', 'hotel-saurat',
  'Hotel familiar 3★ en el centro de Espot, a unos 3 km de las pistas y puerta del parque nacional de Aigüestortes. Precio orientativo; foto ilustrativa de archivo.',
  'Espot', 'España', 3.0, 90, 'wifi,restaurante,salón con chimenea,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Saurat+Espot', 'https://www.hotelsaurat.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Roca Blanca', 'hotel-roca-blanca',
  'Hotel de montaña en el pueblo de Espot con habitaciones con bañera de hidromasaje. Precio orientativo; foto ilustrativa de archivo.',
  'Espot', 'España', 3.0, 95, 'wifi,restaurante,jacuzzi,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Roca+Blanca+Espot', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Espot%'
ON CONFLICT (slug) DO NOTHING;

-- ─── Port Ainé ──────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Port Ainé 2000', 'hotel-port-aine-2000',
  'Hotel 3★ a 2.000 m, a pie de pistas de Port Ainé (Skipallars), con alquiler de material y cocina catalana. Precio orientativo; foto ilustrativa de archivo.',
  'Rialp', 'España', 0.1, 85, 'wifi,restaurante,gimnasio,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Port+Aine+2000+Rialp', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Condes del Pallars', 'hotel-condes-del-pallars',
  'Gran hotel clásico de Rialp, base habitual para esquiar en Port Ainé y Espot. Precio orientativo; foto ilustrativa de archivo.',
  'Rialp', 'España', 15.0, 75, 'wifi,restaurante,piscina,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Condes+del+Pallars+Rialp', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Port Ain%'
ON CONFLICT (slug) DO NOTHING;

-- ─── Boí Taüll ──────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Taüll', 'hotel-taull',
  'Hotel funcional y acogedor en el Pla de l''Ermita, en el corazón de la Vall de Boí, a pocos km de las pistas. Precio orientativo; foto ilustrativa de archivo.',
  'La Vall de Boí', 'España', 6.0, 85, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Taull+Vall+de+Boi', 'https://www.boitaullresort.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('SNÖ Vall de Boí', 'sno-vall-de-boi',
  'Aparthotel de la cadena SNÖ en el Pla de l''Ermita, uno de los alojamientos mejor situados para Boí Taüll. Precio orientativo; foto ilustrativa de archivo.',
  'La Vall de Boí', 'España', 6.0, 95, 'wifi,cocina,parking,admite mascotas',
  'https://www.booking.com/searchresults.es.html?ss=SNO+Vall+de+Boi', 'https://www.snohotels.com/',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Romànic', 'hotel-romanic-boi',
  'Hotel familiar en la Vall de Boí, cerca del conjunto románico Patrimonio de la Humanidad y de la estación. Precio orientativo; foto ilustrativa de archivo.',
  'La Vall de Boí', 'España', 12.0, 80, 'wifi,restaurante',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Romanic+Vall+de+Boi', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Bo% Ta%'
ON CONFLICT (slug) DO NOTHING;

-- ─── Tavascan ───────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Estanys Blaus', 'hotel-estanys-blaus',
  'Hotel 3★ en el pueblo de Tavascan, en pleno Parc Natural de l''Alt Pirineu, cerca de la estación. Precio orientativo; foto ilustrativa de archivo.',
  'Tavascan', 'España', 6.0, 85, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Estanys+Blaus+Tavascan', 'https://hotelestanysblaus.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Llacs de Cardós', 'hotel-llacs-de-cardos',
  'Hotel familiar con jardín junto al río en Tavascan; organiza raquetas y esquí nórdico en invierno. Precio orientativo; foto ilustrativa de archivo.',
  'Tavascan', 'España', 6.0, 70, 'wifi,desayuno,jardín',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Llacs+de+Cardos+Tavascan', 'https://hotelllacsdecardos.com/',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Marxant', 'hotel-marxant',
  'Pequeño hotel de montaña en Tavascan, en la Vall de Cardós, muy bien valorado. Precio orientativo; foto ilustrativa de archivo.',
  'Tavascan', 'España', 6.0, 75, 'wifi,restaurante',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Marxant+Tavascan', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Tavascan'
ON CONFLICT (slug) DO NOTHING;

-- ─── Port del Comte ─────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Port 1730', 'hotel-port-1730',
  'Hotel a pie de pistas de Port del Comte, a unos metros del remonte. Precio orientativo; foto ilustrativa de archivo.',
  'La Coma i la Pedra', 'España', 0.1, 80, 'wifi,restaurante,bar,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Port+1730+Port+del+Comte', 'https://www.litthotels.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Serhs Ski Port del Comte', 'hotel-serhs-ski-port-del-comte',
  'Hotel a pie de pistas de Port del Comte, pensado para grupos y familias. Precio orientativo; foto ilustrativa de archivo.',
  'La Coma i la Pedra', 'España', 0.2, 75, 'wifi,restaurante,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=Serhs+Ski+Port+del+Comte', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Hostal Piteus', 'hostal-piteus',
  'Alojamiento con encanto en un edificio del s. XIV en Sant Llorenç de Morunys, a unos 20 km de la estación. Precio orientativo; foto ilustrativa de archivo.',
  'Sant Llorenç de Morunys', 'España', 20.0, 60, 'wifi,restaurante',
  'https://www.booking.com/searchresults.es.html?ss=Hostal+Piteus+Sant+Llorenc+de+Morunys', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Port del Comte'
ON CONFLICT (slug) DO NOTHING;

-- ─── Candanchú ──────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Edelweiss Candanchú', 'hotel-edelweiss-candanchu',
  'Hotel 3★ a unos 50 m del telesilla de Candanchú, con terrazas acristaladas, gimnasio y sauna. Precio orientativo; foto ilustrativa de archivo.',
  'Candanchú', 'España', 0.1, 95, 'wifi,restaurante,gimnasio,sauna,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Edelweiss+Candanchu', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Tobazo', 'hotel-tobazo',
  'Hotel 3★ con acceso directo a pistas en Candanchú, junto al Parque Nacional de los Pirineos. Precio orientativo; foto ilustrativa de archivo.',
  'Candanchú', 'España', 0.2, 90, 'wifi,restaurante,bar,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Tobazo+Candanchu', 'https://www.hoteltobazo.es/',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Candanchú', 'hotel-candanchu',
  'Hotel clásico de la estación, con vistas al valle del Aragón y a las pistas de Candanchú. Precio orientativo; foto ilustrativa de archivo.',
  'Candanchú', 'España', 0.3, 85, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Candanchu', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Candanch%'
ON CONFLICT (slug) DO NOTHING;

-- ─── Astún ──────────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Europa Astún', 'hotel-europa-astun',
  'El hotel de la propia estación de Astún, a pie de pistas, con guardaesquís gratuito y bar con solárium. Precio orientativo; foto ilustrativa de archivo.',
  'Astún', 'España', 0.2, 100, 'wifi,restaurante,bar,guardaesquís,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Europa+Astun', 'https://www.hoteleuropa-astun.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Canfranc Estación, a Royal Hideaway Hotel', 'canfranc-estacion-royal-hideaway',
  'Hotel 5★ en la histórica estación internacional de Canfranc, a unos 10 km de Astún y Candanchú. Precio orientativo; foto ilustrativa de archivo.',
  'Canfranc', 'España', 10.0, 250, 'spa,wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Canfranc+Estacion+Royal+Hideaway', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Ast%n'
ON CONFLICT (slug) DO NOTHING;

-- ─── Formigal ───────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('SNÖ Hotel Formigal', 'sno-hotel-formigal',
  'Hotel en la urbanización de Formigal con traslado gratuito a pistas en invierno. Precio orientativo; foto ilustrativa de archivo.',
  'Formigal', 'España', 0.5, 110, 'wifi,restaurante,bar,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=SNO+Hotel+Formigal', 'https://www.snohotelformigal.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Abba Formigal', 'abba-formigal',
  'Hotel 4★ en Formigal con vistas a la montaña, spa y ambiente après-ski. Precio orientativo; foto ilustrativa de archivo.',
  'Formigal', 'España', 0.6, 130, 'spa,wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Abba+Formigal', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('SNÖ Villa de Sallent', 'sno-villa-de-sallent',
  'Hotel 4★ de esencia tradicional en Sallent de Gállego, a unos minutos de los accesos de Formigal. Precio orientativo; foto ilustrativa de archivo.',
  'Sallent de Gállego', 'España', 4.0, 95, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=SNO+Villa+de+Sallent', 'https://www.snovilladesallent.com/',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Formigal'
ON CONFLICT (slug) DO NOTHING;

-- ─── Panticosa ──────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Sabocos', 'hotel-sabocos',
  'Hotel en el pueblo de Panticosa, a pocos minutos a pie de la telecabina de la estación. Precio orientativo; foto ilustrativa de archivo.',
  'Panticosa', 'España', 0.4, 90, 'wifi,restaurante,piscina',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Sabocos+Panticosa', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Continental — Balneario de Panticosa', 'hotel-continental-panticosa',
  'Hotel 4★ del Resort Balneario de Panticosa, con 190 habitaciones y acceso al circuito termal. Precio orientativo; foto ilustrativa de archivo.',
  'Panticosa', 'España', 8.0, 110, 'spa,termas,wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Continental+Balneario+de+Panticosa', 'https://www.panticosa.com/',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Gran Hotel — Balneario de Panticosa', 'gran-hotel-panticosa',
  'Histórico Gran Hotel 4★ (solo adultos) del Balneario de Panticosa, en un valle a 1.636 m. Precio orientativo; foto ilustrativa de archivo.',
  'Panticosa', 'España', 8.0, 140, 'spa,termas,wifi,restaurante,solo adultos',
  'https://www.booking.com/searchresults.es.html?ss=Gran+Hotel+Balneario+de+Panticosa', 'https://www.panticosa.com/',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Panticosa'
ON CONFLICT (slug) DO NOTHING;

-- ─── Cerler ─────────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel HG Cerler', 'hotel-hg-cerler',
  'Hotel con zona wellness y vistas al valle de Benasque, junto a la estación de Cerler. Precio orientativo; foto ilustrativa de archivo.',
  'Cerler', 'España', 0.3, 120, 'spa,wifi,restaurante,guardaesquís,parking',
  'https://www.booking.com/searchresults.es.html?ss=HG+Cerler', 'https://www.hghoteles.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Evenia Monte Alba', 'hotel-evenia-monte-alba',
  'Hotel 3★ en Cerler a unos 200 m de los remontes, con piscina climatizada y animación. Precio orientativo; foto ilustrativa de archivo.',
  'Cerler', 'España', 0.2, 95, 'piscina,sauna,wifi,restaurante,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=Evenia+Monte+Alba+Cerler', 'https://eveniahotels.com/',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Cerler'
ON CONFLICT (slug) DO NOTHING;

-- ─── Grandvalira ────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Sport Hotel Hermitage & Spa', 'sport-hotel-hermitage-spa',
  'El único 5★ a pie de pistas de Grandvalira (Soldeu), con gran spa y gastronomía de chefs con estrella Michelin. Precio orientativo; foto ilustrativa de archivo.',
  'Soldeu', 'Andorra', 0.1, 320, 'spa,wifi,restaurante,gimnasio,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=Sport+Hotel+Hermitage+Soldeu', 'https://www.hotelhermitage.sporthotels.ad/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Park Piolets MountainHotel & Spa', 'park-piolets-mountain-hotel',
  'Hotel 5★ en Soldeu al pie de las pistas; habitaciones con balcón y wellness club. Precio orientativo; foto ilustrativa de archivo.',
  'Soldeu', 'Andorra', 0.2, 180, 'spa,piscina,wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Park+Piolets+Soldeu', 'https://www.parkpiolets.com/',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Grau Roig Boutique Hotel & Spa', 'grau-roig-boutique-hotel',
  'Hotel boutique aislado a pie de pistas en el circo de Grau Roig, a 2.100 m. Precio orientativo; foto ilustrativa de archivo.',
  'Grau Roig', 'Andorra', 0.1, 200, 'spa,wifi,restaurante,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=Grau+Roig+Boutique+Hotel+Andorra', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Grandvalira'
ON CONFLICT (slug) DO NOTHING;

-- ─── Pal Arinsal ────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Princesa Parc', 'hotel-princesa-parc',
  'Hotel 4★ en Arinsal con spa, gimnasio y piscina exterior de temporada, cerca del telecabina. Precio orientativo; foto ilustrativa de archivo.',
  'Arinsal', 'Andorra', 0.3, 110, 'spa,gimnasio,piscina,wifi,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Princesa+Parc+Arinsal', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Xalet Verdú', 'hotel-xalet-verdu',
  'Hotel 3★ de estilo chalet en Arinsal, a pocos minutos de los remontes de Pal Arinsal. Precio orientativo; foto ilustrativa de archivo.',
  'Arinsal', 'Andorra', 0.5, 75, 'wifi,bar,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Xalet+Verdu+Arinsal', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Pal Arinsal'
ON CONFLICT (slug) DO NOTHING;

-- ─── Ordino Arcalís ─────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Coma', 'hotel-coma-ordino',
  'Hotel clásico del pueblo de Ordino, a unos minutos en coche de Ordino Arcalís; organiza actividades de montaña. Precio orientativo; foto ilustrativa de archivo.',
  'Ordino', 'Andorra', 10.0, 95, 'wifi,piscina,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Coma+Ordino', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Ordino', 'hotel-ordino',
  'Hotel en el centro de Ordino, buena base para esquiar en Arcalís y recorrer la parroquia. Precio orientativo; foto ilustrativa de archivo.',
  'Ordino', 'Andorra', 10.0, 80, 'wifi,restaurante',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Ordino+Andorra', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Ordino%'
ON CONFLICT (slug) DO NOTHING;

-- ─── Naturlandia ────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Coma Bella', 'hotel-coma-bella',
  'Hotel rodeado del bosque de La Rabassa, a unos 3 km de Naturlandia, con piscina climatizada y bodega propia. Precio orientativo; foto ilustrativa de archivo.',
  'Sant Julià de Lòria', 'Andorra', 3.0, 85, 'piscina,wifi,restaurante,gimnasio,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Coma+Bella+Sant+Julia', 'https://www.hotelcomabella.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Imperial Atiram', 'hotel-imperial-atiram',
  'Hotel en Sant Julià de Lòria, a un corto trayecto de Naturlandia y de Andorra la Vella. Precio orientativo; foto ilustrativa de archivo.',
  'Sant Julià de Lòria', 'Andorra', 7.0, 90, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Imperial+Atiram+Sant+Julia+de+Loria', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Naturlandia'
ON CONFLICT (slug) DO NOTHING;

-- ─── Sierra Nevada ──────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Meliá Sol y Nieve', 'melia-sol-y-nieve',
  'Hotel 4★ en la plaza de Pradollano, a unos 100 m de los remontes, con spa y kids club. Precio orientativo; foto ilustrativa de archivo.',
  'Sierra Nevada (Monachil)', 'España', 0.1, 150, 'spa,piscina,wifi,restaurante,kids club,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=Melia+Sol+y+Nieve+Sierra+Nevada', 'https://www.melia.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Vincci Selección Rumaykiyya', 'vincci-seleccion-rumaykiyya',
  'Hotel 5★ de los más altos de Europa, con acceso directo a pistas, telesilla a la puerta y spa. Precio orientativo; foto ilustrativa de archivo.',
  'Sierra Nevada (Monachil)', 'España', 0.1, 200, 'spa,wifi,restaurante,guardaesquís,traslados',
  'https://www.booking.com/searchresults.es.html?ss=Vincci+Rumaykiyya+Sierra+Nevada', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('El Lodge Ski & Spa', 'el-lodge-ski-spa',
  'Exclusivo hotel boutique de madera a pie de pista en Sierra Nevada, con piscina exterior climatizada. Precio orientativo; foto ilustrativa de archivo.',
  'Sierra Nevada (Monachil)', 'España', 0.1, 350, 'spa,piscina,wifi,restaurante,guardaesquís',
  'https://www.booking.com/searchresults.es.html?ss=El+Lodge+Sierra+Nevada', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Sierra Nevada'
ON CONFLICT (slug) DO NOTHING;

-- ─── Valdesquí ──────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Pasadoiro', 'hotel-pasadoiro',
  'Hotel de montaña en el Puerto de Navacerrada (1.860 m), el alojamiento clásico más cercano a Valdesquí. Precio orientativo; foto ilustrativa de archivo.',
  'Puerto de Navacerrada', 'España', 7.0, 80, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Pasadoiro+Navacerrada', 'https://www.pasadoiro.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Box Art Hotel — Alpino', 'box-art-hotel-alpino',
  'Hotel boutique en el pueblo de Navacerrada con desayunos muy valorados, a unos 12 km de Valdesquí. Precio orientativo; foto ilustrativa de archivo.',
  'Navacerrada', 'España', 12.0, 110, 'wifi,restaurante,jardín,parking',
  'https://www.booking.com/searchresults.es.html?ss=Box+Art+Hotel+Alpino+Navacerrada', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Arcipreste de Hita', 'hotel-arcipreste-de-hita',
  'Hotel 4★ (solo adultos) en Navacerrada con spa, piscina climatizada y vistas al embalse. Precio orientativo; foto ilustrativa de archivo.',
  'Navacerrada', 'España', 12.0, 120, 'spa,piscina,wifi,restaurante,solo adultos',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Arcipreste+de+Hita+Navacerrada', 'https://www.hotelhita.com/',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Valdesqu%'
ON CONFLICT (slug) DO NOTHING;

-- ─── Puerto de Navacerrada ──────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hotel Pasadoiro (Puerto)', 'hotel-pasadoiro-puerto',
  'Hotel de montaña en el propio Puerto de Navacerrada, a pie de las pistas de la estación. Precio orientativo; foto ilustrativa de archivo.',
  'Puerto de Navacerrada', 'España', 0.3, 80, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Pasadoiro+Navacerrada', 'https://www.pasadoiro.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Arcipreste de Hita (Navacerrada)', 'hotel-arcipreste-de-hita-puerto',
  'Hotel 4★ (solo adultos) con spa en el pueblo de Navacerrada, a unos 7 km del puerto. Precio orientativo; foto ilustrativa de archivo.',
  'Navacerrada', 'España', 7.0, 120, 'spa,piscina,wifi,restaurante,solo adultos',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Arcipreste+de+Hita+Navacerrada', 'https://www.hotelhita.com/',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Box Art Hotel — Alpino (Navacerrada)', 'box-art-hotel-alpino-puerto',
  'Hotel boutique en Navacerrada pueblo, a unos 7 km de las pistas del puerto. Precio orientativo; foto ilustrativa de archivo.',
  'Navacerrada', 'España', 7.0, 110, 'wifi,restaurante,jardín,parking',
  'https://www.booking.com/searchresults.es.html?ss=Box+Art+Hotel+Alpino+Navacerrada', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Puerto de Navacerrada'
ON CONFLICT (slug) DO NOTHING;

-- ─── La Pinilla ─────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hostal La Pinilla', 'hostal-la-pinilla',
  'Hostal a pie de pistas de La Pinilla, a un paso de los remontes, con cafetería. Precio orientativo; foto ilustrativa de archivo.',
  'Cerezo de Arriba', 'España', 0.1, 55, 'wifi,cafetería',
  'https://www.booking.com/searchresults.es.html?ss=Hostal+La+Pinilla+Cerezo+de+Arriba', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Rural La Encantada', 'hotel-rural-la-encantada',
  'Hotel rural cerca de Riaza con terraza, restaurante y punto de venta de forfaits de La Pinilla. Precio orientativo; foto ilustrativa de archivo.',
  'Riaza', 'España', 8.0, 90, 'wifi,restaurante,terraza,parking',
  'https://www.booking.com/searchresults.es.html?ss=La+Encantada+Riaza', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Artesa Suites & Spa', 'artesa-suites-spa',
  'Suites con spa en la zona de Riaza, a un corto trayecto en coche de La Pinilla. Precio orientativo; foto ilustrativa de archivo.',
  'Riaza', 'España', 8.0, 100, 'spa,wifi,parking',
  'https://www.booking.com/searchresults.es.html?ss=Artesa+Suites+Spa+Riaza', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'La Pinilla'
ON CONFLICT (slug) DO NOTHING;

-- ─── Alto Campoo ────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('SNÖ Hotel La Corza Blanca', 'sno-la-corza-blanca',
  'Hotel 3★ en Brañavieja, en el corazón de la estación de Alto Campoo, con traslado gratuito a pistas. Precio orientativo; foto ilustrativa de archivo.',
  'Brañavieja', 'España', 0.5, 85, 'wifi,restaurante,bar,parking',
  'https://www.booking.com/searchresults.es.html?ss=SNO+La+Corza+Blanca+Branavieja', 'https://www.snoaltocampoo.com/',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Vejo', 'hotel-vejo',
  'Hotel clásico de Reinosa, la localidad de servicios más cercana a Alto Campoo (a unos 25 km). Precio orientativo; foto ilustrativa de archivo.',
  'Reinosa', 'España', 25.0, 60, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Vejo+Reinosa', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Alto Campoo'
ON CONFLICT (slug) DO NOTHING;

-- ─── San Isidro ─────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Hostal Pico Agujas', 'hostal-pico-agujas',
  'Hostal de montaña en la propia estación de San Isidro, el alojamiento más cercano a pistas. Precio orientativo; foto ilustrativa de archivo.',
  'San Isidro (Puebla de Lillo)', 'España', 0.5, 60, 'wifi,restaurante',
  'https://www.booking.com/searchresults.es.html?ss=Hostal+Pico+Agujas+San+Isidro', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Peña Pandos', 'hotel-pena-pandos',
  'Hotel en Felechosa (vertiente asturiana), base habitual para San Isidro y Fuentes de Invierno. Precio orientativo; foto ilustrativa de archivo.',
  'Felechosa (Aller)', 'España', 9.0, 70, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Hotel+Pena+Pandos+Felechosa', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop'),
 ('Hotel Parador de Felechosa', 'hotel-parador-de-felechosa',
  'Hotel-restaurante familiar en Felechosa, a unos minutos en coche de los accesos de San Isidro. Precio orientativo; foto ilustrativa de archivo.',
  'Felechosa (Aller)', 'España', 9.0, 65, 'wifi,restaurante,parking',
  'https://www.booking.com/searchresults.es.html?ss=Parador+de+Felechosa', '',
  'https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'San Isidro'
ON CONFLICT (slug) DO NOTHING;

-- ─── Valgrande-Pajares ──────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Posada Real Pajares', 'posada-real-pajares',
  'Posada con encanto en el pueblo de Pajares, a un corto trayecto de la estación de Valgrande-Pajares. Precio orientativo; foto ilustrativa de archivo.',
  'Pajares (Lena)', 'España', 4.0, 70, 'wifi,restaurante',
  'https://www.booking.com/searchresults.es.html?ss=Posada+Real+Pajares', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('Moradas Busdongo', 'moradas-busdongo',
  'Apartamentos en Busdongo de Arbas, en la vertiente leonesa del puerto de Pajares. Precio orientativo; foto ilustrativa de archivo.',
  'Busdongo de Arbas', 'España', 6.0, 65, 'wifi,cocina,parking',
  'https://www.booking.com/searchresults.es.html?ss=Moradas+Busdongo', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Valgrande%'
ON CONFLICT (slug) DO NOTHING;

-- ─── Leitariegos ────────────────────────────────────────────────────────────
INSERT INTO hotels (station_id, name, slug, description, address, city, country,
    distance_to_station_km, price_from, currency, rating, review_count,
    amenities, booking_url, official_url, main_image_url)
SELECT s.id, v.nm, v.sl, v.ds, '', v.ct, v.pa, v.di, v.pr, 'EUR', 0, 0, v.am, v.bk, v.of, v.im
FROM stations s CROSS JOIN (VALUES
 ('Apartahotel Portal de León', 'apartahotel-portal-de-leon',
  'Apartahotel con restaurante en Caboalles de Abajo (Laciana), a unos km del puerto de Leitariegos. Precio orientativo; foto ilustrativa de archivo.',
  'Caboalles de Abajo (Villablino)', 'España', 9.0, 65, 'wifi,restaurante,cocina',
  'https://www.booking.com/searchresults.es.html?ss=Apartahotel+Portal+de+Leon+Caboalles', '',
  'https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop'),
 ('El Cordal de Laciana', 'el-cordal-de-laciana',
  'Alojamiento rural con bar y jardín en Caboalles de Abajo, base para esquiar en Leitariegos. Precio orientativo; foto ilustrativa de archivo.',
  'Caboalles de Abajo (Villablino)', 'España', 9.0, 60, 'wifi,bar,jardín',
  'https://www.booking.com/searchresults.es.html?ss=El+Cordal+de+Laciana+Caboalles', '',
  'https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1200&auto=format&fit=crop')
) AS v(nm, sl, ds, ct, pa, di, pr, am, bk, of, im)
WHERE s.name ILIKE 'Leitariegos'
ON CONFLICT (slug) DO NOTHING;

-- ─── Galería genérica para los hoteles sin imágenes ─────────────────────────
-- Tres fotos ilustrativas de archivo (Unsplash) por hotel. NO son fotos del
-- alojamiento; las fotos reales están en el enlace externo de reserva.
INSERT INTO hotel_images (hotel_id, image_url, alt_text, source_name, source_url, sort_order, is_primary)
SELECT h.id, v.url, v.alt || ' (imagen ilustrativa, no corresponde al hotel)', 'Unsplash (foto genérica)', 'https://unsplash.com', v.ord, v.ord = 0
FROM hotels h
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1600&auto=format&fit=crop', 'Paisaje de montaña nevada', 0),
    ('https://images.unsplash.com/photo-1518602164578-cd0074062767?q=80&w=1600&auto=format&fit=crop', 'Entorno invernal', 1),
    ('https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=1600&auto=format&fit=crop',   'Interior de alojamiento de montaña', 2)
) AS v(url, alt, ord)
WHERE NOT EXISTS (SELECT 1 FROM hotel_images hi WHERE hi.hotel_id = h.id);
