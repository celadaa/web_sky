/* Mapa de estaciones de Snowbreak.
 *
 * Mejoras respecto a la versión anterior:
 *   - Marcadores con icono de esquiador (DivIcon SVG inline) en lugar
 *     del pin azul por defecto de Leaflet.
 *   - Color del marcador según el estado operativo de la estación:
 *       verde     → excelente (≥ 80 % pistas abiertas)
 *       amarillo  → buena     (≥ 50 %)
 *       naranja   → regular   (< 50 %)
 *       rojo      → cerrada   (0 pistas abiertas)
 *   - Popup con tarjeta más completa: imagen, badge de estado, datos
 *     meteo y CTA grande "Ver detalle".
 *   - Marcador opcional "Mi ubicación" si el usuario aceptó la
 *     geolocalización (lat/lng + reverse geocoding con Nominatim).
 *   - Tarjeta de ubicación precisa por encima del mapa (calle, ciudad,
 *     país y coordenadas con 5 decimales).
 *
 * El módulo sigue siendo IIFE clásico, sin dependencias de bundler.
 */
(function () {
  'use strict';

  var LEAFLET_CSS = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css';
  var LEAFLET_JS  = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.js';

  // Tile layer "dark" (Carto) para que combine con la estética azul oscuro
  // del proyecto. Se mantiene OSM como respaldo si Carto cae.
  var TILE_DARK = 'https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png';
  var TILE_OSM  = 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png';
  var ATRIB_DARK = '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/attributions">CARTO</a>';

  var mapa = null;
  var capaEstaciones = null;
  var capaUsuario = null;

  // Estado del filtrado de la barra lateral.
  // listaActual contiene la última respuesta del API y se reutiliza
  // sin volver a hacer fetch al cambiar de filtro o búsqueda.
  var listaActual = [];
  var markersPorId = {};
  var filtroMacizo = '';   // '' = todos
  var filtroTexto  = '';

  function iniciar() {
    var btnLista = document.querySelector('[data-vista="lista"]');
    var btnMapa  = document.querySelector('[data-vista="mapa"]');
    if (!btnLista || !btnMapa) { return; }

    btnLista.addEventListener('click', mostrarLista);
    btnMapa.addEventListener('click', mostrarMapa);
  }

  function mostrarLista() { alternarVista(false); }

  function mostrarMapa() {
    alternarVista(true);
    cargarLeaflet().then(pintarMapa).catch(function (err) {
      var c = document.getElementById('mapa-container');
      if (c) { c.textContent = 'No se pudo cargar el mapa: ' + err.message; }
    });
  }

  function alternarVista(verMapa) {
    var grid = document.getElementById('grid-estaciones');
    var nav  = document.getElementById('filtros-estaciones');
    var cont = document.getElementById('mapa-container');
    if (grid) { grid.classList.toggle('oculto', verMapa); }
    if (nav)  { nav.classList.toggle('oculto',  verMapa); }
    if (cont) { cont.classList.toggle('oculto', !verMapa); }

    Array.prototype.forEach.call(document.querySelectorAll('.filters__view-btn'), function (btn) {
      var activa = btn.getAttribute('data-vista') === (verMapa ? 'mapa' : 'lista');
      btn.classList.toggle('filters__view-btn--active', activa);
      btn.setAttribute('aria-pressed', activa ? 'true' : 'false');
    });
  }

  function cargarLeaflet() {
    if (window.L) { return Promise.resolve(); }
    return new Promise(function (resolve, reject) {
      if (!document.querySelector('link[data-leaflet]')) {
        var link = document.createElement('link');
        link.rel = 'stylesheet';
        link.href = LEAFLET_CSS;
        link.setAttribute('data-leaflet', '1');
        document.head.appendChild(link);
      }
      if (document.querySelector('script[data-leaflet]')) {
        document.querySelector('script[data-leaflet]').addEventListener('load', resolve, { once: true });
        return;
      }
      var script = document.createElement('script');
      script.src = LEAFLET_JS;
      script.setAttribute('data-leaflet', '1');
      script.onload  = function () { resolve(); };
      script.onerror = function () { reject(new Error('no se pudo cargar Leaflet')); };
      document.head.appendChild(script);
    });
  }

  function pintarMapa() {
    var div = document.getElementById('mapa-leaflet');
    if (!div) { return; }

    if (mapa === null) {
      mapa = L.map(div, {
        zoomControl: true,
        scrollWheelZoom: true,
        attributionControl: true
      }).setView([41.0, -2.0], 6);

      L.tileLayer(TILE_DARK, {
        maxZoom: 19, subdomains: 'abcd', attribution: ATRIB_DARK
      }).addTo(mapa).on('tileerror', function () {
        // Si Carto falla cambiamos a OSM clásico (un solo intento).
        L.tileLayer(TILE_OSM, { maxZoom: 19, attribution: ATRIB_DARK }).addTo(mapa);
      });

      capaEstaciones = L.layerGroup().addTo(mapa);
      capaUsuario    = L.layerGroup().addTo(mapa);
      anadirLeyenda(mapa);
    }

    fetch('/api/estaciones', { headers: { Accept: 'application/json' } })
      .then(function (resp) {
        if (!resp.ok) { throw new Error('HTTP ' + resp.status); }
        return resp.json();
      })
      .then(anadirMarkers)
      .catch(function (err) { console.error('Error API estaciones:', err); });

    intentarUbicarUsuario();
  }

  /* ─────────────────  MARCADORES DE ESTACIONES  ───────────────── */

  function anadirMarkers(lista) {
    listaActual  = lista || [];
    markersPorId = {};
    capaEstaciones.clearLayers();

    var puntos = [];
    listaActual.forEach(function (e) {
      if (!e.tiene_coords) { return; }
      var marker = L.marker([e.lat, e.lng], { icon: iconoEsquiador(e.estado) });
      marker.bindPopup(popupHTML(e), { maxWidth: 320, className: 'leaflet-popup-station' });
      marker.addTo(capaEstaciones);
      markersPorId[e.id] = marker;
      puntos.push([e.lat, e.lng]);
    });

    if (puntos.length > 0) {
      mapa.fitBounds(puntos, { padding: [60, 60] });
    }
    setTimeout(function () { mapa.invalidateSize(); }, 50);

    // Construye barra lateral, filtros y conexión con el buscador
    construirSidebar();
    conectarBuscadorYFiltros();
  }

  /** Devuelve un L.divIcon con un esquiador SVG y color de fondo según estado. */
  function iconoEsquiador(estado) {
    var clase = 'station-marker station-marker--' + (estado || 'regular');
    var svg = ''
      + '<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">'
      // Cabeza
      + '<circle cx="14.5" cy="4" r="1.8"/>'
      // Cuerpo + brazo + bastón
      + '<path d="M9 18.5l4.2-3.5 1.8 2.6 3.4-1.4-1-2.2-2.2.9-1.8-2.7 2.2-2.6c.6-.7.4-1.7-.4-2.2L12 5.6c-.5-.3-1.2-.2-1.6.3L7.6 9.6l1.5 1.2 2-2.5 1.5.9-2.6 3.1c-.4.4-.4 1.1 0 1.6l1.6 2.3-3.4 2.8z"/>'
      // Esquí
      + '<path d="M3 19.5l16-2 .3 1.8L3.3 21.3z"/>'
      + '</svg>';
    return L.divIcon({
      className: clase,
      html: '<span class="station-marker__pulse"></span><span class="station-marker__pin">' + svg + '</span>',
      iconSize:   [44, 52],
      iconAnchor: [22, 48],
      popupAnchor:[0, -42]
    });
  }

  /** Construye la tarjeta del popup (DOM real, no string concat) para una estación. */
  function popupHTML(e) {
    var box = document.createElement('div');
    box.className = 'popup-station popup-station--' + (e.estado || 'regular');

    if (e.imagen) {
      var fig = document.createElement('div');
      fig.className = 'popup-station__media';
      var img = document.createElement('img');
      img.src = e.imagen;
      img.alt = 'Vista de ' + e.nombre;
      img.loading = 'lazy';
      fig.appendChild(img);
      var badge = document.createElement('span');
      badge.className = 'popup-station__badge popup-station__badge--' + (e.estado || 'regular');
      badge.appendChild(puntoColor(e.estado));
      badge.appendChild(document.createTextNode(' ' + (e.estado_texto || 'Sin datos')));
      fig.appendChild(badge);
      box.appendChild(fig);
    }

    var body = document.createElement('div');
    body.className = 'popup-station__body';

    var h = document.createElement('h3');
    h.className = 'popup-station__title';
    h.textContent = e.nombre;
    body.appendChild(h);

    if (e.ubicacion) {
      var ub = document.createElement('p');
      ub.className = 'popup-station__location';
      ub.textContent = e.ubicacion + (e.altitud ? ' · ' + e.altitud : '');
      body.appendChild(ub);
    }

    var dl = document.createElement('dl');
    dl.className = 'popup-station__data';
    dl.appendChild(dato('Pistas', e.pistas_abiertas + ' / ' + e.pistas_totales));
    dl.appendChild(dato('Remontes', (e.remontes_op || 0) + ' / ' + (e.remontes_tot || 0)));
    dl.appendChild(dato('Nieve base', (e.nieve_base || 0) + ' cm'));
    dl.appendChild(dato('Temperatura', (typeof e.temperatura === 'number' ? e.temperatura : '–') + ' °C'));
    body.appendChild(dl);

    var cta = document.createElement('a');
    cta.href = '/estacion/' + e.id;
    cta.className = 'popup-station__cta';
    cta.textContent = 'Ver detalle de la estación';
    body.appendChild(cta);

    box.appendChild(body);
    return box;
  }

  function dato(label, valor) {
    var w = document.createElement('div');
    w.className = 'popup-station__data-item';
    var dt = document.createElement('dt'); dt.textContent = label;
    var dd = document.createElement('dd'); dd.textContent = valor;
    w.appendChild(dt); w.appendChild(dd);
    return w;
  }

  function puntoColor(estado) {
    var s = document.createElement('span');
    s.className = 'popup-station__dot popup-station__dot--' + (estado || 'regular');
    return s;
  }

  /* ─────────────────  LEYENDA  ───────────────── */

  function anadirLeyenda(m) {
    var Legend = L.Control.extend({
      options: { position: 'bottomleft' },
      onAdd: function () {
        var div = L.DomUtil.create('div', 'map-legend');
        div.innerHTML = ''
          + '<span class="map-legend__chip map-legend__chip--excelente"><span class="map-legend__dot"></span>Excelente</span>'
          + '<span class="map-legend__chip map-legend__chip--buena"><span class="map-legend__dot"></span>Buena</span>'
          + '<span class="map-legend__chip map-legend__chip--regular"><span class="map-legend__dot"></span>Regular</span>'
          + '<span class="map-legend__chip map-legend__chip--cerrado"><span class="map-legend__dot"></span>Cerrada</span>';
        return div;
      }
    });
    m.addControl(new Legend());
  }

  /* ─────────────────  UBICACIÓN DEL USUARIO  ───────────────── */

  function intentarUbicarUsuario() {
    if (!navigator.geolocation) { return; }

    // Si ya teníamos una ubicación reciente guardada, la usamos sin
    // pedir permiso otra vez (UX).
    var prev = window.SnowbreakGeo && window.SnowbreakGeo.leerPreferencia();
    if (prev && typeof prev.lat === 'number') {
      pintarMarcadorUsuario(prev.lat, prev.lng, prev.etiqueta || 'Tu ubicación');
      reverseGeocode(prev.lat, prev.lng);
    }

    navigator.geolocation.getCurrentPosition(
      function (pos) {
        var lat = pos.coords.latitude, lng = pos.coords.longitude;
        pintarMarcadorUsuario(lat, lng, 'Tu ubicación');
        reverseGeocode(lat, lng);
        if (window.SnowbreakGeo) {
          window.SnowbreakGeo.guardarPreferencia({
            lat: lat, lng: lng, fuente: 'gps', etiqueta: 'Tu ubicación'
          });
        }
      },
      function () { /* permiso denegado: no hacemos nada */ },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 5 * 60 * 1000 }
    );
  }

  function pintarMarcadorUsuario(lat, lng, etiqueta) {
    capaUsuario.clearLayers();
    var icon = L.divIcon({
      className: 'user-marker',
      html: '<span class="user-marker__pulse"></span><span class="user-marker__dot"></span>',
      iconSize: [22, 22], iconAnchor: [11, 11]
    });
    L.marker([lat, lng], { icon: icon, zIndexOffset: 1000 })
      .bindTooltip(etiqueta, { permanent: false, direction: 'top' })
      .addTo(capaUsuario);
  }

  /** Pide a Nominatim la dirección postal correspondiente y la pinta
   *  en la tarjeta #ubicacion-card si existe. Falla silenciosamente. */
  function reverseGeocode(lat, lng) {
    var card = document.getElementById('ubicacion-card');
    if (!card) { return; }
    actualizarCard(card, {
      titulo: 'Ubicándote…',
      direccion: '',
      coords: lat.toFixed(5) + ', ' + lng.toFixed(5)
    });
    var url = 'https://nominatim.openstreetmap.org/reverse?format=json&lat=' + lat
            + '&lon=' + lng + '&accept-language=es&zoom=18&addressdetails=1';
    fetch(url, { headers: { Accept: 'application/json' } })
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (data) {
        if (!data) { return; }
        var a = data.address || {};
        var calle = [a.road, a.house_number].filter(Boolean).join(' ');
        var ciudad = a.city || a.town || a.village || a.municipality || a.county || '';
        var region = a.state || a.province || '';
        var pais = a.country || '';
        actualizarCard(card, {
          titulo: ciudad || 'Tu ubicación',
          direccion: [calle, region, pais].filter(Boolean).join(' · '),
          coords: lat.toFixed(5) + ', ' + lng.toFixed(5)
        });
      })
      .catch(function () { /* sin conexión: dejamos lo que haya */ });
  }

  function actualizarCard(card, info) {
    var t = card.querySelector('[data-ub-titulo]');
    var d = card.querySelector('[data-ub-direccion]');
    var c = card.querySelector('[data-ub-coords]');
    if (t) { t.textContent = info.titulo; }
    if (d) { d.textContent = info.direccion; }
    if (c) { c.textContent = info.coords; }
    card.classList.remove('oculto');
  }

  /* ─────────────────  BARRA LATERAL POR MACIZO  ───────────────── */

  /** Construye los chips de filtro (uno por macizo) y la lista
   *  de estaciones agrupada. Se llama tras cargar las estaciones. */
  function construirSidebar() {
    pintarChips();
    pintarGrupos();
  }

  /** Devuelve las estaciones que pasan los filtros activos. */
  function listaFiltrada() {
    var t = filtroTexto.toLowerCase().trim();
    return listaActual.filter(function (e) {
      if (filtroMacizo && e.macizo_clave !== filtroMacizo) { return false; }
      if (t && (e.nombre + ' ' + (e.ubicacion || '')).toLowerCase().indexOf(t) === -1) { return false; }
      return true;
    });
  }

  /** Pinta los chips de "Todos" + un chip por cada macizo presente. */
  function pintarChips() {
    var ul = document.getElementById('mapa-filtros');
    if (!ul) { return; }
    ul.innerHTML = '';

    var grupos = agruparPorMacizo(listaActual);
    var totalAll = listaActual.length;

    ul.appendChild(chip('', 'Todos', totalAll));
    Object.keys(grupos).forEach(function (clave) {
      var g = grupos[clave];
      ul.appendChild(chip(clave, g.nombre, g.estaciones.length));
    });
  }

  function chip(clave, etiqueta, total) {
    var li = document.createElement('li');
    var b  = document.createElement('button');
    b.type = 'button';
    b.className = 'macizo-chip macizo-chip--' + (clave || 'todos');
    if (filtroMacizo === clave) { b.classList.add('macizo-chip--active'); }
    b.setAttribute('data-macizo', clave);
    b.innerHTML = '<span>' + etiqueta + '</span><span class="macizo-chip__count">' + total + '</span>';
    b.addEventListener('click', function () {
      filtroMacizo = (filtroMacizo === clave ? '' : clave);
      pintarChips();
      pintarGrupos();
    });
    li.appendChild(b);
    return li;
  }

  /** Devuelve un objeto {claveMacizo: {nombre, estaciones[]}} respetando
   *  el orden de aparición en la respuesta del API. */
  function agruparPorMacizo(lista) {
    var orden = ['pirineo-cat', 'pirineo-ara', 'andorra', 'penibetico', 'central', 'cantabrica', 'otros'];
    var grupos = {};
    lista.forEach(function (e) {
      var k = e.macizo_clave || 'otros';
      if (!grupos[k]) { grupos[k] = { nombre: e.macizo || 'Otros', estaciones: [] }; }
      grupos[k].estaciones.push(e);
    });
    var ordenadas = {};
    orden.forEach(function (k) { if (grupos[k]) { ordenadas[k] = grupos[k]; } });
    Object.keys(grupos).forEach(function (k) { if (!ordenadas[k]) { ordenadas[k] = grupos[k]; } });
    return ordenadas;
  }

  /** Pinta los grupos de estaciones en el panel lateral, con cabecera
   *  por macizo y un item clicable por estación. */
  function pintarGrupos() {
    var cont = document.getElementById('mapa-grupos');
    if (!cont) { return; }
    cont.innerHTML = '';

    var lista = listaFiltrada();
    if (lista.length === 0) {
      var v = document.createElement('p');
      v.className = 'map-sidebar__empty';
      v.textContent = 'No hay estaciones con ese filtro.';
      cont.appendChild(v);
      return;
    }

    var grupos = agruparPorMacizo(lista);
    Object.keys(grupos).forEach(function (clave) {
      var g = grupos[clave];

      var sec = document.createElement('section');
      sec.className = 'macizo-group macizo-group--' + clave;

      var head = document.createElement('header');
      head.className = 'macizo-group__head';
      head.innerHTML = ''
        + '<span class="macizo-group__title">' + g.nombre + '</span>'
        + '<span class="macizo-group__count">' + g.estaciones.length + '</span>';
      sec.appendChild(head);

      var ul = document.createElement('ul');
      ul.className = 'macizo-group__list';
      g.estaciones.forEach(function (e) { ul.appendChild(itemEstacion(e)); });
      sec.appendChild(ul);

      cont.appendChild(sec);
    });
  }

  function itemEstacion(e) {
    var li = document.createElement('li');
    li.className = 'macizo-item macizo-item--' + (e.estado || 'regular');
    li.tabIndex = 0;
    li.setAttribute('role', 'button');
    li.setAttribute('aria-label', e.nombre + ' — ' + (e.estado_texto || 'estado desconocido'));

    li.innerHTML = ''
      + '<span class="macizo-item__dot" aria-hidden="true"></span>'
      + '<div class="macizo-item__body">'
      +   '<strong class="macizo-item__name">' + e.nombre + '</strong>'
      +   '<span class="macizo-item__meta">' + (e.ubicacion || '') + ' · ' + e.pistas_abiertas + '/' + e.pistas_totales + ' pistas</span>'
      + '</div>'
      + '<span class="macizo-item__estado">' + (e.estado_texto || '–') + '</span>';

    var volar = function () { centrarEnEstacion(e); };
    li.addEventListener('click', volar);
    li.addEventListener('keypress', function (ev) {
      if (ev.key === 'Enter' || ev.key === ' ') { ev.preventDefault(); volar(); }
    });
    return li;
  }

  /** Centra el mapa en la estación y abre su popup. */
  function centrarEnEstacion(e) {
    if (!e.tiene_coords) { return; }
    mapa.flyTo([e.lat, e.lng], 12, { duration: 0.9 });
    var m = markersPorId[e.id];
    if (m) { setTimeout(function () { m.openPopup(); }, 700); }
  }

  function conectarBuscadorYFiltros() {
    var input = document.getElementById('mapa-search');
    if (input && !input._wired) {
      input._wired = true;
      input.addEventListener('input', function () {
        filtroTexto = input.value || '';
        pintarGrupos();
      });
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', iniciar);
  } else {
    iniciar();
  }
})();
