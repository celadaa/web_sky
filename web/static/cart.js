/**
 * cart.js — Cesta de la compra de Snowbreak.
 *
 * Estructura de un ítem (estado local en localStorage):
 * {
 *   id:               string,   // UUID-like, generado al añadir
 *   id_estacion:      number,
 *   nombre_estacion:  string,
 *   imagen:           string,   // URL de imagen (opcional)
 *   fecha_inicio:     string,   // YYYY-MM-DD
 *   fecha_fin:        string,   // YYYY-MM-DD
 *   tipo_pase:        'adulto' | 'nino' | 'senior',
 *   precio_unitario:  number,   // €/día
 *   cantidad:         number,   // nº de personas
 *   dias:             number,   // nº de días entre fecha_inicio y fecha_fin (mín. 1)
 *   total:            number    // precio_unitario * cantidad * dias
 * }
 */
(function () {
  'use strict';

  var STORAGE_KEY = 'skihub_cart_v1';
  var listeners = [];

  // ---------- Helpers ----------

  function uuid() {
    return 'it_' + Date.now().toString(36) + '_' + Math.random().toString(36).slice(2, 8);
  }

  function safeParse(raw) {
    try {
      var arr = JSON.parse(raw);
      return Array.isArray(arr) ? arr : [];
    } catch (e) {
      return [];
    }
  }

  function load() {
    try {
      return safeParse(localStorage.getItem(STORAGE_KEY) || '[]');
    } catch (e) {
      return [];
    }
  }

  function save(items) {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
    } catch (e) { /* ignore */ }
    notify(items);
  }

  function diasEntre(inicioStr, finStr) {
    if (!inicioStr || !finStr) { return 1; }
    var inicio = new Date(inicioStr + 'T00:00:00');
    var fin = new Date(finStr + 'T00:00:00');
    var ms = fin.getTime() - inicio.getTime();
    var dias = Math.round(ms / (1000 * 60 * 60 * 24)) + 1;
    return dias > 0 ? dias : 1;
  }

  function formatoFecha(s) {
    if (!s) { return ''; }
    var d = new Date(s + 'T00:00:00');
    if (isNaN(d.getTime())) { return s; }
    return d.toLocaleDateString('es-ES', { day: 'numeric', month: 'short' });
  }

  function formatoEuros(n) {
    return n.toFixed(2).replace('.', ',') + ' €';
  }

  function tipoLabel(tipo) {
    switch (tipo) {
      case 'adulto': return 'Adulto';
      case 'nino':   return 'Niño';
      case 'senior': return 'Senior';
      default:       return tipo;
    }
  }

  function notify(items) {
    listeners.forEach(function (fn) {
      try { fn(items); } catch (e) { /* ignore */ }
    });
  }

  // ---------- API pública ----------

  var Cart = {
    items: function () { return load(); },

    count: function () {
      return load().reduce(function (acc, it) { return acc + (it.cantidad || 0); }, 0);
    },

    subtotal: function () {
      return load().reduce(function (acc, it) { return acc + (it.total || 0); }, 0);
    },

    add: function (item) {
      var items = load();
      var dias = diasEntre(item.fecha_inicio, item.fecha_fin);
      var cantidad = Math.max(1, parseInt(item.cantidad || 1, 10));
      var precio = parseFloat(item.precio_unitario || 0);
      items.push({
        id: uuid(),
        id_estacion: item.id_estacion,
        nombre_estacion: item.nombre_estacion,
        imagen: item.imagen || '',
        fecha_inicio: item.fecha_inicio,
        fecha_fin: item.fecha_fin,
        tipo_pase: item.tipo_pase,
        precio_unitario: precio,
        cantidad: cantidad,
        dias: dias,
        total: precio * cantidad * dias
      });
      save(items);
    },

    remove: function (id) {
      save(load().filter(function (it) { return it.id !== id; }));
    },

    updateQty: function (id, qty) {
      var items = load();
      var n = Math.max(1, parseInt(qty || 1, 10));
      items.forEach(function (it) {
        if (it.id === id) {
          it.cantidad = n;
          it.total = it.precio_unitario * n * (it.dias || 1);
        }
      });
      save(items);
    },

    clear: function () { save([]); },

    onChange: function (fn) { listeners.push(fn); },

    formatoEuros: formatoEuros,
    formatoFecha: formatoFecha,
    tipoLabel: tipoLabel
  };

  window.SnowbreakCart = Cart;

  // ---------- UI: header (badge + dropdown) ----------

  function refreshHeader() {
    var badge = document.getElementById('cart-count');
    var btn = document.getElementById('cart-button');
    var listEl = document.getElementById('cart-dropdown-items');
    var emptyEl = document.getElementById('cart-dropdown-empty');
    var footerEl = document.getElementById('cart-dropdown-footer');
    var subtotalEl = document.getElementById('cart-dropdown-subtotal');

    var n = Cart.count();
    if (badge) {
      badge.textContent = n;
      badge.hidden = n === 0;
    }
    if (btn) {
      btn.classList.toggle('cart-button--has-items', n > 0);
    }

    if (!listEl || !emptyEl || !footerEl) { return; }

    var items = Cart.items();
    listEl.innerHTML = '';
    if (items.length === 0) {
      emptyEl.hidden = false;
      footerEl.hidden = true;
      return;
    }
    emptyEl.hidden = true;
    footerEl.hidden = false;
    items.forEach(function (it) {
      var li = document.createElement('li');
      li.className = 'cart-dropdown__item';
      var titulo = it.cantidad + 'x Forfait ' + tipoLabel(it.tipo_pase) + ' ' + it.nombre_estacion;
      var fechas = '(' + formatoFecha(it.fecha_inicio) + ' – ' + formatoFecha(it.fecha_fin) + ')';
      li.innerHTML =
        '<div class="cart-dropdown__item-title">' + escapeHtml(titulo) + '</div>' +
        '<div class="cart-dropdown__item-meta">' +
          '<span>' + escapeHtml(fechas) + '</span>' +
          '<strong>' + formatoEuros(it.total) + '</strong>' +
        '</div>';
      listEl.appendChild(li);
    });
    if (subtotalEl) {
      subtotalEl.textContent = formatoEuros(Cart.subtotal());
    }
  }

  function escapeHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  function initHeader() {
    var btn = document.getElementById('cart-button');
    var dropdown = document.getElementById('cart-dropdown');
    if (!btn || !dropdown) { return; }

    // El dropdown se muestra/oculta vía CSS (hover/focus-within), aquí solo
    // nos aseguramos de que no quede el atributo `hidden` del HTML inicial,
    // ya que rompería la transición CSS.
    dropdown.removeAttribute('hidden');
    btn.setAttribute('aria-expanded', 'false');

    refreshHeader();

    // Click en el botón -> navegamos a /cesta (comportamiento de enlace normal).
    // No interceptamos el click; dejamos que el href haga su trabajo.

    // Otra pestaña actualiza el cart -> refrescar.
    window.addEventListener('storage', function (ev) {
      if (ev.key === STORAGE_KEY) { refreshHeader(); }
    });

    Cart.onChange(refreshHeader);
  }

  // ---------- UI: widget de compra en página de estación ----------

  function initStationWidget() {
    var widget = document.querySelector('[data-forfait-widget]');
    if (!widget) { return; }

    var precios = {
      adulto: parseFloat(widget.getAttribute('data-precio-adulto') || 0),
      nino:   parseFloat(widget.getAttribute('data-precio-nino') || 0),
      senior: parseFloat(widget.getAttribute('data-precio-senior') || 0)
    };
    var estacionId = parseInt(widget.getAttribute('data-estacion-id'), 10);
    var nombre = widget.getAttribute('data-estacion-nombre');
    var imagen = widget.getAttribute('data-estacion-imagen') || '';

    var inicioInput = widget.querySelector('[data-fecha-inicio]');
    var finInput = widget.querySelector('[data-fecha-fin]');
    var tipoSelect = widget.querySelector('[data-tipo-pase]');
    var cantidadInput = widget.querySelector('[data-cantidad]');
    var precioEl = widget.querySelector('[data-precio]');
    var totalEl = widget.querySelector('[data-total]');
    var diasEl = widget.querySelector('[data-dias]');
    var btnAdd = widget.querySelector('[data-add-cart]');
    var feedback = widget.querySelector('[data-feedback]');

    // Fechas por defecto: mañana y +2 días.
    function fechaIso(d) {
      return d.toISOString().slice(0, 10);
    }
    var hoy = new Date();
    var manana = new Date(hoy.getTime() + 24 * 60 * 60 * 1000);
    var fin = new Date(hoy.getTime() + 3 * 24 * 60 * 60 * 1000);
    if (inicioInput && !inicioInput.value) { inicioInput.value = fechaIso(manana); }
    if (finInput && !finInput.value) { finInput.value = fechaIso(fin); }
    if (inicioInput) { inicioInput.min = fechaIso(hoy); }

    function recalc() {
      var tipo = tipoSelect ? tipoSelect.value : 'adulto';
      var precio = precios[tipo] || 0;
      var cantidad = cantidadInput ? Math.max(1, parseInt(cantidadInput.value || '1', 10)) : 1;
      var dias = diasEntre(inicioInput && inicioInput.value, finInput && finInput.value);
      if (precioEl) { precioEl.textContent = formatoEuros(precio); }
      if (diasEl) { diasEl.textContent = dias + (dias === 1 ? ' día' : ' días'); }
      if (totalEl) { totalEl.textContent = formatoEuros(precio * cantidad * dias); }
    }

    [inicioInput, finInput, tipoSelect, cantidadInput].forEach(function (el) {
      if (el) { el.addEventListener('change', recalc); el.addEventListener('input', recalc); }
    });
    recalc();

    if (btnAdd) {
      btnAdd.addEventListener('click', function (ev) {
        ev.preventDefault();
        var inicio = inicioInput ? inicioInput.value : '';
        var finVal = finInput ? finInput.value : '';
        if (!inicio || !finVal) {
          if (feedback) { feedback.textContent = 'Selecciona las fechas del forfait.'; feedback.className = 'forfait-card__feedback forfait-card__feedback--error'; }
          return;
        }
        if (new Date(finVal) < new Date(inicio)) {
          if (feedback) { feedback.textContent = 'La fecha de fin debe ser posterior a la de inicio.'; feedback.className = 'forfait-card__feedback forfait-card__feedback--error'; }
          return;
        }
        var tipo = tipoSelect ? tipoSelect.value : 'adulto';
        var cantidad = cantidadInput ? Math.max(1, parseInt(cantidadInput.value || '1', 10)) : 1;

        Cart.add({
          id_estacion: estacionId,
          nombre_estacion: nombre,
          imagen: imagen,
          fecha_inicio: inicio,
          fecha_fin: finVal,
          tipo_pase: tipo,
          precio_unitario: precios[tipo] || 0,
          cantidad: cantidad
        });

        if (feedback) {
          feedback.textContent = '¡Forfait añadido a la cesta!';
          feedback.className = 'forfait-card__feedback forfait-card__feedback--success';
        }
      });
    }
  }

  // ---------- UI: página de cesta completa ----------

  function initCartPage() {
    var page = document.getElementById('cart-page');
    if (!page) { return; }

    var listEl = document.getElementById('cart-page-list');
    var emptyEl = document.getElementById('cart-page-empty');
    var summaryEl = document.getElementById('cart-page-summary');
    var subtotalEl = document.getElementById('cart-page-subtotal');
    var totalEl = document.getElementById('cart-page-total');
    var clearBtn = document.getElementById('cart-page-clear');

    function render() {
      var items = Cart.items();
      listEl.innerHTML = '';
      if (items.length === 0) {
        emptyEl.hidden = false;
        summaryEl.hidden = true;
        return;
      }
      emptyEl.hidden = true;
      summaryEl.hidden = false;

      items.forEach(function (it) {
        var card = document.createElement('article');
        card.className = 'cart-line';
        card.innerHTML =
          '<div class="cart-line__media">' +
            (it.imagen ? '<img src="' + escapeHtml(it.imagen) + '" alt="' + escapeHtml(it.nombre_estacion) + '">' : '<div class="cart-line__media-placeholder"></div>') +
          '</div>' +
          '<div class="cart-line__body">' +
            '<h3 class="cart-line__title">Forfait ' + escapeHtml(tipoLabel(it.tipo_pase)) + ' · ' + escapeHtml(it.nombre_estacion) + '</h3>' +
            '<dl class="cart-line__meta">' +
              '<div><dt>Fechas</dt><dd>' + escapeHtml(formatoFecha(it.fecha_inicio)) + ' – ' + escapeHtml(formatoFecha(it.fecha_fin)) + '</dd></div>' +
              '<div><dt>Días</dt><dd>' + it.dias + '</dd></div>' +
              '<div><dt>Precio</dt><dd>' + formatoEuros(it.precio_unitario) + ' / día</dd></div>' +
            '</dl>' +
            '<div class="cart-line__controls">' +
              '<label>Cantidad <input type="number" min="1" value="' + it.cantidad + '" data-qty="' + it.id + '" class="cart-line__qty"></label>' +
              '<button type="button" data-remove="' + it.id + '" class="cart-line__remove">Eliminar</button>' +
            '</div>' +
          '</div>' +
          '<div class="cart-line__total"><span>Total</span><strong>' + formatoEuros(it.total) + '</strong></div>';
        listEl.appendChild(card);
      });

      var subtotal = Cart.subtotal();
      subtotalEl.textContent = formatoEuros(subtotal);
      totalEl.textContent = formatoEuros(subtotal);
    }

    listEl.addEventListener('change', function (ev) {
      var t = ev.target;
      if (t && t.dataset && t.dataset.qty) {
        Cart.updateQty(t.dataset.qty, t.value);
      }
    });
    listEl.addEventListener('click', function (ev) {
      var t = ev.target;
      if (t && t.dataset && t.dataset.remove) {
        Cart.remove(t.dataset.remove);
      }
    });
    if (clearBtn) {
      clearBtn.addEventListener('click', function () {
        if (confirm('¿Vaciar la cesta?')) { Cart.clear(); }
      });
    }

    Cart.onChange(render);
    render();
  }

  // ---------- Init ----------

  function init() {
    initHeader();
    initStationWidget();
    initCartPage();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
