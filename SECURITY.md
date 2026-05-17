# Política de seguridad

Gracias por ayudar a mantener SkiHub seguro. Tomamos la seguridad muy en serio
y agradecemos cualquier informe responsable de vulnerabilidades.

## Versiones soportadas

| Versión | Soportada |
| ------- | --------- |
| `main`  | ✅ (rama estable, recibe parches de seguridad)
| Otras   | ❌ (sin soporte)

## Cómo reportar una vulnerabilidad

**No abras un issue público para vulnerabilidades.** En su lugar:

1. Ve a la pestaña **Security** del repositorio → **Report a vulnerability**
   (Private Vulnerability Reporting de GitHub).
2. O envía un correo a `a.celadaortega@gmail.com` con el asunto
   `[SECURITY] SkiHub - <breve descripción>`.

Incluye en el reporte:

- Descripción del problema y su impacto potencial.
- Pasos para reproducirlo (PoC si es posible).
- Versión / commit afectado.
- Tu nombre o handle si quieres reconocimiento público.

## Plazos esperados

- **Acuse de recibo**: en un máximo de 72 horas.
- **Primera evaluación**: 7 días.
- **Parche o mitigación**: dependiendo de la severidad, entre 7 y 30 días.

## Alcance

Esto cubre el código en este repositorio, las plantillas, los workflows de
GitHub Actions y la configuración de despliegue en `deploy/`. La
infraestructura externa (VPS, dominio, proveedores) está fuera del alcance,
aunque si encuentras algo crítico que afecte a usuarios reales, también
agradezco el aviso.

## Buenas prácticas activas en el proyecto

- Análisis estático con **CodeQL** en cada push/PR.
- **Dependabot** semanal para Go modules, GitHub Actions y Docker.
- **Gitleaks** escaneando secretos en commits.
- **Dependency Review** bloqueando dependencias con CVEs `high` en PRs.
- Cookies de sesión `HttpOnly` y contraseñas con `bcrypt` cost 12.
- Variables sensibles cargadas desde `.env` (nunca commiteado).

Gracias por contribuir a la seguridad del proyecto. ❄️
