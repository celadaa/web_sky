(function () {
  'use strict';

  var API = '/api/usuarios';
  var dom = {};
  var temporizadorNotif = null;

  function fetchJSON(url, opciones) {
    opciones = opciones || {};
    var cabeceras = opciones.headers || {};
    cabeceras.Accept = 'application/json';
    if (opciones.body && !cabeceras['Content-Type']) {
      cabeceras['Content-Type'] = 'application/json';
    }
    opciones.headers = cabeceras;

    return fetch(url, opciones).then(function (resp) {
      if (resp.status === 204) {
        return null;
      }
      return resp.text().then(function (txt) {
        var datos = null;
        if (txt) {
          try { datos = JSON.parse(txt); } catch (e) { datos = null; }
        }
        if (!resp.ok) {
          var msg = (datos && datos.error) ? datos.error : ('Error HTTP ' + resp.status);
          var err = new Error(msg);
          err.status = resp.status;
          throw err;
        }
        return datos;
      });
    });
  }

  function cargarUsuarios() {
    fetchJSON(API)
      .then(function (lista) {
        renderizarLista(lista || []);
      })
      .catch(function (err) {
        mostrarNotificacion('No se pudieron cargar los usuarios: ' + err.message);
      });
  }

  function editarUsuario(id, datos) {
    return fetchJSON(API + '/' + encodeURIComponent(id), {
      method: 'PUT',
      body: JSON.stringify(datos)
    }).then(function () {
      mostrarNotificacion('Usuario actualizado.');
      cerrarFormEdicion();
      cargarUsuarios();
    }).catch(function (err) {
      mostrarNotificacion('Error al editar: ' + err.message);
      throw err;
    });
  }

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

  function abrirFormEdicion(u) {
    if (!dom.formEdicion) { return; }
    dom.formEdicion.dataset.id = u.id;
    dom.inputNombre.value = u.nombre || '';
    dom.inputEmail.value = u.email || '';
    dom.formEdicion.classList.add('activo');
  }

  function cerrarFormEdicion() {
    if (!dom.formEdicion) { return; }
    dom.formEdicion.classList.remove('activo');
    delete dom.formEdicion.dataset.id;
  }

  function renderizarLista(usuarios) {
    var tbody = dom.tbody;
    var tpl = dom.template;
    if (!tbody || !tpl) { return; }

    while (tbody.firstChild) {
      tbody.removeChild(tbody.firstChild);
    }

    if (usuarios.length === 0) {
      var fila = document.createElement('tr');
      var celda = document.createElement('td');
      celda.colSpan = 4;
      celda.textContent = 'No hay usuarios registrados.';
      fila.appendChild(celda);
      tbody.appendChild(fila);
      return;
    }

    usuarios.forEach(function (u) {
      var clon = tpl.content.cloneNode(true);
      clon.querySelector('[data-campo="nombre"]').textContent = u.nombre;
      clon.querySelector('[data-campo="email"]').textContent = u.email;
      clon.querySelector('[data-campo="rol"]').textContent = u.es_admin ? 'admin' : 'usuario';

      clon.querySelector('[data-accion="editar"]').addEventListener('click', function () {
        abrirFormEdicion(u);
      });

      clon.querySelector('[data-accion="eliminar"]').addEventListener('click', function () {
        if (confirm('¿Eliminar el usuario "' + u.email + '"? Esta acción es irreversible.')) {
          eliminarUsuario(u.id);
        }
      });

      tbody.appendChild(clon);
    });
  }

  function init() {
    dom.tbody = document.getElementById('lista-usuarios');
    dom.template = document.getElementById('template-usuario');
    dom.notificacion = document.getElementById('notificacion');
    dom.formEdicion = document.getElementById('form-edicion');
    dom.inputNombre = document.getElementById('edit-nombre');
    dom.inputEmail = document.getElementById('edit-email');

    var btnGuardar = document.getElementById('edit-guardar');
    var btnCancelar = document.getElementById('edit-cancelar');

    if (btnGuardar) {
      btnGuardar.addEventListener('click', function (ev) {
        ev.preventDefault();
        if (!dom.formEdicion || !dom.formEdicion.dataset.id) {
          mostrarNotificacion('No hay usuario seleccionado.');
          return;
        }
        editarUsuario(parseInt(dom.formEdicion.dataset.id, 10), {
          nombre: dom.inputNombre.value,
          email: dom.inputEmail.value
        });
      });
    }

    if (btnCancelar) {
      btnCancelar.addEventListener('click', function (ev) {
        ev.preventDefault();
        cerrarFormEdicion();
      });
    }

    cargarUsuarios();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
