# Makefile — atajos para tareas habituales de desarrollo y operación.
# Uso: `make help`

SHELL := /bin/bash
APP_NAME := skihub
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP_NAME)
PKG := ./cmd/servidor
GO ?= go

.DEFAULT_GOAL := help

.PHONY: help
help: ## Lista los targets disponibles
	@awk 'BEGIN {FS = ":.*##"; printf "\nUso: make \033[36m<target>\033[0m\n\nTargets:\n"} \
	     /^[a-zA-Z0-9_.-]+:.*##/ {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## ── Desarrollo ───────────────────────────────────────────────

.PHONY: run
run: ## Arranca la web localmente (lee .env)
	$(GO) run $(PKG)

.PHONY: build
build: ## Compila el binario en bin/
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) $(PKG)
	@echo "✅ Binario en $(BIN)"

.PHONY: tidy
tidy: ## Actualiza go.mod / go.sum
	$(GO) mod tidy

## ── Calidad ──────────────────────────────────────────────────

.PHONY: fmt
fmt: ## Formatea el código (gofmt + goimports si lo tienes)
	gofmt -w .
	@command -v goimports >/dev/null && goimports -w . || true

.PHONY: vet
vet: ## go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## Ejecuta golangci-lint (instálalo: https://golangci-lint.run)
	golangci-lint run ./...

.PHONY: check
check: fmt vet lint test ## Pasa todas las comprobaciones antes de commitear

## ── Tests ────────────────────────────────────────────────────

.PHONY: test
test: ## Ejecuta los tests con race detector
	$(GO) test -race ./...

.PHONY: cover
cover: ## Tests + reporte de cobertura HTML
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "📊 Cobertura: file://$(PWD)/coverage.html"

## ── Base de datos (desarrollo) ───────────────────────────────

.PHONY: db-up
db-up: ## Arranca Postgres en Docker
	docker compose up -d
	@echo "Postgres en localhost:$${DB_PORT:-5432}"

.PHONY: db-down
db-down: ## Para Postgres (los datos persisten)
	docker compose down

.PHONY: db-reset
db-reset: ## ⚠️  Borra el volumen de Postgres
	docker compose down -v

## ── Limpieza ─────────────────────────────────────────────────

.PHONY: clean
clean: ## Elimina binarios y artefactos
	rm -rf $(BIN_DIR) coverage.out coverage.html
