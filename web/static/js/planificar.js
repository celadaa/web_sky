/* ============================================================
   planificar.js — wizard de /planificar-estancia

   - Navega entre los 5 paneles (4 pasos + éxito).
   - Llama window.SBTrip.set(...) cada vez que el usuario elige
     algo para que el ticket de la derecha se actualice.
   - Mocks de alojamiento generados client-side a partir de la
     estación elegida (no hace falta backend nuevo todavía).

   Si trip-planner-ticket.js no carga, window.SBTrip no existe
   y este script degrada silenciosamente sin romper el wizard.
   ============================================================ */
(function () {
  'use strict';

  var $ = function (sel, root) { return (root || document).querySelector(sel); };
  var $$ = function (sel, root) { return (root || document).querySelectorAll(sel); };

  var wizard = $('#wizard');
  if (!wizard) return;

  var current = 1;
  var maxStep = 5; // 1..4 + éxito

  // ---------- Mocks de alojamiento ----------
  function lodgingsFor(station) {
    if (!station) return [];
    var base = Math.max(60, Math.round((station.pricePerDayNino || 30) * 2.2));
    return [
      {
        id: 'apt-' + station.id,
        name: 'Apartamento ' + station.name,
        sub: '2 habitaciones · cocina · WiFi',
        pricePerNight: base
      },
      {
        id: 'hotel-' + station.id,
        name: 'Hotel cerca de pistas',
        sub: 'Desayuno incluido · spa · 4★',
        pricePerNight: Math.round(base * 1.6)
      },
      {
        id: 'cabin-' + station.id,
        name: 'Cabaña de montaña',
        sub: 'Chimenea · vistas · acceso 4×4',
        pricePerNight: Math.round(base * 2.4)
      }
    ];
  }

  function formatEuro(n) {
    return (Math.round(n * 100) / 100).toFixed(2).replace('.', ',') + '€';
  }

  // ---------- Activación de paneles ----------
  var panels = $$('.wizard__panel');
  var pills  = $$('.wizard__steps li[data-step-pill]');

  function showStep(n) {
    if (n < 1) n = 1;
    if (n > maxStep) n = maxStep;
    current = n;
    for (var i = 0; i < panels.length; i++) {
      var step = parseInt(panels[i].getAttribute('data-step'), 10);
      var on = (step === n);
      panels[i].hidden = !on;
      panels[i].classList.toggle('is-active', on);
    }
    for (var j = 0; j < pills.length; j++) {
      var pillStep = parseInt(pills[j].getAttribute('data-step-pill'), 10);
      pills[j].classList.toggle('is-active', pillStep === n);
      pills[j].classList.toggle('is-done', pillStep < n);
    }
    var prevBtn = $('#wizard-prev');
    var nextBtn = $('#wizard-next');
    if (prevBtn) prevBtn.hidden = (n === 1);
    if (nextBtn) {
      nextBtn.hidden = (n === maxStep);
      nextBtn.disabled = !canAdvance(n);
      nextBtn.textContent = (n === 4) ? 'Finalizar ✓' : 'Continuar →';
    }
  }

  function canAdvance(step) {
    var s = (window.SBTrip && window.SBTrip.get) ? window.SBTrip.get() : null;
    if (!s) return true; // si no hay storage, permitimos avanzar
    switch (step) {
      case 1: return !!s.station;
      case 2: return !!s.dates && s.dates.nights > 0;
      case 3: return !!s.lodging; // material es opcional
      case 4: return true; // forfait/extras opcionales
      default: return true;
    }
  }

  // ---------- PASO 1: estaciones ----------
  var stationItems = $$('#wizard-stations .wizard__station');
  var searchInput  = $('#wizard-search');

  function selectStation(item) {
    var data = {
      id:               parseInt(item.getAttribute('data-station-id'), 10),
      name:             item.getAttribute('data-station-name'),
      ubic:             item.getAttribute('data-station-ubic'),
      pricePerDayNino:  parseFloat(item.getAttribute('data-station-price'))
    };
    for (var i = 0; i < stationItems.length; i++) {
      stationItems[i].classList.toggle('is-selected', stationItems[i] === item);
    }
    if (window.SBTrip) {
      window.SBTrip.set({ station: data });
      // Reset alojamiento (cambia el catálogo)
      window.SBTrip.set({ lodging: null });
    }
    populateLodgings(data);
    refreshForfaitUI();
    refreshNext();
  }

  for (var si = 0; si < stationItems.length; si++) {
    (function (item) {
      item.addEventListener('click', function () { selectStation(item); });
      item.addEventListener('keydown', function (e) {
        if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); selectStation(item); }
      });
    })(stationItems[si]);
  }

  if (searchInput) {
    searchInput.addEventListener('input', function () {
      var q = searchInput.value.trim().toLowerCase();
      var visible = 0;
      for (var i = 0; i < stationItems.length; i++) {
        var name = (stationItems[i].getAttribute('data-station-name') || '').toLowerCase();
        var ubic = (stationItems[i].getAttribute('data-station-ubic') || '').toLowerCase();
        var match = !q || name.indexOf(q) !== -1 || ubic.indexOf(q) !== -1;
        stationItems[i].hidden = !match;
        if (match) visible++;
      }
      var empty = $('[data-empty]');
      if (empty) empty.hidden = visible > 0;
    });
  }

  // ---------- PASO 2: fechas ----------
  var checkin  = $('#wizard-checkin');
  var checkout = $('#wizard-checkout');
  var guests   = $('#wizard-guests');
  var datesErr = $('[data-dates-error]');

  function readDates() {
    if (!checkin || !checkout) return null;
    var ci = checkin.value;
    var co = checkout.value;
    if (!ci || !co) return null;
    var a = new Date(ci + 'T00:00:00');
    var b = new Date(co + 'T00:00:00');
    if (isNaN(a) || isNaN(b)) return null;
    var nights = Math.round((b - a) / 86400000);
    if (nights <= 0) {
      if (datesErr) datesErr.textContent = 'La salida debe ser posterior a la entrada.';
      return null;
    }
    if (datesErr) datesErr.textContent = '';
    return { checkin: ci, checkout: co, nights: nights };
  }
  function readGuests() {
    if (!guests) return 1;
    var n = parseInt(guests.value, 10);
    return n > 0 ? n : 1;
  }

  function persistDates() {
    var d = readDates();
    var g = readGuests();
    if (window.SBTrip) {
      window.SBTrip.set({ dates: d, guests: g });
    }
    refreshNext();
  }
  if (checkin)  checkin.addEventListener('change', persistDates);
  if (checkout) checkout.addEventListener('change', persistDates);
  if (guests)   guests.addEventListener('input', persistDates);

  // ---------- PASO 3: alojamiento + material ----------
  var lodgingsList = $('#wizard-lodgings');
  function populateLodgings(station) {
    if (!lodgingsList) return;
    lodgingsList.innerHTML = '';
    var arr = lodgingsFor(station);
    arr.forEach(function (l) {
      var li = document.createElement('li');
      li.className = 'wizard__lodging page-planificar__lodging';
      li.setAttribute('role', 'option');
      li.setAttribute('tabindex', '0');
      li.setAttribute('data-lodging-id', l.id);
      li.setAttribute('data-lodging-name', l.name);
      li.setAttribute('data-lodging-price', String(l.pricePerNight));
      li.innerHTML =
        '<div class="page-planificar__lodging-body">' +
          '<strong>' + escapeHtml(l.name) + '</strong>' +
          '<span>' + escapeHtml(l.sub) + '</span>' +
        '</div>' +
        '<div class="page-planificar__lodging-price">' +
          '<small>desde</small>' +
          '<strong>' + formatEuro(l.pricePerNight) + '<small>/noche</small></strong>' +
        '</div>';
      li.addEventListener('click', function () { selectLodging(li); });
      li.addEventListener('keydown', function (e) {
        if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); selectLodging(li); }
      });
      lodgingsList.appendChild(li);
    });
  }
  function selectLodging(li) {
    var siblings = lodgingsList.querySelectorAll('.page-planificar__lodging');
    for (var i = 0; i < siblings.length; i++) siblings[i].classList.toggle('is-selected', siblings[i] === li);
    var data = {
      id: li.getAttribute('data-lodging-id'),
      name: li.getAttribute('data-lodging-name'),
      pricePerNight: parseFloat(li.getAttribute('data-lodging-price'))
    };
    if (window.SBTrip) window.SBTrip.set({ lodging: data });
    refreshNext();
  }

  // Material
  var materialItems = $$('#wizard-materials .page-planificar__material');
  function selectMaterial(item) {
    for (var i = 0; i < materialItems.length; i++) materialItems[i].classList.toggle('is-selected', materialItems[i] === item);
    var id    = item.getAttribute('data-material-id');
    var name  = item.getAttribute('data-material-name');
    var price = parseFloat(item.getAttribute('data-material-price')) || 0;
    if (window.SBTrip) {
      window.SBTrip.set({ material: id === 'none' ? { id: 'none', name: 'Sin material', pricePerDay: 0 }
                                                  : { id: id, name: name, pricePerDay: price } });
    }
  }
  for (var mi = 0; mi < materialItems.length; mi++) {
    (function (item) {
      item.addEventListener('click', function () { selectMaterial(item); });
      item.addEventListener('keydown', function (e) {
        if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); selectMaterial(item); }
      });
    })(materialItems[mi]);
  }

  // ---------- PASO 4: forfait + extras ----------
  var forfaitRadios = $$('#wizard-forfait input[name="forfait-tipo"]');
  var forfaitQty    = $('#wizard-forfait-qty');
  var forfaitStat   = $('[data-forfait-station]');

  function refreshForfaitUI() {
    var s = window.SBTrip && window.SBTrip.get();
    if (forfaitStat) {
      forfaitStat.textContent = (s && s.station) ? s.station.name : 'Selecciona una estación primero.';
    }
    // Precios sintéticos relativos a precioNino — adulto 1.2x, senior 0.9x
    var base = (s && s.station) ? s.station.pricePerDayNino : 0;
    var pAdulto = base ? formatEuro(base * 1.2) : '—';
    var pNino   = base ? formatEuro(base)       : '—';
    var pSenior = base ? formatEuro(base * 0.9) : '—';
    var byType = { adulto: pAdulto, nino: pNino, senior: pSenior };
    var labels = $$('[data-price-adulto]'); for (var i = 0; i < labels.length; i++) labels[i].textContent = pAdulto;
    labels = $$('[data-price-nino]');       for (i = 0; i < labels.length; i++) labels[i].textContent   = pNino;
    labels = $$('[data-price-senior]');     for (i = 0; i < labels.length; i++) labels[i].textContent   = pSenior;
  }

  function persistForfait() {
    var s = window.SBTrip && window.SBTrip.get();
    var base = (s && s.station) ? s.station.pricePerDayNino : 0;
    var tipo = 'ninguno';
    for (var i = 0; i < forfaitRadios.length; i++) if (forfaitRadios[i].checked) tipo = forfaitRadios[i].value;
    var price = 0;
    if (tipo === 'adulto') price = base * 1.2;
    if (tipo === 'nino')   price = base;
    if (tipo === 'senior') price = base * 0.9;
    var qty = parseInt(forfaitQty ? forfaitQty.value : '1', 10) || 0;
    if (window.SBTrip) window.SBTrip.set({ forfait: { tipo: tipo, price: price, qty: qty } });
  }
  for (var fr = 0; fr < forfaitRadios.length; fr++) forfaitRadios[fr].addEventListener('change', persistForfait);
  if (forfaitQty) forfaitQty.addEventListener('input', persistForfait);

  // Extras
  var extrasInputs = $$('#wizard-extras input[type="checkbox"]');
  function persistExtras() {
    var arr = [];
    for (var i = 0; i < extrasInputs.length; i++) {
      if (extrasInputs[i].checked) {
        arr.push({
          id:    extrasInputs[i].getAttribute('data-extra-id'),
          name:  extrasInputs[i].getAttribute('data-extra-name'),
          price: parseFloat(extrasInputs[i].getAttribute('data-extra-price')) || 0,
          daily: extrasInputs[i].getAttribute('data-extra-id') !== 'wax'
        });
      }
    }
    if (window.SBTrip) window.SBTrip.set({ extras: arr });
  }
  for (var ei = 0; ei < extrasInputs.length; ei++) extrasInputs[ei].addEventListener('change', persistExtras);

  // ---------- Navegación ----------
  function refreshNext() {
    var nextBtn = $('#wizard-next');
    if (nextBtn) nextBtn.disabled = !canAdvance(current);
  }
  var prevBtn = $('#wizard-prev');
  var nextBtn = $('#wizard-next');
  if (prevBtn) prevBtn.addEventListener('click', function () { showStep(current - 1); });
  if (nextBtn) nextBtn.addEventListener('click', function () {
    if (!canAdvance(current)) return;
    if (current === 4) persistForfait();
    showStep(current + 1);
  });

  // ---------- Restaurar selección visual desde state ----------
  function hydrateFromState() {
    if (!window.SBTrip) return;
    var s = window.SBTrip.get();
    if (s.station) {
      for (var i = 0; i < stationItems.length; i++) {
        if (parseInt(stationItems[i].getAttribute('data-station-id'), 10) === s.station.id) {
          stationItems[i].classList.add('is-selected');
        }
      }
      populateLodgings(s.station);
    }
    if (s.dates) {
      if (checkin)  checkin.value  = s.dates.checkin || '';
      if (checkout) checkout.value = s.dates.checkout || '';
    }
    if (s.guests && guests) guests.value = s.guests;
    if (s.lodging && lodgingsList) {
      var lis = lodgingsList.querySelectorAll('[data-lodging-id="' + s.lodging.id + '"]');
      if (lis.length) lis[0].classList.add('is-selected');
    }
    if (s.material) {
      for (i = 0; i < materialItems.length; i++) {
        if (materialItems[i].getAttribute('data-material-id') === s.material.id) {
          materialItems[i].classList.add('is-selected');
        }
      }
    }
    if (s.forfait) {
      for (i = 0; i < forfaitRadios.length; i++) {
        forfaitRadios[i].checked = (forfaitRadios[i].value === s.forfait.tipo);
      }
      if (forfaitQty && typeof s.forfait.qty === 'number') forfaitQty.value = s.forfait.qty;
    }
    if (s.extras && s.extras.length) {
      var ids = s.extras.map(function (e) { return e.id; });
      for (i = 0; i < extrasInputs.length; i++) {
        extrasInputs[i].checked = ids.indexOf(extrasInputs[i].getAttribute('data-extra-id')) !== -1;
      }
    }
    refreshForfaitUI();
  }

  // ---------- Helpers ----------
  function escapeHtml(s) {
    return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  // ---------- Boot ----------
  function init() {
    hydrateFromState();
    showStep(1);
    refreshForfaitUI();
    // Cuando el ticket cambia, refresca el botón "Continuar"
    if (window.SBTrip && window.SBTrip.subscribe) {
      window.SBTrip.subscribe(function () { refreshNext(); });
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
