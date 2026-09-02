.PHONY: help bootstrap test lint verify check-harness build demo

export PATH := /usr/local/go/bin:$(PATH)
export GOROOT := /usr/local/go
GO := /usr/local/go/bin/go

help:
	@echo "ChaosSQL (Go 1.23+) Harness Commands:"
	@echo "  make bootstrap   - Baixa e verifica todas as dependencias Go"
	@echo "  make test        - Executa suite de testes unitarios e integracao"
	@echo "  make lint        - Executa go vet e analise estatica"
	@echo "  make check-harness - Valida integridade dos documentos do harness"
	@echo "  make build       - Compila o binario chaossql"
	@echo "  make demo        - Executa os 3 cenarios de demonstracao"
	@echo "  make verify      - Gate unificado (check + vet + test)"

check-harness:
	@$(GO) run tools/harness_check.go

bootstrap:
	@$(GO) mod tidy

test:
	@$(GO) test -v -cover ./internal/...

lint:
	@$(GO) vet ./...

build:
	@$(GO) build -o bin/chaossql cmd/chaossql/main.go

demo: build
	@./bin/chaossql demo banking
	@./bin/chaossql demo inventory
	@./bin/chaossql demo hospital

verify: check-harness lint test
	@echo ""
	@echo "✔ Gate de verificacao concluido com sucesso!"
