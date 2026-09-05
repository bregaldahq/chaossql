GO ?= $(shell which /usr/local/go/bin/go go 2>/dev/null | head -n 1)

.PHONY: help bootstrap test lint verify check-harness build wasm demo bench matrix diff replay serve-site

help:
	@echo "ChaosSQL (Go 1.23+) Harness Commands:"
	@echo "  make bootstrap     - Baixa e verifica todas as dependencias Go"
	@echo "  make test          - Executa suite de testes unitarios e de integracao (-race)"
	@echo "  make lint          - Executa go vet e analise estatica"
	@echo "  make check-harness - Valida integridade dos documentos do harness"
	@echo "  make build         - Compila o binario chaossql (Zero CGO)"
	@echo "  make wasm          - Compila o binario WebAssembly chaossql.wasm (Zero CGO)"
	@echo "  make demo          - Executa as 10 demonstracoes interativas"
	@echo "  make bench         - Executa benchmarks de performance e throughput"
	@echo "  make matrix        - Executa a matriz empirica de isolamento Hermitage"
	@echo "  make serve-site    - Inicia servidor HTTP local para o portal de documentacao (porta 8080)"
	@echo "  make verify        - Gate unificado (check-harness + lint + test)"

wasm:
	@echo "Compilando ChaosSQL Core para WebAssembly (Zero CGO)..."
	@mkdir -p site/assets
	@CGO_ENABLED=0 GOOS=js GOARCH=wasm $(GO) build -ldflags="-s -w -X main.version=1.3.0" -trimpath -o site/assets/chaossql.wasm ./cmd/chaossql-wasm
	@ls -lh site/assets/chaossql.wasm

check-harness:
	@$(GO) run tools/harness_check.go

bootstrap:
	@$(GO) mod tidy

test:
	@$(GO) test -v -race ./internal/... ./cmd/... ./pkg/...

lint:
	@$(GO) vet ./...

build:
	@mkdir -p bin
	@$(GO) build -o bin/chaossql ./cmd/chaossql

bench: build
	@./bin/chaossql bench

matrix: build
	@./bin/chaossql matrix

serve-site:
	@echo "Iniciando portal de documentacao do ChaosSQL em http://localhost:8080 ..."
	@python3 -m http.server 8080 --directory site

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
	@echo ""
	@echo "=== 8. Demonstrando Ticket Seat Reservation Anti-Dependency (G2) ==="
	@./bin/chaossql demo ticket || true
	@echo ""
	@echo "=== 9. Demonstrando Deadlock Cycle & Timeout Diagnostics ==="
	@./bin/chaossql demo deadlock || true
	@echo ""
	@echo "=== 10. Demonstrando Foreign Key Cascade Deadlock & Referential Integrity ==="
	@./bin/chaossql demo fk || true

verify: check-harness lint test
	@node tools/test_wasm_worker.js && node tools/test_playground_ui.js
	@echo ""
	@echo "✔ Gate de verificacao concluido com sucesso!"
