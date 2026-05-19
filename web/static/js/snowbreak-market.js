/* ============================================================
   snowbreak-market.js
   Renderiza la sección "Snowbreak Market" desde una lista de
   productos definida aquí. Lógica de carrito vive en cart.js.

   Arquitectura preparada para Three.js: cada producto declara
   `model` (ruta a un .glb) y el div .sb-product__viz expone
   data-model="ski-basic.glb". Cuando integremos Three.js, este
   div se sustituye por un <canvas> sin tocar nada del resto.
   ============================================================ */
(function () {
  'use strict';

  // ---------- Datos ----------
  // Iconos SVG en línea para no depender de assets externos.
  // (sustituibles 1:1 por modelos 3D)
  var ICONS = {
    ski: '<svg viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M14 50 L48 14"/><path d="M18 50 L52 14"/><path d="M12 52 L22 52"/><path d="M44 16 L54 16"/><path d="M16 46 L20 42"/><path d="M44 22 L48 18"/></svg>',
    premium: '<svg viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M10 52 L52 10"/><path d="M14 52 L56 10"/><path d="M8 54 L20 54"/><path d="M48 12 L60 12"/><path d="M22 40 L26 36"/><path d="M40 22 L44 18"/><circle cx="48" cy="48" r="4"/></svg>',
    board: '<svg viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M20 8 C 36 8, 56 28, 56 44 C 56 52, 50 56, 44 56 C 28 56, 8 36, 8 20 C 8 12, 14 8, 20 8 Z"/><path d="M24 18 L40 34"/></svg>',
    kid: '<svg viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M18 52 L42 16"/><path d="M22 52 L46 16"/><circle cx="32" cy="14" r="4"/><path d="M20 28 L28 32 L24 40"/></svg>',
    helmet: '<svg viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M14 38 C 14 22, 24 14, 32 14 C 40 14, 50 22, 50 38 L 50 44 L 14 44 Z"/><path d="M14 44 L50 44"/><path d="M22 38 L42 38"/></svg>',
    insurance: '<svg viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M32 8 L52 16 V32 C 52 44, 42 52, 32 56 C 22 52, 12 44, 12 32 V16 Z"/><path d="M24 32 L30 38 L42 26"/></svg>',
    wax: '<svg viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><rect x="14" y="12" width="36" height="20"/><path d="M14 32 L50 32 L46 52 L18 52 Z"/><path d="M22 22 L42 22"/></svg>'
  };

  var PRODUCTS = [
    {
      id: 'ski-basic',
      lot: 'LOT 01',
      name: 'Pack Esquí Básico',
      desc: 'Esquís, botas y bastones. Material certificado de inicio para escapadas de fin de semana.',
      price: 24,
      icon: ICONS.ski,
      model: 'ski-basic.glb',
      barcode: '8 412 SB 0001',
      nutri: {
        'Nivel recomendado': 'Principiante',
        'Material incluido': 'Esquís + Botas + Bastones',
        'Comodidad': '7 / 10',
        'Agarre': '6 / 10',
        'Flexibilidad': 'Alta',
        'Ideal para': 'Aprender'
      }
    },
    {
      id: 'ski-premium',
      lot: 'LOT 02',
      name: 'Pack Esquí Premium',
      desc: 'Carving avanzado, perfil radial y botas térmicas. Pensado para esquiadores de nivel intermedio/alto.',
      price: 49,
      icon: ICONS.premium,
      model: 'ski-premium.glb',
      barcode: '8 412 SB 0002',
      nutri: {
        'Nivel recomendado': 'Avanzado',
        'Material incluido': 'Esquís Carving + Botas Térmicas',
        'Comodidad': '9 / 10',
        'Agarre': '9 / 10',
        'Flexibilidad': 'Media',
        'Ideal para': 'Pista y velocidad'
      }
    },
    {
      id: 'snowboard',
      lot: 'LOT 03',
      name: 'Pack Snowboard',
      desc: 'Tabla all-mountain, fijaciones y botas blandas. Versatilidad pista + freestyle.',
      price: 39,
      icon: ICONS.board,
      model: 'snowboard.glb',
      barcode: '8 412 SB 0003',
      nutri: {
        'Nivel recomendado': 'Intermedio',
        'Material incluido': 'Tabla + Fijaciones + Botas',
        'Comodidad': '8 / 10',
        'Agarre': '7 / 10',
        'Flexibilidad': 'Alta',
        'Ideal para': 'All-mountain'
      }
    },
    {
      id: 'kids-pack',
      lot: 'LOT 04',
      name: 'Pack Infantil',
      desc: 'Equipo escalado para peques. Esquís cortos, botas confortables y bastones ligeros.',
      price: 18,
      icon: ICONS.kid,
      model: 'kids-pack.glb',
      barcode: '8 412 SB 0004',
      nutri: {
        'Nivel recomendado': 'Iniciación · 4–12 años',
        'Material incluido': 'Esquís + Botas + Bastones',
        'Comodidad': '9 / 10',
        'Agarre': '7 / 10',
        'Flexibilidad': 'Muy alta',
        'Ideal para': 'Familias'
      }
    },
    {
      id: 'helmet',
      lot: 'LOT 05',
      name: 'Casco + Seguridad',
      desc: 'Casco homologado, gafas con lente fotocromática y espaldera. Cero excusas para no protegerte.',
      price: 9,
      icon: ICONS.helmet,
      model: 'helmet.glb',
      barcode: '8 412 SB 0005',
      nutri: {
        'Nivel recomendado': 'Todos',
        'Material incluido': 'Casco + Gafas + Espaldera',
        'Comodidad': '9 / 10',
        'Agarre': '—',
        'Flexibilidad': '—',
        'Ideal para': 'Seguridad'
      }
    },
    {
      id: 'insurance',
      lot: 'LOT 06',
      name: 'Seguro de Nieve',
      desc: 'Cobertura de rescate, RC y rotura de material. Activable por días, sin letra pequeña.',
      price: 5,
      icon: ICONS.insurance,
      model: 'insurance.glb',
      barcode: '8 412 SB 0006',
      nutri: {
        'Nivel recomendado': 'Todos',
        'Material incluido': 'Cobertura completa',
        'Comodidad': '—',
        'Agarre': '—',
        'Flexibilidad': '—',
        'Ideal para': 'Tranquilidad total'
      }
    },
    {
      id: 'wax',
      lot: 'LOT 07',
      name: 'Mantenimiento / Encerado',
      desc: 'Afilado de cantos, reparación de suela y encerado en caliente. Tu material como nuevo cada día.',
      price: 12,
      icon: ICONS.wax,
      model: 'wax.glb',
      barcode: '8 412 SB 0007',
      nutri: {
        'Nivel recomendado': 'Todos',
        'Material incluido': 'Afilado + Suela + Encerado',
        'Comodidad': '—',
        'Agarre': '+15%',
        'Flexibilidad': 'Sin cambio',
        'Ideal para': 'Optimizar deslizamiento'
      }
    }
  ];

  // Patrón de barcode aleatorio pero estable por producto.
  function genBarcodePattern(seed) {
    var s = String(seed);
    var widths = [];
    for (var i = 0; i < 32; i++) {
      var w = ((s.charCodeAt(i % s.length) + i * 31) % 4) + 1;
      widths.push(w);
    }
    var bg = 'repeating-linear-gradient(to right';
    var pos = 0;
    var ink = getInk();
    widths.forEach(function (w, idx) {
      var bar = w + 'px';
      var color = idx % 2 === 0 ? ink : 'transparent';
      bg += ', ' + color + ' ' + pos + 'px ' + (pos + w) + 'px';
      pos += w;
    });
    bg += ')';
    return bg;
  }
  function getInk() {
    try {
      var c = getComputedStyle(document.documentElement).getPropertyValue('--sb-ink').trim();
      return c || '#0a0a0a';
    } catch (e) { return '#0a0a0a'; }
  }

  // ---------- Render ----------
  function renderNutri(parent, data) {
    var keys = Object.keys(data);
    var html = '';
    for (var i = 0; i < keys.length; i++) {
      var k = keys[i];
      html += '<tr><th scope="row">' + escapeHtml(k) + '</th><td>' + escapeHtml(data[k]) + '</td></tr>';
    }
    parent.innerHTML = html;
  }

  function escapeHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function formatEuro(n) {
    return n.toFixed(2).replace('.', ',') + ' €';
  }

  function renderProduct(p) {
    var tpl = document.getElementById('sb-product-template');
    if (!tpl || !tpl.content) { return null; }
    var node = tpl.content.firstElementChild.cloneNode(true);

    node.setAttribute('data-product-id', p.id);
    node.querySelector('[data-product-lot]').textContent = p.lot;
    node.querySelector('[data-product-name]').textContent = p.name;
    node.querySelector('[data-product-desc]').textContent = p.desc;
    node.querySelector('[data-product-price]').textContent = formatEuro(p.price);

    var viz = node.querySelector('[data-product-viz]');
    viz.setAttribute('data-model', p.model || '');
    var icon = node.querySelector('[data-product-icon]');
    icon.innerHTML = p.icon;

    renderNutri(node.querySelector('[data-product-nutri]'), p.nutri);

    var bars = node.querySelector('[data-barcode]');
    bars.style.setProperty('--barcode', genBarcodePattern(p.barcode));
    node.querySelector('[data-barcode-num]').textContent = p.barcode;

    var btn = node.querySelector('[data-add-to-cart]');
    btn.addEventListener('click', function () {
      if (!window.SBCart) { return; }
      window.SBCart.add({
        id: p.id,
        name: p.name,
        price: p.price,
        lot: p.lot
      });
      btn.classList.add('is-added');
      var label = btn.querySelector('[data-cta-label]');
      var prev = label.textContent;
      label.textContent = 'AÑADIDO ✓';
      setTimeout(function () {
        label.textContent = prev;
        btn.classList.remove('is-added');
      }, 1200);
    });

    return node;
  }

  function renderAll() {
    var grid = document.getElementById('sb-market-grid');
    if (!grid) { return; }
    var frag = document.createDocumentFragment();
    PRODUCTS.forEach(function (p) {
      var node = renderProduct(p);
      if (node) { frag.appendChild(node); }
    });
    grid.appendChild(frag);

    // Expose count for the header meta
    var counter = document.querySelector('[data-market-count]');
    if (counter) { counter.textContent = PRODUCTS.length; }
  }

  // ---------- Three.js hook (placeholder) ----------
  // window.SBMarket.mount3D({ root, productId, modelUrl }) será donde
  // pongamos el código de Three.js cuando llegue. Por ahora es un
  // no-op: el div placeholder ya está listo para ser reemplazado.
  window.SBMarket = {
    products: PRODUCTS,
    mount3D: function (_opts) { /* TODO: integrar Three.js + GLTFLoader */ }
  };

  // ---------- Boot ----------
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', renderAll);
  } else {
    renderAll();
  }
})();
