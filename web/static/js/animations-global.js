/* ============================================================
   animations-global.js  (versión HOTFIX, mínima y segura)

   REGLA INNEGOCIABLE:
   Ningún contenido depende de este script para verse. Si este
   archivo falla, no se carga, o las librerías externas no están,
   la web sigue siendo 100% visible y usable.

   Lo único que hace esta capa:
     1. Reloj para cualquier [data-clock] del documento.
     2. Reveal sutil opt-in para [data-reveal] vía IntersectionObserver.
        (El elemento ya es visible; al entrar en pantalla se le añade
         la clase .reveal-in que dispara una pequeña animación CSS
         de translateY 20px → 0. Si JS falla, no pasa nada.)
   ============================================================ */
(function () {
  'use strict';

  function ready(fn) {
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', fn);
    } else {
      fn();
    }
  }

  // ---- Reloj ----
  function initClock() {
    var els = document.querySelectorAll('[data-clock]');
    if (!els.length) return;
    function pad(n) { return n < 10 ? '0' + n : '' + n; }
    function tick() {
      var d = new Date();
      var s = pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds());
      for (var i = 0; i < els.length; i++) els[i].textContent = s;
    }
    tick();
    setInterval(tick, 1000);
  }

  // ---- Reveal opt-in (NO oculta nada, solo añade clase tras IO) ----
  function initReveals() {
    if (!('IntersectionObserver' in window)) return;

    var prefersReduce =
      window.matchMedia &&
      window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (prefersReduce) return;

    var els = document.querySelectorAll('[data-reveal]');
    if (!els.length) return;

    var io = new IntersectionObserver(function (entries) {
      for (var i = 0; i < entries.length; i++) {
        var e = entries[i];
        if (e.isIntersecting) {
          e.target.classList.add('reveal-in');
          io.unobserve(e.target);
        }
      }
    }, { threshold: 0.1, rootMargin: '0px 0px -8% 0px' });

    for (var i = 0; i < els.length; i++) io.observe(els[i]);
  }

  ready(function () {
    try { initClock();   } catch (e) { /* never block render */ }
    try { initReveals(); } catch (e) { /* never block render */ }
  });

  // API mínima (vacía a propósito; las animaciones grandes
  // viven en archivos por página si se necesitan).
  window.SBAnim = { version: 'hotfix-1' };
})();
