# Roadmap

Estado vivo de qué se ha hecho, qué viene y qué está aparcado.
Para granularidad fina mira [GitHub Projects](https://github.com/celadaa/web_sky/projects)
y los [Issues](https://github.com/celadaa/web_sky/issues).

## ✅ Hecho

- Estructura del backend Go (handlers / services / repository / models).
- Autenticación con cookies + bcrypt.
- Scraper de pistas en directo (`internal/infonieve`).
- Plantillas para todas las páginas principales.
- Migraciones SQL versionadas.
- Despliegue manual en VPS Ubuntu con Nginx + systemd.

## 🟡 En curso / próximo

- CI/CD completo en GitHub Actions (build + tests + deploy automático).
- Backups automatizados de Postgres a almacenamiento externo.
- Endpoint `/healthz` que devuelva versión + estado de BD.
- Métricas básicas (peticiones por minuto, latencia p50/p99) — Prometheus + Grafana.
- Sitemap XML y `robots.txt` afinados.
- Mejorar SEO de páginas de estación (Open Graph, meta description).

## 🔵 Backlog

- Comparador de forfaits multi-estación.
- Sistema de cupones / descuentos.
- App móvil (PWA primero, nativo después si hay tracción).
- Internacionalización (catalán, francés para Pirineo francés).
- Modo "previsión de nieve" con datos AEMET / Meteoblue.
- Pasarela de pago real (Stripe / Redsys).

## ❄️ Aparcado

- Migración a Postgres 17 (esperar a que esté estable en distros).
- SSR híbrido con HTMX (interesante pero el coste/beneficio no compensa aún).

## Cómo se mueven cosas de aquí

1. Si quieres trabajar en algo del backlog, abre un issue describiendo el alcance.
2. Cuando empiezas, mueve la tarjeta en el Project a "In progress".
3. Al mergear, el bot mueve la tarjeta a "Done" automáticamente.
