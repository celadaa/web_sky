(function () {
  'use strict';

  var COOKIE_NAME = 'skihub_csrf';
  var FIELD_NAME = 'csrf_token';
  var HEADER_NAME = 'X-CSRF-Token';
  var UNSAFE = { POST: true, PUT: true, PATCH: true, DELETE: true };

  function csrfToken() {
    return document.cookie
      .split(';')
      .map(function (part) { return part.trim(); })
      .filter(function (part) { return part.indexOf(COOKIE_NAME + '=') === 0; })
      .map(function (part) { return decodeURIComponent(part.slice(COOKIE_NAME.length + 1)); })[0] || '';
  }

  function sameOrigin(url) {
    try {
      return new URL(url, window.location.href).origin === window.location.origin;
    } catch (_) {
      return false;
    }
  }

  document.addEventListener('submit', function (event) {
    var form = event.target;
    if (!(form instanceof HTMLFormElement)) return;

    var method = (form.getAttribute('method') || 'GET').toUpperCase();
    var action = form.getAttribute('action') || window.location.href;

    if (!UNSAFE[method] || !sameOrigin(action)) return;

    var token = csrfToken();
    if (!token) return;

    var input = form.querySelector('input[name="' + FIELD_NAME + '"]');
    if (!input) {
      input = document.createElement('input');
      input.type = 'hidden';
      input.name = FIELD_NAME;
      form.appendChild(input);
    }
    input.value = token;
  }, true);

  if (window.fetch) {
    var originalFetch = window.fetch;
    window.fetch = function (input, init) {
      init = init || {};

      var url = typeof input === 'string' ? input : input.url;
      var method = (init.method || (typeof input !== 'string' && input.method) || 'GET').toUpperCase();

      if (UNSAFE[method] && sameOrigin(url)) {
        var token = csrfToken();
        if (token) {
          var headers = new Headers(init.headers || (typeof input !== 'string' ? input.headers : undefined));
          if (!headers.has(HEADER_NAME)) {
            headers.set(HEADER_NAME, token);
          }
          init = Object.assign({}, init, { headers: headers });
        }
      }

      return originalFetch(input, init);
    };
  }
})();
