/* ===========================================================
 * SkiHub — Visor de mapa de estaciones.
 *
 * Activa los botones "Lista" / "Mapa" del bloque "Estaciones cercanas"
 * de la página inicio. Al pulsar "Mapa":
 *   1. Oculta la rejilla de tarjetas y los filtros.
 *   2. Carga Leaflet (mapa OpenStreetMap) bajo demanda desde CDN.
 *   3. Llama a GET /api/estaciones para obtener los puntos.
 *   4. Pinta un marker por estación con popup informativo.
 *
 * fetch + .then() (sin async/await), manipulación SOLO con classList,
 * sin estilos en línea desde JS.
 * =========================================================== */

(function () {
  'use strict';

  // CDN de Leaflet (versión estable). CSS + JS.
  var LEAFLET_CSS = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css';
  var LEAFLET_JS  = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.js';

  // Estado interno: instancia del mapa Leaflet (null si no se ha creado).
  var mapa = null;

  function iniciar() {
    var btnLista = document.querySelector('[data-vista="lista"]');
    var btnMapa  = document.querySelector('[data-vista="mapa"]');
    if (!btnLista || !btnMapa) { return; }

    btnLista.addEventListener('click', mostrarLista);
    btnMapa.addEventListener('click', mostrarMapa);
  }

  // -----------------------------------------------------------
  // Cambio de vista: clases CSS, sin tocar style.
  // -----------------------------------------------------------

  function mostrarLista() {
    var grid = document.getElementById('grid-estaciones');
    var nav  = document.getElementById('filtros-estaciones');
    var cont = document.getElementById('mapa-container');
    if (grid) { grid.classList.remove('oculto'); }
    if (nav)  { nav.classList.remove('oculto'); }
    if (cont) { cont.classList.add('oculto'); }
    marcarBotonActivo('lista');
  }

  function mostrarMapa() {
    var grid = document.getElementById('grid-estaciones');
    var nav  = document.getElementById('filtros-estaciones');
    var cont = document.getElementById('mapa-container');
    if (grid) { grid.classList.add('oculto'); }
    if (nav)  { nav.classList.add('oculto'); }
    if (cont) { cont.classList.remove('oculto'); }
    marcarBotonActivo('mapa');

    // Carga perezosa: solo cuando el usuario lo pide la primera vez.
    cargarLeaflet().then(pintarMapa).catch(function (err) {
      var c = document.getElementById('mapa-container');
      if (c) {
        c.textContent = 'No se pudo cargar el mapa: ' + err.message;
      }
    });
  }

  function marcarBotonActivo(cual) {
    var todos = document.querySelectorAll('.filters__view-btn');
    for (var i = 0; i < todos.length; i++) {
      todos[i].classList.remove('filters__view-btn--active');
    }
    var btn = document.querySelector('[data-vista="' + cual + '"]');
    if (btn) { btn.classList.add('filters__view-btn--active'); }
  }

  // -----------------------------------------------------------
  // Carga dinámica de Leaflet desde CDN (CSS + JS).
  // -----------------------------------------------------------

  function cargarLeaflet() {
    // Si ya está cargado en window.L, no hace falta volver a meterlo.
    if (window.L) { return Promise.resolve(); }

    return new Promise(function (resolve, reject) {
      // CSS.
      if (!document.querySelector('link[data-leaflet]')) {
        var link = document.createElement('link');
        link.rel = 'stylesheet';
        link.href = LEAFLET_CSS;
        link.setAttribute('data-leaflet', '1');
        document.head.appendChild(link);
      }
      // JS.
      var script = document.createElement('script');
      script.src = LEAFLET_JS;
      script.setAttribute('data-leaflet', '1');
      script.onload = function () { resolve(); };
      script.onerror = function () { reject(new Error('no se pudo cargar Leaflet')); };
      document.head.appendChild(script);
    });
  }

  // -----------------------------------------------------------
  // Pintar el mapa: una sola vez, luego solo se hace fitBounds.
  // -----------------------------------------------------------

  function pintarMapa() {
    var div = document.getElementById('mapa-leaflet');
    if (!div) { return; }

    if (mapa === null) {
      // Centro inicial aproximado en la península; el fitBounds posterior
      // ajusta el zoom para que entren todos los markers.
      mapa = L.map(div).setView([41.0, -2.0], 6);
      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 18,
        attribution: '© OpenStreetMap'
      }).addTo(mapa);
    }

    fetch('/api/estaciones', { headers: { 'Accept': 'application/json' } })
      .then(function (resp) {
        if (!resp.ok) { throw new Error('HTTP ' + resp.status); }
        return resp.json();
      })
      .then(function (lista) {
        anadirMarkers(lista);
      })
      .catch(function (err) {
        console.error('Error API estaciones:', err);
      });
  }

  function anadirMarkers(lista) {
    var puntos = [];
    lista.forEach(function (e) {
      if (!e.tiene_coords) { return; } // saltamos las que no tienen lat/lng
      var marker = L.marker([e.lat, e.lng]).addTo(mapa);
      marker.bindPopup(popupHTML(e));
      puntos.push([e.lat, e.lng]);
    });
    // Encuadra el mapa para que entren todos los puntos.
    if (puntos.length > 0) {
      mapa.fitBounds(puntos, { padding: [40, 40] });
    }
    // El contenedor cambia de tamaño al pasar de oculto a visible:
    // forzamos a Leaflet a recalcular dimensiones.
    setTimeout(function () { mapa.invalidateSize(); }, 50);
  }

  // popupHTML construye el contenido del popup de cada marker.
  // Texto plano; los datos vienen de la API y no son confiables al 100%,
  // así que usamos textContent para evitar inyección.
  function popupHTML(e) {
    var box = document.createElement('div');
    box.className = 'mapa-popup';

    var h = document.createElement('strong');
    h.textContent = e.nombre;
    box.appendChild(h);

    box.appendChild(saltoLinea());
    box.appendChild(textoPequeno(e.ubicacion || ''));

    if (e.altitud) {
      box.appendChild(saltoLinea());
      box.appendChild(textoPequeno('Altitud: ' + e.altitud));
    }
    box.appendChild(saltoLinea());
    box.appendChild(textoPequeno(
      'Pistas: ' + e.pistas_abiertas + '/' + e.pistas_totales +
      ' · Nieve base: ' + e.nieve_base + ' cm'
    ));

    box.appendChild(saltoLinea());
    var enlace = document.createElement('a');
    enlace.href = '/estacion/' + e.id;
    enlace.textContent = 'Ver detalle →';
    box.appendChild(enlace);

    return box;
  }

  function saltoLinea() { return document.createElement('br'); }
  function textoPequeno(t) {
    var s = document.createElement('small');
    s.textContent = t;
    return s;
  }

  // -----------------------------------------------------------
  // Arranque
  // -----------------------------------------------------------
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', iniciar);
  } else {
    iniciar();
  }
})();
