-- Migración 012: soporte de OAuth (Google) y verificación de email.
-- ----------------------------------------------------------------------------
-- Añade columnas a la tabla `users` para:
--   * Verificación de email vía token enviado por correo (registro local).
--   * Vinculación con cuentas Google (auth_provider, google_id, avatar_url).
--
-- Reglas:
--   * Idempotente lo razonable: usamos ADD COLUMN IF NOT EXISTS para que
--     re-aplicar la migración a mano no rompa.
--   * Sin pérdida de datos: defaults seguros para usuarios ya creados
--     (email_verified=FALSE, auth_provider='local').
--   * UNIQUE en google_id usando un índice parcial: permite múltiples NULL
--     y garantiza unicidad cuando hay un sub real de Google.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email_verified                BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS email_verified_at             TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS google_id                     TEXT        NULL,
    ADD COLUMN IF NOT EXISTS avatar_url                    TEXT        NULL,
    ADD COLUMN IF NOT EXISTS auth_provider                 TEXT        NOT NULL DEFAULT 'local',
    ADD COLUMN IF NOT EXISTS verification_token_hash       TEXT        NULL,
    ADD COLUMN IF NOT EXISTS verification_token_expires_at TIMESTAMPTZ NULL;

-- Constraint: auth_provider sólo acepta valores conocidos.
-- Lo añadimos con DROP+ADD para que la migración sea aplicable también si
-- alguien hizo cambios a mano que dejaron un constraint con el mismo nombre.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_auth_provider_chk;
ALTER TABLE users ADD CONSTRAINT users_auth_provider_chk
    CHECK (auth_provider IN ('local', 'google'));

-- Índice parcial UNIQUE: una cuenta Google sólo puede vincularse a un usuario.
-- NULLs no compiten entre sí gracias al WHERE.
CREATE UNIQUE INDEX IF NOT EXISTS users_google_id_uniq
    ON users (google_id)
    WHERE google_id IS NOT NULL;

-- Índice para la búsqueda por hash de token al confirmar email. Pequeño
-- (suele haber muy pocas filas con token activo) pero acelera la lookup.
CREATE INDEX IF NOT EXISTS users_verification_token_idx
    ON users (verification_token_hash)
    WHERE verification_token_hash IS NOT NULL;

COMMENT ON COLUMN users.email_verified               IS 'TRUE cuando la dirección de correo ha sido confirmada (link o Google verified).';
COMMENT ON COLUMN users.email_verified_at            IS 'Instante en que se marcó email_verified=TRUE.';
COMMENT ON COLUMN users.google_id                    IS 'sub de Google OIDC. Único cuando no es NULL.';
COMMENT ON COLUMN users.avatar_url                   IS 'URL del avatar (picture de Google, en otros providers podría ser otro).';
COMMENT ON COLUMN users.auth_provider                IS 'local | google. Identifica cómo se creó la cuenta originalmente.';
COMMENT ON COLUMN users.verification_token_hash      IS 'SHA-256 hex del token enviado por correo. Nunca se almacena el token en claro.';
COMMENT ON COLUMN users.verification_token_expires_at IS 'Caducidad del token de verificación (24h por defecto).';
