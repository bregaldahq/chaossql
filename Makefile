GO ?= $(shell which /usr/local/go/bin/go go 2>/dev/null | head -n 1)

.PHONY: help bootstrap test lint verify check-harness build wasm demo bench matrix diff replay serve-site stress-wasm test-wasm-stress

help:
	@echo "ChaosSQL (Go 1.23+) Harness Commands:"
	@echo "  make bootstrap     - Download and verify all Go dependencies"
	@echo "  make test          - Run unit and integration test suite (-race)"
	@echo "  make lint          - Run go vet and static analysis"
	@echo "  make check-harness - Validate harness document integrity"
	@echo "  make build         - Compile chaossql binary (Zero CGO)"
	@echo "  make wasm          - Compile WebAssembly chaossql.wasm binary (Zero CGO)"
	@echo "  make demo          - Run 10 interactive demonstrations"
	@echo "  make bench         - Run performance and throughput benchmarks"
	@echo "  make matrix        - Run Hermitage empirical isolation matrix"
	@echo "  make stress-wasm   - Run headless WebAssembly and Web Worker stress harness"
	@echo "  make serve-site    - Start local HTTP server for documentation portal (port 8080)"
	@echo "  make verify        - Unified quality gate (check-harness + lint + test)"

wasm:
	@echo "Compiling ChaosSQL Core to WebAssembly (Zero CGO)..."
	@mkdir -p site/assets
	@CGO_ENABLED=0 GOOS=js GOARCH=wasm $(GO) build -ldflags="-s -w -X main.version=1.4.0" -trimpath -o site/assets/chaossql.wasm ./cmd/chaossql-wasm
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
	@CGO_ENABLED=0 $(GO) build -o bin/chaossql ./cmd/chaossql

bench: build
	@./bin/chaossql bench

matrix: build
	@./bin/chaossql matrix

serve-site:
	@echo "Starting ChaosSQL documentation portal at http://localhost:8080 ..."
	@python3 -m http.server 8080 --directory site

demo: build
	@echo "=== 1. Demonstrating Banking Lost Update (P4) ==="
	@./bin/chaossql demo banking || true
	@echo ""
	@echo "=== 2. Demonstrating Inventory Oversell (A3) ==="
	@./bin/chaossql demo inventory || true
	@echo ""
	@echo "=== 3. Demonstrating Hospital Write Skew (A5B) ==="
	@./bin/chaossql demo hospital || true
	@echo ""
	@echo "=== 4. Demonstrating Financial Audit Read Skew (A5A) ==="
	@./bin/chaossql demo financial || true
	@echo ""
	@echo "=== 5. Demonstrating Auction Bidding Dirty Write (G0) ==="
	@./bin/chaossql demo auction || true
	@echo ""
	@echo "=== 6. Demonstrating Crypto Arbitrage Circular Info (G1c) ==="
	@./bin/chaossql demo crypto || true
	@echo ""
	@echo "=== 7. Demonstrating Flash Crash Liquidation Dirty Read (G1a) ==="
	@./bin/chaossql demo flash_crash || true
	@echo ""
	@echo "=== 8. Demonstrating Ticket Seat Reservation Anti-Dependency (G2) ==="
	@./bin/chaossql demo ticket || true
	@echo ""
	@echo "=== 9. Demonstrating Deadlock Cycle & Timeout Diagnostics ==="
	@./bin/chaossql demo deadlock || true
	@echo ""
	@echo "=== 10. Demonstrating Foreign Key Cascade Deadlock & Referential Integrity ==="
	@./bin/chaossql demo fk || true

stress-wasm: ## Run headless WebAssembly & worker stress harness
	@node tools/headless_worker_stress.js

test-wasm-stress: stress-wasm

verify: check-harness lint test
	@node tools/test_english_purity.js && node tools/test_wasm_worker.js && node tools/test_playground_ui.js && node tools/test_wasm_bench.js && node tools/headless_worker_stress.js
	@echo ""
	@echo "✔ Verification gate completed successfully!"
