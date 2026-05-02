# Imágenes de estaciones — pendiente de conseguir

Las migraciones 007 y 009 usaban URLs de Unsplash que no podíamos garantizar
que fueran fotos coherentes (en algunos casos servían fotos absurdas: arte,
retratos, ciudades sin nieve…). La migración **010_clear_station_images.sql**
ha vaciado el campo `image_url` de todas las estaciones.

Mientras no tengamos fotos verificadas, la web muestra automáticamente el
**cover de marca**: gradiente azul/blanco + silueta de montañas + nombre de
la estación, con un tono de color distinto para cada zona geográfica. Es
limpio, coherente y nunca enseña una imagen rota o descontextualizada.

Cuando consigas fotos reales (oficiales, turísticas o de Wikimedia Commons),
crea una nueva migración (`db/migrations/011_set_real_station_images.sql`)
con un `UPDATE stations SET image_url = '...' WHERE name = '...';` por cada
estación.

## Criterios obligatorios para cada imagen

- Debe mostrar nieve, pistas, montaña, remontes, esquiadores o paisaje de
  estación de esquí.
- Debe estar relacionada con la estación concreta cuando sea posible.
- Calidad suficiente para una card web (>= 1200px de ancho).
- Orientación horizontal (16:10 o 16:9).
- URL estable y pública (oficial, turística o Wikimedia Commons).
- `alt` recomendado: `Estación de esquí <Nombre>`.
- **Cada estación debe tener una URL distinta**. Cero duplicados.

No vale: personas famosas, cuadros/arte/museos/esculturas, ciudades sin
nieve, logos como imagen principal, fotos de baja calidad, URLs rotas,
imágenes repetidas o genéricas iguales para varias estaciones.

## Listado completo (27 estaciones)

Para cada una hay una sugerencia de búsqueda y la fuente preferente. Una
buena estrategia: ir a Wikipedia ES de cada estación y reutilizar la imagen
principal (suelen estar en Commons con licencia compatible).

### Pirineo Catalán (10)

- [ ] **Baqueira Beret** — *Lleida*
  - Buscar: `Baqueira Beret pista` o `Baqueira Beret valle de Arán nieve`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Estaci%C3%B3n_de_esqu%C3%AD_Baqueira_Beret>
  - Foto oficial: <https://www.baqueira.es>

- [ ] **La Molina** — *Girona*
  - Buscar: `La Molina pistas esquí` o `La Molina Cerdaña nieve`.
  - Wikipedia: <https://es.wikipedia.org/wiki/La_Molina_(estaci%C3%B3n_de_esqu%C3%AD)>
  - Oficial: <https://www.lamolina.cat>

- [ ] **Masella** — *Girona*
  - Buscar: `Masella pistas` o `Masella Cerdaña esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Masella>
  - Oficial: <https://www.masella.com>

- [ ] **Vall de Núria** — *Girona*
  - Buscar: `Vall de Núria invierno` o `santuario Núria nieve`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Vall_de_N%C3%BAria>
  - Oficial: <https://www.valldenuria.cat>

- [ ] **Vallter 2000** — *Girona*
  - Buscar: `Vallter 2000 pistas` o `circo Ulldeter esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Vallter_2000>
  - Oficial: <https://vallter.cat>

- [ ] **Espot Esquí** — *Lleida*
  - Buscar: `Espot Esquí pistas` o `Aigüestortes invierno`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Espot_Esqu%C3%AD>
  - Oficial: <https://www.espotesqui.cat>

- [ ] **Port Ainé** — *Lleida*
  - Buscar: `Port Ainé esquí` o `Port Ainé Pallars Sobirà`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Port_Ain%C3%A9>
  - Oficial: <https://www.portaine.cat>

- [ ] **Boí Taüll** — *Lleida*
  - Buscar: `Boí Taüll pistas` o `Vall de Boí esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Bo%C3%AD_Ta%C3%BCll>
  - Oficial: <https://www.boitaullresort.com>

- [ ] **Tavascan** — *Lleida*
  - Buscar: `Tavascan estación esquí` o `Pallars Sobirà invierno`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Tavascan>
  - Oficial: <https://www.tavascan.net>

- [ ] **Port del Comte** — *Lleida*
  - Buscar: `Port del Comte pistas` o `Solsonès esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Port_del_Comte>
  - Oficial: <https://www.portdelcomte.net>

### Pirineo Aragonés (5)

- [ ] **Candanchú** — *Huesca*
  - Buscar: `Candanchú pistas` o `valle del Aragón esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Candanch%C3%BA>
  - Oficial: <https://www.candanchu.com>

- [ ] **Astún** — *Huesca*
  - Buscar: `Astún pistas` o `Astún anfiteatro esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Ast%C3%BAn>
  - Oficial: <https://www.astun.com>

