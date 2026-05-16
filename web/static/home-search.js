(function () {
  'use strict';

  /* ── Helpers ───────────────────────────────────────────────── */

  function norm(s) {
    return String(s).toLowerCase()
      .normalize('NFD').replace(/[̀-ͯ]/g, '');
  }

  function escHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;')
      .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  function isoToEs(iso) {
    if (!iso) return '';
    var p = iso.split('-');
    var d = new Date(+p[0], +p[1] - 1, +p[2]);
    return d.toLocaleDateString('es-ES', { day: 'numeric', month: 'short' });
  }

  function todayIso() {
    var d = new Date();
    return d.getFullYear() + '-' +
      String(d.getMonth() + 1).padStart(2, '0') + '-' +
      String(d.getDate()).padStart(2, '0');
  }

  /* ── Station data ──────────────────────────────────────────── */

  var stations = [];
  var dataEl = document.getElementById('hero-estaciones-data');
  if (dataEl) {
    dataEl.querySelectorAll('[data-id]').forEach(function (s) {
      stations.push({
        id:        s.dataset.id,
        nombre:    s.dataset.nombre    || '',
        ubicacion: s.dataset.ubicacion || ''
      });
    });
  }

  /* ══════════════════════════════════════════════════════════
     AUTOCOMPLETE — position:fixed, posicionado con JS
  ══════════════════════════════════════════════════════════ */

  var acInput   = document.getElementById('hero-destino');
  var acList    = document.getElementById('hero-ac-list');
  var acIdInput = document.getElementById('hero-estacion-id');

  if (!acInput || !acList || !acIdInput) return;

  var acActive = -1;
  var acItems  = [];

  function positionAc() {
    var rect = acInput.getBoundingClientRect();
    acList.style.top   = (rect.bottom + 4) + 'px';
    acList.style.left  = rect.left + 'px';
    acList.style.width = Math.max(rect.width, 220) + 'px';
  }

  function acShow(open) {
    acInput.setAttribute('aria-expanded', open ? 'true' : 'false');
    if (open) { positionAc(); acList.hidden = false; }
    else       { acList.hidden = true; }
  }

  function acClear() {
    acItems = []; acActive = -1;
    acList.innerHTML = '';
    acShow(false);
  }

  function acSelect(item) {
    acInput.value   = item.nombre;
    acIdInput.value = item.id;
    acClear();
  }

  function acRender(matches) {
    acList.innerHTML = '';
    acItems = matches; acActive = -1;
    if (!matches.length) { acShow(false); return; }
    matches.forEach(function (st, idx) {
      var li = document.createElement('li');
      li.className = 'home-ac__item';
      li.setAttribute('role', 'option');
      li.setAttribute('aria-selected', 'false');
      li.id = 'hero-ac-opt-' + idx;
      li.innerHTML =
        '<span class="home-ac__nombre">' + escHtml(st.nombre) + '</span>' +
        (st.ubicacion ? '<span class="home-ac__ubi">' + escHtml(st.ubicacion) + '</span>' : '');
      li.addEventListener('mousedown', function (e) { e.preventDefault(); acSelect(st); });
      acList.appendChild(li);
    });
    acShow(true);
  }

  function acHighlight(idx) {
    var items = acList.querySelectorAll('.home-ac__item');
    items.forEach(function (el, i) {
      el.classList.toggle('home-ac__item--active', i === idx);
      el.setAttribute('aria-selected', i === idx ? 'true' : 'false');
    });
    if (idx >= 0 && items[idx]) {
      acInput.setAttribute('aria-activedescendant', 'hero-ac-opt-' + idx);
      items[idx].scrollIntoView({ block: 'nearest' });
    } else {
      acInput.removeAttribute('aria-activedescendant');
    }
  }

  acInput.addEventListener('input', function () {
    acIdInput.value = '';
    var q = norm(acInput.value.trim());
    if (!q) { acClear(); return; }
    acRender(
      stations.filter(function (st) {
        return norm(st.nombre).includes(q) || norm(st.ubicacion).includes(q);
      }).slice(0, 8)
    );
  });

  acInput.addEventListener('keydown', function (e) {
    if (acList.hidden) return;
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      acActive = Math.min(acActive + 1, acItems.length - 1);
      acHighlight(acActive);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      acActive = Math.max(acActive - 1, -1);
      acHighlight(acActive);
    } else if (e.key === 'Enter' && acActive >= 0 && acItems[acActive]) {
      e.preventDefault(); acSelect(acItems[acActive]);
    } else if (e.key === 'Escape') {
      acClear();
    }
  });

  acInput.addEventListener('blur', function () { setTimeout(acClear, 160); });
  window.addEventListener('scroll', function () { if (!acList.hidden) positionAc(); }, { passive: true });
  window.addEventListener('resize', function () { if (!acList.hidden) positionAc(); }, { passive: true });

  /* ══════════════════════════════════════════════════════════
     DATE MODAL — creado dinámicamente en <body>
     Usa dos <input type="date"> dentro de un panel premium.
     Sin stacking context conflicts: está fuera del hero.
  ══════════════════════════════════════════════════════════ */

  var dpTrigger     = document.getElementById('hero-fechas-display');
  var dpHiddenStart = document.getElementById('hero-fecha-inicio');
  var dpHiddenEnd   = document.getElementById('hero-fecha-fin');

  if (!dpTrigger || !dpHiddenStart || !dpHiddenEnd) return;

  /* References created once in buildModal() */
  var hsOverlay  = null;
  var hsModal    = null;
  var hsStart    = null;
  var hsEnd      = null;

  function buildModal() {
    if (hsModal) return;

    /* Overlay */
    hsOverlay = document.createElement('div');
    hsOverlay.className = 'hs-overlay';
    hsOverlay.hidden    = true;
    hsOverlay.addEventListener('click', closeModal);
    document.body.appendChild(hsOverlay);

    /* Modal panel */
    hsModal = document.createElement('div');
    hsModal.className = 'hs-modal';
    hsModal.hidden    = true;
    hsModal.setAttribute('role',       'dialog');
    hsModal.setAttribute('aria-modal', 'true');
    hsModal.setAttribute('aria-label', 'Seleccionar fechas');
    hsModal.innerHTML =
      '<div class="hs-modal__header">' +
        '<div class="hs-modal__title-wrap">' +
          '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>' +
          '<h3 class="hs-modal__title">Selecciona tus fechas</h3>' +
        '</div>' +
        '<button type="button" class="hs-modal__close" aria-label="Cerrar">' +
          '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" aria-hidden="true"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>' +
        '</button>' +
      '</div>' +
      '<div class="hs-modal__body">' +
        '<div class="hs-date-row">' +
          '<div class="hs-date-field">' +
            '<label class="hs-date-label" for="hs-start-input">Fecha de inicio</label>' +
            '<input type="date" id="hs-start-input" class="hs-date-input">' +
          '</div>' +
          '<span class="hs-date-arrow" aria-hidden="true">' +
            '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" stroke-width="1.8" stroke-linecap="round" aria-hidden="true"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>' +
          '</span>' +
          '<div class="hs-date-field">' +
            '<label class="hs-date-label" for="hs-end-input">Fecha de fin</label>' +
            '<input type="date" id="hs-end-input" class="hs-date-input">' +
          '</div>' +
        '</div>' +
        '<p class="hs-date-hint">Elige el rango completo de tu estancia en la estación.</p>' +
      '</div>' +
      '<div class="hs-modal__footer">' +
        '<button type="button" class="hs-btn-clear">Borrar fechas</button>' +
        '<button type="button" class="hs-btn-confirm">Confirmar</button>' +
      '</div>';

    document.body.appendChild(hsModal);

    hsStart = document.getElementById('hs-start-input');
    hsEnd   = document.getElementById('hs-end-input');

    /* min = today */
    var t = todayIso();
    hsStart.min = t;
    hsEnd.min   = t;

    /* when start changes, adjust end min */
    hsStart.addEventListener('change', function () {
      if (hsStart.value) {
        hsEnd.min = hsStart.value;
        if (hsEnd.value && hsEnd.value < hsStart.value) hsEnd.value = '';
      }
    });

    hsModal.querySelector('.hs-modal__close').addEventListener('click', closeModal);

    hsModal.querySelector('.hs-btn-clear').addEventListener('click', function () {
      hsStart.value = '';
      hsEnd.value   = '';
    });

    hsModal.querySelector('.hs-btn-confirm').addEventListener('click', confirmDates);

    /* Escape inside modal */
    hsModal.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') closeModal();
    });
  }

  function openModal() {
    buildModal();
    /* pre-fill with existing values */
    hsStart.value = dpHiddenStart.value || '';
    hsEnd.value   = dpHiddenEnd.value   || '';
    if (hsStart.value) {
      hsEnd.min = hsStart.value;
    }
    hsOverlay.hidden = false;
    hsModal.hidden   = false;
    dpTrigger.setAttribute('aria-expanded', 'true');
    /* trap focus on first relevant input */
    setTimeout(function () {
      var f = hsStart.value ? hsEnd : hsStart;
      f.focus();
    }, 60);
  }

  function closeModal() {
    if (!hsModal) return;
    hsOverlay.hidden = true;
    hsModal.hidden   = true;
    dpTrigger.setAttribute('aria-expanded', 'false');
    dpTrigger.focus();
  }

  function confirmDates() {
    var s = hsStart ? hsStart.value : '';
    var e = hsEnd   ? hsEnd.value   : '';
    if (s && e && e < s) {
      hsEnd.setCustomValidity('La fecha de fin debe ser posterior a la de inicio.');
      hsEnd.reportValidity();
      return;
    }
    if (hsEnd) hsEnd.setCustomValidity('');
    dpHiddenStart.value = s;
    dpHiddenEnd.value   = e;
    dpTrigger.value =
      (s && e) ? isoToEs(s) + ' – ' + isoToEs(e) :
      s         ? isoToEs(s) + ' – …'             : '';
    closeModal();
  }

  dpTrigger.addEventListener('click', openModal);
  dpTrigger.addEventListener('keydown', function (e) {
    if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); openModal(); }
  });

  /* Global Escape closes modal */
  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape' && hsModal && !hsModal.hidden) closeModal();
  });

  /* ══════════════════════════════════════════════════════════
     FORM SUBMIT
  ══════════════════════════════════════════════════════════ */

  var form = document.getElementById('hero-search-form');
  if (form) {
    form.addEventListener('submit', function (e) {
      var sid = acIdInput.value.trim();
      if (sid) {
        e.preventDefault();
        window.location.href = '/estacion/' + encodeURIComponent(sid);
      }
      /* else: GET /estaciones?q=...&fecha_inicio=...&fecha_fin=...&tipo=... */
    });
  }

})();
