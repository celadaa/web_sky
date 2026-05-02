-- Migración 001: tabla `users` y trigger genérico de updated_at.
-- ----------------------------------------------------------------------------
-- Crea la tabla principal de cuentas y una función reutilizable que
-- actualiza la columna updated_at de cualquier tabla que la incluya.

-- Función set_updated_at: trigger BEFORE UPDATE para tocar updated_at.
-- La definimos una vez aquí y la reutilizan el resto de tablas.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    name          TEXT        NOT NULL,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    is_admin      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Validaciones de integridad básica.
    CONSTRAINT users_name_len   CHECK (char_length(name)  BETWEEN 2 AND 60),
    CONSTRAINT users_email_fmt  CHECK (email ~* '^[^@\s]+@[^@\s]+\.[^@\s]+$'),
    CONSTRAINT users_email_len  CHECK (char_length(email) <= 120)
);

-- Índice por email ya viene implícito por UNIQUE; no añadimos extra.

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE  users               IS 'Cuentas de usuario de SkiHub (auth + perfil).';
COMMENT ON COLUMN users.password_hash IS 'Hash bcrypt (nunca contraseña en claro).';
COMMENT ON COLUMN users.is_admin      IS 'TRUE si la cuenta puede acceder al panel /admin.';
