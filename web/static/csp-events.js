/* csp-events.js
 *
 * Sustituye los antiguos `onclick`, `onerror`, `onsubmit` inline (bloqueados
 * por la CSP que no incluye 'unsafe-inline' en script-src) por delegación
 * via data-attributes. Reglas implementadas:
 *
 *   [data-confirm="texto"]       → al click pide confirm(); si se cancela,
 *                                  preventDefault + stopPropagation.
 *   <img data-broken-parent>     → si la imagen falla, añade la clase al
 *                                  padre. Usar con data-remove-on-error
 *                                  para retirarla del DOM.
 *   <form data-no-submit>        → previene el submit nativo (equivalente
 *                                  a onsubmit="return false;"). Útil para
 *                                  formularios de filtrado controlados por JS.
 */
(function () {
  'use strict';

  document.addEventListener('click', function (event) {
    var el = event.target.closest('[data-confirm]');
    if (!el) return;

    var message = el.getAttribute('data-confirm') || '';
    message = message.replace(/\\n/g, '\n');

    if (!window.confirm(message)) {
      event.preventDefault();
      event.stopPropagation();
    }
  }, true);

  document.addEventListener('error', function (event) {
    var img = event.target;
    if (!(img instanceof HTMLImageElement)) return;

    var parentClass = img.getAttribute('data-broken-parent');
    if (parentClass && img.parentNode && img.parentNode.classList) {
      img.parentNode.classList.add(parentClass);
    }

    if (img.hasAttribute('data-remove-on-error')) {
      img.remove();
    }
  }, true);

  // Forms decorativos (filtros) que NUNCA deben hacer submit nativo.
  document.addEventListener('submit', function (event) {
    var form = event.target;
    if (!(form instanceof HTMLFormElement)) return;
    if (form.hasAttribute('data-no-submit')) {
      event.preventDefault();
    }
  }, true);
})();
