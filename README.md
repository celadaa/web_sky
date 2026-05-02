# SkiHub (Snowbreak)

Portal web para comparar estaciones de esquí, comprar forfaits y consultar el estado de las pistas en tiempo real. Backend escrito en Go, persistencia en PostgreSQL, frontend con plantillas `html/template` + JavaScript vanilla.

---

## Tabla de contenidos

1. Stack y arquitectura
2. Requisitos
3. Instalar PostgreSQL
4. Crear la base de datos
5. Variables de entorno
6. Ejecutar migraciones y seed
7. Arrancar la aplicación
8. Comprobar que la conexión funciona
9. Despliegue en Ubuntu detrás de Nginx
10. Docker (opcional)
11. Estructura del proyecto

---

## 1. Stack y arquitectura

- **Lenguaje**: Go 1.26
- **Persistencia**: PostgreSQL 14+ accedido desde Go vía `database/sql` con el driver `github.com/jackc/pgx/v5/stdlib`.
- **Auth**: cookies HttpOnly persistidas en BD; contraseñas almacenadas como hash bcrypt (cost 12).
- **Configuración**: variables de entorno + `.env` opcional cargado con `joho/godotenv`.
- **Capas**: `handlers` → `services` → `repositories` → `models`. Toda operación que toca la BD recibe `context.Context`.

```
cmd/servidor/main.go        ← punto de entrada
internal/config/            ← lectura de variables de entorno
internal/db/                ← conexión, runner de migraciones, bootstrap admin
internal/handlers/          ← controladores HTTP (web + API)
internal/services/          ← lógica de negocio
internal/repository/        ← acceso a Postgres
internal/models/            ← entidades del dominio
internal/infonieve/         ← scraper externo (datos en directo de pistas)
db/migrations/              ← *.sql versionados
web/templates/              ← plantillas html/template
web/static/                 ← CSS, JS, imágenes
```

---

## 2. Requisitos

- **Go ≥ 1.26**
- **PostgreSQL ≥ 14** (instalación nativa o vía Docker)
- **git** para clonar el repositorio

> Si prefieres no instalar PostgreSQL en tu máquina, salta al apartado **Docker** y vuelve aquí cuando lo tengas levantado.

---

## 3. Instalar PostgreSQL

### Ubuntu / Debian

```bash
sudo apt update
sudo apt install -y postgresql postgresql-contrib
sudo systemctl enable --now postgresql
```

### macOS (Homebrew)

```bash
brew install postgresql@16
brew services start postgresql@16
```

### Windows

