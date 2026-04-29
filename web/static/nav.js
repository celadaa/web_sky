(function () {
  'use strict';

  function initNav() {
    var toggle = document.querySelector('.navbar__toggle');
    var nav = document.getElementById('navbar-nav');
    if (!toggle || !nav) { return; }

    toggle.addEventListener('click', function () {
      var expanded = toggle.getAttribute('aria-expanded') === 'true';
      toggle.setAttribute('aria-expanded', expanded ? 'false' : 'true');
      nav.classList.toggle('is-open', !expanded);
    });

    document.addEventListener('click', function (ev) {
      if (!nav.classList.contains('is-open')) { return; }
      if (nav.contains(ev.target) || toggle.contains(ev.target)) { return; }
      toggle.setAttribute('aria-expanded', 'false');
      nav.classList.remove('is-open');
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initNav);
  } else {
    initNav();
  }
})();
