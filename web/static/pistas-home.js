/* Widget en la home: pinta las 3 estaciones con datos en directo más
 * cercanas a la ubicación del usuario. No pide permiso de ubicación de
 * forma activa: si el usuario ya lo había concedido en /pistas (queda
 * en localStorage), reutilizamos ese valor; si no, usamos Madrid como
 * fallback silencioso para no molestar.
 */
(function () {
  'use strict';

  var ENDPOINT = '/api/nieve/estaciones';

  function iniciar() {
    var contenedor = document.getElementById('home-pistas-directo');
    if (!contenedor) { return; }
    var lista = contenedor.querySelector('[data-home-pistas-lista]');
    var aviso = contenedor.querySelector('[data-home-pistas-aviso]');
    if (!lista) { return; }

    var ubic = (window.SnowbreakGeo && window.SnowbreakGeo.leerPreferencia()) || null;
    if (!ubic) {
      ubic = { lat: 40.4168, lng: -3.7038, etiqueta: 'Madrid' };
    }

    if (aviso) {
      aviso.textContent = 'Mostrando datos cercanos a ' + ubic.etiqueta + '.';
    }

    var url = ENDPOINT
      + '?lat=' + encodeURIComponent(ubic.lat)
      + '&lng=' + encodeURIComponent(ubic.lng)
      + '&limit=3';

    fetch(url, { headers: { 'Accept': 'application/json' } })
      .then(function (r) { if (!r.ok) { throw new Error('HTTP ' + r.status); } return r.json(); })
      .then(function (j) {
        var data = (j && j.data) || [];
        if (data.length === 0) {
          if (aviso) { aviso.textContent = 'No hay datos en directo disponibles ahora mismo.'; }
          return;
        }
        renderTarjetas(lista, data);
      })
      .catch(function (err) {
        if (aviso) {
          aviso.textContent = 'No se pudieron obtener datos en directo (' + (err.message || 'error') + ').';
        }
      });
  }

  function renderTarjetas(contenedor, data) {
    contenedor.innerHTML = '';
    data.forEach(function (e) {
      var li = document.createElement('li');
      li.className = 'pistas-home__item';

      var card = document.createElement('article');
      card.className = 'pistas-home__card';

      var h = document.createElement('h3');
      h.className = 'pistas-home__name';
      h.textContent = e.nombre;
      card.appendChild(h);

      var estado = document.createElement('p');
      estado.className = 'pistas-home__status pista-card__status--' + (e.estado || 'desconocido');
      estado.innerHTML = '<span class="pista-card__status-dot" aria-hidden="true"></span>' +
        (e.estado === 'abierta' ? 'Abierta'
         : e.estado === 'parcial' ? 'Apertura parcial'
         : e.estado === 'cerrada' ? 'Cerrada' : 'Sin datos');
      card.appendChild(estado);

      var datos = document.createElement('p');
      datos.className = 'pistas-home__data';
      var partes = [];
      if (typeof e.distancia_km === 'number') {
        partes.push((Math.round(e.distancia_km * 10) / 10).toString().replace('.', ',') + ' km');
      }
      if (e.kilometros && e.kilometros.abiertos != null && e.kilometros.total != null) {
        partes.push(e.kilometros.abiertos + '/' + e.kilometros.total + ' km abiertos');
      }
      if (typeof e.nieve_cm === 'number') {
        partes.push(e.nieve_cm + ' cm de nieve');
      }
      datos.textContent = partes.join(' · ');
      card.appendChild(datos);

      var enlace = document.createElement('a');
      enlace.className = 'pistas-home__link text-link';
      enlace.href = '/pistas';
      enlace.textContent = 'Ver detalle →';
      card.appendChild(enlace);

      li.appendChild(card);
      contenedor.appendChild(li);
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', iniciar);
  } else {
    iniciar();
  }
})();
