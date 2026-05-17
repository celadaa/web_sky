# Modelo de datos

Esquema Postgres del proyecto. Para el detalle exacto, mira siempre las
migraciones en `db/migrations/`, que son la fuente de verdad.

## Convenciones

- Nombres de tabla en **plural** y `snake_case`: `usuarios`, `estaciones`, `forfaits`.
- Claves primarias `id BIGSERIAL` salvo cuando hay una clave natural clara.
- Timestamps `created_at` / `updated_at` con `TIMESTAMPTZ DEFAULT now()`.
- Borrados lógicos (`deleted_at IS NULL`) sólo cuando es necesario; preferimos
  borrado físico con `ON DELETE CASCADE`.
- Texto: `TEXT` (no `VARCHAR(N)` salvo que tengas razón fuerte).
- Dinero: `NUMERIC(10,2)` siempre, jamás `FLOAT`.

## Entidades principales

```
usuarios          (id, email, password_hash, nombre, rol, created_at, ...)
sesiones          (token, usuario_id, expira_en, created_at)

estaciones        (id, slug, nombre, comunidad, altitud_max, ...)
pistas            (id, estacion_id, nombre, dificultad, longitud_m, ...)
estado_pistas     (id, pista_id, abierta, scrapeado_en)

forfaits          (id, estacion_id, nombre, precio_eur, dias, ...)
pedidos           (id, usuario_id, total_eur, estado, created_at)
pedido_lineas     (id, pedido_id, forfait_id, cantidad, precio_unit)

favoritos         (usuario_id, estacion_id, created_at)
                  PRIMARY KEY (usuario_id, estacion_id)

noticias          (id, titulo, slug, cuerpo_html, publicada_en)
```

> Esto es un esquema indicativo. Mira `db/migrations/*.sql` para los
> nombres y columnas exactos.

## Relaciones (resumen)

- `pistas.estacion_id → estaciones.id` (FK, cascade)
- `forfaits.estacion_id → estaciones.id` (FK, cascade)
- `pedido_lineas.pedido_id → pedidos.id` (FK, cascade)
- `pedido_lineas.forfait_id → forfaits.id` (FK, restrict — no borrar forfait si está en un pedido)
- `favoritos` es una tabla puente (PK compuesta).

## Migraciones

- Cada migración es un fichero `.sql` con prefijo numérico: `00007_add_indice_pistas.sql`.
- El runner aplica en orden alfabético y registra cuáles ya corrieron.
- **Las migraciones no se editan después de mergearse a `main`**. Si te equivocaste,
  crea una nueva que arregle el estado.

## Índices

Mira `db/migrations/` para los `CREATE INDEX`. Por defecto creamos:

- `(estacion_id)` en `pistas` y `forfaits`.
- `(usuario_id, created_at DESC)` en `pedidos` para listados.
- `(slug)` único en `estaciones` y `noticias`.

## Consultas pesadas

Si una query empieza a doler, **antes de optimizar el SQL**:

1. `EXPLAIN ANALYZE` la consulta real con datos de producción (o un dump).
2. Comprueba si falta un índice.
3. Considera materializar o cachear si los datos cambian poco.

## Backup

Ver [Runbook](Runbook-Operacional#backups).
