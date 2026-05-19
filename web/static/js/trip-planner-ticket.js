/* ============================================================
   trip-planner-ticket.js
   Estado del viaje + render del ticket de /planificar-estancia.

   - Storage:   localStorage["snowbreak_trip_planner_v1"]
   - API:       window.SBTrip (set, get, reset, render)
   - Listeners: el wizard (planificar.js) llama SBTrip.set(parche)
                cada vez que el usuario elige algo.

   REGLA: el ticket es visible por defecto (HTML inicial muestra
   "Estación pendiente", "Fechas pendientes", etc.). Si este JS
   falla, no pasa nada — el ticket sigue legible.
   ============================================================ */
(function () {
  'use strict';

  // Clave del navegador, no es un secreto. gitleaks:allow
  var STORAGE_KEY = 'snowbreak_trip_planner_v1'; // gitleaks:allow
  var IVA_RATIO = 0.21;

  // ---------- Estado inicial ----------
  var defaultState = {
    station:  null,   // { id, name, ubic, pricePerDayNino }
    dates:    null,   // { checkin: 'YYYY-MM-DD', checkout, nights }
    guests:   null,   // entero
    lodging:  null,   // { id, name, pricePerNight }
    material: null,   // { id, name, pricePerDay }
    forfait:  null,   // { tipo, price, qty }
    extras:   []      // [{ id, name, price, daily }]
  };

  function loadState() {
    try {
      var raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return clone(defaultState);
      var parsed = JSON.parse(raw);
      // Merge defensivo con defaults
      return Object.assign(clone(defaultState), parsed);
    } catch (e) {
      return clone(defaultState);
    }
  }

  function saveState() {
    try { localStorage.setItem(STORAGE_KEY, JSON.stringify(state)); }
    catch (e) { /* almacenamiento lleno o privado */ }
  }

  function clone(obj) { return JSON.parse(JSON.stringify(obj)); }

  var state = loadState();

  // ---------- Helpers ----------
  function nightsBetween(checkin, checkout) {
    if (!checkin || !checkout) return 0;
    var a = new Date(checkin + 'T00:00:00');
    var b = new Date(checkout + 'T00:00:00');
    var diff = Math.round((b - a) / 86400000);
    return diff > 0 ? diff : 0;
  }

  function formatEuro(n) {
    if (typeof n !== 'number' || isNaN(n)) return '0,00€';
    return n.toFixed(2).replace('.', ',') + '€';
  }

  function formatDateRange(checkin, checkout) {
    if (!checkin || !checkout) return 'Fechas pendientes';
    var fmt = function (s) {
      var d = new Date(s + 'T00:00:00');
      if (isNaN(d.getTime())) return s;
      return d.toLocaleDateString('es-ES', { day: 'numeric', month: 'short' });
    };
    return fmt(checkin) + ' — ' + fmt(checkout);
  }

  // ---------- Cálculos ----------
  function getNights() {
    if (!state.dates) return 0;
    return state.dates.nights || 0;
  }

  function lodgingTotal() {
    if (!state.lodging || !getNights()) return 0;
    return state.lodging.pricePerNight * getNights() * (state.guests || 1);
  }

  function materialTotal() {
    if (!state.material || !getNights()) return 0;
    if (state.material.pricePerDay === 0) return 0;
    return state.material.pricePerDay * getNights() * (state.guests || 1);
  }

  function forfaitTotal() {
    if (!state.forfait) return 0;
    if (state.forfait.tipo === 'ninguno') return 0;
    var nights = getNights() || 1;
    return state.forfait.price * (state.forfait.qty || 1) * nights;
  }

  function extrasTotal() {
    if (!state.extras || !state.extras.length) return 0;
    var nights = getNights() || 1;
    var sum = 0;
    for (var i = 0; i < state.extras.length; i++) {
      var e = state.extras[i];
      sum += e.daily ? e.price * nights : e.price;
    }
    return sum;
  }

  function subtotal() {
    return lodgingTotal() + materialTotal() + forfaitTotal() + extrasTotal();
  }
  function iva()      { return subtotal() * IVA_RATIO; }
  function total()    { return subtotal() + iva(); }

  // ---------- Código de barras pseudoaleatorio estable ----------
  function barcodeFromSeed(seed) {
    var s = String(seed || 'SB000');
    var bg = 'repeating-linear-gradient(to right';
    var pos = 0;
    for (var i = 0; i < 60; i++) {
      var w = ((s.charCodeAt(i % s.length) + i * 13) % 4) + 1;
      var color = i % 2 === 0 ? '#0a0a0a' : 'transparent';
      bg += ', ' + color + ' ' + pos + 'px ' + (pos + w) + 'px';
      pos += w;
    }
    bg += ')';
    return bg;
  }
  function barcodeNumFromState() {
    var parts = [];
    parts.push('SB');
    parts.push(state.station ? pad4(state.station.id || 0) : '0000');
    parts.push(state.dates && state.dates.nights ? pad2(state.dates.nights) : '00');
    parts.push(state.guests ? pad2(state.guests) : '00');
    parts.push(state.lodging ? pad3(Math.round(state.lodging.pricePerNight)) : '000');
    return parts.join(' ');
  }
  function pad2(n) { n = parseInt(n, 10) || 0; return n < 10 ? '0' + n : '' + n; }
  function pad3(n) { n = parseInt(n, 10) || 0; return ('000' + n).slice(-3); }
  function pad4(n) { n = parseInt(n, 10) || 0; return ('0000' + n).slice(-4); }

  // ---------- Render ----------
  function $$(sel) { return document.querySelectorAll(sel); }
  function setText(sel, text) {
    var els = $$(sel);
    for (var i = 0; i < els.length; i++) els[i].textContent = text;
  }

  function render() {
    var ticket = document.querySelector('[data-trip-ticket]');
    if (!ticket) return;

    // Fecha del ticket
    var d = new Date();
    var pad = function (n) { return n < 10 ? '0' + n : '' + n; };
    setText('[data-trip-date]',
      pad(d.getDate()) + '/' + pad(d.getMonth() + 1) + '/' + d.getFullYear() +
      ' · ' + pad(d.getHours()) + ':' + pad(d.getMinutes()));

    // ESTACIÓN
    if (state.station) {
      setText('[data-trip-station]', state.station.name);
      setText('[data-trip-station-meta]', state.station.ubic || 'Incluido');
    } else {
      setText('[data-trip-station]', 'Estación pendiente');
      setText('[data-trip-station-meta]', '—');
    }

    // FECHAS
    if (state.dates && state.dates.checkin && state.dates.checkout) {
      setText('[data-trip-dates]', formatDateRange(state.dates.checkin, state.dates.checkout));
      setText('[data-trip-nights]', getNights() + ' ' + (getNights() === 1 ? 'noche' : 'noches'));
    } else {
      setText('[data-trip-dates]', 'Fechas pendientes');
      setText('[data-trip-nights]', '—');
    }

    // HUÉSPEDES
    setText('[data-trip-guests]',
      state.guests ? state.guests + ' ' + (state.guests === 1 ? 'persona' : 'personas') : '—');

    // ALOJAMIENTO
    if (state.lodging) {
      setText('[data-trip-lodging]', state.lodging.name);
      setText('[data-trip-lodging-price]', formatEuro(lodgingTotal()));
    } else {
      setText('[data-trip-lodging]', 'Alojamiento pendiente');
      setText('[data-trip-lodging-price]', '—');
    }

    // MATERIAL
    if (state.material && state.material.id) {
      var matLabel = state.material.name;
      setText('[data-trip-material]', matLabel);
      setText('[data-trip-material-price]', materialTotal() ? formatEuro(materialTotal()) : '0,00€');
    } else {
      setText('[data-trip-material]', 'Material pendiente');
      setText('[data-trip-material-price]', '—');
    }

    // FORFAIT
    if (state.forfait && state.forfait.tipo !== 'ninguno' && state.forfait.qty > 0) {
      var label = 'Forfait ' + state.forfait.tipo + ' × ' + state.forfait.qty;
      setText('[data-trip-forfait]', label);
      setText('[data-trip-forfait-price]', formatEuro(forfaitTotal()));
    } else {
      setText('[data-trip-forfait]', 'Sin forfait');
      setText('[data-trip-forfait-price]', '0,00€');
    }

    // EXTRAS
    if (state.extras && state.extras.length) {
      var names = state.extras.map(function (e) { return e.name; }).join(' + ');
      setText('[data-trip-extras]', names);
      setText('[data-trip-extras-price]', formatEuro(extrasTotal()));
    } else {
      setText('[data-trip-extras]', 'Ninguno');
      setText('[data-trip-extras-price]', '0,00€');
    }

    // TOTALES
    setText('[data-trip-subtotal]', formatEuro(subtotal()));
    setText('[data-trip-iva]',      formatEuro(iva()));
    setText('[data-trip-total]',    formatEuro(total()));

    // Barcode
    var bars = ticket.querySelector('[data-trip-barcode]');
    if (bars) bars.style.setProperty('--barcode', barcodeFromSeed(barcodeNumFromState()));
    setText('[data-trip-barcode-num]', barcodeNumFromState());

    // Estado completado: clase utilitaria por si se quiere estilar el CTA
    var done = state.station && state.dates && state.lodging;
    ticket.classList.toggle('is-complete', !!done);
  }

  // ---------- API ----------
  function set(patch) {
    if (!patch || typeof patch !== 'object') return;
    // Merge superficial campo a campo
    Object.keys(patch).forEach(function (k) {
      state[k] = patch[k];
      // Cuando cambian las fechas, recalculamos noches automáticamente
      if (k === 'dates' && patch.dates) {
        state.dates.nights = nightsBetween(patch.dates.checkin, patch.dates.checkout);
      }
    });
    saveState();
    render();
    notify();
  }

  function reset() {
    state = clone(defaultState);
    try { localStorage.removeItem(STORAGE_KEY); } catch (e) {}
    render();
    notify();
  }

  // ---------- Suscriptores externos (planificar.js) ----------
  var subs = [];
  function subscribe(fn) { subs.push(fn); }
  function notify() {
    for (var i = 0; i < subs.length; i++) {
      try { subs[i](state); } catch (e) {}
    }
  }

  // ---------- Botón reset + CTA ----------
  function bindUI() {
    document.addEventListener('click', function (e) {
      var t = e.target;
      if (!t) return;
      if (t.matches && t.matches('[data-trip-reset]')) {
        if (window.confirm('¿Reiniciar la planificación? Se perderá el ticket actual.')) {
          reset();
        }
      }
    });
  }

  // ---------- Boot ----------
  function boot() {
    bindUI();
    render();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', boot);
  } else {
    boot();
  }

  // API pública
  window.SBTrip = {
    set: set,
    reset: reset,
    get: function () { return clone(state); },
    subscribe: subscribe,
    render: render
  };
})();
