/* Lógica de la página /pistas: carga estaciones desde /api/nieve/estaciones,
 * gestiona el flujo de geolocalización y dibuja una tarjeta por estación.
 *
 * Consume window.SnowbreakGeo (geolocation.js) que debe haberse cargado
 * antes en la plantilla.
 */
(function () {
  'use strict';

  var ENDPOINT = '/api/nieve/estaciones';

  var state = {
    ubicacion: null,
    estaciones: [],
    filtro: 'todas',
    cargando: false,
    error: null
  };

  var refs = {};

  function cachearRefs() {
    refs.estado    = document.getElementById('pistas-estado');
    refs.grid      = document.getElementById('pistas-grid');
    refs.filtros   = document.getElementById('pistas-filtros');
    refs.ciudades  = document.getElementById('pistas-ciudades');
    refs.formCiu   = refs.ciudades && refs.ciudades.querySelector('[data-ciudad-form]');
    refs.presets   = refs.ciudades && refs.ciudades.querySelector('[data-ciudad-presets]');
    refs.tpl       = document.getElementById('tpl-pista-card');
    refs.meta      = document.getElementById('pistas-meta');
    refs.metaUbic  = document.getElementById('pistas-meta-ubicacion');
    refs.metaTotal = document.getElementById('pistas-meta-total');
    refs.metaAct   = document.getElementById('pistas-meta-actualizado');
    refs.btnRecarg = document.querySelector('[data-pistas-recargar]');
    refs.btnCiudad = document.querySelector('[data-pistas-elegir]');
    refs.resumen   = document.getElementById('pistas-resumen');
  }

  function iniciar() {
    cachearRefs();
    if (!refs.estado || !refs.grid) { return; }

    pintarPresets();
    cablearEventos();

    var prev = window.SnowbreakGeo && window.SnowbreakGeo.leerPreferencia();
    if (prev) {
      usarUbicacion(prev);
    } else {
      pedirUbicacion();
    }
  }

  function cablearEventos() {
    refs.filtros.addEventListener('click', function (ev) {
      var btn = ev.target.closest('button[data-filtro]');
      if (!btn) { return; }
      state.filtro = btn.dataset.filtro;
      Array.prototype.forEach.call(refs.filtros.querySelectorAll('.filters__btn'), function (b) {
        var activo = b === btn;
        b.classList.toggle('filters__btn--active', activo);
        b.setAttribute('aria-pressed', activo ? 'true' : 'false');
      });
      pintarLista();
    });

    if (refs.btnRecarg) {
      refs.btnRecarg.addEventListener('click', function () {
        window.SnowbreakGeo && window.SnowbreakGeo.olvidarPreferencia();
        pedirUbicacion();
      });
    }
    if (refs.btnCiudad) {
      refs.btnCiudad.addEventListener('click', function () {
        refs.ciudades.hidden = !refs.ciudades.hidden;
        if (!refs.ciudades.hidden) {
          var input = refs.formCiu.querySelector('input[name="ciudad"]');
          if (input) { input.focus(); }
        }
      });
    }

    if (refs.formCiu) {
      refs.formCiu.addEventListener('submit', function (ev) {
        ev.preventDefault();
        var input = refs.formCiu.querySelector('input[name="ciudad"]');
        if (!input || !input.value.trim()) { return; }
        mostrarEstado('Buscando estaciones cercanas a "' + input.value + '"…', 'cargando');
        window.SnowbreakGeo.desdeCiudad(input.value)
          .then(function (u) { usarUbicacion(u); refs.ciudades.hidden = true; })
          .catch(function (err) { mostrarErrorGeo(err); });
      });
    }
  }

  function pintarPresets() {
    if (!refs.presets || !window.SnowbreakGeo) { return; }
    refs.presets.innerHTML = '';
    window.SnowbreakGeo.CIUDADES_PRESET.forEach(function (c) {
      var b = document.createElement('button');
      b.type = 'button';
      b.className = 'pistas-cities__chip';
      b.textContent = c.etiqueta;
      b.addEventListener('click', function () {
        usarUbicacion({ lat: c.lat, lng: c.lng, etiqueta: c.etiqueta, fuente: 'preset' });
        refs.ciudades.hidden = true;
      });
      refs.presets.appendChild(b);
    });
  }

  function pedirUbicacion() {
    mostrarEstado('Buscando estaciones cercanas a tu ubicación…', 'cargando');
    window.SnowbreakGeo.obtener()
      .then(usarUbicacion)
      .catch(function (err) {
        // Fallback automático a Madrid + sugerencia de ciudades.
        var madrid = window.SnowbreakGeo.CIUDADES_PRESET[0];
        usarUbicacion(Object.assign({ etiqueta: 'Madrid (por defecto)' }, madrid));
        if (refs.ciudades) { refs.ciudades.hidden = false; }
        mostrarAvisoFallback(err);
      });
  }

  function usarUbicacion(ubic) {
    state.ubicacion = ubic;
    if (window.SnowbreakGeo) { window.SnowbreakGeo.guardarPreferencia(ubic); }
    if (refs.resumen) {
      refs.resumen.textContent = 'Mostrando estaciones cercanas a ' + (ubic.etiqueta || 'tu ubicación') + '.';
    }
    cargar();
  }

  function cargar() {
    if (state.cargando || !state.ubicacion) { return; }
    state.cargando = true;
    state.error = null;
    mostrarEstado('Consultando el estado actualizado de las pistas…', 'cargando');

    var url = ENDPOINT
      + '?lat=' + encodeURIComponent(state.ubicacion.lat)
      + '&lng=' + encodeURIComponent(state.ubicacion.lng);

    fetch(url, { headers: { 'Accept': 'application/json' } })
      .then(function (r) {
        if (!r.ok) {
          return r.json().catch(function () { return null; }).then(function (j) {
            var msg = (j && j.error) ? j.error : 'HTTP ' + r.status;
            throw new Error(msg);
          });
        }
        return r.json();
      })
      .then(function (j) {
        state.estaciones = (j && j.data) || [];
        state.cargando = false;
        actualizarMeta(j);
        pintarLista();
      })
      .catch(function (err) {
        state.cargando = false;
        state.error = err && err.message ? err.message : 'Error de red';
        mostrarEstado('No se han podido obtener los datos: ' + state.error +
          '. Vuelve a intentarlo en unos minutos.', 'error');
      });
  }

  function actualizarMeta(resp) {
    if (!refs.meta) { return; }
    refs.meta.hidden = false;
    var ubic = (state.ubicacion && state.ubicacion.etiqueta) || 'tu ubicación';
    refs.metaUbic.textContent  = 'Ubicación: ' + ubic;
    refs.metaTotal.textContent = (resp && resp.total ? resp.total : state.estaciones.length) + ' estaciones';
    refs.metaAct.textContent   = 'Actualizado: ' + new Date().toLocaleTimeString('es-ES', {
      hour: '2-digit', minute: '2-digit'
    });
  }

  function pintarLista() {
    if (!refs.grid || !refs.tpl) { return; }
    refs.grid.innerHTML = '';
    var lista = filtrar(state.estaciones, state.filtro);
    if (lista.length === 0) {
      mostrarEstado('No hay estaciones que cumplan ese filtro.', 'vacio');
      refs.grid.hidden = true;
      return;
    }
    refs.estado.hidden = true;
    refs.grid.hidden = false;
    var frag = document.createDocumentFragment();
    lista.forEach(function (e) { frag.appendChild(crearCard(e)); });
    refs.grid.appendChild(frag);
  }

  function filtrar(lista, filtro) {
    if (!filtro || filtro === 'todas') { return lista; }
    return lista.filter(function (e) { return (e.estado || '') === filtro; });
  }

  function crearCard(e) {
    var nodo = refs.tpl.content.firstElementChild.cloneNode(true);

    nodo.querySelector('.pista-card__name').textContent = e.nombre || e.slug;

    // Estado
    var estadoBox  = nodo.querySelector('[data-estado]');
    var estadoText = estadoBox.querySelector('.pista-card__status-text');
    var estado = e.estado || 'desconocido';
    estadoBox.classList.add('pista-card__status--' + estado);
    estadoText.textContent = etiquetaEstado(estado);

    // Distancia
    if (typeof e.distancia_km === 'number') {
      var dist = nodo.querySelector('[data-distancia]');
      dist.hidden = false;
      dist.querySelector('span').textContent = formatKm(e.distancia_km) + ' km';
    }

    // Métricas
    nodo.querySelector('[data-pistas]').textContent    = formatFraccion(e.pistas);
    nodo.querySelector('[data-km]').textContent        = formatKmFraccion(e.kilometros);
    nodo.querySelector('[data-remontes]').textContent  = formatFraccion(e.remontes);
    nodo.querySelector('[data-nieve]').textContent     = (e.nieve_cm == null ? '—' : e.nieve_cm + ' cm');

    var calidad = nodo.querySelector('[data-calidad]');
    if (e.calidad_nieve) {
      calidad.textContent = e.calidad_nieve;
    } else if (e.temperatura) {
      calidad.textContent = e.temperatura;
    } else {
      calidad.textContent = '';
    }

    var enlace = nodo.querySelector('[data-detalle]');
    enlace.href = e.url || ('https://www.infonieve.es/estacion-esqui/' + (e.slug || ''));
    enlace.setAttribute('aria-label', 'Ver ' + (e.nombre || e.slug) + ' en infonieve.es (se abre en otra pestaña)');

    return nodo;
  }

  function etiquetaEstado(s) {
    switch (s) {
      case 'abierta': return 'Abierta';
      case 'cerrada': return 'Cerrada';
      case 'parcial': return 'Apertura parcial';
      default:        return 'Sin datos';
    }
  }

  function formatKm(n) {
    return (Math.round(n * 10) / 10).toString().replace('.', ',');
  }

  function formatFraccion(f) {
    if (!f) { return '—'; }
    var a = (f.abiertos == null ? '—' : f.abiertos);
    var t = (f.total    == null ? '—' : f.total);
    if (a === '—' && t === '—') { return '—'; }
    return a + ' / ' + t;
  }

  function formatKmFraccion(f) {
    if (!f) { return '—'; }
    var a = (f.abiertos == null ? '—' : f.abiertos);
    var t = (f.total    == null ? '—' : f.total);
    if (a === '—' && t === '—') { return '—'; }
    return a + ' / ' + t + ' km';
  }

  function mostrarEstado(texto, clase) {
    if (!refs.estado) { return; }
    refs.estado.hidden = false;
    refs.estado.className = 'pistas-state pistas-state--' + (clase || 'info');
    refs.estado.innerHTML = '';
    var p = document.createElement('p');
    p.className = 'pistas-state__msg';
    if (clase === 'cargando') {
      var spin = document.createElement('span');
      spin.className = 'pistas-state__spinner';
      spin.setAttribute('aria-hidden', 'true');
      p.appendChild(spin);
    }
    p.appendChild(document.createTextNode(texto));
    refs.estado.appendChild(p);
    if (refs.grid) {
      if (clase === 'cargando') {
        refs.grid.hidden = false;
        refs.grid.innerHTML = '';
        for (var i = 0; i < 6; i++) {
          var li = document.createElement('li');
          li.className = 'pista-card-item';
          var sk = document.createElement('div');
          sk.className = 'skeleton skeleton--card';
          sk.setAttribute('aria-hidden', 'true');
          li.appendChild(sk);
          refs.grid.appendChild(li);
        }
      } else {
        refs.grid.hidden = true;
      }
    }
  }

  function mostrarErrorGeo(err) {
    var msg = 'No pudimos obtener tu ubicación.';
    if (err && err.codigo) {
      switch (err.codigo) {
        case 'denegado':       msg = 'Has denegado el permiso. Elige una ciudad o introdúcela manualmente.'; break;
        case 'no_soportado':   msg = 'Tu navegador no permite geolocalización.'; break;
        case 'timeout':        msg = 'La ubicación tardó demasiado. Prueba a elegir una ciudad.'; break;
        case 'no_encontrada':  msg = 'No encontramos esa ciudad. Prueba con otra distinta.'; break;
        case 'red':            msg = 'No pudimos contactar con el servicio de geocoding. Inténtalo más tarde.'; break;
      }
    }
    mostrarEstado(msg, 'error');
  }

  function mostrarAvisoFallback(err) {
    var aviso = document.createElement('p');
    aviso.className = 'pistas-state__aviso';
    if (err && err.codigo === 'denegado') {
      aviso.textContent = 'Como no nos has dado permiso de ubicación, mostramos las estaciones más cercanas a Madrid.';
    } else {
      aviso.textContent = 'No pudimos obtener tu ubicación; usamos Madrid como referencia. Puedes cambiar la ciudad arriba.';
    }
    if (refs.estado) {
      refs.estado.appendChild(aviso);
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', iniciar);
  } else {
    iniciar();
  }
})();
