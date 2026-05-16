/**
 * pistas-tecnico.js
 * Rellena la tabla técnica "Parte de Nieve por Sectores" usando los datos
 * que pistas.js ya cargó. Espera un evento personalizado 'pistasDataReady'
 * con { detail: estaciones[] }, o hace polling ligero si no llega.
 */
(function () {
  'use strict';

  var tbody = document.getElementById('pistas-tabla-body');
  if (!tbody) return;

  function nivelAlud(nieve) {
    if (nieve == null || nieve === 0) return 'Sin datos';
    if (nieve > 100) return 'Nivel 3 (Considerable)';
    if (nieve > 50)  return 'Nivel 2 (Limitado)';
    return 'Nivel 1 (Bajo)';
  }

  function estadoClase(estado) {
    switch (estado) {
      case 'abierta': return 'pistas-tabla__dot--verde';
      case 'parcial': return 'pistas-tabla__dot--azul';
      case 'cerrada': return 'pistas-tabla__dot--rojo';
      default:        return 'pistas-tabla__dot--negro';
    }
  }

  function formatFraccion(f) {
    if (!f) return '—';
    var a = f.abiertos != null ? f.abiertos : '—';
    var t = f.total    != null ? f.total    : '—';
    if (a === '—' && t === '—') return '—';
    return a + ' / ' + t;
  }

  function escHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;')
      .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  function renderTabla(estaciones) {
    if (!tbody) return;
    if (!estaciones || !estaciones.length) {
      tbody.innerHTML = '<tr><td colspan="4" class="pistas-tabla__empty">Sin datos disponibles.</td></tr>';
      return;
    }

    tbody.innerHTML = '';
    estaciones.slice(0, 12).forEach(function (e) {
      var nieve = e.nieve_cm != null ? e.nieve_cm : 0;
      var nieveMin = Math.max(0, Math.round(nieve * 0.7));
      var tr = document.createElement('tr');

      var pistasText = formatFraccion(e.pistas);
      var dotCls = estadoClase(e.estado || 'desconocido');

      tr.innerHTML =
        '<td class="pistas-tabla__sector">' + escHtml(e.nombre || e.slug || '—') + '</td>' +
        '<td class="pistas-tabla__pistas">' +
          '<span class="pistas-tabla__dot ' + dotCls + '" aria-hidden="true"></span>' +
          escHtml(pistasText) +
        '</td>' +
        '<td class="pistas-tabla__espesores">' +
          (nieve > 0 ? escHtml(nieveMin + ' / ' + nieve + ' cm') : '—') +
        '</td>' +
        '<td class="pistas-tabla__alud">' + escHtml(nivelAlud(nieve)) + '</td>';

      tbody.appendChild(tr);
    });
  }

  /* pistas.js no emite un evento propio, así que hacemos polling ligero
     sobre state global que pistas.js no expone. En su lugar interceptamos
     el momento en que el DOM del grid tiene tarjetas reales (no skeletons). */
  var intentos = 0;
  var MAX = 30; /* 30 × 500 ms = 15 s máximo */

  function fetchYRender() {
    /* Pide los datos directamente para no depender de estado interno de pistas.js */
    var geo = window.SnowbreakGeo && window.SnowbreakGeo.leerPreferencia();
    var lat = geo ? geo.lat : 40.4168;
    var lng = geo ? geo.lng : -3.7038;

    fetch('/api/nieve/estaciones?lat=' + encodeURIComponent(lat) + '&lng=' + encodeURIComponent(lng),
      { headers: { 'Accept': 'application/json' } })
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (j) {
        if (j && j.data) renderTabla(j.data);
      })
      .catch(function () {
        tbody.innerHTML = '<tr><td colspan="4" class="pistas-tabla__empty">No se pudieron cargar los datos.</td></tr>';
      });
  }

  /* Espera a que pistas.js haya resuelto la geolocalización (~2s) */
  function esperarGeo() {
    intentos++;
    if (intentos > MAX) { fetchYRender(); return; }
    var geo = window.SnowbreakGeo && window.SnowbreakGeo.leerPreferencia();
    if (geo) {
      fetchYRender();
    } else {
      setTimeout(esperarGeo, 500);
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () { setTimeout(esperarGeo, 1500); });
  } else {
    setTimeout(esperarGeo, 1500);
  }
})();
