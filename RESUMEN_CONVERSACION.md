# Resumen completo — Proyecto Snowbreak (web_esqui)

## Contexto del proyecto

**Snowbreak** (snowbreak.es) es una web de planificación de viajes de esquí.  
- Repositorio: https://github.com/celadaa/web_sky  
- Módulo Go: `skihub`  
- VPS Ubuntu con systemd, usuario `javier`  
- Dominio registrado en IONOS  
- CI/CD: GitHub Actions (ci.yml → deploy.yml), el deploy solo se lanza si el CI pasa  
- Base de datos: PostgreSQL  
- Email: Resend (SMTP smtp.resend.com:587 con STARTTLS)  
- DNS verificado en Resend, dominio snowbreak.es  

---

## Stack técnico

| Capa | Tecnología |
|------|-----------|
| Backend | Go (net/http, html/template) |
| DB | PostgreSQL |
| Auth | Sesiones propias + Google OAuth 2.0 (OIDC) |
| Email | Resend vía SMTP (multipart MIME: text/plain + text/html) |
| Frontend | HTML/CSS/JS vanilla (sin frameworks) |
| Deploy | GitHub Actions → SSH → systemd restart |
| Config | Variables de entorno en `/etc/systemd/system/skihub.service.d/override.conf` |

---

## Estructura de ficheros relevantes

```
web_esqui/
├── cmd/servidor/main.go                  ← rutas y arranque
├── internal/
│   ├── config/config.go                  ← variables de entorno (SMTP, OAuth, APP_BASE_URL)
│   ├── handlers/
│   │   ├── oauth_google.go               ← login con Google
│   │   ├── verificar_email.go            ← GET /confirmar-email?token=...
│   │   └── ...
│   ├── models/usuario.go                 ← struct Usuario
│   └── services/
│       ├── email.go                      ← envío SMTP multipart
│       └── email_test.go                 ← tests de primerNombre y construirMensaje
├── web/
│   ├── static/
│   │   ├── css/pages.css                 ← estilos wizard, .is-just-selected, .is-flash, toast
│   │   └── js/
│   │       ├── trip-planner-ticket.js    ← estado del planificador (SBTrip), toast, flashLine
│   │       └── planificar.js             ← wizard /planificar-estancia, markJustSelected
│   └── templates/
│       ├── layout.tmpl                   ← layout base, navbar requiere campo .Usuario
│       ├── email/
│       │   ├── verification.html.tmpl    ← email confirmación (rediseñado)
│       │   ├── verification.txt.tmpl     ← versión texto plano
│       │   ├── welcome.html.tmpl         ← email bienvenida Google (rediseñado)
│       │   └── welcome.txt.tmpl          ← versión texto plano
│       └── ...
├── db/migrations/                        ← 12 migraciones aplicadas en VPS
└── .github/
    ├── workflows/ci.yml + deploy.yml
    └── dependabot.yml                    ← actualiza deps Go + GitHub Actions (sin docker)
```

---

## Lo que se hizo en esta sesión

### 1. Feedback visual en /planificar-estancia

Se conectaron tres efectos visuales que estaban preparados en CSS pero sin JS:

- **Toast** (`showPlannerToast`): aparece abajo al guardar cambios en el planificador. Usa `data-trip-toast` y `data-trip-toast-msg`. Clase `is-visible` con transición CSS.
- **Flash de línea** (`flashTicketLine`): resalta la línea del ticket que cambia. Usa `data-line="..."` y clase `is-flash` con animación.
- **is-just-selected** (`markJustSelected`): destaca el elemento recién seleccionado (estación, alojamiento, material, forfait, extras). Clase `is-just-selected` con @keyframes `sbJustSelected`.

Los tres efectos respetan `prefers-reduced-motion`.  
El módulo global es `window.SBTrip` (en `trip-planner-ticket.js`).

### 2. Google OAuth

El código de Google OAuth (oauth_google.go) existía localmente pero **nunca se había commiteado**. El VPS solo despliega desde GitHub, así que no tenía esa funcionalidad.

- Se commiteó oauth_google.go + rutas `/auth/google` y `/auth/google/callback`
- Variables necesarias en override.conf:
  ```
  GOOGLE_CLIENT_ID=...
  GOOGLE_CLIENT_SECRET=...
  ```
- La app de Google está en modo prueba (solo funciona con el correo del desarrollador). Para abrirla al público hay que publicar la app en Google Cloud Console → OAuth consent screen → Publish.

### 3. Email con Resend