Descarga el instalador oficial de [postgresql.org/download/windows](https://www.postgresql.org/download/windows/) y sigue el asistente. Marca la opción de instalar `psql`.

---

## 4. Crear la base de datos

Conéctate como superusuario `postgres` y crea un usuario y una base de datos dedicados a la aplicación:

```sql
-- Como superusuario (en Linux: `sudo -u postgres psql`)
CREATE USER skihub WITH PASSWORD 'una_contrasena_fuerte';
CREATE DATABASE skihub OWNER skihub;
GRANT ALL PRIVILEGES ON DATABASE skihub TO skihub;
```

Después prueba la conexión con:

```bash
psql -h localhost -U skihub -d skihub
```

> ⚠️ **No uses la cuenta `postgres` directamente**. Crear un usuario propio es buena práctica de seguridad y deja la BD aislada.

---

## 5. Variables de entorno

Copia `.env.example` a `.env` y rellena los valores reales:

```bash
cp .env.example .env
```

Variables obligatorias:

| Variable      | Descripción                                                      |
|---------------|------------------------------------------------------------------|
| `DB_HOST`     | Host del servidor Postgres (`localhost` en desarrollo)           |
| `DB_PORT`     | Puerto (típicamente `5432`)                                      |
| `DB_USER`     | Usuario de la BD (`skihub`)                                      |
| `DB_PASSWORD` | Contraseña del usuario                                           |
| `DB_NAME`     | Nombre de la base de datos (`skihub`)                            |
| `DB_SSLMODE`  | `disable` (dev), `require` (prod), `verify-full` (prod estricta) |

Variables opcionales con defaults sensatos: `DB_MAX_OPEN_CONNS=25`, `DB_MAX_IDLE_CONNS=5`, `DB_CONN_MAX_LIFETIME_MIN=30`, `APP_PORT=:8080`, `APP_TEMPLATES=web/templates`, `APP_STATIC=web/static`, `ADMIN_EMAIL`, `ADMIN_NAME`, `ADMIN_PASSWORD`.

> 🔐 El fichero `.env` ya está en `.gitignore`. **Nunca lo subas al repositorio.** En producción exporta las variables desde systemd, Docker o el orquestador correspondiente.

---

## 6. Ejecutar migraciones y seed

Las migraciones SQL viven en `db/migrations/` y son ficheros versionados (`001_create_users.sql`, `002_create_stations.sql`, …). El runner está embebido en la aplicación: **se ejecutan automáticamente al arrancar el servidor**, dentro de una transacción por fichero, y solo se aplican las que faltan (control con la tabla `schema_migrations`).

Eso significa que:

- En primera instalación, basta con arrancar la aplicación: creará todo el esquema y sembrará las 27 estaciones + las 6 noticias iniciales.
- En despliegues sucesivos, cualquier `*.sql` nuevo en `db/migrations/` se aplicará automáticamente al siguiente arranque.

Los seeds (`007_seed_stations.sql`, `008_seed_news.sql`) son **idempotentes** (`ON CONFLICT (name) DO NOTHING` para estaciones), así que volver a ejecutarlos no duplica datos.

> Si prefieres aplicarlas manualmente con `psql` antes de arrancar la app:
> ```bash
> for f in db/migrations/*.sql; do
>   psql -h localhost -U skihub -d skihub -v ON_ERROR_STOP=1 -f "$f"
> done
> ```

### Crear un administrador inicial

Si en `.env` defines `ADMIN_EMAIL`, `ADMIN_NAME` y `ADMIN_PASSWORD`, al arrancar el servidor:

- Si **no existe ningún admin**, crea uno con esos datos.
- Si ya existe un usuario con ese email, lo promociona a admin.
- Si ya hay algún admin en la BD, no hace nada.

Cambia esa contraseña en cuanto entres por primera vez.

---

## 7. Arrancar la aplicación

```bash
cd skihub
go mod download
go run ./cmd/servidor
```

Salida esperada:

```
[SKIHUB] Conectado a PostgreSQL skihub@localhost:5432/skihub (sslmode=disable)
[SKIHUB] [migraciones] ✓ aplicada: 001_create_users.sql
...
[SKIHUB] Administrador inicial creado: admin@skihub.local
[SKIHUB] Servidor escuchando en http://localhost:8080
```

Para producción se compila un binario:

```bash
go build -o bin/skihub ./cmd/servidor
./bin/skihub
```

---

## 8. Comprobar que la conexión funciona

Abre en el navegador `http://localhost:8080`. Deberías ver la home con las estaciones cargadas desde Postgres. Otras pruebas rápidas:

```bash
# API pública: lista de estaciones (debe devolver 27)
curl -s http://localhost:8080/api/estaciones | jq 'length'

# Comprobación directa de la conexión y los datos
psql -h localhost -U skihub -d skihub -c "SELECT COUNT(*) FROM stations;"
psql -h localhost -U skihub -d skihub -c "SELECT filename FROM schema_migrations ORDER BY filename;"
```

Si la app no arranca, los mensajes más típicos son:

| Síntoma                                                              | Solución                                                |
|----------------------------------------------------------------------|---------------------------------------------------------|
| `faltan variables de entorno obligatorias: DB_HOST, ...`             | Crea o corrige `.env` (mira `.env.example`).            |
| `ping Postgres localhost:5432: connect: connection refused`          | El servicio no está arrancado: `sudo systemctl start postgresql`. |
| `password authentication failed for user "skihub"`                   | Revisa `DB_USER` / `DB_PASSWORD` en `.env`.             |
| `database "skihub" does not exist`                                   | Vuelve al apartado 4 y crea la BD.                      |

---

## 9. Despliegue en Ubuntu detrás de Nginx

La aplicación escucha en `:8080`. Para servirla en `:80`/`:443` con un dominio propio se usa Nginx como reverse-proxy.

`/etc/nginx/sites-available/skihub.conf`:

```nginx
server {
    listen 80;
    server_name tu-dominio.com;

    # Si usas Let's Encrypt, certbot añadirá aquí el bloque HTTPS.

    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    # Cache estática (CSS / JS / imágenes), opcional
    location /static/ {
        proxy_pass         http://127.0.0.1:8080;
        expires            7d;
        add_header         Cache-Control "public, max-age=604800, immutable";
    }
}
```

```bash
sudo ln -s /etc/nginx/sites-available/skihub.conf /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

Ejemplo de unit de systemd (`/etc/systemd/system/skihub.service`):

```ini
[Unit]
Description=SkiHub web server
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=skihub
WorkingDirectory=/opt/skihub
EnvironmentFile=/opt/skihub/.env
ExecStart=/opt/skihub/bin/skihub
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now skihub
sudo journalctl -u skihub -f
```

---

## 10. Docker (opcional)

Si no quieres instalar PostgreSQL nativo, usa el `docker-compose.yml` incluido para levantarlo en contenedor durante el desarrollo:

```bash
docker compose up -d        # arranca Postgres en background
docker compose ps           # comprueba que está "healthy"
go run ./cmd/servidor       # ejecuta la web (apuntando a localhost:5432)
docker compose down         # para Postgres (los datos persisten)
docker compose down -v      # para Postgres y BORRA los datos
```

El compose lee del mismo `.env` que la aplicación (`DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`). Los datos se guardan en un volumen Docker llamado `pgdata`.

> Este compose **no levanta la app Go** dentro de un contenedor: la app sigue ejecutándose en tu máquina (más cómodo para iterar). Si quieres ese paso, créate un `Dockerfile` propio o pídelo y lo añadimos.

---

## 11. Estructura del proyecto

```
.
├── cmd/servidor/main.go            ← arranque, wiring, rutas
├── db/
│   └── migrations/                 ← *.sql versionados
│       ├── 001_create_users.sql
│       ├── 002_create_stations.sql
│       ├── 003_create_news.sql
│       ├── 004_create_sessions.sql
│       ├── 005_create_favorites.sql
│       ├── 006_create_orders.sql
│       ├── 007_seed_stations.sql
│       └── 008_seed_news.sql
├── internal/
│   ├── config/                     ← LoadFromEnv + DSN
│   ├── db/                         ← Conectar, EjecutarMigraciones, AsegurarAdmin
│   ├── handlers/                   ← controladores HTTP (web + API)
│   ├── services/                   ← lógica de negocio (con context.Context)
│   ├── repository/                 ← acceso a PostgreSQL
│   ├── models/                     ← entidades del dominio
│   └── infonieve/                  ← scraper externo (datos en directo)
├── web/
│   ├── templates/                  ← html/template
│   └── static/                     ← CSS / JS / imágenes
├── .env.example
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

---

## Cambios respecto a la versión SQLite

Para constancia académica de la migración:

- Driver: `modernc.org/sqlite` ❌ → `github.com/jackc/pgx/v5/stdlib` ✅
- Placeholders: `?` → `$1, $2, …`
- Auto-IDs: `INTEGER PRIMARY KEY AUTOINCREMENT` + `LastInsertId()` → `BIGSERIAL` + `INSERT … RETURNING id`
- Detección UNIQUE: parseo de texto → `pgconn.PgError.Code == "23505"`
- Hash de contraseña: SHA-256 + salt → **bcrypt cost 12** (migración completa)
- Fichero único `data/skihub.db` → BD relacional con `BOOLEAN`, `NUMERIC`, `TIMESTAMPTZ`, FKs `ON DELETE CASCADE`, CHECKs.
- Esquema gestionado con migraciones SQL versionadas + tabla `schema_migrations`.
- Configuración por variables de entorno (no más rutas hardcodeadas).
- Nuevas tablas `orders` y `order_items` + endpoint `POST /api/cesta/checkout` que persiste compras de forfait con precio congelado en cada línea (snapshot histórico).

Funcionalidad que **NO ha cambiado** desde fuera: rutas, contratos JSON existentes, plantillas, cesta en `localStorage`, scraper de pistas en directo (sigue siendo cache en memoria).
