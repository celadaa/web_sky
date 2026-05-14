/* Snowbreak — Refresco del parte de nieve.
 *
 * Engancha al botón [data-parte-refresh] dentro de [data-parte] y, al
 * pulsarlo, hace fetch a /api/estacion/{id}/parte para obtener los
 * valores recalculados y actualizar la UI sin recargar la página.
 *
 * Maneja los tres estados profesionales:
 *  - cargando: muestra el loader y deshabilita el botón.
 *  - éxito: actualiza valores y muestra toast.
 *  - error: muestra mensaje en el bloque [data-parte-error] y deja la
 *    UI con los valores anteriores intactos (no rompe la página).
 */
(function () {
  'use strict';

  function init() {
    var parte = document.querySelector('[data-parte]');
    if (!parte) { return; }
    var btn = parte.querySelector('[data-parte-refresh]');
    if (!btn) { return; }

    btn.addEventListener('click', function () { refrescar(parte, btn); });
  }

  function refrescar(parte, btn) {
    var id = parte.getAttribute('data-estacion-id');
    if (!id) { return; }

    mostrarCarga(parte, btn, true);
    ocultarError(parte);

    fetch('/api/estacion/' + encodeURIComponent(id) + '/parte', {
      headers: { 'Accept': 'application/json' },
      cache:   'no-cache'
    })
    .then(function (r) {
      if (!r.ok) { throw new Error('HTTP ' + r.status); }
      return r.json();
    })
    .then(function (data) {
      aplicar(parte, data);
      pulso(parte);
      mostrarToast('Parte actualizado a las ' + (data.parte_hora || ''));
    })
    .catch(function (err) {
      console.warn('refresco parte:', err);
      mostrarError(parte);
    })
    .finally(function () {
      mostrarCarga(parte, btn, false);
    });
  }

  function aplicar(parte, d) {
    setText(parte, '[data-parte-estado]',   d.estado);
    setText(parte, '[data-parte-temp]',     formatear(d.temperatura, ' °C'));
    setText(parte, '[data-parte-nieve]',    d.nieve_min + '–' + d.nieve_max + ' cm');
    setText(parte, '[data-parte-nueva]',    '+ ' + d.nieve_nueva + ' cm');
    setText(parte, '[data-parte-viento]',   d.viento || '—');
    setText(parte, '[data-parte-ultima]',   d.ultima_nevada || '—');
    setText(parte, '[data-parte-pistas]',   d.pistas_abiertas + '/' + d.pistas_totales);
    setText(parte, '[data-parte-remontes]', d.remontes_op + '/' + d.remontes_tot);

    var tag = parte.querySelector('[data-parte-tag]');
    if (tag) {
      tag.innerHTML =
        '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg> ' +
        'Parte actualizado a las ' + escapar(d.parte_hora || '');
    }
    setText(parte, '[data-parte-fecha]', 'Última actualización: ' + (d.parte_actualizado || ''));
  }

  function setText(root, sel, txt) {
    var el = root.querySelector(sel);
    if (el) { el.textContent = txt; }
  }

  function formatear(n, suffix) {
    if (n == null || n === '' || (typeof n === 'number' && !isFinite(n))) { return '—'; }
    return String(n) + (suffix || '');
  }

  function mostrarCarga(parte, btn, on) {
    var ld = parte.querySelector('[data-parte-loading]');
    if (ld) { ld.hidden = !on; }
    if (btn) {
      btn.disabled = !!on;
      btn.classList.toggle('is-loading', !!on);
    }
  }

  function mostrarError(parte) {
    var er = parte.querySelector('[data-parte-error]');
    if (er) { er.hidden = false; }
  }
  function ocultarError(parte) {
    var er = parte.querySelector('[data-parte-error]');
    if (er) { er.hidden = true; }
  }

  function pulso(el) {
    el.classList.remove('parte--pulse');
    /* trigger reflow para reiniciar la animación */
    void el.offsetWidth;
    el.classList.add('parte--pulse');
  }

  function mostrarToast(msg) {
    var t = document.getElementById('snow-toast');
    if (!t) { return; }
    t.textContent = msg;
    t.classList.add('snow-toast--visible');
    clearTimeout(t._timer);
    t._timer = setTimeout(function () { t.classList.remove('snow-toast--visible'); }, 2400);
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
