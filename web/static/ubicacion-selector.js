/* Snowbreak — Selector de ubicación
 *
 * Pinta una píldora "Estoy en: <ciudad>" + botón "Cambiar ubicación".
 * Al pulsar el botón abre un panel con:
 *   - Botón "Usar mi ubicación" (Geolocation API).
 *   - Input para ciudad o código postal.
 *   - Lista rápida de presets (Madrid, Barcelona…).
 *
 * Cuando el usuario fija una ubicación nueva, dispara un CustomEvent
 * "snowbreak:ubicacion:cambio" con detail = {lat, lng, etiqueta, fuente}.
 * Persiste la elección con SnowbreakGeo.guardarPreferencia.
 *
 * Madrid (40.4168, -3.7038) es el fallback si el usuario no tiene
 * preferencia guardada.
 */
(function () {
  'use strict';

  var FALLBACK = { lat: 40.4168, lng: -3.7038, etiqueta: 'Madrid', fuente: 'preset' };

  function init() {
    var anclas = document.querySelectorAll('[data-ubicacion-selector]');
    if (!anclas.length) { return; }

    // Cargamos la última ubicación conocida (si existe) o Madrid por defecto.
    var actual = leerActual();

    // Pintamos cada ancla con la UI completa y guardamos el contexto.
    var instancias = [];
    anclas.forEach(function (ancla) {
      var ui = pintar(ancla, actual);
      instancias.push(ui);
      ui.boton.addEventListener('click', function () { abrirPanel(ui, actual); });
    });

    // Difundimos la ubicación inicial para que filtros/sort se enteren.
    publicar(actual);

    // Si en algún sitio cambian la ubicación, sincronizamos todas las píldoras.
    document.addEventListener('snowbreak:ubicacion:cambio', function (ev) {
      actual = ev.detail;
      instancias.forEach(function (ui) { ui.titulo.textContent = actual.etiqueta; });
    });
  }

  function leerActual() {
    if (window.SnowbreakGeo && SnowbreakGeo.leerPreferencia) {
      var p = SnowbreakGeo.leerPreferencia();
      if (p && typeof p.lat === 'number') {
        return { lat: p.lat, lng: p.lng, etiqueta: p.etiqueta || 'Tu ubicación', fuente: p.fuente || 'manual' };
      }
    }
    return Object.assign({}, FALLBACK);
  }

  function pintar(ancla, actual) {
    ancla.classList.add('ubic-selector');
    ancla.innerHTML =
      '<span class="ubic-selector__icon" aria-hidden="true">' +
        '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>' +
      '</span>' +
      '<span class="ubic-selector__label">Estoy en: <strong data-titulo>' + escapar(actual.etiqueta) + '</strong></span>' +
      '<button type="button" class="ubic-selector__btn" data-cambiar>Cambiar ubicación</button>';

    return {
      raiz:   ancla,
      titulo: ancla.querySelector('[data-titulo]'),
      boton:  ancla.querySelector('[data-cambiar]')
    };
  }

  function abrirPanel(ui, actual) {
    cerrarPanel();
    var panel = document.createElement('div');
    panel.className = 'ubic-modal';
    panel.setAttribute('role', 'dialog');
    panel.setAttribute('aria-modal', 'true');
    panel.setAttribute('aria-label', 'Cambiar ubicación');
    panel.innerHTML = (
      '<div class="ubic-modal__card">' +
        '<header class="ubic-modal__header">' +
          '<h2>Cambiar ubicación</h2>' +
          '<button type="button" class="ubic-modal__close" data-cerrar aria-label="Cerrar">×</button>' +
        '</header>' +
        '<button type="button" class="ubic-modal__action" data-gps>' +
          '<span class="ubic-modal__action-ico" aria-hidden="true">📍</span>' +
          '<span><strong>Usar mi ubicación</strong><small>Permite acceso a tu posición</small></span>' +
        '</button>' +
        '<form class="ubic-modal__form" data-form-ciudad autocomplete="off">' +
          '<label class="ubic-modal__label">Introduce tu ciudad o código postal</label>' +
          '<div class="ubic-modal__row">' +
            '<input type="text" name="ciudad" placeholder="Ej. Zaragoza o 28013" required>' +
            '<button type="submit" class="ubic-modal__submit">Buscar</button>' +
          '</div>' +
          '<small class="ubic-modal__hint">Aceptamos nombre de ciudad o código postal español.</small>' +
        '</form>' +
        '<div class="ubic-modal__quick" data-presets></div>' +
        '<p class="ubic-modal__error" data-error hidden></p>' +
      '</div>'
    );
    document.body.appendChild(panel);
    document.body.classList.add('snow-no-scroll');

    var presets = panel.querySelector('[data-presets]');
    var ciudades = (window.SnowbreakGeo && SnowbreakGeo.CIUDADES_PRESET) || [];
    presets.innerHTML = ciudades.map(function (c) {
      return '<button type="button" class="ubic-modal__chip" data-preset=\'' + JSON.stringify(c).replace(/'/g, '&apos;') + '\'>' + escapar(c.etiqueta) + '</button>';
    }).join('');

    function cerrar() { cerrarPanel(); }

    panel.addEventListener('click', function (ev) {
      if (ev.target === panel) { cerrar(); }
    });
    panel.querySelector('[data-cerrar]').addEventListener('click', cerrar);

    panel.querySelector('[data-gps]').addEventListener('click', function () {
      mostrarError(panel, 'Buscando tu ubicación…');
      if (!window.SnowbreakGeo) {
        mostrarError(panel, 'No hemos podido acceder a tu ubicación. Puedes introducirla manualmente.');
        return;
      }
      SnowbreakGeo.obtener({ timeout: 8000 })
        .then(function (ubic) {
          aplicar(ubic);
          cerrar();
        })
        .catch(function () {
          mostrarError(panel, 'No hemos podido acceder a tu ubicación. Puedes introducirla manualmente.');
        });
    });

    panel.querySelector('[data-form-ciudad]').addEventListener('submit', function (ev) {
      ev.preventDefault();
      var input = ev.target.querySelector('input[name="ciudad"]');
      var v = input.value.trim();
      if (!v) { return; }
      buscarCiudadOPostal(v)
        .then(function (ubic) {
          aplicar(ubic);
          cerrar();
        })
        .catch(function () {
          mostrarError(panel, 'No hemos encontrado “' + v + '”. Prueba con otro término.');
        });
    });

    presets.addEventListener('click', function (ev) {
      var btn = ev.target.closest('[data-preset]');
      if (!btn) { return; }
      try {
        var raw = btn.getAttribute('data-preset').replace(/&apos;/g, "'");
        var p = JSON.parse(raw);
        aplicar({ lat: p.lat, lng: p.lng, etiqueta: p.etiqueta, fuente: 'preset' });
        cerrar();
      } catch (e) {}
    });

    document.addEventListener('keydown', escListener);
    function escListener(ev) {
      if (ev.key === 'Escape') { cerrar(); }
    }

    function cerrarPanelLocal() {
      document.removeEventListener('keydown', escListener);
    }
    panel._cleanup = cerrarPanelLocal;
  }

  function cerrarPanel() {
    var panel = document.querySelector('.ubic-modal');
    if (!panel) { return; }
    if (panel._cleanup) { panel._cleanup(); }
    panel.remove();
    document.body.classList.remove('snow-no-scroll');
  }

  /** Busca ciudad o código postal. Si parece postal (5 dígitos), añade ", España". */
  function buscarCiudadOPostal(valor) {
    if (!window.SnowbreakGeo) {
      return Promise.reject(new Error('sin geo'));
    }
    var v = valor.trim();
    var esPostal = /^\d{4,5}$/.test(v);
    var consulta = esPostal ? (v + ', España') : v;
    return SnowbreakGeo.desdeCiudad(consulta);
  }

  function aplicar(ubic) {
    if (!ubic || typeof ubic.lat !== 'number') { return; }
    if (window.SnowbreakGeo && SnowbreakGeo.guardarPreferencia) {
      SnowbreakGeo.guardarPreferencia(ubic);
    }
    publicar(ubic);
  }

  function publicar(ubic) {
    document.dispatchEvent(new CustomEvent('snowbreak:ubicacion:cambio', { detail: ubic }));
  }

  function mostrarError(panel, msg) {
    var p = panel.querySelector('[data-error]');
    if (!p) { return; }
    p.textContent = msg;
    p.hidden = false;
  }

  function escapar(s) {
    return String(s == null ? '' : s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
