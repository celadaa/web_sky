-- Migración 013: columnas para noticias externas (RSS/API).
--
-- Añade a la tabla `news` los campos necesarios para almacenar noticias
-- obtenidas de fuentes externas (RSS, APIs), manteniendo compatibilidad
-- total con las noticias manuales existentes.

ALTER TABLE news
    ADD COLUMN IF NOT EXISTS source_name  TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS source_url   TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS original_url TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS slug         TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS is_external  BOOLEAN     NOT NULL DEFAULT FALSE;

-- Índice único para deduplicación: si original_url no está vacía, no se
-- permite insertar la misma noticia externa dos veces.
-- Las noticias manuales tienen original_url = '' por lo que no chocan.
CREATE UNIQUE INDEX IF NOT EXISTS idx_news_original_url
    ON news (original_url)
    WHERE original_url <> '';

-- Índice para filtrado por categoría (usado en /noticias?categoria=X).
CREATE INDEX IF NOT EXISTS idx_news_category_class
    ON news (category_class);

COMMENT ON COLUMN news.source_name  IS 'Nombre de la fuente (p.ej. "Nevasport")';
COMMENT ON COLUMN news.source_url   IS 'URL raíz de la fuente';
COMMENT ON COLUMN news.original_url IS 'URL original del artículo; clave única para deduplicación';
COMMENT ON COLUMN news.slug         IS 'Slug URL-friendly generado del título';
COMMENT ON COLUMN news.is_external  IS 'true = noticia obtenida de fuente externa';
