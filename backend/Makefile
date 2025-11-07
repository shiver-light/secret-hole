# -------- settings --------
GO ?= go
ADDR ?= :8080
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/shudong?sslmode=disable
CMD ?= ./cmd/api

.PHONY: run dev-db up fmt tidy help

help:
	@echo "make run      - run api locally"
	@echo "make dev-db   - start db via docker compose"
	@echo "make up       - compose up (build)"
	@echo "make fmt      - gofmt -w ."
	@echo "make tidy     - go mod tidy"

run:
	ADDR=$(ADDR) DATABASE_URL=$(DATABASE_URL) $(GO) run $(CMD)

dev-db:
	docker compose up -d db

up:
	docker compose up --build

fmt:
	gofmt -w .

tidy:
	$(GO) mod tidy
