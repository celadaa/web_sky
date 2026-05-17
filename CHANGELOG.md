# Changelog

Todos los cambios notables de este proyecto se documentan aquí.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/)
y este proyecto se adhiere a [Semantic Versioning](https://semver.org/lang/es/).

## [Unreleased]

### Añadido
- Workflows de CI con build, vet, tests con race detector, gofmt y golangci-lint.
- Workflow de despliegue automático al VPS por SSH con health check y rollback manual.
- Análisis de seguridad: CodeQL, gitleaks, dependency review.
- Dependabot semanal para Go modules, GitHub Actions y Docker.
- Plantillas para Issues (bug / feature) y Pull Requests.
- `CODEOWNERS`, `SECURITY.md`, `CONTRIBUTING.md`.
- `.golangci.yml`, `.editorconfig` y `Makefile` con targets comunes.
- Wiki inicial en `docs/wiki/` (Arquitectura, Despliegue, Runbook, Datos).

### Cambiado
- README ampliado con badges de estado.

## [0.1.0] - 2026-05-17

### Añadido
- Versión inicial de SkiHub: portal Go + PostgreSQL para comparar estaciones,
  comprar forfaits y consultar pistas.

<!--
Ejemplo de plantilla para futuras releases:

## [0.2.0] - YYYY-MM-DD

### Añadido
- ...

### Cambiado
- ...

### Arreglado
- ...

### Eliminado
- ...

### Seguridad
- ...
-->

[Unreleased]: https://github.com/celadaa/web_sky/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/celadaa/web_sky/releases/tag/v0.1.0
