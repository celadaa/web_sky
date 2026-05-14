/* Snowbreak — Caja de herramientas de estaciones
 *
 * Centraliza la lógica de:
 *   - Recalcular distancias en cliente cuando cambia la ubicación.
 *   - Filtros avanzados (distancia, precio, abiertas, nieve, país, dificultad, pistas).
 *   - Ordenación (cercana, nieve, barata, pistas, valoración).
 *   - Comparador de hasta 3 estaciones (modal).
 *   - Favoritos: para usuarios logueados usa el endpoint existente
 *     /favorito/toggle; para invitados guarda en localStorage y
 *     muestra un mensaje invitando a iniciar sesión.
 *
 * Trabaja sobre cualquier <ul data-grid-estaciones> cuyos <li> tengan
 * los data-attributes esperados (ver estaciones.tmpl / index.tmpl).
 */
(function () {
  'use strict';

  var FALLBACK = { lat: 40.4168, lng: -3.7038, etiqueta: 'Madrid', fuente: 'preset' };
  var STORAGE_COMPARADOR = 'snowbreak.comparador.v1';
  var STORAGE_FAVS_INVITADO = 'snowbreak.favoritos.invitado.v1';

  var estado = {
    ubicacion: leerUbicacion(),
    filtros: filtrosVacios(),
    orden: 'cercana',
    comparador: leerComparador()
  };

  function init() {
    var grids = Array.prototype.slice.call(document.querySelectorAll('[data-grid-estaciones]'));
    if (!grids.length) { return; }

    grids.forEach(prepararGrid);
    inicializarPanelFiltros();
    inicializarSelectorOrden();
    inicializarComparadorBar();
    inicializarFavoritos();
    inicializarTostadas();

    // Cuando cambia la ubicación, recalculamos distancias y reaplicamos.
    document.addEventListener('snowbreak:ubicacion:cambio', function (ev) {
      estado.ubicacion = ev.detail || FALLBACK;
      grids.forEach(recalcularDistancias);
      aplicar();
    });

    // Primer pase para que coincidan las cards con la ubicación inicial.
    grids.forEach(recalcularDistancias);
    aplicar();
  }

  /* ------------------------------ Estado ------------------------------ */

  function filtrosVacios() {
    return {
      distanciaMax: '',
      precioMax:    '',
      soloAbiertas: false,
      soloFavs:     false,
      nieveMin:     '',
      pais:         '',
      dificultad:   '',
      pistasMin:    ''
    };
  }

  function leerUbicacion() {
    if (window.SnowbreakGeo && SnowbreakGeo.leerPreferencia) {
      var p = SnowbreakGeo.leerPreferencia();
      if (p && typeof p.lat === 'number') {
        return { lat: p.lat, lng: p.lng, etiqueta: p.etiqueta || 'Tu ubicación' };
      }
    }
    return Object.assign({}, FALLBACK);
  }

  /* ----------------------------- Distancia ---------------------------- */

  function prepararGrid(grid) {
    var items = Array.prototype.slice.call(grid.querySelectorAll('[data-estacion]'));
    items.forEach(function (li) {
      // Si ya hay coords del template, perfecto; si no, intentamos lookup.
      var lat = parseFloat(li.getAttribute('data-lat'));
      var lng = parseFloat(li.getAttribute('data-lng'));
      if (!isFinite(lat) || !isFinite(lng)) {
        var coords = window.SnowbreakCoords && SnowbreakCoords.lookup(li.getAttribute('data-nombre') || '');
        if (coords) {
          li.setAttribute('data-lat', coords.lat);
          li.setAttribute('data-lng', coords.lng);
          if (!li.getAttribute('data-pais') && coords.pais) {
            li.setAttribute('data-pais', coords.pais);
          }
        }
      }
    });
  }

  function recalcularDistancias(grid) {
    if (!window.SnowbreakCoords || !estado.ubicacion) { return; }
    var origen = { lat: estado.ubicacion.lat, lng: estado.ubicacion.lng };
    var items = grid.querySelectorAll('[data-estacion]');
    items.forEach(function (li) {
      var lat = parseFloat(li.getAttribute('data-lat'));
      var lng = parseFloat(li.getAttribute('data-lng'));
      if (!isFinite(lat) || !isFinite(lng)) { return; }
      var km = SnowbreakCoords.distancia(origen, { lat: lat, lng: lng });
      if (km == null || !isFinite(km)) { return; }
      li.setAttribute('data-distancia', km.toFixed(1));
      // Actualizamos los textos visibles si existen.
      var badge = li.querySelector('[data-distancia-badge]');
      if (badge) { badge.textContent = Math.round(km) + ' km'; }
      var metric = li.querySelector('[data-distancia-metric]');
      if (metric) { metric.textContent = Math.round(km) + ' km'; }
    });
  }

  /* ------------------------------ Filtros ----------------------------- */

  function inicializarPanelFiltros() {
    var panel = document.getElementById('panel-filtros');
    if (!panel) { return; }

    // Drawer en móvil.
    var btnAbrir = document.getElementById('btn-abrir-filtros');
    if (btnAbrir) {
      btnAbrir.addEventListener('click', function () {
        panel.classList.toggle('panel-filtros--open');
        var abierto = panel.classList.contains('panel-filtros--open');
        btnAbrir.setAttribute('aria-expanded', abierto ? 'true' : 'false');
      });
    }
    var btnCerrar = panel.querySelector('[data-cerrar-filtros]');
    if (btnCerrar) {
      btnCerrar.addEventListener('click', function () {
        panel.classList.remove('panel-filtros--open');
        if (btnAbrir) { btnAbrir.setAttribute('aria-expanded', 'false'); }
      });
    }

    panel.addEventListener('input', recogerYAplicar);
    panel.addEventListener('change', recogerYAplicar);

    var btnLimpiar = panel.querySelector('[data-limpiar-filtros]');
    if (btnLimpiar) {
      btnLimpiar.addEventListener('click', function () {
        estado.filtros = filtrosVacios();
        // Reset de inputs visibles.
        panel.querySelectorAll('input, select').forEach(function (el) {
          if (el.type === 'checkbox') { el.checked = false; }
          else { el.value = ''; }
        });
        aplicar();
      });
    }

    function recogerYAplicar() {
      estado.filtros.distanciaMax = leerNumero(panel, '[name="filtro-distancia"]');
      estado.filtros.precioMax    = leerNumero(panel, '[name="filtro-precio"]');
      estado.filtros.nieveMin     = leerNumero(panel, '[name="filtro-nieve"]');
      estado.filtros.pistasMin    = leerNumero(panel, '[name="filtro-pistas"]');
      estado.filtros.pais         = leerValor(panel,  '[name="filtro-pais"]');
      estado.filtros.dificultad   = leerValor(panel,  '[name="filtro-dificultad"]');
      var ab = panel.querySelector('[name="filtro-abiertas"]');
      estado.filtros.soloAbiertas = !!(ab && ab.checked);
      var favs = panel.querySelector('[name="filtro-favoritas"]');
      estado.filtros.soloFavs = !!(favs && favs.checked);
      aplicar();
    }
  }

  function leerNumero(root, sel) {
    var el = root.querySelector(sel);
    if (!el) { return ''; }
    var n = parseFloat(el.value);
    return isFinite(n) ? n : '';
  }
  function leerValor(root, sel) {
    var el = root.querySelector(sel);
    return el ? (el.value || '') : '';
  }

  /* ------------------------------ Orden ------------------------------- */

  function inicializarSelectorOrden() {
    var sel = document.getElementById('selector-orden');
    if (!sel) { return; }
    sel.value = estado.orden;
    sel.addEventListener('change', function () {
      estado.orden = sel.value;
      aplicar();
    });
  }

  /* ----------------------- Aplicar filtros + orden -------------------- */

  function aplicar() {
    var grids = document.querySelectorAll('[data-grid-estaciones]');
    var totalVisibles = 0, totalRestantes = 0;
    grids.forEach(function (grid) {
      var items = Array.prototype.slice.call(grid.children);
      var visibles = items.filter(function (li) { return pasaFiltros(li); });
      totalVisibles += visibles.length;
      totalRestantes += items.length;

      // Aplicar orden a los visibles.
      visibles.sort(comparadorOrden(estado.orden));

      // Reordenamos en el DOM y ocultamos los que no pasan filtros.
      visibles.forEach(function (li) { grid.appendChild(li); });
      items.forEach(function (li) {
        var visible = visibles.indexOf(li) !== -1;
        li.classList.toggle('oculto', !visible);
      });
    });

    actualizarContador(totalVisibles, totalRestantes);
    actualizarEstadoVacio(totalVisibles);
  }

  function pasaFiltros(li) {
    var f = estado.filtros;
    var distancia = num(li, 'data-distancia');
    var precio    = num(li, 'data-precio');
    var nieve     = num(li, 'data-nieve');
    var abiertas  = num(li, 'data-pistas-abiertas');
    var totales   = num(li, 'data-pistas-totales');
    var pais      = (li.getAttribute('data-pais') || '').toLowerCase();
    var dif       = (li.getAttribute('data-dificultad') || '').toLowerCase();
    var fav       = li.getAttribute('data-favorita') === '1';

    if (f.distanciaMax !== '' && distancia > f.distanciaMax) { return false; }
    if (f.precioMax    !== '' && precio    > f.precioMax)    { return false; }
    if (f.nieveMin     !== '' && nieve     < f.nieveMin)     { return false; }
    if (f.pistasMin    !== '' && abiertas  < f.pistasMin)    { return false; }
    if (f.soloAbiertas && abiertas <= 0) { return false; }
    if (f.pais       && pais.indexOf(f.pais.toLowerCase()) === -1) { return false; }
    if (f.dificultad && dif  !== f.dificultad.toLowerCase())       { return false; }
    if (f.soloFavs   && !fav) { return false; }
    return true;
  }

  function comparadorOrden(orden) {
    return function (a, b) {
      switch (orden) {
        case 'nieve':    return num(b, 'data-nieve') - num(a, 'data-nieve');
        case 'precio':   return seguro(num(a, 'data-precio'), 99999) - seguro(num(b, 'data-precio'), 99999);
        case 'pistas':   return num(b, 'data-pistas-abiertas') - num(a, 'data-pistas-abiertas');
        case 'rating':   return num(b, 'data-rating') - num(a, 'data-rating');
        case 'cercana':
        default:         return seguro(num(a, 'data-distancia'), 99999) - seguro(num(b, 'data-distancia'), 99999);
      }
    };
  }

  function num(li, attr) {
    var n = parseFloat(li.getAttribute(attr));
    return isFinite(n) ? n : 0;
  }
  function seguro(v, fallback) { return (v === 0 || isNaN(v)) ? fallback : v; }

  function actualizarContador(visibles, total) {
    var c = document.querySelectorAll('[data-resultados-contador]');
    c.forEach(function (el) {
      el.textContent = visibles + ' ' + (visibles === 1 ? 'estación' : 'estaciones') + ' encontradas';
    });
  }

  function actualizarEstadoVacio(visibles) {
    var vacio = document.getElementById('estaciones-vacio');
    if (!vacio) { return; }
    vacio.hidden = visibles > 0;
  }

  /* --------------------------- Comparador ----------------------------- */

  function leerComparador() {
    try {
      var raw = localStorage.getItem(STORAGE_COMPARADOR);
      if (!raw) { return []; }
      var arr = JSON.parse(raw);
      if (!Array.isArray(arr)) { return []; }
      return arr.slice(0, 3);
    } catch (e) { return []; }
  }
  function guardarComparador() {
    try { localStorage.setItem(STORAGE_COMPARADOR, JSON.stringify(estado.comparador)); }
    catch (e) {}
  }

  function inicializarComparadorBar() {
    pintarComparadorBar();

    document.body.addEventListener('click', function (ev) {
      var btn = ev.target.closest('[data-toggle-comparar]');
      if (btn) {
        ev.preventDefault();
        var li = btn.closest('[data-estacion]');
        if (li) { toggleComparar(extraerInfo(li)); }
        return;
      }
      var abrir = ev.target.closest('[data-abrir-comparador]');
      if (abrir) {
        ev.preventDefault();
        abrirComparador();
      }
      var vaciar = ev.target.closest('[data-vaciar-comparador]');
      if (vaciar) {
        ev.preventDefault();
        estado.comparador = [];
        guardarComparador();
        pintarComparadorBar();
        sincronizarBotones();
      }
    });
  }

  function extraerInfo(li) {
    return {
      id:        li.getAttribute('data-id'),
      nombre:    li.getAttribute('data-nombre'),
      ubicacion: li.getAttribute('data-ubicacion') || '',
      pais:      li.getAttribute('data-pais') || '',
      distancia: parseFloat(li.getAttribute('data-distancia')) || null,
      precio:    parseFloat(li.getAttribute('data-precio')) || null,
      nieve:     parseFloat(li.getAttribute('data-nieve')) || 0,
      nieveMin:  parseFloat(li.getAttribute('data-nieve-min')) || 0,
      nieveMax:  parseFloat(li.getAttribute('data-nieve-max')) || 0,
      nieveNueva: parseFloat(li.getAttribute('data-nieve-nueva')) || 0,
      viento:    li.getAttribute('data-viento') || '',
      kmEsq:     parseFloat(li.getAttribute('data-km-esquiables')) || null,
      pistasAbiertas: parseInt(li.getAttribute('data-pistas-abiertas'), 10) || 0,
      pistasTotales:  parseInt(li.getAttribute('data-pistas-totales'), 10) || 0,
      dificultad: li.getAttribute('data-dificultad') || '',
      rating:    parseFloat(li.getAttribute('data-rating')) || null,
      parteHora: li.getAttribute('data-parte-hora') || '',
      parteFecha: li.getAttribute('data-parte-fecha') || '',
      url:       li.getAttribute('data-url') || ('/estacion/' + li.getAttribute('data-id'))
    };
  }

  function toggleComparar(info) {
    if (!info || !info.id) { return; }
    var i = estado.comparador.findIndex(function (e) { return String(e.id) === String(info.id); });
    if (i !== -1) {
      estado.comparador.splice(i, 1);
      tostada('Quitado del comparador');
    } else {
      if (estado.comparador.length >= 3) {
        tostada('Máximo 3 estaciones en el comparador');
        return;
      }
      estado.comparador.push(info);
      tostada('Añadido al comparador');
    }
    guardarComparador();
    pintarComparadorBar();
    sincronizarBotones();
  }

  function pintarComparadorBar() {
    var bar = document.getElementById('comparador-bar');
    if (!bar) { return; }
    var n = estado.comparador.length;
    if (n === 0) {
      bar.hidden = true;
      bar.innerHTML = '';
      return;
    }
    bar.hidden = false;
    bar.innerHTML =
      '<div class="comparador-bar__chips">' +
        estado.comparador.map(function (e) {
          return '<span class="comparador-bar__chip">' + escapar(e.nombre) +
                 '<button type="button" data-quitar="' + escapar(e.id) + '" aria-label="Quitar ' + escapar(e.nombre) + '">×</button></span>';
        }).join('') +
      '</div>' +
      '<div class="comparador-bar__actions">' +
        '<button type="button" class="comparador-bar__btn" data-abrir-comparador>Comparar (' + n + ')</button>' +
        '<button type="button" class="comparador-bar__btn comparador-bar__btn--ghost" data-vaciar-comparador>Vaciar</button>' +
      '</div>';

    bar.querySelectorAll('[data-quitar]').forEach(function (b) {
      b.addEventListener('click', function () {
        var id = b.getAttribute('data-quitar');
        estado.comparador = estado.comparador.filter(function (e) { return String(e.id) !== String(id); });
        guardarComparador();
        pintarComparadorBar();
        sincronizarBotones();
      });
    });
  }

  function sincronizarBotones() {
    var ids = estado.comparador.map(function (e) { return String(e.id); });
    document.querySelectorAll('[data-toggle-comparar]').forEach(function (b) {
      var li = b.closest('[data-estacion]');
      if (!li) { return; }
      var activo = ids.indexOf(li.getAttribute('data-id')) !== -1;
      b.classList.toggle('is-active', activo);
      b.setAttribute('aria-pressed', activo ? 'true' : 'false');
      var lbl = b.querySelector('[data-comparar-label]');
      if (lbl) { lbl.textContent = activo ? 'En comparador' : 'Comparar'; }
    });
  }

  function abrirComparador() {
    var modal = document.createElement('div');
    modal.className = 'comparador-modal';
    modal.setAttribute('role', 'dialog');
    modal.setAttribute('aria-modal', 'true');
    modal.setAttribute('aria-label', 'Comparador de estaciones');
    if (estado.comparador.length < 1) {
      modal.innerHTML =
        '<div class="comparador-modal__card comparador-modal__card--empty">' +
          '<header class="comparador-modal__header">' +
            '<h2>Comparador de estaciones</h2>' +
            '<button type="button" class="comparador-modal__close" data-cerrar aria-label="Cerrar">×</button>' +
          '</header>' +
          '<p>Selecciona 2 o 3 estaciones para compararlas.</p>' +
        '</div>';
    } else {
      modal.innerHTML =
        '<div class="comparador-modal__card">' +
          '<header class="comparador-modal__header">' +
            '<h2>Comparador de estaciones</h2>' +
            '<button type="button" class="comparador-modal__close" data-cerrar aria-label="Cerrar">×</button>' +
          '</header>' +
          '<p class="comparador-modal__hint">Máximo 3 estaciones — añade o quita desde cualquier ficha.</p>' +
          tablaComparador(estado.comparador) +
          '<p class="parte__disclaimer parte__disclaimer--inline">Datos orientativos para demostración académica.</p>' +
          '<footer class="comparador-modal__footer">' +
            '<button type="button" class="ubic-modal__submit" data-vaciar-comparador>Vaciar comparador</button>' +
          '</footer>' +
        '</div>';
    }
    document.body.appendChild(modal);
    document.body.classList.add('snow-no-scroll');
    modal.querySelector('[data-cerrar]').addEventListener('click', cerrar);
    modal.addEventListener('click', function (ev) { if (ev.target === modal) { cerrar(); } });
    document.addEventListener('keydown', escListener);
    function escListener(ev) { if (ev.key === 'Escape') { cerrar(); } }
    function cerrar() {
      document.removeEventListener('keydown', escListener);
      modal.remove();
      document.body.classList.remove('snow-no-scroll');
    }
    // botones quitar dentro de la tabla
    modal.querySelectorAll('[data-quitar-celda]').forEach(function (b) {
      b.addEventListener('click', function () {
        var id = b.getAttribute('data-quitar-celda');
        estado.comparador = estado.comparador.filter(function (e) { return String(e.id) !== String(id); });
        guardarComparador();
        pintarComparadorBar();
        sincronizarBotones();
        cerrar();
        if (estado.comparador.length > 0) { abrirComparador(); }
      });
    });
  }

  function tablaComparador(estaciones) {
    var filas = [
      { etiq: 'Distancia',     fmt: function (e) { return e.distancia != null ? Math.round(e.distancia) + ' km' : '—'; } },
      { etiq: 'Precio desde',  fmt: function (e) { return e.precio != null ? Math.round(e.precio) + ' €' : '—'; } },
      { etiq: 'Nieve mín/máx', fmt: function (e) { return (e.nieveMin || e.nieveMax) ? e.nieveMin + '–' + e.nieveMax + ' cm' : '—'; } },
      { etiq: 'Nieve reciente',fmt: function (e) { return e.nieveNueva ? e.nieveNueva + ' cm' : '0 cm'; } },
      { etiq: 'Viento',        fmt: function (e) { return e.viento || '—'; } },
      { etiq: 'Km esquiables', fmt: function (e) { return e.kmEsq != null ? e.kmEsq + ' km' : '—'; } },
      { etiq: 'Pistas abiertas', fmt: function (e) { return e.pistasAbiertas + '/' + e.pistasTotales; } },
      { etiq: 'Estado',        fmt: function (e) { return etiquetaEstado(e); } },
      { etiq: 'País',          fmt: function (e) { return e.pais || '—'; } },
      { etiq: 'Dificultad',    fmt: function (e) { return e.dificultad || '—'; } },
      { etiq: 'Valoración',    fmt: function (e) { return e.rating ? e.rating.toFixed(1) + ' / 5' : '—'; } },
      { etiq: 'Parte',         fmt: function (e) { return e.parteHora ? 'Actualizado ' + e.parteHora : '—'; } }
    ];
    var thead = '<th scope="col">Dato</th>' +
      estaciones.map(function (e) {
        return '<th scope="col">' +
                 '<div class="comparador-modal__th">' +
                   '<a href="' + escapar(e.url) + '" class="text-link">' + escapar(e.nombre) + '</a>' +
                   '<button type="button" class="comparador-modal__quitar" data-quitar-celda="' + escapar(e.id) + '" aria-label="Quitar ' + escapar(e.nombre) + ' del comparador">Quitar</button>' +
                 '</div>' +
               '</th>';
      }).join('');
    var tbody = filas.map(function (f) {
      return '<tr><th scope="row">' + f.etiq + '</th>' +
        estaciones.map(function (e) { return '<td>' + escapar(f.fmt(e)) + '</td>'; }).join('') +
        '</tr>';
    }).join('');
    return '<div class="comparador-modal__scroll"><table class="comparador-tabla"><thead><tr>' + thead + '</tr></thead><tbody>' + tbody + '</tbody></table></div>';
  }

  function etiquetaEstado(e) {
    if (e.pistasAbiertas <= 0) { return 'Cerrada'; }
    // Mismo umbral que el modelo Go: ≥80% abiertas → "Abierta".
    if (e.pistasTotales > 0 && e.pistasAbiertas * 5 >= e.pistasTotales * 4) {
      return 'Abierta';
    }
    return 'Apertura parcial';
  }

  /* ----------------------------- Favoritos ---------------------------- */

  function inicializarFavoritos() {
    document.body.addEventListener('click', function (ev) {
      var btn = ev.target.closest('[data-toggle-favorita]');
      if (!btn) { return; }

      var logueado = btn.getAttribute('data-logueado') === '1';
      if (logueado) {
        // Confiamos en la submission natural del <form> que envuelve al
        // botón. Mostramos una tostada inmediata; el cuajado real ocurre
        // tras el redirect del backend.
        tostada(btn.classList.contains('is-active')
          ? 'Estación eliminada de favoritos'
          : 'Estación añadida a favoritos');
        return;
      }

      // Invitado — interceptamos y mostramos modal.
      ev.preventDefault();
      mostrarModalLoginFav();
    });
  }

  function mostrarModalLoginFav() {
    var modal = document.createElement('div');
    modal.className = 'login-fav-modal';
    modal.setAttribute('role', 'dialog');
    modal.setAttribute('aria-modal', 'true');
    modal.setAttribute('aria-label', 'Inicia sesión para guardar favoritas');
    modal.innerHTML =
      '<div class="login-fav-modal__card">' +
        '<button type="button" class="ubic-modal__close" data-cerrar aria-label="Cerrar">×</button>' +
        '<h2>Inicia sesión para guardar tus estaciones favoritas</h2>' +
        '<p>Crea una cuenta para tener tus favoritas siempre a mano, en cualquier dispositivo.</p>' +
        '<div class="login-fav-modal__actions">' +
          '<a href="/login" class="ubic-modal__submit">Iniciar sesión</a>' +
          '<a href="/registro" class="ubic-modal__submit ubic-modal__submit--ghost">Crear cuenta</a>' +
        '</div>' +
      '</div>';
    document.body.appendChild(modal);
    document.body.classList.add('snow-no-scroll');
    modal.addEventListener('click', function (ev) { if (ev.target === modal) { cerrar(); } });
    modal.querySelector('[data-cerrar]').addEventListener('click', cerrar);
    document.addEventListener('keydown', escListener);
    function escListener(ev) { if (ev.key === 'Escape') { cerrar(); } }
    function cerrar() {
      document.removeEventListener('keydown', escListener);
      modal.remove();
      document.body.classList.remove('snow-no-scroll');
    }
  }

  /* ------------------------------ Tostadas ---------------------------- */

  function inicializarTostadas() {
    if (document.getElementById('snow-toast')) { return; }
    var div = document.createElement('div');
    div.id = 'snow-toast';
    div.className = 'snow-toast';
    div.setAttribute('role', 'status');
    div.setAttribute('aria-live', 'polite');
    document.body.appendChild(div);
  }

  var tostadaTimer = null;
  function tostada(msg) {
    var el = document.getElementById('snow-toast');
    if (!el) { return; }
    el.textContent = msg;
    el.classList.add('snow-toast--visible');
    clearTimeout(tostadaTimer);
    tostadaTimer = setTimeout(function () { el.classList.remove('snow-toast--visible'); }, 2400);
  }

  /* ------------------------------ Utils ------------------------------- */

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
