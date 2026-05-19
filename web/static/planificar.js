/* SnowBreak — planificar.js
 *
 * Asistente "Planifica tu estancia" en 4 pasos:
 *   1) Destino   → carga /api/estaciones, click para seleccionar
 *   2) Fechas    → input dates + huespedes, validacion en cliente
 *   3) Alojamiento → carga /api/alojamientos/estacion/{id}
 *   4) Forfait   → muestra precios de la estacion elegida, anade al
 *                  carrito existente via window.SnowbreakCart.add()
 *
 * Al confirmar el paso 3 se hace POST /api/alojamientos/reservar
 * (la reserva se persiste en BD). Despues, el paso 4 es opcional:
 * el usuario puede anadir forfait al carrito o saltarse.
 *
 * Sin librerias externas. CSRF lo inyecta csrf.js automaticamente
 * en cualquier fetch unsafe del mismo origen.
 */
(function () {
  'use strict';

  // ----- Estado del wizard ------------------------------------------------
  var state = {
    step: 1,                      // 1..4 + 5 (success)
    estacion: null,               // objeto de /api/estaciones
    fechas: { entrada: '', salida: '', noches: 0, huespedes: 2 },
    alojamiento: null,            // objeto de /api/alojamientos/estacion/X
    totalAlojamiento: 0,
    reservaId: null,
    estacionesCache: null         // se cachea para no re-fetchear
  };
  var userLogged = window.__SNOWBREAK_USER_LOGGED__ === true;

  // ----- Helpers ----------------------------------------------------------
  function $(sel) { return document.querySelector(sel); }
  function $$(sel) { return Array.prototype.slice.call(document.querySelectorAll(sel)); }
  function pad(n) { return n < 10 ? '0' + n : '' + n; }
  function isoDate(d) { return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()); }
  function parseISO(s) {
    if (!s) return null;
    var p = s.split('-');
    if (p.length !== 3) return null;
    var d = new Date(parseInt(p[0], 10), parseInt(p[1], 10) - 1, parseInt(p[2], 10));
    return isNaN(d.getTime()) ? null : d;
  }
  function fmtEuros(n) {
    return n.toLocaleString('es-ES', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) + ' €';
  }
  function fmtFechaCorta(s) {
    var d = parseISO(s);
    if (!d) return s;
    return pad(d.getDate()) + '/' + pad(d.getMonth() + 1) + '/' + d.getFullYear();
  }

  // ----- Render del estado en pasos + resumen ----------------------------
  function showStep(n) {
    state.step = n;
    $$('.wizard__panel').forEach(function (p) {
      var s = parseInt(p.getAttribute('data-step'), 10);
      var active = s === n;
      p.classList.toggle('is-active', active);
      p.hidden = !active;
    });
    $$('.wizard__step').forEach(function (li) {
      var s = parseInt(li.getAttribute('data-step-pill'), 10);
      li.classList.toggle('is-active', s === n);
      li.classList.toggle('is-done', s < n);
    });
    var prev = $('#wizard-prev');
    var next = $('#wizard-next');
    prev.hidden = (n <= 1 || n === 5);
    next.hidden = (n === 5);
    next.textContent = etiquetaBoton(n);
    actualizarBotonNext();
  }
  function etiquetaBoton(n) {
    if (n === 1) return 'Continuar';
    if (n === 2) return 'Ver alojamientos';
    if (n === 3) return userLogged ? 'Reservar alojamiento' : 'Inicia sesión para reservar';
    if (n === 4) return 'Finalizar';
    return 'Continuar';
  }
  function actualizarBotonNext() {
    var next = $('#wizard-next');
    var ok = false;
    switch (state.step) {
      case 1: ok = !!state.estacion; break;
      case 2: ok = !!(state.fechas.entrada && state.fechas.salida && state.fechas.noches > 0); break;
      case 3: ok = !!state.alojamiento && userLogged; break;
      case 4: ok = true; break;
    }
    next.disabled = !ok;
  }
  function actualizarResumen() {
    $('[data-summary-station]').textContent = state.estacion ? state.estacion.nombre : '—';
    if (state.fechas.entrada && state.fechas.salida) {
      $('[data-summary-dates]').textContent = fmtFechaCorta(state.fechas.entrada) + ' → ' + fmtFechaCorta(state.fechas.salida) +
        ' · ' + state.fechas.noches + (state.fechas.noches === 1 ? ' noche' : ' noches');
    } else {
      $('[data-summary-dates]').textContent = '—';
    }
    $('[data-summary-guests]').textContent = state.fechas.huespedes || '—';
    $('[data-summary-lodging]').textContent = state.alojamiento ? state.alojamiento.nombre : '—';
    $('[data-summary-total]').textContent = state.totalAlojamiento > 0 ? fmtEuros(state.totalAlojamiento) : '—';
  }

  // ----- PASO 1: estaciones ----------------------------------------------
  function cargarEstaciones() {
    if (state.estacionesCache) {
      renderEstaciones(state.estacionesCache, '');
      return;
    }
    var loading = $('section[data-step="1"] [data-loading]');
    loading.hidden = false;
    fetch('/api/estaciones', { credentials: 'same-origin' })
      .then(function (r) { return r.json(); })
      .then(function (lista) {
        loading.hidden = true;
        state.estacionesCache = lista;
        renderEstaciones(lista, '');
      })
      .catch(function () {
        loading.textContent = 'No se han podido cargar las estaciones.';
      });
  }
  function renderEstaciones(lista, filtro) {
    var cont = $('#wizard-stations');
    var f = (filtro || '').trim().toLowerCase();
    var visibles = lista.filter(function (e) {
      if (!f) return true;
      return (e.nombre + ' ' + (e.ubicacion || '')).toLowerCase().indexOf(f) !== -1;
    });
    if (!visibles.length) {
      cont.innerHTML = '<li class="wizard__empty">Sin resultados para "' + escapeHTML(filtro) + '".</li>';
      return;
    }
    cont.innerHTML = visibles.map(function (e) {
      var selectedCls = (state.estacion && state.estacion.id === e.id) ? ' is-selected' : '';
      return '<li class="wizard-station' + selectedCls + '" role="option" tabindex="0" data-id="' + e.id + '">' +
             '  <div class="wizard-station__media">' +
             (e.imagen ? '<img src="' + escapeAttr(e.imagen) + '" alt="" loading="lazy" class="wizard-station__img" data-remove-on-error>' : '') +
             '  </div>' +
             '  <div class="wizard-station__body">' +
             '    <h3 class="wizard-station__name">' + escapeHTML(e.nombre) + '</h3>' +
             '    <p class="wizard-station__loc">' + escapeHTML(e.ubicacion || '') + '</p>' +
             '    <span class="wizard-station__badge wizard-station__badge--' + escapeAttr(e.estado || 'cerrado') + '">' + escapeHTML(e.estado_texto || '') + '</span>' +
             '  </div>' +
             '</li>';
    }).join('');
    cont.querySelectorAll('.wizard-station').forEach(function (el) {
      el.addEventListener('click', function () { seleccionarEstacion(parseInt(el.getAttribute('data-id'), 10)); });
      el.addEventListener('keydown', function (ev) {
        if (ev.key === 'Enter' || ev.key === ' ') { ev.preventDefault(); seleccionarEstacion(parseInt(el.getAttribute('data-id'), 10)); }
      });
    });
  }
  function seleccionarEstacion(id) {
    var enc = (state.estacionesCache || []).find(function (e) { return e.id === id; });
    if (!enc) return;
    state.estacion = enc;
    state.alojamiento = null;
    state.totalAlojamiento = 0;
    renderEstaciones(state.estacionesCache, $('#wizard-search').value);
    actualizarResumen();
    actualizarBotonNext();
  }

  // ----- PASO 2: fechas --------------------------------------------------
  function initFechas() {
    var hoy = new Date(); hoy.setHours(0,0,0,0);
    var manana = new Date(hoy); manana.setDate(manana.getDate() + 1);
    var t3 = new Date(manana); t3.setDate(t3.getDate() + 3);

    var ci = $('#wizard-checkin'); var co = $('#wizard-checkout'); var g = $('#wizard-guests');
    ci.min = isoDate(hoy);
    co.min = isoDate(manana);
    if (!ci.value) ci.value = isoDate(manana);
    if (!co.value) co.value = isoDate(t3);
    if (!g.value) g.value = state.fechas.huespedes || 2;

    [ci, co, g].forEach(function (el) {
      el.addEventListener('change', actualizarFechas);
      el.addEventListener('input', actualizarFechas);
    });
    actualizarFechas();
  }
  function actualizarFechas() {
    var ci = $('#wizard-checkin'); var co = $('#wizard-checkout'); var g = $('#wizard-guests');
    var err = $('[data-dates-error]');
    err.textContent = '';
    var entrada = parseISO(ci.value); var salida = parseISO(co.value);
    if (!entrada || !salida) { state.fechas.noches = 0; actualizarResumen(); actualizarBotonNext(); return; }
    var noches = Math.round((salida - entrada) / 86400000);
    if (noches < 1) {
      err.textContent = 'La fecha de salida tiene que ser posterior a la de entrada.';
      state.fechas.noches = 0; actualizarResumen(); actualizarBotonNext(); return;
    }
    if (noches > 30) {
      err.textContent = 'Máximo 30 noches por reserva.';
      state.fechas.noches = 0; actualizarResumen(); actualizarBotonNext(); return;
    }
    var huespedes = parseInt(g.value, 10);
    if (!isFinite(huespedes) || huespedes < 1 || huespedes > 10) {
      err.textContent = 'Huéspedes entre 1 y 10.';
      state.fechas.noches = 0; actualizarResumen(); actualizarBotonNext(); return;
    }
    // Actualizar min de salida si entrada cambió
    var minSalida = new Date(entrada); minSalida.setDate(minSalida.getDate() + 1);
    co.min = isoDate(minSalida);

    state.fechas.entrada = ci.value;
    state.fechas.salida = co.value;
    state.fechas.noches = noches;
    state.fechas.huespedes = huespedes;
    if (state.alojamiento) state.totalAlojamiento = state.alojamiento.precio_noche * noches;
    actualizarResumen();
    actualizarBotonNext();
  }

  // ----- PASO 3: alojamientos -------------------------------------------
  function cargarAlojamientos() {
    if (!state.estacion) return;
    var cont = $('#wizard-lodgings');
    var loading = $('section[data-step="3"] [data-loading]');
    cont.innerHTML = '';
    loading.hidden = false;
    fetch('/api/alojamientos/estacion/' + state.estacion.id, { credentials: 'same-origin' })
      .then(function (r) { return r.json(); })
      .then(function (lista) {
        loading.hidden = true;
        if (!lista || !lista.length) {
          cont.innerHTML = '<li class="wizard__empty">No hay alojamientos para esta estación.</li>';
          return;
        }
        renderAlojamientos(lista);
      })
      .catch(function () {
        loading.textContent = 'No se han podido cargar los alojamientos.';
      });
  }
  function renderAlojamientos(lista) {
    var cont = $('#wizard-lodgings');
    cont.innerHTML = lista.map(function (a) {
      var selCls = (state.alojamiento && state.alojamiento.id === a.id) ? ' is-selected' : '';
      var totalPrev = (a.precio_noche * state.fechas.noches).toFixed(2);
      return '<li class="wizard-lodging' + selCls + '" role="option" tabindex="0" data-id="' + a.id + '">' +
             '  <div class="wizard-lodging__media">' +
             (a.imagen ? '<img src="' + escapeAttr(a.imagen) + '" alt="" loading="lazy" class="wizard-lodging__img" data-remove-on-error>' : '') +
             '    <span class="wizard-lodging__badge">' + escapeHTML(a.tipo_texto) + '</span>' +
             '  </div>' +
             '  <div class="wizard-lodging__body">' +
             '    <h3 class="wizard-lodging__name">' + escapeHTML(a.nombre) + '</h3>' +
             '    <p class="wizard-lodging__zone">' + escapeHTML(a.zona) + ' · a ' + a.distancia_km.toFixed(1) + ' km</p>' +
             '    <p class="wizard-lodging__desc">' + escapeHTML(a.descripcion) + '</p>' +
             '    <div class="wizard-lodging__price">' +
             '      <strong>' + fmtEuros(a.precio_noche) + '</strong> / noche · total ' + fmtEuros(parseFloat(totalPrev)) +
             '    </div>' +
             '    <div class="wizard-lodging__actions">' +
             '      <a href="/alojamiento/' + a.id + '" target="_blank" rel="noopener" class="wizard-lodging__details">Ver detalle</a>' +
             '    </div>' +
             '  </div>' +
             '</li>';
    }).join('');
    cont._lista = lista;
    cont.querySelectorAll('.wizard-lodging').forEach(function (el) {
      el.addEventListener('click', function (ev) {
        // No interceptar el click del enlace "Ver detalle"
        if (ev.target.closest('.wizard-lodging__details')) return;
        seleccionarAlojamiento(parseInt(el.getAttribute('data-id'), 10), cont._lista);
      });
      el.addEventListener('keydown', function (ev) {
        if (ev.key === 'Enter' || ev.key === ' ') { ev.preventDefault(); seleccionarAlojamiento(parseInt(el.getAttribute('data-id'), 10), cont._lista); }
      });
    });
  }
  function seleccionarAlojamiento(id, lista) {
    var enc = lista.find(function (a) { return a.id === id; });
    if (!enc) return;
    state.alojamiento = enc;
    state.totalAlojamiento = enc.precio_noche * state.fechas.noches;
    renderAlojamientos(lista);
    actualizarResumen();
    actualizarBotonNext();
  }

  // ----- Confirmar reserva (al pasar de paso 3 a 4) ----------------------
  function confirmarReserva() {
    var next = $('#wizard-next');
    next.disabled = true; next.textContent = 'Reservando…';
    return fetch('/api/alojamientos/reservar', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
      credentials: 'same-origin',
      body: JSON.stringify({
        alojamiento_id: state.alojamiento.id,
        fecha_entrada: state.fechas.entrada,
        fecha_salida: state.fechas.salida,
        huespedes: state.fechas.huespedes
      })
    }).then(function (res) {
      return res.json().then(function (body) { return { ok: res.ok, status: res.status, body: body }; });
    }).then(function (r) {
      next.disabled = false; next.textContent = etiquetaBoton(state.step);
      if (!r.ok) {
        if (r.status === 401) {
          window.location.href = '/login';
          return false;
        }
        alert((r.body && r.body.error) || 'No se ha podido reservar.');
        return false;
      }
      state.reservaId = r.body.reserva_id;
      return true;
    }).catch(function (err) {
      next.disabled = false; next.textContent = etiquetaBoton(state.step);
      alert('Error de red: ' + err.message);
      return false;
    });
  }

  // ----- PASO 4: forfait -------------------------------------------------
  function pintarForfait() {
    if (!state.estacion) return;
    $('[data-forfait-station]').textContent = 'Forfaits para ' + state.estacion.nombre;
    var noches = state.fechas.noches;
    var dias = noches; // un dia de forfait por noche (aprox)
    $('[data-forfait-nights]').textContent = dias + (dias === 1 ? ' día' : ' días');
    $('[data-price-adulto]').textContent = fmtEuros((state.estacion.precio_adulto || 0) * dias);
    $('[data-price-nino]').textContent   = fmtEuros((state.estacion.precio_nino   || 0) * dias);
    $('[data-price-senior]').textContent = fmtEuros((state.estacion.precio_senior || 0) * dias);
  }
  function anadirForfaitAlCarrito() {
    if (!window.SnowbreakCart || typeof window.SnowbreakCart.add !== 'function') return false;
    var tipo = (document.querySelector('input[name="forfait-tipo"]:checked') || {}).value || 'adulto';
    var cantidad = parseInt($('#wizard-forfait-qty').value, 10) || 1;
    var precio = state.estacion['precio_' + (tipo === 'nino' ? 'nino' : (tipo === 'senior' ? 'senior' : 'adulto'))] || 0;
    if (!precio) return false;
    window.SnowbreakCart.add({
      id_estacion: state.estacion.id,
      nombre_estacion: state.estacion.nombre,
      imagen: state.estacion.imagen || '',
      fecha_inicio: state.fechas.entrada,
      fecha_fin: state.fechas.salida,
      tipo_pase: tipo,
      precio_unitario: precio,
      cantidad: cantidad
    });
    return true;
  }

  // ----- Final: pantalla de éxito ---------------------------------------
  function mostrarExito(opciones) {
    var msg = $('[data-success-msg]');
    var goCart = $('[data-go-cart]');
    msg.textContent = opciones.mensaje;
    goCart.hidden = !opciones.mostrarCart;
    showStep(5);
  }

  // ----- Navegación -----------------------------------------------------
  function ir(n) {
    if (n === 1) cargarEstaciones();
    if (n === 2) initFechas();
    if (n === 3) cargarAlojamientos();
    if (n === 4) pintarForfait();
    showStep(n);
  }
  function alSiguiente() {
    if (state.step === 3) {
      // Confirmar reserva primero
      confirmarReserva().then(function (ok) { if (ok) ir(4); });
      return;
    }
    if (state.step === 4) {
      // Añadir forfait al carrito y terminar
      var anadido = anadirForfaitAlCarrito();
      mostrarExito({
        mensaje: 'Tu reserva #' + state.reservaId + ' está confirmada. ' +
                 (anadido ? 'El forfait se ha añadido a tu cesta para que completes el pago.' : ''),
        mostrarCart: anadido
      });
      return;
    }
    ir(state.step + 1);
  }
  function alAtras() { if (state.step > 1) ir(state.step - 1); }

  // ----- Wire up --------------------------------------------------------
  document.addEventListener('DOMContentLoaded', function () {
    $('#wizard-next').addEventListener('click', alSiguiente);
    $('#wizard-prev').addEventListener('click', alAtras);
    $('#wizard-search').addEventListener('input', function () {
      if (state.estacionesCache) renderEstaciones(state.estacionesCache, this.value);
    });
    ir(1);
  });

  // ----- Anti-XSS helpers -----------------------------------------------
  function escapeHTML(s) {
    return String(s == null ? '' : s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
  }
  function escapeAttr(s) {
    return String(s == null ? '' : s).replace(/"/g, '&quot;');
  }
})();
