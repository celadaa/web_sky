/* ============================================================
   animations-global.js  (post-hotfix, mínimo y seguro)

   REGLA INNEGOCIABLE:
   Ningún contenido depende de este script para verse. Si este
   archivo falla, no se carga, o las librerías externas no están,
   la web sigue siendo 100% visible y usable.

   Solo:
     1. Reloj [data-clock]
     2. Reveal opt-in [data-reveal] vía IntersectionObserver
        (CSS añade transición a .reveal-in; base visible siempre)
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

  function initReveals() {
    if (!('IntersectionObserver' in window)) return;
    var prefersReduce = window.matchMedia &&
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
    try { initClock();   } catch (e) {}
    try { initReveals(); } catch (e) {}
  });

  window.SBAnim = { version: 'hotfix-1' };
})();
