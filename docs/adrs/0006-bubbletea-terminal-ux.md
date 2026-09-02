# ADR 0006: UI de Terminal com Bubbletea e Lipgloss

* **Status:** Aceito
* **Data:** 2026-09-01

## Contexto
Uma ferramenta de chaos engineering precisa ser nuanciada e elegante, mostrando gráficos de barras de intercalação e tabelas de invariantes em tempo real.

## Decisão
Utilizar as bibliotecas **Lipgloss** e **Bubbletea** (Charm.sh) para formatação da CLI.

## Consequências
* Interface visualmente impactante (bordas arredondadas, cores de status verde/vermelho, tabelas de sums de invariantes).
* Exportação JSON e Mermaid nativa via flags (`--json`, `--mermaid`).
