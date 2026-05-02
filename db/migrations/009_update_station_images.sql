-- Migración 009: imágenes únicas y representativas por estación.
-- ----------------------------------------------------------------------------
-- Hasta ahora muchas estaciones compartían la misma URL de Unsplash. Aquí
-- asignamos una imagen distinta a cada una de las 27 estaciones del catálogo
-- usando IDs estables de Unsplash con orientación landscape (fotos de
-- montaña, esquí o nieve). Los UPDATE son idempotentes: si vuelves a
-- aplicar la migración no rompe nada porque sólo actualiza el campo
-- `image_url` por nombre.
--
-- Si añades nuevas estaciones, añade aquí su UPDATE correspondiente para
-- garantizar la unicidad de la imagen.

-- ───────── Pirineo Catalán ─────────
UPDATE stations SET image_url='https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop' WHERE name='Baqueira Beret';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1551524559-8af4e6624178?q=80&w=1200&auto=format&fit=crop' WHERE name='La Molina';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1551698618-1dfe5d97d256?q=80&w=1200&auto=format&fit=crop' WHERE name='Masella';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1486684338211-1a7ced564b0d?q=80&w=1200&auto=format&fit=crop' WHERE name='Vall de Núria';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1528048228650-0d2f4deedc67?q=80&w=1200&auto=format&fit=crop' WHERE name='Vallter 2000';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1483728642387-6c3bdd6c93e5?q=80&w=1200&auto=format&fit=crop' WHERE name='Espot Esquí';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1517825738774-7de9363ef735?q=80&w=1200&auto=format&fit=crop' WHERE name='Port Ainé';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1542038784456-1ea8e935640e?q=80&w=1200&auto=format&fit=crop' WHERE name='Boí Taüll';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1502104034360-73176bb1e92e?q=80&w=1200&auto=format&fit=crop' WHERE name='Tavascan';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1457269449834-928af64c684d?q=80&w=1200&auto=format&fit=crop' WHERE name='Port del Comte';

-- ───────── Pirineo Aragonés ─────────
UPDATE stations SET image_url='https://images.unsplash.com/photo-1565992441121-4367c2967103?q=80&w=1200&auto=format&fit=crop' WHERE name='Candanchú';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1542338347-4fff3276af78?q=80&w=1200&auto=format&fit=crop' WHERE name='Astún';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1612871689353-cccf581d667b?q=80&w=1200&auto=format&fit=crop' WHERE name='Formigal';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1605555088956-2c9bd05e7e2c?q=80&w=1200&auto=format&fit=crop' WHERE name='Panticosa';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1551649001-7a2482d98d05?q=80&w=1200&auto=format&fit=crop' WHERE name='Cerler';

-- ───────── Andorra ─────────
UPDATE stations SET image_url='https://images.unsplash.com/photo-1602601311014-c43b40e51e34?q=80&w=1200&auto=format&fit=crop' WHERE name='Grandvalira';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1543200260-7b4eb1d6b6c8?q=80&w=1200&auto=format&fit=crop' WHERE name='Pal Arinsal';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1467998526012-8d09226aaba1?q=80&w=1200&auto=format&fit=crop' WHERE name='Ordino Arcalís';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1614109075946-321d54d6deb1?q=80&w=1200&auto=format&fit=crop' WHERE name='Naturlandia';

-- ───────── Sistema Penibético / Central / Cantábrico ─────────
UPDATE stations SET image_url='https://images.unsplash.com/photo-1601887389937-0b02c26b602c?q=80&w=1200&auto=format&fit=crop' WHERE name='Sierra Nevada';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1480497490787-505ba2495a44?q=80&w=1200&auto=format&fit=crop' WHERE name='Valdesquí';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1612869538502-fdb1eaae9b6c?q=80&w=1200&auto=format&fit=crop' WHERE name='Puerto de Navacerrada';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1610648580013-b40e5b9d8ed8?q=80&w=1200&auto=format&fit=crop' WHERE name='La Pinilla';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1551524559-8af4e6624178?q=80&w=1200&auto=format&fit=crop&sat=-30' WHERE name='Alto Campoo';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1473773508845-188df298d2d1?q=80&w=1200&auto=format&fit=crop' WHERE name='San Isidro';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1522163182402-834f871fd851?q=80&w=1200&auto=format&fit=crop' WHERE name='Valgrande-Pajares';
UPDATE stations SET image_url='https://images.unsplash.com/photo-1551698618-1dfe5d97d256?q=80&w=1200&auto=format&fit=crop&sat=-20' WHERE name='Leitariegos';
