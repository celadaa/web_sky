/* ============================================================
   animations.js  — específico de /premium  (versión HOTFIX)

   REGLA: el hero y todo el contenido de /premium debe verse
   siempre, aunque este script falle o las librerías externas
   (GSAP/SplitType/Lenis) no carguen.

   Por eso este archivo:
     - NO usa SplitType.
     - NO toca opacity/visibility/display de nada.
     - Solo activa Lenis (smooth scroll) si está disponible.
     - Solo activa el cursor magnético si está en desktop con
       pointer fino y NO sobre inputs/selects/textareas.
     - Reloj y reveal generales ya los hace animations-global.js.

   Si querés volver a meter animaciones del título estilo
   "letra a letra", el patrón correcto es:
     1) JS confirma que SplitType+GSAP están listos.
     2) JS clona el elemento, oculta la copia inline,
        anima la copia, y al final restaura el original.
   Nunca dependas de una clase global tipo .js-ready para
   ocultar nada en CSS.
   ============================================================ */
(function () {
  'use strict';

  var root = document.querySelector('[data-premium-root]');
  if (!root) return;

  var prefersReduce =
    window.matchMedia &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  var isTouch =
    window.matchMedia &&
    window.matchMedia('(hover: none), (pointer: coarse)').matches;

  // ---------- Lenis (opcional, solo en /premium) ----------
  function initLenis() {
    if (prefersReduce) return;
    if (typeof window.Lenis === 'undefined') return;
    try {
      var lenis = new window.Lenis({
        duration: 1.05,
        smoothWheel: true,
        smoothTouch: false,
        prevent: function (node) {
          if (!node || !node.closest) return false;
          return !!node.closest(
            '[data-lenis-prevent], .leaflet-container, .leaflet-popup, .cart-dropdown'
          );
        }
      });
      function raf(t) { lenis.raf(t); requestAnimationFrame(raf); }
      requestAnimationFrame(raf);
      window.SBLenis = lenis;
    } catch (e) {
      /* scroll nativo sigue */
    }
  }

  // ---------- Cursor magnético (opcional, solo desktop) ----------
  function shouldDisableCursor(target) {
    if (!target) return true;
    if (target.closest && target.closest('.no-magnetic, [data-no-magnetic]')) return true;
    var tag = (target.tagName || '').toLowerCase();
    if (tag === 'input' || tag === 'select' || tag === 'textarea' || tag === 'option') return true;
    if (target.closest && target.closest('.leaflet-container')) return true;
    return false;
  }

  function initCursor() {
    if (prefersReduce || isTouch) return;

    // Inyectamos estilos mínimos del cursor para no depender de CSS
    if (!document.getElementById('sb-cursor-base-styles')) {
      var style = document.createElement('style');
      style.id = 'sb-cursor-base-styles';
      style.textContent =
        '.sb-cursor{position:fixed;top:0;left:0;width:14px;height:14px;border-radius:50%;' +
        'background:#0a0a0a;pointer-events:none;z-index:9999;transform:translate(-50%,-50%);' +
        'opacity:0;mix-blend-mode:difference;filter:invert(1);' +
        'transition:width .25s,height .25s,background .25s,opacity .2s}' +
        '.sb-cursor.is-visible{opacity:1}' +
        '.sb-cursor.is-hover{width:48px;height:48px;background:#e8ff45}' +
        '.sb-cursor.is-suppressed{opacity:0}';
      document.head.appendChild(style);
    }

    var cursor = document.createElement('div');
    cursor.className = 'sb-cursor';
    cursor.setAttribute('aria-hidden', 'true');
    document.body.appendChild(cursor);

    var x = window.innerWidth / 2, y = window.innerHeight / 2;
    var tx = x, ty = y;

    document.addEventListener('mousemove', function (e) {
      tx = e.clientX; ty = e.clientY;
      cursor.classList.add('is-visible');
    }, { passive: true });

    (function loop() {
      x += (tx - x) * 0.22;
      y += (ty - y) * 0.22;
      cursor.style.transform = 'translate(' + x + 'px,' + y + 'px) translate(-50%,-50%)';
      requestAnimationFrame(loop);
    })();

    var hoverSel = 'a, button, [role="button"], [data-magnetic], .sb-btn, .sb-product__cta';
    document.addEventListener('mouseover', function (e) {
      if (shouldDisableCursor(e.target)) {
        cursor.classList.add('is-suppressed');
        cursor.classList.remove('is-hover');
        return;
      }
      cursor.classList.remove('is-suppressed');
      if (e.target.closest && e.target.closest(hoverSel)) {
        cursor.classList.add('is-hover');
      }
    });
    document.addEventListener('mouseout', function (e) {
      if (e.target.closest && e.target.closest(hoverSel)) {
        cursor.classList.remove('is-hover');
      }
    });
  }

  // ---------- Reloj del topbar ----------
  // El reloj global ya lo hace animations-global.js. Este script
  // se ejecuta DESPUÉS de él, así que no toca [data-clock].

  function init() {
    try { initLenis();  } catch (e) { /* nunca bloquear */ }
    try { initCursor(); } catch (e) { /* nunca bloquear */ }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
