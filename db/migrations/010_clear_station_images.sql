-- Migración 010: limpiar las URLs de imagen de estaciones.
-- ----------------------------------------------------------------------------
-- Las migraciones 007 y 009 asignaban URLs de Unsplash que no podemos
-- garantizar al 100% que sean fotos coherentes (esquí/nieve/montaña). En
-- algunos casos Unsplash sirve fotos no relacionadas (arte, ciudades,
-- retratos…). Para evitar que el usuario vea fotos absurdas, vaciamos
-- el campo `image_url` y dejamos que la web muestre el "cover" de marca
-- por defecto: gradiente azul + silueta de montañas + nombre de la estación.
--
-- Cuando se consigan fotos reales y verificadas (ver
-- station-images-needed.md en la raíz del proyecto), añade una nueva
-- migración (011_set_real_station_images.sql) con UPDATE por estación.

UPDATE stations SET image_url = '' WHERE name IN (
  'Baqueira Beret','La Molina','Masella','Vall de Núria','Vallter 2000',
  'Espot Esquí','Port Ainé','Boí Taüll','Tavascan','Port del Comte',
  'Candanchú','Astún','Formigal','Panticosa','Cerler',
  'Grandvalira','Pal Arinsal','Ordino Arcalís','Naturlandia',
  'Sierra Nevada','Valdesquí','Puerto de Navacerrada','La Pinilla',
  'Alto Campoo','San Isidro','Valgrande-Pajares','Leitariegos'
);
