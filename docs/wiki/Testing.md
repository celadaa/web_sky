# Testing

## Filosofía

- Tests **rápidos** > tests exhaustivos. Si tarda más de 30s, hay que repensarlo.
- **Tabla** > muchas funciones individuales (`tests := []struct{...}{...}`).
- Tests de unidad para `services` y `models`. Tests de integración para `repository`
  contra Postgres real (CI ya lo monta).
- No mockear lo que es trivial de usar de verdad.

## Comandos

```bash
make test            # tests con -race
make cover           # genera coverage.html
go test ./internal/services/... -run TestRegistrarUsuario -v
go test ./... -count=1     # invalida caché
```

## Convenciones

- Fichero al lado del código: `usuarios.go` ↔ `usuarios_test.go`.
- Funciones: `func TestNombre_Escenario_ResultadoEsperado(t *testing.T)`.
- Subtests con `t.Run("descripción", func(t *testing.T) { ... })`.
- Helpers en `internal/testutil/`.

## Tests contra BD real

El workflow de CI levanta un Postgres como service. Localmente:

```bash
make db-up
DB_HOST=127.0.0.1 DB_NAME=skihub_test go test ./internal/repository/...
```

Cada test que escribe debería:

1. Empezar una transacción.
2. Hacer su trabajo.
3. `tx.Rollback()` en defer — no contamina los siguientes.

O alternativamente: truncar tablas afectadas al final del test.

## Cobertura

No perseguimos un % concreto. Apuntar a:

- **Críticos** (`services/auth`, `services/pago`): > 80%.
- **Resto**: > 50% es un objetivo razonable, pero un test que aporta poco vale menos
  que ninguno.

## Tests E2E (futuro)

Cuando haya tiempo: Playwright contra un binario levantado en CI. De momento no es
prioridad — los tests de handler con `httptest` cubren lo crítico.
