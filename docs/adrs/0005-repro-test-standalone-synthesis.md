# ADR 0005: Síntese de Teste de Reprodução Autocontido (repro_test.go)

* **Status:** Aceito
* **Data:** 2026-09-01

## Contexto
Quando um bug é encontrado e reduzido, outros desenvolvedores e o pipeline de CI precisam reproduz�-lo sem necessitar da instalação do ChaosSQL completo.

## Decisão
Gerar um arquivo **`repro_test.go`** autocontido que inclui:
1. O schema e seed embeddados.
2. As goroutines exatas das 2 ou 3 transações do trace mínimo.
3. A asserção da invariante violada.

## Consequências
* Qualquer dev pode executar `go test -v repro_test.go` e ver o bug em 0.1s.
* Serve como prova irrefutável em Pull Requests e Issues.
