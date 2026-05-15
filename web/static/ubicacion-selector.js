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
    panel.setAttribute('aria-labelledby', 'ubic-modal-titulo');
    panel.setAttribute('aria-describedby', 'ubic-modal-sub');
    panel.innerHTML = (
      '<div class="ubic-modal__card">' +
        '<header class="ubic-modal__header">' +
          '<span class="ubic-modal__header-ico" aria-hidden="true">' +
            '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>' +
          '</span>' +
          '<div class="ubic-modal__heading">' +
            '<h2 id="ubic-modal-titulo">Cambiar ubicación</h2>' +
            '<p id="ubic-modal-sub" class="ubic-modal__subtitle">Calculamos distancias a las estaciones desde aquí.</p>' +
          '</div>' +
          '<button type="button" class="ubic-modal__close" data-cerrar aria-label="Cerrar diálogo">×</button>' +
        '</header>' +

        '<section class="ubic-modal__section">' +
          '<button type="button" class="ubic-modal__action" data-gps>' +
            '<span class="ubic-modal__action-ico" aria-hidden="true">' +
              '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="M4.93 4.93l1.41 1.41"/><path d="M17.66 17.66l1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="M6.34 17.66l-1.41 1.41"/><path d="M19.07 4.93l-1.41 1.41"/></svg>' +
            '</span>' +
            '<span><strong>Usar mi ubicación</strong><small>El navegador pedirá permiso</small></span>' +
          '</button>' +
        '</section>' +

        '<hr class="ubic-modal__divider">' +

        '<section class="ubic-modal__section">' +
          '<h3 class="ubic-modal__section-title">Buscar por nombre</h3>' +
          '<form class="ubic-modal__form" data-form-ciudad autocomplete="off" novalidate>' +
            '<label class="ubic-modal__label" for="ubic-input-ciudad">Ciudad o código postal</label>' +
            '<div class="ubic-modal__row">' +
              '<input id="ubic-input-ciudad" type="text" name="ciudad" placeholder="Ej. Zaragoza o 28013" required autocomplete="off" inputmode="search">' +
              '<button type="submit" class="ubic-modal__submit">' +
                '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/></svg>' +
                'Buscar' +
              '</button>' +
            '</div>' +
            '<small class="ubic-modal__hint">Aceptamos nombre de ciudad o código postal español.</small>' +
          '</form>' +
        '</section>' +

        '<hr class="ubic-modal__divider">' +

        '<section class="ubic-modal__section">' +
          '<h3 class="ubic-modal__section-title">Atajos</h3>' +
          '<div class="ubic-modal__quick" data-presets></div>' +
        '</section>' +

        '<p class="ubic-modal__status" data-status hidden role="status" aria-live="polite"></p>' +
        '<p class="ubic-modal__error" data-error hidden role="alert"></p>' +
      '</div>'
    );

    document.body.appendChild(panel);
    document.body.classList.add('snow-no-scroll');

    // Pintamos los presets resaltando el actual.
    var presets = panel.querySelector('[data-presets]');
    var ciudades = (window.SnowbreakGeo && SnowbreakGeo.CIUDADES_PRESET) || [];
    presets.innerHTML = ciudades.map(function (c) {
      var activo = actual && actual.etiqueta === c.etiqueta ? ' ubic-modal__chip--active' : '';
      var aria = activo ? ' aria-current="true"' : '';
      return (
        '<button type="button" class="ubic-modal__chip' + activo + '"' + aria +
        ' data-preset=\'' + JSON.stringify(c).replace(/'/g, '&apos;') + '\'>' +
        escapar(c.etiqueta) + '</button>'
      );
    }).join('');

    function cerrar() { cerrarPanel(); }

    // Cierre al clicar overlay.
    panel.addEventListener('click', function (ev) {
      if (ev.target === panel) { cerrar(); }
    });
    panel.querySelector('[data-cerrar]').addEventListener('click', cerrar);

    panel.querySelector('[data-gps]').addEventListener('click', function () {
      mostrarStatus(panel, 'Buscando tu ubicación…');
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
      if (!v) { input.focus(); return; }
      mostrarStatus(panel, 'Buscando “' + v + '”…');
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

    // Foco automático: el input es lo más útil para empezar.
    var primerInput = panel.querySelector('#ubic-input-ciudad');
    if (primerInput) {
      // setTimeout para esperar a que la animación de entrada no lo cancele
      // en algunos navegadores móviles.
      setTimeout(function () { primerInput.focus(); }, 60);
    }

    // Cierre con Escape + focus trap básico con Tab.
    function teclas(ev) {
      if (ev.key === 'Escape') {
        cerrar();
        return;
      }
      if (ev.key !== 'Tab') { return; }
      var focuseables = panel.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      if (!focuseables.length) { return; }
      var primero = focuseables[0];
      var ultimo = focuseables[focuseables.length - 1];
      if (ev.shiftKey && document.activeElement === primero) {
        ev.preventDefault();
        ultimo.focus();
      } else if (!ev.shiftKey && document.activeElement === ultimo) {
        ev.preventDefault();
        primero.focus();
      }
    }
    document.addEventListener('keydown', teclas);
    panel._cleanup = function () { document.removeEventListener('keydown', teclas); };
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

  function mostrarStatus(panel, msg) {
    var s = panel.querySelector('[data-status]');
    var e = panel.querySelector('[data-error]');
    if (e) { e.hidden = true; e.textContent = ''; }
    if (!s) { return; }
    s.textContent = msg;
    s.hidden = false;
  }

  function mostrarError(panel, msg) {
    var s = panel.querySelector('[data-status]');
    var p = panel.querySelector('[data-error]');
    if (s) { s.hidden = true; s.textContent = ''; }
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
