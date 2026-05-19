/* ============================================================
   cart.js (premium) — Carrito del Snowbreak Market
   Independiente del cart.js de Snowbreak (forfaits) para no
   chocar con la funcionalidad existente. localStorage key
   diferente: "snowbreak_market_v1".
   ============================================================ */
(function () {
  'use strict';

  var STORAGE_KEY = 'snowbreak_market_v1';
  var IVA = 0.21;

  // ---------- Persistencia ----------
  function load() {
    try {
      var raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return [];
      var arr = JSON.parse(raw);
      return Array.isArray(arr) ? arr : [];
    } catch (e) { return []; }
  }

  function save(items) {
    try { localStorage.setItem(STORAGE_KEY, JSON.stringify(items)); }
    catch (e) { /* almacenamiento lleno o privado */ }
  }

  // ---------- Estado ----------
  var items = load();

  // ---------- Helpers ----------
  function formatEuro(n) {
    return n.toFixed(2).replace('.', ',') + ' €';
  }

  function escapeHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function subtotal() {
    return items.reduce(function (sum, it) {
      return sum + (it.price * it.qty);
    }, 0);
  }

  function iva() { return subtotal() * IVA; }
  function total() { return subtotal() + iva(); }

  function findIndex(id) {
    for (var i = 0; i < items.length; i++) {
      if (items[i].id === id) return i;
    }
    return -1;
  }

  // ---------- Mutaciones ----------
  function add(product) {
    if (!product || !product.id) return;
    var idx = findIndex(product.id);
    if (idx >= 0) {
      items[idx].qty += 1;
    } else {
      items.push({
        id: product.id,
        name: product.name,
        price: Number(product.price) || 0,
        lot: product.lot || '',
        qty: 1
      });
    }
    save(items);
    render();
  }

  function remove(id) {
    var idx = findIndex(id);
    if (idx < 0) return;
    items.splice(idx, 1);
    save(items);
    render();
  }

  function clear() {
    items = [];
    save(items);
    render();
  }

  // ---------- Barcode aleatorio dinámico ----------
  function makeBarcodePattern() {
    var ink = getInkColor();
    var widths = [];
    for (var i = 0; i < 60; i++) {
      widths.push((Math.floor(Math.random() * 4) + 1));
    }
    var bg = 'repeating-linear-gradient(to right';
    var pos = 0;
    widths.forEach(function (w, idx) {
      var color = idx % 2 === 0 ? ink : 'transparent';
      bg += ', ' + color + ' ' + pos + 'px ' + (pos + w) + 'px';
      pos += w;
    });
    bg += ')';
    return bg;
  }
  function getInkColor() {
    try {
      var c = getComputedStyle(document.documentElement).getPropertyValue('--sb-ink').trim();
      return c || '#0a0a0a';
    } catch (e) { return '#0a0a0a'; }
  }
  function randomBarcodeNum() {
    var s = '';
    for (var i = 0; i < 13; i++) {
      if (i === 1 || i === 4 || i === 8) s += ' ';
      s += Math.floor(Math.random() * 10);
    }
    return 'SB ' + s;
  }

  // ---------- Render del ticket ----------
  function render() {
    var ticket = document.querySelector('[data-ticket]');
    if (!ticket) return;

    var list = ticket.querySelector('[data-ticket-items]');
    var empty = ticket.querySelector('[data-ticket-empty]');
    var totals = ticket.querySelector('[data-ticket-totals]');
    var subEl = ticket.querySelector('[data-ticket-subtotal]');
    var ivaEl = ticket.querySelector('[data-ticket-iva]');
    var totalEl = ticket.querySelector('[data-ticket-total]');
    var dateEl = ticket.querySelector('[data-ticket-date]');
    var barEl = ticket.querySelector('[data-ticket-barcode]');
    var barNumEl = ticket.querySelector('[data-ticket-barcode-num]');

    // Fecha tipo ticket
    if (dateEl) {
      var d = new Date();
      var pad = function (n) { return n < 10 ? '0' + n : '' + n; };
      var stamp = pad(d.getDate()) + '/' + pad(d.getMonth() + 1) + '/' + d.getFullYear() +
                  ' · ' + pad(d.getHours()) + ':' + pad(d.getMinutes());
      dateEl.textContent = stamp;
    }

    // Líneas
    list.innerHTML = '';
    if (items.length === 0) {
      empty.hidden = false;
      totals.hidden = true;
    } else {
      empty.hidden = true;
      totals.hidden = false;
      items.forEach(function (it) {
        var li = document.createElement('li');
        li.className = 'sb-ticket__item';
        li.innerHTML =
          '<span class="sb-ticket__item-name">' + escapeHtml(it.name) + '</span>' +
          '<span class="sb-ticket__item-qty">x' + it.qty + ' · ' + formatEuro(it.price) + '</span>' +
          '<span class="sb-ticket__item-price">' + formatEuro(it.price * it.qty) + '</span>' +
          '<button type="button" class="sb-ticket__item-remove" data-remove="' + escapeHtml(it.id) + '">eliminar</button>';
        list.appendChild(li);
      });

      subEl.textContent = formatEuro(subtotal());
      ivaEl.textContent = formatEuro(iva());
      totalEl.textContent = formatEuro(total());
    }

    // Barcode dinámico
    if (barEl) barEl.style.setProperty('--barcode', makeBarcodePattern());
    if (barNumEl) barNumEl.textContent = randomBarcodeNum();
  }

  // ---------- Eventos delegados (eliminar) ----------
  function bindDelegated() {
    document.addEventListener('click', function (e) {
      var t = e.target;
      if (t && t.matches && t.matches('[data-remove]')) {
        remove(t.getAttribute('data-remove'));
      }
    });
  }

  // ---------- Boot ----------
  function init() {
    bindDelegated();
    render();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

  // ---------- API pública ----------
  window.SBCart = {
    add: add,
    remove: remove,
    clear: clear,
    items: function () { return items.slice(); },
    subtotal: subtotal,
    total: total,
    render: render
  };
})();
