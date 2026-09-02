GO ?= $(shell which /usr/local/go/bin/go go 2>/dev/null | head -n 1)

.PHONY: help bootstrap test lint verify check-harness build demo bench matrix diff replay

help:
	@echo "ChaosSQL (Go 1.23+) Harness Commands:"
	@echo "  make bootstrap     - Baixa e verifica todas as dependencias Go"
	@echo "  make test          - Executa suite de testes unitarios e de integracao (-race)"
	@echo "  make lint          - Executa go vet e analise estatica"
	@echo "  make check-harness - Valida integridade dos 27 documentos do harness"
	@echo "  make build         - Compila o binario chaossql (Zero CGO)"
	@echo "  make demo          - Executa as 7 demonstracoes interativas"
	@echo "  make bench         - Executa benchmarks de performance e throughput"
	@echo "  make matrix        - Executa a matriz empirica de isolamento Hermitage"
	@echo "  make verify        - Gate unificado (check-harness + lint + test)"

check-harness:
	@$(GO) run tools/harness_check.go

bootstrap:
	@$(GO) mod tidy

test:
	@$(GO) test -v -race ./internal/... ./cmd/...

lint:
	@$(GO) vet ./...

build:
	@mkdir -p bin
	@$(GO) build -o bin/chaossql ./cmd/chaossql

bench: build
	@./bin/chaossql bench

matrix: build
	@./bin/chaossql matrix

demo: build
	@echo "=== 1. Demonstrando Banking Lost Update (P4) ==="
	@./bin/chaossql demo banking || true
	@echo ""
	@echo "=== 2. Demonstrando Inventory Oversell (A3) ==="
	@./bin/chaossql demo inventory || true
	@echo ""
	@echo "=== 3. Demonstrando Hospital Write Skew (A5B) ==="
	@./bin/chaossql demo hospital || true
	@echo ""
	@echo "=== 4. Demonstrando Financial Audit Read Skew (A5A) ==="
	@./bin/chaossql demo financial || true
	@echo ""
	@echo "=== 5. Demonstrando Auction Bidding Dirty Write (G0) ==="
	@./bin/chaossql demo auction || true
	@echo ""
	@echo "=== 6. Demonstrando Crypto Arbitrage Circular Info (G1c) ==="
	@./bin/chaossql demo crypto || true
	@echo ""
	@echo "=== 7. Demonstrando Flash Crash Liquidation Dirty Read (G1a) ==="
	@./bin/chaossql demo flash_crash || true

verify: check-harness lint test
	@echo ""
	@echo "✔ Gate de verificacao concluido com sucesso!"
