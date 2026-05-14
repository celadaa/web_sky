/* Snowbreak — coordenadas conocidas de estaciones de esquí
 *
 * Replica en JavaScript el mapa coordsBySlug que usa el backend en
 * internal/infonieve/coords.go. Lo usamos en el cliente para poder
 * recalcular distancias cuando el usuario cambia su ubicación, sin
 * tener que pegar al backend ni guardar lat/lng en la BD.
 *
 * window.SnowbreakCoords ofrece:
 *   - lookup(nombre) → {lat,lng,pais} | null
 *   - distancia(a,b) → km (haversine)
 *   - normalizar(s)  → string sin tildes, en minúsculas
 */
(function () {
  'use strict';

  /** Mapa nombre-normalizado → {lat,lng,pais}. */
  var ESTACIONES = {
    // Pirineo aragonés
    'astun':                     { lat: 42.8125, lng: -0.5253, pais: 'España' },
    'candanchu':                 { lat: 42.7833, lng: -0.5167, pais: 'España' },
    'formigal':                  { lat: 42.7811, lng: -0.4108, pais: 'España' },
    'formigal-panticosa':        { lat: 42.7811, lng: -0.4108, pais: 'España' },
    'panticosa':                 { lat: 42.7167, lng: -0.2917, pais: 'España' },
    'cerler':                    { lat: 42.5400, lng:  0.4225, pais: 'España' },
    // Pirineo catalán
    'baqueira':                  { lat: 42.7000, lng:  0.9333, pais: 'España' },
    'baqueira-beret':            { lat: 42.7000, lng:  0.9333, pais: 'España' },
    'boi-taull':                 { lat: 42.5256, lng:  0.8667, pais: 'España' },
    'espot':                     { lat: 42.5750, lng:  1.1100, pais: 'España' },
    'espot-esqui':               { lat: 42.5750, lng:  1.1100, pais: 'España' },
    'la-molina':                 { lat: 42.3406, lng:  1.9469, pais: 'España' },
    'masella':                   { lat: 42.3675, lng:  1.9347, pais: 'España' },
    'alp-2500':                  { lat: 42.3506, lng:  1.9408, pais: 'España' },
    'port-aine':                 { lat: 42.6219, lng:  1.2358, pais: 'España' },
    'port-del-comte':            { lat: 42.1942, lng:  1.5292, pais: 'España' },
    'tavascan':                  { lat: 42.6717, lng:  1.2347, pais: 'España' },
    'vall-de-nuria':             { lat: 42.4006, lng:  2.1531, pais: 'España' },
    'nuria':                     { lat: 42.4006, lng:  2.1531, pais: 'España' },
    'vallter-2000':              { lat: 42.4322, lng:  2.2725, pais: 'España' },
    'vallter':                   { lat: 42.4322, lng:  2.2725, pais: 'España' },
    // Sistema Central
    'la-pinilla':                { lat: 41.2917, lng: -3.4583, pais: 'España' },
    'navacerrada':               { lat: 40.7842, lng: -4.0094, pais: 'España' },
    'puerto-de-navacerrada':     { lat: 40.7842, lng: -4.0094, pais: 'España' },
    'valdesqui':                 { lat: 40.7867, lng: -3.9000, pais: 'España' },
    'sierra-de-bejar-la-covatilla': { lat: 40.3450, lng: -5.7233, pais: 'España' },
    'la-covatilla':              { lat: 40.3450, lng: -5.7233, pais: 'España' },
    // Cordillera Cantábrica
    'alto-campoo':               { lat: 43.0686, lng: -4.4400, pais: 'España' },
    'fuentes-de-invierno':       { lat: 43.0700, lng: -5.4500, pais: 'España' },
    'leitariegos':               { lat: 43.0058, lng: -6.4242, pais: 'España' },
    'san-isidro':                { lat: 43.0467, lng: -5.4117, pais: 'España' },
    'valgrande-pajares':         { lat: 43.0450, lng: -5.7567, pais: 'España' },
    'pajares':                   { lat: 43.0450, lng: -5.7567, pais: 'España' },
    // Sierra Nevada
    'sierra-nevada':             { lat: 37.0928, lng: -3.3953, pais: 'España' },
    // Sistema Ibérico
    'javalambre':                { lat: 40.0944, lng: -1.0125, pais: 'España' },
    'valdelinares':              { lat: 40.4133, lng: -0.6233, pais: 'España' },
    'valdezcaray':               { lat: 42.2275, lng: -3.0344, pais: 'España' },
    // Galicia
    'manzaneda':                 { lat: 42.2675, lng: -7.2467, pais: 'España' },
    // Andorra
    'grandvalira':               { lat: 42.5392, lng:  1.7308, pais: 'Andorra' },
    'pas-de-la-casa':            { lat: 42.5450, lng:  1.7378, pais: 'Andorra' },
    'soldeu':                    { lat: 42.5778, lng:  1.6611, pais: 'Andorra' },
    'soldeu-el-tarter':          { lat: 42.5778, lng:  1.6611, pais: 'Andorra' },
    'ordino-arcalis':            { lat: 42.6342, lng:  1.5158, pais: 'Andorra' },
    'vallnord':                  { lat: 42.5786, lng:  1.4917, pais: 'Andorra' },
    'pal-arinsal':               { lat: 42.5786, lng:  1.4917, pais: 'Andorra' },
    // Pirineo francés
    'saint-lary':                { lat: 42.7842, lng:  0.3242, pais: 'Francia' },
    'font-romeu':                { lat: 42.5158, lng:  2.0367, pais: 'Francia' },
    'piau-engaly':               { lat: 42.7717, lng:  0.1611, pais: 'Francia' },
    'peyragudes':                { lat: 42.8014, lng:  0.4500, pais: 'Francia' },
    'les-angles':                { lat: 42.5681, lng:  2.0689, pais: 'Francia' },
    'porte-puymorens':           { lat: 42.5500, lng:  1.7967, pais: 'Francia' },
    'pyrenees-2000':             { lat: 42.5000, lng:  2.1000, pais: 'Francia' },
    // Portugal
    'serra-da-estrela':          { lat: 40.3217, lng: -7.6131, pais: 'Portugal' }
  };

  /** Normaliza un nombre: minúsculas, sin acentos, no-letra → guion. */
  function normalizar(s) {
    return String(s || '')
      .toLowerCase()
      .normalize('NFD')
      .replace(/[̀-ͯ]/g, '')
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '');
  }

  /** Devuelve {lat,lng,pais} o null. Tolera nombres parciales. */
  function lookup(nombre) {
    var clave = normalizar(nombre);
    if (!clave) { return null; }
    if (ESTACIONES[clave]) { return ESTACIONES[clave]; }
    // búsqueda permisiva: la clave del mapa contiene parte del nombre.
    var keys = Object.keys(ESTACIONES);
    for (var i = 0; i < keys.length; i++) {
      var k = keys[i];
      if (clave.indexOf(k) !== -1 || k.indexOf(clave) !== -1) {
        return ESTACIONES[k];
      }
    }
    return null;
  }

  /** Distancia en km entre dos puntos {lat,lng} (Haversine). */
  function distancia(a, b) {
    if (!a || !b) { return null; }
    var R = 6371;
    var rad = Math.PI / 180;
    var dLat = (b.lat - a.lat) * rad;
    var dLng = (b.lng - a.lng) * rad;
    var lat1 = a.lat * rad;
    var lat2 = b.lat * rad;
    var s = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
            Math.sin(dLng / 2) * Math.sin(dLng / 2) * Math.cos(lat1) * Math.cos(lat2);
    var c = 2 * Math.atan2(Math.sqrt(s), Math.sqrt(1 - s));
    return R * c;
  }

  window.SnowbreakCoords = {
    lookup: lookup,
    distancia: distancia,
    normalizar: normalizar
  };
})();
