/* ===========================================================
 * SkiHub — Cliente JavaScript de la API REST de usuarios.
 * PEC 3 (Tema 4) — Redes y Sistemas Web.
 *
 * Reglas de implementación:
 *   - JS puro: cero dependencias externas.
 *   - fetch + .then() (promesas), nunca async/await.
 *   - El render de filas usa <template id="template-usuario">.
 *   - Toda manipulación visual se hace con classList.
 *     Nunca se asigna element.style = ... ni element.style.X = ...
 * =========================================================== */

(function () {
  'use strict';

  // URL base de la API (no se concatenan dominios: misma origen).
  var API = '/api/usuarios';

  // Cache de referencias a nodos del DOM. Se rellena en init().
  var dom = {};

  // -----------------------------------------------------------
  // Utilidades genéricas
  // -----------------------------------------------------------

  /**
   * fetchJSON — wrapper de fetch que:
   *   1. fija Accept y Content-Type según convenga,
   *   2. parsea siempre la respuesta como JSON (si la hay),
   *   3. rechaza la promesa si HTTP status >= 400, propagando
   *      el mensaje de error que devuelve la API.
   */
  function fetchJSON(url, opciones) {
    opciones = opciones || {};
    var cabeceras = opciones.headers || {};
    cabeceras['Accept'] = 'application/json';
    if (opciones.body && !cabeceras['Content-Type']) {
      cabeceras['Content-Type'] = 'application/json';
    }
    opciones.headers = cabeceras;

    return fetch(url, opciones).then(function (resp) {
      // 204 No Content → sin cuerpo.
      if (resp.status === 204) {
        return null;
      }
      return resp.text().then(function (txt) {
        var datos = null;
        if (txt) {
          try { datos = JSON.parse(txt); } catch (e) { datos = null; }
        }
        if (!resp.ok) {
          var msg = (datos && datos.error) ? datos.error
                  : ('Error HTTP ' + resp.status);
          var err = new Error(msg);
          err.status = resp.status;
          throw err;
        }
        return datos;
      });
    });
  }

  // -----------------------------------------------------------
  // CRUD — funciones públicas pedidas por el enunciado
  // -----------------------------------------------------------

  /**
   * cargarUsuarios — GET /api/usuarios
   * Vacía el <tbody> y vuelca cada usuario clonando <template>.
   */
  function cargarUsuarios() {
    fetchJSON(API)
      .then(function (lista) {
        renderizarLista(lista || []);
      })
      .catch(function (err) {
        mostrarNotificacion('No se pudieron cargar los usuarios: ' + err.message);
      });
  }

  /**
   * crearUsuario — POST /api/usuarios
   * @param {{nombre:string,email:string,password:string,password2?:string}} datos
   */
  function crearUsuario(datos) {
    return fetchJSON(API, {
      method: 'POST',
      body: JSON.stringify(datos)
    })
    .then(function (nuevo) {
      mostrarNotificacion('Usuario creado: ' + (nuevo.email || nuevo.nombre));
      cargarUsuarios();
      return nuevo;
    })
    .catch(function (err) {
      mostrarNotificacion('Error al crear: ' + err.message);
      throw err;
    });
  }

  /**
   * editarUsuario — PUT /api/usuarios/{id}
   * @param {number} id
   * @param {{nombre:string,email:string}} datos
   */
  function editarUsuario(id, datos) {
    return fetchJSON(API + '/' + encodeURIComponent(id), {
      method: 'PUT',
      body: JSON.stringify(datos)
    })
    .then(function (actualizado) {
      mostrarNotificacion('Usuario actualizado: ' + (actualizado.email || actualizado.nombre));
      cerrarFormEdicion();
      cargarUsuarios();
      return actualizado;
    })
    .catch(function (err) {
      mostrarNotificacion('Error al editar: ' + err.message);
      throw err;
    });
  }

  /**
   * eliminarUsuario — DELETE /api/usuarios/{id}
   * @param {number} id
   */
  function eliminarUsuario(id) {
    return fetchJSON(API + '/' + encodeURIComponent(id), { method: 'DELETE' })
      .then(function () {
        mostrarNotificacion('Usuario eliminado.');
        cargarUsuarios();
      })
      .catch(function (err) {
        mostrarNotificacion('Error al eliminar: ' + err.message);
      });
  }

  // -----------------------------------------------------------
  // UI — notificaciones y formulario de edición
  // -----------------------------------------------------------

  // Temporizador para auto-ocultar la notificación.
  var temporizadorNotif = null;

  /**
   * mostrarNotificacion — añade el texto al div#notificacion y
   * gestiona la clase CSS "visible" para mostrarlo/ocultarlo.
   * NUNCA usa style= directamente.
   */
  function mostrarNotificacion(msg) {
    var caja = dom.notificacion;
    if (!caja) { return; }
    caja.textContent = msg;
    caja.classList.add('visible');

    if (temporizadorNotif) {
      clearTimeout(temporizadorNotif);
    }
    temporizadorNotif = setTimeout(function () {
      caja.classList.remove('visible');
    }, 3500);
  }

  /**
   * abrirFormEdicion — muestra el formulario de edición precargado
   * con los datos del usuario indicado. Usa la clase "activo" (no style).
   */
  function abrirFormEdicion(u) {
    if (!dom.formEdicion) { return; }
    dom.formEdicion.dataset.id = u.id;
    if (dom.inputNombre) { dom.inputNombre.value = u.nombre || ''; }
    if (dom.inputEmail)  { dom.inputEmail.value  = u.email  || ''; }
    dom.formEdicion.classList.add('activo');
  }

  /**
   * cerrarFormEdicion — oculta el formulario quitándole la clase "activo".
   */
  function cerrarFormEdicion() {
    if (!dom.formEdicion) { return; }
    dom.formEdicion.classList.remove('activo');
    delete dom.formEdicion.dataset.id;
  }

  // -----------------------------------------------------------
  // Render con <template>
  // -----------------------------------------------------------

  /**
   * renderizarLista — vacía el tbody y, por cada usuario del array,
   * clona el contenido del <template id="template-usuario">.
   */
  function renderizarLista(usuarios) {
    var tbody = dom.tbody;
    var tpl = dom.template;
    if (!tbody || !tpl) { return; }

    // Vaciar.
    while (tbody.firstChild) {
      tbody.removeChild(tbody.firstChild);
    }

    if (usuarios.length === 0) {
      var fila = document.createElement('tr');
      var celda = document.createElement('td');
      celda.colSpan = 5;
      celda.textContent = 'No hay usuarios registrados.';
      fila.appendChild(celda);
      tbody.appendChild(fila);
      return;
    }

    usuarios.forEach(function (u) {
      var clon = tpl.content.cloneNode(true);

      var celdaNombre = clon.querySelector('[data-campo="nombre"]');
      var celdaEmail  = clon.querySelector('[data-campo="email"]');
      var celdaRol    = clon.querySelector('[data-campo="rol"]');
      var btnEditar   = clon.querySelector('[data-accion="editar"]');
      var btnBorrar   = clon.querySelector('[data-accion="eliminar"]');

      if (celdaNombre) { celdaNombre.textContent = u.nombre; }
      if (celdaEmail)  { celdaEmail.textContent  = u.email; }
      if (celdaRol)    { celdaRol.textContent    = u.es_admin ? 'admin' : 'usuario'; }

      if (btnEditar) {
        btnEditar.addEventListener('click', function () {
          abrirFormEdicion(u);
        });
      }
      if (btnBorrar) {
        btnBorrar.addEventListener('click', function () {
          if (confirm('¿Eliminar el usuario "' + u.email + '"? Esta acción es irreversible.')) {
            eliminarUsuario(u.id);
          }
        });
      }

      tbody.appendChild(clon);
    });
  }

  // -----------------------------------------------------------
  // Inicialización
  // -----------------------------------------------------------

  function init() {
    dom.tbody        = document.getElementById('lista-usuarios');
    dom.template     = document.getElementById('template-usuario');
    dom.notificacion = document.getElementById('notificacion');
    dom.formEdicion  = document.getElementById('form-edicion');
    dom.inputNombre  = document.getElementById('edit-nombre');
    dom.inputEmail   = document.getElementById('edit-email');

    // Botones del formulario de edición (Guardar / Cancelar).
    var btnGuardar  = document.getElementById('edit-guardar');
    var btnCancelar = document.getElementById('edit-cancelar');

    if (btnGuardar) {
      btnGuardar.addEventListener('click', function (ev) {
        ev.preventDefault();
        if (!dom.formEdicion) { return; }
        var id = dom.formEdicion.dataset.id;
        if (!id) {
          mostrarNotificacion('No hay usuario seleccionado.');
          return;
        }
        editarUsuario(parseInt(id, 10), {
          nombre: dom.inputNombre ? dom.inputNombre.value : '',
          email:  dom.inputEmail  ? dom.inputEmail.value  : ''
        });
      });
    }

    if (btnCancelar) {
      btnCancelar.addEventListener('click', function (ev) {
        ev.preventDefault();
        cerrarFormEdicion();
      });
    }

    // Carga inicial.
    cargarUsuarios();
  }

  // Lanzamos init en DOMContentLoaded (o ya, si el DOM está listo).
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

  // Exposición opcional para depurar desde la consola del navegador.
  window.SkiHubAPI = {
    cargarUsuarios: cargarUsuarios,
    crearUsuario: crearUsuario,
    editarUsuario: editarUsuario,
    eliminarUsuario: eliminarUsuario,
    mostrarNotificacion: mostrarNotificacion,
    abrirFormEdicion: abrirFormEdicion,
    cerrarFormEdicion: cerrarFormEdicion
  };
})();
