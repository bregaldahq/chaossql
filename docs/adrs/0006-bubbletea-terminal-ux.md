# ADR 0006: Terminal UI with Bubbletea and Lipgloss

* **Status:** Accepted
* **Date:** 2026-09-01

## Context
A chaos engineering tool requires clear, nuanced terminal visualization, displaying real-time interleaving graphs and invariant verification tables.

## Decision
Use the **Lipgloss** and **Bubbletea** libraries (Charm.sh) for CLI formatting and interactive terminal UI.

## Consequences
* Visually impactful terminal interface (rounded borders, green/red status indicators, live invariant tables).
* Native JSON and Mermaid diagram export via CLI flags (`--json`, `--mermaid`).
