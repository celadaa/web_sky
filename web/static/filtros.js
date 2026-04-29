/* ===========================================================
 * SkiHub — Filtros y ordenación cliente.
 *
 * Activa los botones que ya existían en el index (estaciones)
 * y en /noticias. Ambos comparten la clase .filters__btn y se
 * detectan por el id del <ul> contenedor.
 *
 * - Estaciones: ordena las tarjetas in-place según data-*.
 * - Noticias:   muestra/oculta las tarjetas según data-categoria.
 *
 * JS puro, sin dependencias. Toda la animación visual va por
 * classList ("filters__btn--active"); jamás se toca element.style.
 * =========================================================== */

(function () {
  'use strict';

  // -----------------------------------------------------------
  // Estaciones: orden por distancia / nieve / pistas / nieve nueva.
  // -----------------------------------------------------------

  function inicializarFiltrosEstaciones() {
    var grid = document.getElementById('grid-estaciones');
    var nav  = document.getElementById('filtros-estaciones');
    if (!grid || !nav) { return; }

    // Comparadores: cada criterio devuelve un Array.sort callback.
    // Se lee el valor de data-* y se convierte a número.
    var criterios = {
      cercanas: function (a, b) {
        return numero(a, 'distancia') - numero(b, 'distancia');
      },
      nieve: function (a, b) {
        return numero(b, 'nieve') - numero(a, 'nieve');
      },
      pistas: function (a, b) {
        return numero(b, 'pistas') - numero(a, 'pistas');
      },
      recien: function (a, b) {
        // "Recién nevado" = más nieve nueva en las últimas 24 h.
        var dn = numero(b, 'nieve-nueva') - numero(a, 'nieve-nueva');
        if (dn !== 0) { return dn; }
        // Empate: cae a más cercanas para que quede estable.
        return numero(a, 'distancia') - numero(b, 'distancia');
      }
    };

    nav.addEventListener('click', function (ev) {
      var btn = ev.target.closest('button[data-orden]');
      if (!btn) { return; }
      var orden = btn.dataset.orden;
      var cmp = criterios[orden];
      if (!cmp) { return; }

      ordenarHijos(grid, cmp);
      marcarActivo(nav, btn);
    });
  }

  // -----------------------------------------------------------
  // Noticias: filtrado por categoría.
  // -----------------------------------------------------------

  function inicializarFiltrosNoticias() {
    var grid = document.getElementById('grid-noticias');
    var nav  = document.getElementById('filtros-noticias');
    if (!grid || !nav) { return; }

    nav.addEventListener('click', function (ev) {
      var btn = ev.target.closest('button[data-categoria]');
      if (!btn) { return; }
      var cat = btn.dataset.categoria;
      filtrarHijos(grid, cat);
      marcarActivo(nav, btn);
    });
  }

  // -----------------------------------------------------------
  // Helpers
  // -----------------------------------------------------------

  // numero — lee data-<clave> de un <li> y lo devuelve como Number.
  // Devuelve 0 si el dato falta o no es válido.
  function numero(li, clave) {
    var raw = li.getAttribute('data-' + clave);
    var n = parseFloat(raw);
    return isFinite(n) ? n : 0;
  }

  // ordenarHijos — ordena los <li> del contenedor según el comparador
  // y los reinserta en el DOM (no recrea, solo recoloca).
  function ordenarHijos(contenedor, cmp) {
    var hijos = Array.prototype.slice.call(contenedor.children);
    hijos.sort(cmp);
    hijos.forEach(function (li) {
      contenedor.appendChild(li);
    });
  }

  // filtrarHijos — muestra/oculta los <li> según data-categoria.
  // "todas" muestra todos. La visibilidad usa la clase "oculto" (CSS),
  // no se manipula style en línea.
  function filtrarHijos(contenedor, categoria) {
    var hijos = contenedor.children;
    for (var i = 0; i < hijos.length; i++) {
      var li = hijos[i];
      var cat = li.getAttribute('data-categoria') || '';
      if (categoria === 'todas' || cat === categoria) {
        li.classList.remove('oculto');
      } else {
        li.classList.add('oculto');
      }
    }
  }

  // marcarActivo — quita la clase --active a los hermanos y la pone
  // en el botón pulsado.
  function marcarActivo(nav, btn) {
    var botones = nav.querySelectorAll('.filters__btn');
    for (var i = 0; i < botones.length; i++) {
      botones[i].classList.remove('filters__btn--active');
    }
    btn.classList.add('filters__btn--active');
  }

  // -----------------------------------------------------------
  // Arranque
  // -----------------------------------------------------------

  function iniciar() {
    inicializarFiltrosEstaciones();
    inicializarFiltrosNoticias();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', iniciar);
  } else {
    iniciar();
  }
})();
