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
})();
