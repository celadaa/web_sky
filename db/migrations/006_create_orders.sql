-- Migración 006: pedidos de forfait + sus líneas de detalle.
-- ----------------------------------------------------------------------------
-- La cesta del cliente vive en localStorage del navegador. Cuando el
-- usuario confirma la compra (POST /api/cesta/checkout) se persiste un
-- pedido (`orders`) con sus líneas (`order_items`) en una transacción.
--
-- El precio de cada línea se congela en order_items.unit_price_eur para
-- que un cambio futuro de precios en `stations` no afecte a pedidos
-- pasados (snapshot histórico).

CREATE TABLE orders (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    total_eur    NUMERIC(10,2) NOT NULL DEFAULT 0,
    status       TEXT        NOT NULL DEFAULT 'paid',  -- paid | pending | cancelled
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT orders_total_nn  CHECK (total_eur >= 0),
    CONSTRAINT orders_status_ok CHECK (status IN ('paid','pending','cancelled'))
);

CREATE INDEX idx_orders_user_id    ON orders (user_id);
CREATE INDEX idx_orders_created_at ON orders (created_at DESC);

CREATE TRIGGER trg_orders_updated_at
BEFORE UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE orders IS 'Cabecera de pedido confirmado por un usuario.';

-- ----------------------------------------------------------------------------

CREATE TABLE order_items (
    id              BIGSERIAL PRIMARY KEY,
    order_id        BIGINT      NOT NULL REFERENCES orders(id)   ON DELETE CASCADE,
    station_id      BIGINT      NOT NULL REFERENCES stations(id) ON DELETE RESTRICT,

    -- Snapshots para historial: aunque cambien stations.* o desaparezca
    -- la estación, el ticket sigue siendo legible.
    station_name    TEXT        NOT NULL,
    pass_type       TEXT        NOT NULL,         -- adult | child | senior
    quantity        INTEGER     NOT NULL,
    days            INTEGER     NOT NULL,
    start_date      DATE        NOT NULL,
    end_date        DATE        NOT NULL,
    unit_price_eur  NUMERIC(8,2) NOT NULL,
    line_total_eur  NUMERIC(10,2) NOT NULL,

    CONSTRAINT order_items_qty_pos     CHECK (quantity > 0),
    CONSTRAINT order_items_days_pos    CHECK (days     > 0),
    CONSTRAINT order_items_unit_nn     CHECK (unit_price_eur >= 0),
    CONSTRAINT order_items_total_nn    CHECK (line_total_eur >= 0),
    CONSTRAINT order_items_pass_ok     CHECK (pass_type IN ('adult','child','senior')),
    CONSTRAINT order_items_dates_order CHECK (end_date >= start_date)
);

CREATE INDEX idx_order_items_order_id ON order_items (order_id);

COMMENT ON TABLE order_items IS 'Líneas de detalle de un pedido (un forfait por línea).';
