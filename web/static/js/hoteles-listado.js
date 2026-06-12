/* ============================================================
   hoteles-listado.js — mejoras del listado /hoteles

   - El selector de estación re-envía el formulario GET al cambiar
     (con JS desactivado queda el botón "Filtrar" del <noscript>).
   - Sin librerías externas. Script externo por la CSP (script-src 'self').
   ============================================================ */
(function () {
  'use strict';

  var select = document.querySelector('[data-autosubmit]');
  if (!select || !select.form) return;

  select.addEventListener('change', function () {
    select.form.submit();
  });
})();
