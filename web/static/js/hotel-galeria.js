/* ============================================================
   hotel-galeria.js — galería + lightbox de /hoteles/{slug}

   - Clic en miniatura → cambia la imagen principal.
   - Clic (o Enter/Espacio) en la imagen principal → abre el modal.
   - Modal: anterior/siguiente/cerrar, Escape y flechas de teclado,
     clic en el fondo para cerrar. Devuelve el foco al elemento
     que abrió el modal. Sin librerías externas.
   ============================================================ */
(function () {
  'use strict';

  var gallery = document.getElementById('hotel-gallery');
  if (!gallery) return;

  var mainImg = document.getElementById('hotel-gallery-main');
  var thumbs = Array.prototype.slice.call(
    gallery.querySelectorAll('[data-gallery-thumb]')
  );

  // Lista de fotos: de las miniaturas si hay varias; si no, la principal.
  var photos = thumbs.length
    ? thumbs.map(function (b) {
        return { url: b.getAttribute('data-url'), alt: b.getAttribute('data-alt') || '' };
      })
    : (mainImg ? [{ url: mainImg.getAttribute('src'), alt: mainImg.getAttribute('alt') || '' }] : []);

  if (!photos.length) return;

  var currentIndex = 0;

  // ---------- Miniaturas → imagen principal ----------
  function setMain(index) {
    if (!mainImg || index < 0 || index >= photos.length) return;
    currentIndex = index;
    mainImg.src = photos[index].url;
    mainImg.alt = photos[index].alt;
    mainImg.setAttribute('data-index', String(index));
    thumbs.forEach(function (b, i) {
      b.classList.toggle('is-active', i === index);
    });
  }

  thumbs.forEach(function (btn, i) {
    btn.addEventListener('click', function () { setMain(i); });
  });

  // ---------- Lightbox ----------
  var box = document.getElementById('hotel-lightbox');
  var boxImg = document.getElementById('hotel-lightbox-img');
  var boxCaption = document.getElementById('hotel-lightbox-caption');
  if (!box || !boxImg) return;

  var btnClose = box.querySelector('[data-lightbox-close]');
  var btnPrev = box.querySelector('[data-lightbox-prev]');
  var btnNext = box.querySelector('[data-lightbox-next]');
  var openerEl = null; // a quién devolver el foco al cerrar

  function renderBox(index) {
    currentIndex = ((index % photos.length) + photos.length) % photos.length;
    boxImg.src = photos[currentIndex].url;
    boxImg.alt = photos[currentIndex].alt;
    if (boxCaption) {
      boxCaption.textContent =
        photos[currentIndex].alt + ' (' + (currentIndex + 1) + ' de ' + photos.length + ')';
    }
    // Con una sola foto no tiene sentido navegar.
    var multi = photos.length > 1;
    if (btnPrev) btnPrev.hidden = !multi;
    if (btnNext) btnNext.hidden = !multi;
  }

  function openBox(index, opener) {
    openerEl = opener || null;
    renderBox(index);
    box.hidden = false;
    document.body.classList.add('hotel-lightbox-open');
    if (btnClose) btnClose.focus();
  }

  function closeBox() {
    box.hidden = true;
    document.body.classList.remove('hotel-lightbox-open');
    setMain(currentIndex); // sincroniza la principal con lo último visto
    if (openerEl && openerEl.focus) openerEl.focus();
    openerEl = null;
  }

  function prev() { renderBox(currentIndex - 1); }
  function next() { renderBox(currentIndex + 1); }

  if (mainImg) {
    mainImg.addEventListener('click', function () { openBox(currentIndex, mainImg); });
    mainImg.addEventListener('keydown', function (e) {
      if (e.key !== 'Enter' && e.key !== ' ') return;
      e.preventDefault();
      openBox(currentIndex, mainImg);
    });
  }

  if (btnClose) btnClose.addEventListener('click', closeBox);
  if (btnPrev) btnPrev.addEventListener('click', prev);
  if (btnNext) btnNext.addEventListener('click', next);

  // Clic en el fondo oscuro (no en la foto ni los botones) cierra.
  box.addEventListener('click', function (e) {
    if (e.target === box) closeBox();
  });

  // Teclado: Escape cierra; flechas navegan. Solo con el modal abierto.
  document.addEventListener('keydown', function (e) {
    if (box.hidden) return;
    switch (e.key) {
      case 'Escape':
        e.preventDefault();
        closeBox();
        break;
      case 'ArrowLeft':
        e.preventDefault();
        prev();
        break;
      case 'ArrowRight':
        e.preventDefault();
        next();
        break;
      case 'Tab':
        // Foco confinado al modal mientras está abierto (3 botones).
        trapFocus(e);
        break;
    }
  });

  function trapFocus(e) {
    var focusables = [btnClose, btnPrev, btnNext].filter(function (b) {
      return b && !b.hidden;
    });
    if (!focusables.length) return;
    var first = focusables[0];
    var last = focusables[focusables.length - 1];
    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault();
      first.focus();
    }
  }
})();
