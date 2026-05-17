# Cómo contribuir a SkiHub

Gracias por querer colaborar. Esta guía explica cómo trabajar con el repo de forma
ordenada para que las contribuciones se acepten rápido.

## Setup en 5 minutos

```bash
# 1. Fork + clone
git clone https://github.com/<tu-usuario>/web_sky.git
cd web_sky

# 2. Variables de entorno
cp .env.example .env
# Edita .env con tus credenciales locales

# 3. Postgres en Docker (opcional, también vale uno nativo)
make db-up

# 4. Arrancar
make run
```

La app debería estar disponible en `http://localhost:8080` (revisa `.env` para el puerto exacto).

## Flujo de trabajo

1. **Crea un issue antes de codear** si el cambio no es trivial. Así
   podemos discutir el enfoque antes de que inviertas tiempo.
2. **Crea una rama** desde `main` con un nombre descriptivo:
   - `feat/filtro-por-altitud`
   - `fix/login-rompe-en-safari`
   - `chore/actualizar-pgx`
3. **Haz commits pequeños y atómicos**. Usamos [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat: añadir filtro por altitud en /estaciones`
   - `fix: validar email en /registro antes de crear usuario`
   - `docs: explicar variables de entorno`
   - `refactor:`, `test:`, `chore:`, `perf:`, `ci:`
4. **Antes de abrir el PR**:
   ```bash
   make check       # fmt + vet + lint + test
   ```
5. **Abre el PR contra `main`**. La plantilla te guía. CI corre solo.

## Estándares de código

- **Formato**: `gofmt` obligatorio. Lo enforce el CI.
- **Lint**: `golangci-lint` con la config en `.golangci.yml`.
- **Tests**: cualquier feature nueva o bug fix debería traer tests. Si es
  imposible o desproporcionado, explícalo en el PR.
- **Errores**: nunca `panic` en código de petición HTTP. Devuelve errores
  envueltos con `fmt.Errorf("contexto: %w", err)`.
- **SQL**: parámetros siempre con `$1`, `$2`, etc. **Nunca** concatenes strings.
- **Cookies**: `HttpOnly`, `Secure` en producción, `SameSite=Lax`.

## Estructura del proyecto

```
cmd/servidor/      ← entry point
internal/config/   ← lectura de variables de entorno
internal/db/       ← conexión, migraciones
internal/handlers/ ← controladores HTTP
internal/services/ ← lógica de negocio
internal/repository/ ← acceso a Postgres
internal/models/   ← entidades de dominio
internal/infonieve/ ← scraper de pistas en vivo
db/migrations/     ← *.sql versionados
web/templates/     ← html/template
web/static/        ← CSS, JS, imágenes
```

Respeta la dirección de dependencias: `handlers → services → repository → models`.
Nada va al revés.

## Cómo añadir una migración

```bash
# Crea el archivo con un nombre incremental
touch db/migrations/000NN_descripcion_corta.sql
```

Cada migración es idempotente (`CREATE TABLE IF NOT EXISTS ...`) o usa
transacciones para no dejar la BD a medias.

## Buenas prácticas con secretos

- **Nunca** comitees `.env`, claves SSH, tokens, contraseñas.
- Si alguna vez ocurre: avisa inmediatamente; gitleaks lo cazará en CI,
  pero hay que rotar la credencial filtrada en cualquier caso.

## Reportar bugs y vulnerabilidades

- Bugs: abre un issue con la plantilla "Bug report".
- Vulnerabilidades de seguridad: lee [`SECURITY.md`](SECURITY.md) — usa el
  reporte privado, no el issue público.

## Licencia

Al contribuir aceptas que tu código se publique bajo la misma licencia del
proyecto (ver [`LICENSE`](LICENSE)).
