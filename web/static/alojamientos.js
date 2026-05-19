/* SnowBreak — alojamientos.js
 *
 * Lógica de la página /alojamiento/{id}:
 *   - Inicializa las dos fechas (hoy y mañana) si están vacías.
 *   - Recalcula noches y total en vivo cuando cambia cualquier input.
 *   - Valida en cliente antes de enviar (entrada >= hoy, salida > entrada,
 *     huéspedes en [1,10]). El backend valida igual: nunca confiamos en
 *     el cliente.
 *   - Envía POST /api/alojamientos/reservar con JSON. El CSRF lo añade
 *     csrf.js automáticamente al header X-CSRF-Token.
 *
 * Sin librerías externas. Compatible con navegadores modernos
 * (input type="date", fetch, JSON nativo).
 */
(function () {
  'use strict';

  var root = document.querySelector('.lodging-booking');
  if (!root) return;

  var form = document.getElementById('lodging-booking-form');
  if (!form) return; // usuario no logueado: solo CTA de login, no hay form

  var alojamientoId = parseInt(root.getAttribute('data-alojamiento-id'), 10);
  var precioNoche = parseFloat(root.getAttribute('data-precio-noche'));
  if (!Number.isFinite(alojamientoId) || alojamientoId <= 0) return;
  if (!Number.isFinite(precioNoche) || precioNoche < 0) return;

  var inCheckIn   = document.getElementById('lodging-checkin');
  var inCheckOut  = document.getElementById('lodging-checkout');
  var inGuests    = document.getElementById('lodging-guests');
  var elNoches    = form.querySelector('[data-noches]');
  var elTotal     = form.querySelector('[data-total]');
  var elFeedback  = form.querySelector('[data-feedback]');
  var btnSubmit   = form.querySelector('button[type="submit"]');

  function pad(n) { return n < 10 ? '0' + n : '' + n; }
  function isoDate(d) {
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate());
  }

  // Inicializa fechas: hoy + 1 día como entrada por defecto, salida 3 noches después.
  // Pone también el atributo min para que el selector nativo no permita pasado.
  var hoy = new Date();
  hoy.setHours(0, 0, 0, 0);
  var manana = new Date(hoy);
  manana.setDate(manana.getDate() + 1);
  var tresNochesDespues = new Date(manana);
  tresNochesDespues.setDate(tresNochesDespues.getDate() + 3);

  inCheckIn.min  = isoDate(hoy);
  inCheckOut.min = isoDate(manana);
  if (!inCheckIn.value)  inCheckIn.value  = isoDate(manana);
  if (!inCheckOut.value) inCheckOut.value = isoDate(tresNochesDespues);

  function parseFechaISO(s) {
    // Construye una Date en hora local desde "YYYY-MM-DD" sin tocar zona horaria.
    if (!s) return null;
    var partes = s.split('-');
    if (partes.length !== 3) return null;
    var d = new Date(parseInt(partes[0], 10), parseInt(partes[1], 10) - 1, parseInt(partes[2], 10));
    return isNaN(d.getTime()) ? null : d;
  }

  function calcular() {
    var entrada = parseFechaISO(inCheckIn.value);
    var salida  = parseFechaISO(inCheckOut.value);

    if (!entrada || !salida) {
      elNoches.textContent = '—';
      elTotal.textContent  = '—';
      return null;
    }
    var ms = salida.getTime() - entrada.getTime();
    var noches = Math.round(ms / (1000 * 60 * 60 * 24));
    if (noches < 1) {
      elNoches.textContent = '—';
      elTotal.textContent  = '—';
      return null;
    }
    var total = noches * precioNoche;
    elNoches.textContent = noches + (noches === 1 ? ' noche' : ' noches');
    elTotal.textContent  = total.toLocaleString('es-ES', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) + ' €';

    // Actualiza min de salida según entrada (no puede ser anterior).
    var minSalida = new Date(entrada);
    minSalida.setDate(minSalida.getDate() + 1);
    inCheckOut.min = isoDate(minSalida);

    return { entrada: inCheckIn.value, salida: inCheckOut.value, noches: noches, total: total };
  }

  function setFeedback(msg, tipo) {
    elFeedback.textContent = msg || '';
    elFeedback.className = 'lodging-booking__feedback';
    if (tipo) elFeedback.classList.add('lodging-booking__feedback--' + tipo);
  }

  // Cálculo en vivo.
  inCheckIn.addEventListener('change', calcular);
  inCheckOut.addEventListener('change', calcular);
  calcular();

  // Envío.
  form.addEventListener('submit', function (e) {
    e.preventDefault();
    setFeedback('', null);

    var datos = calcular();
    if (!datos) {
      setFeedback('Comprueba las fechas: la salida tiene que ser posterior a la entrada.', 'error');
      return;
    }
    var huespedes = parseInt(inGuests.value, 10);
    if (!Number.isFinite(huespedes) || huespedes < 1 || huespedes > 10) {
      setFeedback('El número de huéspedes debe estar entre 1 y 10.', 'error');
      return;
    }

    btnSubmit.disabled = true;
    btnSubmit.textContent = 'Reservando…';

    fetch('/api/alojamientos/reservar', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
      credentials: 'same-origin',
      body: JSON.stringify({
        alojamiento_id: alojamientoId,
        fecha_entrada:  datos.entrada,
        fecha_salida:   datos.salida,
        huespedes:      huespedes
      })
    })
    .then(function (res) {
      return res.json().then(function (body) { return { ok: res.ok, status: res.status, body: body }; });
    })
    .then(function (r) {
      btnSubmit.disabled = false;
      btnSubmit.textContent = 'Confirmar reserva';
      if (!r.ok) {
        if (r.status === 401) {
          setFeedback('Inicia sesión para reservar. Te redirijo...', 'error');
          setTimeout(function () { window.location.href = '/login'; }, 1200);
          return;
        }
        var mensaje = (r.body && r.body.error) ? r.body.error : 'No se ha podido completar la reserva.';
        setFeedback(mensaje, 'error');
        return;
      }
      var body = r.body || {};
      var total = (body.total_eur || 0).toLocaleString('es-ES', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
      setFeedback(
        '¡Reserva confirmada! ' + body.noches + ' noche' + (body.noches === 1 ? '' : 's') +
        ' por ' + total + ' € (referencia #' + body.reserva_id + ').',
        'success'
      );
      btnSubmit.disabled = true;
      btnSubmit.textContent = 'Reserva confirmada';
    })
    .catch(function (err) {
      btnSubmit.disabled = false;
      btnSubmit.textContent = 'Confirmar reserva';
      setFeedback('Error de red: ' + err.message, 'error');
    });
  });
})();
