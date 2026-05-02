-- Migración 003: tabla `news` (entradas del blog de noticias).

CREATE TABLE news (
    id             BIGSERIAL PRIMARY KEY,
    title          TEXT        NOT NULL,
    summary        TEXT        NOT NULL,
    category       TEXT        NOT NULL,
    category_class TEXT        NOT NULL,    -- clase CSS: nevada | consejos | evento | general
    published_at   DATE        NOT NULL,
    image_url      TEXT        NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT news_title_len CHECK (char_length(title) BETWEEN 1 AND 200)
);

CREATE INDEX idx_news_published_at ON news (published_at DESC);

CREATE TRIGGER trg_news_updated_at
BEFORE UPDATE ON news
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE news IS 'Entradas del blog de SkiHub.';
