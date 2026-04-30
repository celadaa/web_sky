/* Helper de geolocalización para Snowbreak.
 *
 * Expone window.SnowbreakGeo con:
 *   - obtener({timeout}) → Promise<{lat,lng,fuente,etiqueta}>
 *     pide permiso al navegador; si el usuario lo deniega o el navegador
 *     no soporta la API, rechaza con un Error con .codigo conocido.
 *   - desdeCiudad(nombre) → Promise<{lat,lng,fuente,etiqueta}>
 *     busca coordenadas con OpenStreetMap (Nominatim).
 *   - guardarPreferencia(obj) / leerPreferencia()
 *     persiste la última ubicación elegida en localStorage para no
 *     volver a pedir permiso en cada visita.
 *   - CIUDADES_PRESET — lista corta para fallback con un click.
 *
 * Códigos de error normalizados (.codigo):
 *   "no_soportado"   — el navegador no implementa Geolocation.
 *   "denegado"       — el usuario rechazó el permiso.
 *   "no_disponible"  — no se pudo obtener (sin GPS, sin red…).
 *   "timeout"        — tardó demasiado.
 *   "no_encontrada"  — la ciudad introducida no se encontró.
 *   "red"            — fallo de red al geocodificar.
 */
(function () {
  'use strict';

  var CLAVE_LOCAL = 'snowbreak.ubicacion.v1';
  var TTL_HORAS   = 24;

  /** Ciudades de fallback frecuentes en el contexto de la web. */
  var CIUDADES_PRESET = [
    { etiqueta: 'Madrid',     lat: 40.4168, lng: -3.7038 },
    { etiqueta: 'Barcelona',  lat: 41.3851, lng: 2.1734 },
    { etiqueta: 'Zaragoza',   lat: 41.6488, lng: -0.8891 },
    { etiqueta: 'Pamplona',   lat: 42.8125, lng: -1.6458 },
    { etiqueta: 'Bilbao',     lat: 43.2630, lng: -2.9350 },
    { etiqueta: 'Granada',    lat: 37.1773, lng: -3.5986 },
    { etiqueta: 'Valencia',   lat: 39.4699, lng: -0.3763 },
    { etiqueta: 'Andorra la Vella', lat: 42.5063, lng: 1.5218 }
  ];

  /** Crea un error con un código de los normalizados arriba. */
  function errorGeo(codigo, mensaje) {
    var e = new Error(mensaje || codigo);
    e.codigo = codigo;
    return e;
  }

  /** Pide la posición al navegador. */
  function obtener(opciones) {
    opciones = opciones || {};
    return new Promise(function (resolve, reject) {
      if (!navigator.geolocation) {
        reject(errorGeo('no_soportado', 'Tu navegador no permite localización'));
        return;
      }
      navigator.geolocation.getCurrentPosition(
        function (pos) {
          resolve({
            lat: pos.coords.latitude,
            lng: pos.coords.longitude,
            fuente: 'gps',
            etiqueta: 'Tu ubicación'
          });
        },
        function (err) {
          var codigo = 'no_disponible';
          if (err.code === err.PERMISSION_DENIED) { codigo = 'denegado'; }
          else if (err.code === err.TIMEOUT)      { codigo = 'timeout'; }
          reject(errorGeo(codigo, err.message || 'no se pudo obtener la ubicación'));
        },
        {
          enableHighAccuracy: false,
          timeout: opciones.timeout || 8000,
          maximumAge: 5 * 60 * 1000
        }
      );
    });
  }

  /** Busca coordenadas para una ciudad. Usa Nominatim (OpenStreetMap). */
  function desdeCiudad(nombre) {
    var q = String(nombre || '').trim();
    if (!q) { return Promise.reject(errorGeo('no_encontrada', 'Ciudad vacía')); }
    var preset = encontrarPreset(q);
    if (preset) { return Promise.resolve(preset); }
    var url = 'https://nominatim.openstreetmap.org/search?format=json&limit=1&accept-language=es&q=' + encodeURIComponent(q);
    return fetch(url, { headers: { 'Accept': 'application/json' } })
      .then(function (r) {
        if (!r.ok) { throw errorGeo('red', 'HTTP ' + r.status); }
        return r.json();
      })
      .then(function (data) {
        if (!Array.isArray(data) || data.length === 0) {
          throw errorGeo('no_encontrada', 'No se encontró "' + q + '"');
        }
        var d = data[0];
        return {
          lat: parseFloat(d.lat),
          lng: parseFloat(d.lon),
          fuente: 'nominatim',
          etiqueta: d.display_name ? d.display_name.split(',')[0] : q
        };
      })
      .catch(function (err) {
        if (err && err.codigo) { throw err; }
        throw errorGeo('red', err && err.message ? err.message : 'fallo de red');
      });
  }

  /** Devuelve el preset cuyo nombre coincide ignorando mayúsculas/tildes. */
  function encontrarPreset(nombre) {
    var sin = normalizar(nombre);
    for (var i = 0; i < CIUDADES_PRESET.length; i++) {
      if (normalizar(CIUDADES_PRESET[i].etiqueta) === sin) {
        return Object.assign({ fuente: 'preset' }, CIUDADES_PRESET[i]);
      }
    }
    return null;
  }

  function normalizar(s) {
    return String(s)
      .toLowerCase()
      .trim()
      .normalize('NFD')
      .replace(/[̀-ͯ]/g, '');
  }

  function guardarPreferencia(ubic) {
    if (!ubic || !ventanaConStorage()) { return; }
    try {
      var payload = {
        lat: ubic.lat, lng: ubic.lng,
        fuente: ubic.fuente, etiqueta: ubic.etiqueta,
        ts: Date.now()
      };
      window.localStorage.setItem(CLAVE_LOCAL, JSON.stringify(payload));
    } catch (e) { /* ignorar quotas */ }
  }

  function leerPreferencia() {
    if (!ventanaConStorage()) { return null; }
    try {
      var raw = window.localStorage.getItem(CLAVE_LOCAL);
      if (!raw) { return null; }
      var v = JSON.parse(raw);
      if (!v || typeof v.lat !== 'number') { return null; }
      var edadH = (Date.now() - (v.ts || 0)) / 3600000;
      if (edadH > TTL_HORAS) { return null; }
      return v;
    } catch (e) { return null; }
  }

  function olvidarPreferencia() {
    if (!ventanaConStorage()) { return; }
    try { window.localStorage.removeItem(CLAVE_LOCAL); } catch (e) {}
  }

  function ventanaConStorage() {
    try { return typeof window !== 'undefined' && !!window.localStorage; }
    catch (e) { return false; }
  }

  window.SnowbreakGeo = {
    CIUDADES_PRESET: CIUDADES_PRESET,
    obtener: obtener,
    desdeCiudad: desdeCiudad,
    guardarPreferencia: guardarPreferencia,
    leerPreferencia: leerPreferencia,
    olvidarPreferencia: olvidarPreferencia
  };
})();
