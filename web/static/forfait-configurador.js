(function () {
  'use strict';

  /* Usa la misma clave que cart.js para leer el carrito real */
  var CART_KEY = 'skihub_cart_v1';

  function getCart() {
    try {
      var raw = localStorage.getItem(CART_KEY);
      var arr = JSON.parse(raw || '[]');
      return Array.isArray(arr) ? arr : [];
    } catch (_) { return []; }
  }

  function formatEuros(n) {
    return n.toFixed(2).replace('.', ',') + ' €';
  }

  function updateResumen() {
    var vacio = document.getElementById('forfait-resumen-vacio');
    var datos = document.getElementById('forfait-resumen-datos');
    if (!vacio || !datos) return;

    var items = getCart();

    if (!items.length) {
      vacio.hidden = false;
      datos.hidden = true;
      return;
    }

    /* it.total ya es precio_unitario * cantidad * dias — calculado en cart.js */
    var subtotal = items.reduce(function (acc, it) { return acc + (it.total || 0); }, 0);

    var seguroCheck = document.getElementById('forfait-seguro-check');
    var seguroRadio = document.querySelector('input[name="seguro"]:checked');
    var seguroCost = 0;
    if (seguroCheck && seguroCheck.checked && seguroRadio) {
      var rates = { basico: 15, completo: 30, premium: 50 };
      var personas = items.reduce(function (acc, it) { return acc + (it.cantidad || 1); }, 0);
      seguroCost = (rates[seguroRadio.value] || 0) * personas;
    }

    var total = subtotal + seguroCost;

    var elPrecio = document.getElementById('forfait-resumen-total-precio');
    var elSeguro = document.getElementById('forfait-resumen-seguro');
    var elTotal  = document.getElementById('forfait-resumen-total');

    if (elPrecio) elPrecio.textContent = formatEuros(subtotal);
    if (elSeguro) elSeguro.textContent = seguroCost > 0 ? formatEuros(seguroCost) : 'No incluido';
    if (elTotal)  elTotal.textContent  = formatEuros(total);

    vacio.hidden = true;
    datos.hidden = false;
  }

  function initRadioHighlight() {
    document.querySelectorAll('input[name="seguro"]').forEach(function (radio) {
      radio.addEventListener('change', function () {
        document.querySelectorAll('.forfait-radio').forEach(function (el) {
          el.classList.remove('forfait-radio--selected');
        });
        radio.closest('.forfait-radio').classList.add('forfait-radio--selected');
        updateResumen();
      });
    });
  }

  function initSeguroToggle() {
    var check = document.getElementById('forfait-seguro-check');
    var opciones = document.getElementById('forfait-seguro-opciones');
    if (!check || !opciones) return;
    check.addEventListener('change', function () {
      opciones.hidden = !check.checked;
      updateResumen();
    });
  }

  function initGuardar() {
    var btn = document.getElementById('forfait-guardar');
    if (!btn) return;
    btn.addEventListener('click', function () {
      btn.textContent = '✓ Configuración guardada';
      btn.disabled = true;
      setTimeout(function () {
        btn.innerHTML = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg> Guardar Configuración';
        btn.disabled = false;
      }, 2000);
    });
  }

  document.addEventListener('DOMContentLoaded', function () {
    initRadioHighlight();
    initSeguroToggle();
    initGuardar();
    updateResumen();

    /* Escucha cambios del carrito en esta misma pestaña (SnowbreakCart.onChange)
       y también desde otras pestañas (storage event) */
    if (window.SnowbreakCart) {
      window.SnowbreakCart.onChange(updateResumen);
    }
    window.addEventListener('storage', function (e) {
      if (e.key === CART_KEY) updateResumen();
    });
  });
})();