- Proveedor: **Resend** (https://resend.com)
- Protocolo: SMTP con STARTTLS en **puerto 587** (NO 465 — ese usa TLS implícito y fallaba)
- DNS en IONOS: se añadieron 3 registros (DKIM TXT, MX, SPF TXT) — verificados en Resend
- Variables en override.conf:
  ```
  SMTP_HOST=smtp.resend.com
  SMTP_PORT=587
  SMTP_USER=resend
  SMTP_PASS=<API_KEY_de_Resend>
  SMTP_FROM_EMAIL=hola@snowbreak.es
  SMTP_FROM_NAME=Snowbreak
  ```

### 4. Fix: crash al confirmar email

- **Error**: al hacer clic en el enlace de confirmación daba 500.
- **Causa**: `datosConfirmar` en `verificar_email.go` no tenía el campo `Usuario *models.Usuario`, que el layout.tmpl necesita para el navbar.
- **Fix**: añadir el campo y pasar `Usuario: a.UsuarioActual(r)` en los dos `render()`.

### 5. Fix: APP_BASE_URL crash en producción

- Tras un deploy grande, el servicio entró en bucle de reinicios.
- **Causa**: `config.go` valida que `APP_BASE_URL` empiece por `https://` en producción, y la variable no estaba en override.conf.
- **Fix**: añadir `Environment="APP_BASE_URL=https://snowbreak.es"` al override.

### 6. Fix: dependabot.yml

- Dependabot fallaba porque el yml incluía el ecosistema `docker` pero no hay Dockerfile.
- Se eliminó ese bloque. Ahora solo gestiona `gomod` y `github-actions`.

### 7. Rediseño de emails

Los cuatro templates de email se rediseñaron por completo:

**verification.html.tmpl** — email de confirmación de cuenta (registro normal):
- Header negro con SNOWBREAK + "La nieve, a tu ritmo"
- Badge azul "Un paso más"
- Saludo serif con nombre del usuario
- Botón negro "Confirmar email →"
- Bloque de fallback con URL copiable
- Footer con dominio + emoji

**welcome.html.tmpl** — email de bienvenida (registro con Google):
- Mismo diseño base
- Badge verde "Cuenta activa"
- Texto adaptado: no necesita confirmar nada (Google ya verificó)
- Botón "Ver estaciones →"

Los `.txt.tmpl` se actualizaron en tono y estructura para coincidir.

---

## Variables de entorno en el VPS (override.conf)

Ruta: `/etc/systemd/system/skihub.service.d/override.conf`

```ini
[Service]
Environment="APP_BASE_URL=https://snowbreak.es"
Environment="DATABASE_URL=postgres://..."
Environment="SESSION_SECRET=..."
Environment="GOOGLE_CLIENT_ID=..."
Environment="GOOGLE_CLIENT_SECRET=..."
Environment="SMTP_HOST=smtp.resend.com"
Environment="SMTP_PORT=587"
Environment="SMTP_USER=resend"
Environment="SMTP_PASS=<API_KEY>"
Environment="SMTP_FROM_EMAIL=hola@snowbreak.es"
Environment="SMTP_FROM_NAME=Snowbreak"
```

Después de editar: `sudo systemctl daemon-reload && sudo systemctl restart skihub.service`

---

## Comandos útiles en el VPS

```bash
# Ver logs en tiempo real
sudo journalctl -u skihub.service -f

# Estado del servicio
sudo systemctl status skihub.service

# Reiniciar
sudo systemctl restart skihub.service

# Editar variables de entorno
sudo nano /etc/systemd/system/skihub.service.d/override.conf
sudo systemctl daemon-reload && sudo systemctl restart skihub.service
```

---

## Pendiente / próximos pasos

1. **Publicar la app de Google OAuth**: en Google Cloud Console → OAuth consent screen → cambiar de "Testing" a "Production" para que cualquier usuario pueda registrarse con Google (no solo el correo del desarrollador).

2. **Remitente del email en Resend**: verificar que los emails se envían con `hola@snowbreak.es` como remitente (y no una dirección de Resend por defecto). Confirmar en el dashboard de Resend → Domains.

3. **gofmt**: los archivos `oauth_google.go`, `verificar_email.go` y `email_test.go` necesitan pasar `gofmt -w` antes de commitear si el CI tiene el check de formato activado. Hacerlo en local con:
   ```powershell
   gofmt -w internal/handlers/oauth_google.go internal/handlers/verificar_email.go internal/services/email_test.go
   ```

4. **Dependabot PR #11**: hay un PR abierto con 5 actualizaciones de dependencias Go (menores/patch). Revisar que el CI lo apruebe y mergearlo cuando convenga.

5. **Modo producción Google OAuth**: si se quiere abrir el registro a todos, verificar que el redirect URI en Google Cloud Console apunte a `https://snowbreak.es/auth/google/callback`.

---

## Flujo de deploy

```
git push → GitHub Actions CI (go build + go test + gofmt) → si pasa → deploy.yml → SSH al VPS → git pull + go build + systemctl restart
```

El CI falla si:
- `go build ./...` da error
- `go test ./...` falla
- `gofmt` detecta archivos mal formateados

---

## Notas importantes

- El módulo Go se llama `skihub` (no snowbreak) — así está en `go.mod`
- El layout.tmpl **siempre** necesita el campo `.Usuario` en los datos del template para el navbar
- El SMTP usa STARTTLS (puerto 587), NO TLS implícito (puerto 465)
- Los emails tienen siempre dos partes: `text/plain` + `text/html` (multipart/alternative)
- El estado del planificador se guarda en localStorage con la clave `snowbreak_trip_planner_v1`
