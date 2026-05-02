-- Migración 011: imágenes reales y verificadas de cada estación.
-- ----------------------------------------------------------------------------
-- Cada URL apunta a Wikimedia Commons (extraídas via la API de Wikipedia ES,
-- summary REST endpoint, campo originalimage.source). Son fotos oficiales
-- de la propia estación: pistas, valle, esquiadores, remontes, etc.
--
-- Las URLs de upload.wikimedia.org son extremadamente estables (años o
-- décadas) y respetan el dominio público/CC de cada imagen. Idempotente:
-- volver a aplicar la migración no cambia nada.
--
-- 24 de 27 estaciones tienen foto. Las 3 sin foto verificable (Tavascán,
-- Naturlandia y Leitariegos) se dejan en blanco intencionadamente para que
-- la web muestre el "cover de marca" (gradiente + silueta + nombre).

-- ───────── Pirineo Catalán ─────────
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/7/7a/Baqueira_1800.jpg' WHERE name='Baqueira Beret';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/f/f6/Molina_La2.JPG' WHERE name='La Molina';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/8/84/Coma_Oriola.JPG' WHERE name='Masella';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/7/7c/Estaci%C3%B3_d%27esqu%C3%AD_de_Vall_de_N%C3%BAria.JPG' WHERE name='Vall de Núria';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/4/41/Vallter.jpg' WHERE name='Vallter 2000';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/d/d8/Espot_Esqu%C3%AD_des_del_coll_de_Fogueruix.jpg' WHERE name='Espot Esquí';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/d/d2/Port_ain%C3%A9.JPG' WHERE name='Port Ainé';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/d/df/Ski_resort_Bo%C3%AD-Ta%C3%BCll.jpg' WHERE name='Boí Taüll';
-- Tavascán: sin foto de la estación en Wikipedia (solo iglesia del pueblo); usa cover de marca.
UPDATE stations SET image_url='' WHERE name='Tavascan';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/b/b6/Port_del_Comte-Estivella.JPG' WHERE name='Port del Comte';

-- ───────── Pirineo Aragonés ─────────
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/4/45/Candanchu.jpg' WHERE name='Candanchú';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/8/89/Zaloa_Etxaniz_-_31654177534.jpg' WHERE name='Astún';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/e/ec/Formigal_2007.jpg' WHERE name='Formigal';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/2/26/Views_of_Panticosa.jpg' WHERE name='Panticosa';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/9/96/Cerler%2C_telesilla.jpg' WHERE name='Cerler';

-- ───────── Andorra ─────────
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/b/bf/Grandvalira_ski_resort%2C_Andorra4.jpg' WHERE name='Grandvalira';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/8/88/Casamanya_03.jpg' WHERE name='Pal Arinsal';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/9/99/View_down_red_run_to_Arcalis_Ski_Center_Andorra_Mar_2011.jpg' WHERE name='Ordino Arcalís';
-- Naturlandia: artículo sin foto en Wikipedia ES; usa cover de marca.
UPDATE stations SET image_url='' WHERE name='Naturlandia';

-- ───────── Sistema Penibético / Central ─────────
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/1/12/Pradollano08b.jpg' WHERE name='Sierra Nevada';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/7/71/P1271366.JPG' WHERE name='Valdesquí';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/a/a5/Puerto_de_Navacerrada_14-10-2006.jpg' WHERE name='Puerto de Navacerrada';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/c/ce/La-pinilla-2-100110.jpg' WHERE name='La Pinilla';

-- ───────── Cordillera Cantábrica ─────────
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/d/da/Remonte_de_los_Asnos.jpg' WHERE name='Alto Campoo';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/2/27/Estaci%C3%B3n_de_esqu%C3%AD_de_San_Isidro.JPG' WHERE name='San Isidro';
UPDATE stations SET image_url='https://upload.wikimedia.org/wikipedia/commons/f/f3/Estaci%C3%B3n_de_Valgrande-Pajares.JPG' WHERE name='Valgrande-Pajares';
-- Leitariegos: solo hay un mapa en relieve en Wikipedia, no foto real; usa cover de marca.
UPDATE stations SET image_url='' WHERE name='Leitariegos';
