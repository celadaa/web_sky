# Arquitectura

SkiHub es un monolito Go con plantillas server-side renderizadas y PostgreSQL como
única fuente de verdad. Sin frameworks pesados: stdlib + librerías mínimas.

## Visión general

```
                ┌─────────────┐
                │  Navegador  │
                └─────┬───────┘
                      │ HTTPS
                ┌─────▼───────┐
                │    Nginx    │  ← TLS, gzip, cabeceras, rate limit
                └─────┬───────┘
                      │ HTTP loopback
                ┌─────▼───────┐
                │  skihub     │  ← binario Go (systemd)
                │  (Go)       │
                └─────┬───────┘
                      │ TCP/5432
                ┌─────▼───────┐
                │  Postgres   │
                └─────────────┘

  Scraping ────► infonieve  ────► repository ────► Postgres
                (goroutines)
```

## Capas internas

```
internal/handlers/     ← HTTP. Parsean request, llaman a services, renderizan template.
        │
        ▼
internal/services/     ← Lógica de negocio. Coordinan repos. Sin tocar HTTP ni SQL.
        │
        ▼
internal/repository/   ← Acceso a Postgres. Devuelven models.
        │
        ▼
internal/models/       ← Structs del dominio. Sin lógica.
```

**Regla**: las flechas nunca van al revés. `models` no importa `repository`,
`repository` no importa `services`, etc.

## Componentes adicionales

| Paquete                   | Responsabilidad |
| ------------------------- | --------------- |
| `internal/config`         | Lee `.env` y variables de entorno; expone tipos tipados. |
| `internal/db`             | Pool de conexiones (`pgx/stdlib`), runner de migraciones, bootstrap del admin. |
| `internal/infonieve`      | Scraper de pistas en directo (HTML → DB) con `goquery`. |
| `cmd/servidor/main.go`    | Wire-up: config → db → repos → services → handlers → http.Server. |
| `cmd/servidor/security_helpers.go` | Middleware de seguridad (CSRF, headers, sesión). |

## Flujo de una petición

1. Nginx recibe HTTPS, valida certificado, aplica rate limit, pasa a `127.0.0.1:8080`.
2. Go `http.ServeMux` (o router) matchea la ruta.
3. Middleware: logger, recovery, seguridad, sesión.
4. Handler parsea inputs, llama a `service.X(ctx, ...)`.
5. Service compone llamadas a uno o varios `repository`.
6. Repository ejecuta SQL parametrizado, mapea filas a `models`.
7. Handler decide: redirect, JSON o `template.ExecuteTemplate(layout, ...)`.

## Sesiones y autenticación

- Cookie `HttpOnly`, `Secure` (prod), `SameSite=Lax`.
- El valor es un token aleatorio; la sesión real vive en BD.
- Contraseñas: `bcrypt` cost 12. Nunca se loguean ni se devuelven al cliente.

## Decisiones notables

- **Sin ORM**: SQL crudo con `pgx`, parametrizado. Más explícito, sin magia,
  performance predecible.
- **Sin SPA**: las páginas son `html/template` + JS vanilla para interacciones puntuales.
  Render rápido, indexable, sin overhead de hidratación.
- **Migraciones idempotentes en SQL plano**: nada de Atlas/Goose para mantenerlo simple.

## Diagrama de carpetas

```
cmd/servidor/main.go
internal/
├── config/
├── db/
├── handlers/
├── services/
├── repository/
├── models/
└── infonieve/
db/migrations/
deploy/nginx/
web/
├── templates/
└── static/
```