- [ ] **Formigal** — *Huesca*
  - Buscar: `Formigal pistas` o `Tena Formigal esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Formigal>
  - Oficial: <https://www.formigal-panticosa.com>

- [ ] **Panticosa** — *Huesca*
  - Buscar: `Panticosa estación esquí` o `Panticosa balneario nieve`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Estaci%C3%B3n_de_esqu%C3%AD_de_Panticosa>
  - Oficial: <https://www.formigal-panticosa.com>

- [ ] **Cerler** — *Huesca*
  - Buscar: `Cerler pistas` o `valle de Benasque esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Cerler>
  - Oficial: <https://www.cerler.com>

### Andorra (4)

- [ ] **Grandvalira** — *Andorra*
  - Buscar: `Grandvalira pistas` o `Pas de la Casa Soldeu esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Grandvalira>
  - Oficial: <https://www.grandvalira.com>

- [ ] **Pal Arinsal** — *Andorra*
  - Buscar: `Pal Arinsal pistas` o `Vallnord Pal Arinsal`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Pal_Arinsal>
  - Oficial: <https://www.palarinsal.com>

- [ ] **Ordino Arcalís** — *Andorra*
  - Buscar: `Ordino Arcalís pistas` o `Vallnord Arcalís freeride`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Ordino_Arcal%C3%ADs>
  - Oficial: <https://www.ordinoarcalis.com>

- [ ] **Naturlandia** — *Andorra*
  - Buscar: `Naturlandia La Rabassa nieve` (es parque, no resort grande).
  - Wikipedia: <https://es.wikipedia.org/wiki/Naturlandia>
  - Oficial: <https://www.naturlandia.ad>

### Sistema Central / Penibético (4)

- [ ] **Sierra Nevada** — *Granada*
  - Buscar: `Sierra Nevada pistas` o `Pradollano Sierra Nevada esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Estaci%C3%B3n_de_esqu%C3%AD_Sierra_Nevada>
  - Oficial: <https://sierranevada.es>

- [ ] **Valdesquí** — *Madrid*
  - Buscar: `Valdesquí pistas` o `Valdesquí Sierra Guadarrama nieve`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Valdesqu%C3%AD>
  - Oficial: <https://www.valdesqui.es>

- [ ] **Puerto de Navacerrada** — *Madrid*
  - Buscar: `Puerto de Navacerrada nieve` o `Navacerrada estación esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Puerto_de_Navacerrada_(estaci%C3%B3n_de_esqu%C3%AD)>

- [ ] **La Pinilla** — *Segovia*
  - Buscar: `La Pinilla pistas` o `Sierra de Ayllón esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/La_Pinilla>
  - Oficial: <https://www.lapinilla.es>

### Cordillera Cantábrica (4)

- [ ] **Alto Campoo** — *Cantabria*
  - Buscar: `Alto Campoo pistas` o `Pico Tres Mares esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Alto_Campoo>
  - Oficial: <https://www.altocampoo.com>

- [ ] **San Isidro** — *León*
  - Buscar: `San Isidro estación esquí` o `San Isidro León nieve`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Estaci%C3%B3n_invernal_y_de_monta%C3%B1a_San_Isidro>
  - Oficial: <https://www.dipuleon.es/sanisidro>

- [ ] **Valgrande-Pajares** — *Asturias*
  - Buscar: `Valgrande Pajares pistas` o `Pajares estación esquí`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Valgrande-Pajares>
  - Oficial: <https://www.valgrande-pajares.com>

- [ ] **Leitariegos** — *Asturias*
  - Buscar: `Leitariegos pistas` o `Leitariegos puerto nieve`.
  - Wikipedia: <https://es.wikipedia.org/wiki/Leitariegos_(estaci%C3%B3n_de_esqu%C3%AD)>
  - Oficial: <https://www.leitariegos.net>

## Cómo aplicar las imágenes una vez conseguidas

1. Crea `db/migrations/011_set_real_station_images.sql`.
2. Por cada estación añade una línea:

   ```sql
   UPDATE stations SET image_url = 'https://…' WHERE name = 'Baqueira Beret';
   ```

3. Asegúrate de que **ninguna URL se repite** entre estaciones.
4. Arranca el servidor; el runner aplicará la migración una vez.
5. Limpia caché del navegador (Ctrl+F5) para ver las imágenes nuevas.

Si una URL futura llega a romperse, no pasa nada: el `<img>` tiene
`onerror="…"` que la oculta y deja visible el cover de marca.
