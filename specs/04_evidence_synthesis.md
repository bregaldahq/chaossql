# Spec 04: Evidence Synthesis (Mermaid & Repro Script)

## Objetivo
Transformar o trace m�nimo de falha em artefatos visuais e executáveis para diagnóstico imediato.

## Requisitos Verificáveis
1. **Diagrama Mermaid:** Gera um `sequenceDiagram` com atores por worker, mostrando a ordem de intercalação e a nota de violação no banco.
2. **Script Standalone (repro_test.py):** Gera um script Python autocontido (sem depender do ChaosSQL instalado) que reproduz a falha em 1 segundo.
3. **Relatório no Terminal:** Painel Rich com tabela de invariantes, mostrando valores esperados vs reais.
