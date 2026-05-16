(function () {
  'use strict';

  /* Misma clave que cart.js */
  var CART_KEY = 'skihub_cart_v1';
  var TASA_TURISTICA = 5;

  function getCart() {
    try {
      var raw = localStorage.getItem(CART_KEY);
      var arr = JSON.parse(raw || '[]');
      return Array.isArray(arr) ? arr : [];
    } catch (_) { return []; }
  }

  function clearCart() {
    try { localStorage.removeItem(CART_KEY); } catch (_) {}
  }

  function formatEuros(n) {
    return n.toFixed(2).replace('.', ',') + ' €';
  }

  function escHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;')
      .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  function formatFecha(s) {
    if (!s) return '';
    var d = new Date(s + 'T00:00:00');
    if (isNaN(d.getTime())) return s;
    return d.toLocaleDateString('es-ES', { day: 'numeric', month: 'short' });
  }

  function tipoLabel(t) {
    return { adulto: 'Adulto', nino: 'Niño', senior: 'Sénior' }[t] || t;
  }

  function renderResumen() {
    var items = getCart();
    var lista   = document.getElementById('pago-resumen-lista');
    var vacio   = document.getElementById('pago-resumen-vacio');
    var totales = document.getElementById('pago-resumen-totales');
    var btn     = document.getElementById('pago-btn');
    if (!lista) return;

    lista.innerHTML = '';

    if (!items.length) {
      if (vacio)   vacio.hidden   = false;
      if (totales) totales.hidden = true;
      var tb = document.getElementById('pago-total-block');
      if (tb) tb.hidden = true;
      if (btn)     btn.disabled   = true;
      return;
    }
    if (vacio) vacio.hidden = true;

    var subtotalForfaits = 0;
    items.forEach(function (it) {
      /* it.total = precio_unitario * cantidad * dias  (calculado en cart.js) */
      var precio = it.total || 0;
      subtotalForfaits += precio;

      var diasStr = (it.dias || 1) + (it.dias === 1 ? ' día' : ' días');
      var personasStr = (it.cantidad || 1) + ' ' + tipoLabel(it.tipo_pase || 'adulto');
      var fechasStr = it.fecha_inicio && it.fecha_fin
        ? formatFecha(it.fecha_inicio) + ' – ' + formatFecha(it.fecha_fin)
        : '';

      var div = document.createElement('div');
      div.className = 'pago-resumen__item';
      div.innerHTML =
        '<dt>' +
          '<span class="pago-resumen__item-nombre">Forfait ' + escHtml(tipoLabel(it.tipo_pase)) + ' · ' + escHtml(it.nombre_estacion || '') + '</span>' +
          '<span class="pago-resumen__item-sub">' + escHtml(personasStr) + ', ' + escHtml(diasStr) +
          (fechasStr ? ' &bull; ' + escHtml(fechasStr) : '') + '</span>' +
        '</dt>' +
        '<dd>' + formatEuros(precio) + '</dd>';
      lista.appendChild(div);
    });

    var iva   = subtotalForfaits * 0.10;
    var total = subtotalForfaits + iva + TASA_TURISTICA;

    if (totales) totales.hidden = false;
    var totalBlock = document.getElementById('pago-total-block');
    if (totalBlock) totalBlock.hidden = false;

    var elSub   = document.getElementById('pago-subtotal');
    var elIva   = document.getElementById('pago-iva');
    var elTasa  = document.getElementById('pago-tasa');
    var elTotal = document.getElementById('pago-total');

    if (elSub)   elSub.textContent   = formatEuros(subtotalForfaits);
    if (elIva)   elIva.textContent   = formatEuros(iva);
    if (elTasa)  elTasa.textContent  = formatEuros(TASA_TURISTICA);
    if (elTotal) elTotal.textContent = formatEuros(total);

    if (btn) btn.disabled = false;
  }

  function initMetodos() {
    document.querySelectorAll('.pago-metodo input[type="radio"]').forEach(function (radio) {
      radio.addEventListener('change', function () {
        document.querySelectorAll('.pago-metodo').forEach(function (el) {
          el.classList.remove('pago-metodo--selected');
        });
        radio.closest('.pago-metodo').classList.add('pago-metodo--selected');
      });
    });
  }

  function initFormatoTarjeta() {
    var numero = document.getElementById('pago-numero-tarjeta');
    if (numero) {
      numero.addEventListener('input', function () {
        var val = numero.value.replace(/\D/g, '').substring(0, 16);
        numero.value = val.replace(/(.{4})/g, '$1 ').trim();
      });
    }
    var cad = document.getElementById('pago-caducidad');
    if (cad) {
      cad.addEventListener('input', function () {
        var val = cad.value.replace(/\D/g, '').substring(0, 4);
        if (val.length >= 3) val = val.substring(0, 2) + '/' + val.substring(2);
        cad.value = val;
      });
    }
  }

  function initPagar() {
    var btn = document.getElementById('pago-btn');
    var terminosCheck = document.getElementById('pago-terminos-check');
    if (!btn) return;

    btn.addEventListener('click', function () {
      if (terminosCheck && !terminosCheck.checked) {
        alert('Debes aceptar los Términos y Condiciones para continuar.');
        terminosCheck.focus();
        return;
      }
      var items = getCart();
      if (!items.length) {
        alert('Tu cesta está vacía.');
        return;
      }

      btn.disabled = true;
      btn.textContent = 'Procesando…';

      setTimeout(function () {
        clearCart();
        /* Notifica a cart.js si está disponible */
        if (window.SnowbreakCart) {
          window.SnowbreakCart.clear();
        }

        var overlay = document.createElement('div');
        overlay.className = 'pago-confirmacion-overlay';
        overlay.innerHTML =
          '<div class="pago-confirmacion">' +
            '<div class="pago-confirmacion__icono">✓</div>' +
            '<h2>¡Reserva confirmada!</h2>' +
            '<p>Hemos recibido tu pedido correctamente.</p>' +
            '<p class="pago-confirmacion__nota">Recibirás un correo de confirmación en breve.</p>' +
            '<p class="pago-confirmacion__demo"><strong>Demo académica:</strong> No se ha procesado ningún pago real.</p>' +
            '<a href="/" class="btn-forfait-reservar">Volver al inicio</a>' +
          '</div>';
        document.body.appendChild(overlay);
      }, 1600);
    });
  }

  document.addEventListener('DOMContentLoaded', function () {
    renderResumen();
    initMetodos();
    initFormatoTarjeta();
    initPagar();

    if (window.SnowbreakCart) {
      window.SnowbreakCart.onChange(renderResumen);
    }
    window.addEventListener('storage', function (e) {
      if (e.key === CART_KEY) renderResumen();
    });
  });
})();
