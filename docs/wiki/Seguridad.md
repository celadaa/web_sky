# Seguridad

Resumen de las medidas activas y políticas de hardening del proyecto.

## En el código

- **Contraseñas**: `bcrypt` cost 12. Nunca se loguean ni se devuelven al cliente.
- **Sesiones**: token aleatorio en cookie `HttpOnly`, `Secure`, `SameSite=Lax`.
  El estado real vive en BD y caduca.
- **CSRF**: token doble (cookie + campo oculto) en formularios mutantes.
- **SQL Injection**: parámetros `$1`, `$2`... con `pgx`. Nunca concatenación.
- **XSS**: `html/template` escapa por defecto. Prohibido `template.HTML(...)` salvo
  con input ya saneado y validado.
- **Open redirects**: en redirects con parámetro `next=`, validar que apunte
  a un path interno (`startsWith("/")`).
- **Rate limiting**: en Nginx (`limit_req_zone`) y a nivel app para `/login`, `/registro`.

## En infraestructura

- **HTTPS**: Let's Encrypt + redirección 301 desde HTTP.
- **HSTS**: cabecera `Strict-Transport-Security: max-age=63072000; includeSubDomains; preload`.
- **Cabeceras seguras** en Nginx:
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `Referrer-Policy: strict-origin-when-cross-origin`
  - `Content-Security-Policy: default-src 'self'; ...` (ajustar al stack real)
- **Firewall (ufw)**: solo 22, 80, 443.
- **SSH**: solo claves, sin password auth (`PasswordAuthentication no`).
- **Postgres**: bind a `127.0.0.1`, no expuesto al exterior.

## En el pipeline

- **CodeQL** en cada PR y semanalmente: detecta CWEs comunes en Go.
- **Gitleaks** escanea historial completo en cada push.
- **Dependency Review** bloquea PRs con CVEs `high`.
- **Dependabot** abre PRs semanales con actualizaciones de seguridad.
- **CI** corre tests con race detector — detecta data races que en producción
  son timing bombs.

## Política de incidentes

Si se filtra una credencial (en `.env` commiteado, en logs, en un PR público):

1. **Rotar inmediatamente** la credencial afectada (cambiar contraseña/token).
2. Re-escribir historial si es viable (`git filter-repo`) y forzar push.
   Si no, asumir que la credencial vieja está comprometida para siempre.
3. Revisar logs de acceso del servicio afectado por uso anómalo.
4. Post-mortem en `docs/postmortems/` con timeline y prevención.

## Reportar vulnerabilidades

Ver [`SECURITY.md`](https://github.com/celadaa/web_sky/blob/main/SECURITY.md).
