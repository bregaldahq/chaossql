# ADR 0004: SQLite Puro em Go (modernc.org/sqlite) vs CGO

* **Status:** Aceito
* **Data:** 2026-09-01

## Contexto
O driver tradicional `github.com/mattn/go-sqlite3` exige CGO, o que dificulta cross-compilação (Linux, MacOS, Windows), desativa binários estáticos e introduz overhead de troca de contexto goroutine->C.

## Decisão
Adotar `modernc.org/sqlite`, uma transpilação direta do SQLite para Go puro.

## Consequências
* **Compilação Cruzada Instantânea:** O binário do ChaosSQL pode ser gerado para qualquer SO com (`CGO_ENABLED=0 go build`).
* **Segurança de Memória:** Gerenciado diretamente pelo runtime do Go.
