(function () {
  'use strict';

  function inicializarFiltrosEstaciones() {
    var grid = document.getElementById('grid-estaciones');
    var nav = document.getElementById('filtros-estaciones');
    if (!grid || !nav) { return; }

    var criterios = {
      cercanas: function (a, b) { return numero(a, 'distancia') - numero(b, 'distancia'); },
      nieve: function (a, b) { return numero(b, 'nieve') - numero(a, 'nieve'); },
      pistas: function (a, b) { return numero(b, 'pistas') - numero(a, 'pistas'); },
      recien: function (a, b) {
        var dn = numero(b, 'nieve-nueva') - numero(a, 'nieve-nueva');
        if (dn !== 0) { return dn; }
        return numero(a, 'distancia') - numero(b, 'distancia');
      }
    };

    nav.addEventListener('click', function (ev) {
      var btn = ev.target.closest('button[data-orden]');
      if (!btn) { return; }
      var cmp = criterios[btn.dataset.orden];
      if (!cmp) { return; }
      ordenarHijos(grid, cmp);
      marcarActivo(nav, btn, '.filters__btn');
    });
  }

  function inicializarFiltrosNoticias() {
    var grid = document.getElementById('grid-noticias');
    var nav = document.getElementById('filtros-noticias');
    if (!grid || !nav) { return; }

    nav.addEventListener('click', function (ev) {
      var btn = ev.target.closest('button[data-categoria]');
      if (!btn) { return; }
      filtrarHijos(grid, btn.dataset.categoria);
      marcarActivo(nav, btn, '.filters__btn');
    });
  }

  function numero(li, clave) {
    var raw = li.getAttribute('data-' + clave);
    var n = parseFloat(raw);
    return isFinite(n) ? n : 0;
  }

  function ordenarHijos(contenedor, cmp) {
    Array.prototype.slice.call(contenedor.children).sort(cmp).forEach(function (li) {
      contenedor.appendChild(li);
    });
  }

  function filtrarHijos(contenedor, categoria) {
    Array.prototype.forEach.call(contenedor.children, function (li) {
      var cat = li.getAttribute('data-categoria') || '';
      li.classList.toggle('oculto', categoria !== 'todas' && cat !== categoria);
    });
  }

  function marcarActivo(nav, btn, selector) {
    Array.prototype.forEach.call(nav.querySelectorAll(selector), function (item) {
      item.classList.remove('filters__btn--active');
      item.setAttribute('aria-pressed', 'false');
    });
    btn.classList.add('filters__btn--active');
    btn.setAttribute('aria-pressed', 'true');
  }

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
