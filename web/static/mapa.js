(function () {
  'use strict';

  var LEAFLET_CSS = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css';
  var LEAFLET_JS = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.js';
  var mapa = null;
  var capaMarkers = null;

  function iniciar() {
    var btnLista = document.querySelector('[data-vista="lista"]');
    var btnMapa = document.querySelector('[data-vista="mapa"]');
    if (!btnLista || !btnMapa) { return; }

    btnLista.addEventListener('click', mostrarLista);
    btnMapa.addEventListener('click', mostrarMapa);
  }

  function mostrarLista() {
    alternarVista(false);
  }

  function mostrarMapa() {
    alternarVista(true);
    cargarLeaflet().then(pintarMapa).catch(function (err) {
      var c = document.getElementById('mapa-container');
      if (c) {
        c.textContent = 'No se pudo cargar el mapa: ' + err.message;
      }
    });
  }

  function alternarVista(verMapa) {
    var grid = document.getElementById('grid-estaciones');
    var nav = document.getElementById('filtros-estaciones');
    var cont = document.getElementById('mapa-container');
    if (grid) { grid.classList.toggle('oculto', verMapa); }
    if (nav) { nav.classList.toggle('oculto', verMapa); }
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
      script.onload = function () { resolve(); };
      script.onerror = function () { reject(new Error('no se pudo cargar Leaflet')); };
      document.head.appendChild(script);
    });
  }

  function pintarMapa() {
    var div = document.getElementById('mapa-leaflet');
    if (!div) { return; }

    if (mapa === null) {
      mapa = L.map(div).setView([41.0, -2.0], 6);
      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 18,
        attribution: '© OpenStreetMap'
      }).addTo(mapa);
      capaMarkers = L.layerGroup().addTo(mapa);
    }

    fetch('/api/estaciones', { headers: { Accept: 'application/json' } })
      .then(function (resp) {
        if (!resp.ok) { throw new Error('HTTP ' + resp.status); }
        return resp.json();
      })
      .then(anadirMarkers)
      .catch(function (err) {
        console.error('Error API estaciones:', err);
      });
  }

  function anadirMarkers(lista) {
    var puntos = [];
    capaMarkers.clearLayers();

    lista.forEach(function (e) {
      if (!e.tiene_coords) { return; }
      var marker = L.marker([e.lat, e.lng]).addTo(capaMarkers);
      marker.bindPopup(popupHTML(e));
      puntos.push([e.lat, e.lng]);
    });

    if (puntos.length > 0) {
      mapa.fitBounds(puntos, { padding: [40, 40] });
    }
    setTimeout(function () { mapa.invalidateSize(); }, 50);
  }

  function popupHTML(e) {
    var box = document.createElement('div');
    var h = document.createElement('strong');
    h.textContent = e.nombre;
    box.appendChild(h);
    box.appendChild(document.createElement('br'));
    box.appendChild(textoPequeno(e.ubicacion || ''));
    if (e.altitud) {
      box.appendChild(document.createElement('br'));
      box.appendChild(textoPequeno('Altitud: ' + e.altitud));
    }
    box.appendChild(document.createElement('br'));
    box.appendChild(textoPequeno('Pistas: ' + e.pistas_abiertas + '/' + e.pistas_totales + ' · Nieve base: ' + e.nieve_base + ' cm'));
    box.appendChild(document.createElement('br'));
    var enlace = document.createElement('a');
    enlace.href = '/estacion/' + e.id;
    enlace.textContent = 'Ver detalle';
    box.appendChild(enlace);
    return box;
  }

  function textoPequeno(t) {
    var s = document.createElement('small');
    s.textContent = t;
    return s;
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', iniciar);
  } else {
    iniciar();
  }
})();
