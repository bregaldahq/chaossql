# AGENTS.md — ChaosSQL Harness Engineering Protocol

Este documento define as regras operacionais, limites arquiteturais e garantias que humanos e agentes de IA devem seguir ao contribuir para o **ChaosSQL**.

---

## 1. Filosofia e Princípios Não-Negociáveis

1. **Determinismo Estrito:**
   Dada a mesma especificação (`chaos.yaml`) e a mesma `seed`, o ChaosSQL deve gerar exatamente a mesma sequência de operações, parâmetros e agendamento de workers. A aleatoriedade nunca deve depender de relc3�gio de parede não controlado ou estado global mutável.
2. **Verificabilidade por Invariantes:**
   Todo teste de concorrência deve expressar uma invariante matemática ou de negócio observável no banco. Não aceitamos testes baseados em "esperar 5 segundos e torcer para dar certo".
3. **Minimização Obrigatória (Delta-Debugging):**
   Ao encontrar uma violação em um trace de $N$ operações, o sistema tem o dever de reduzir o trace até isolar o subconjunto $1$-minimal reproduz�ivel.
4. **Isolamento de Camadas (Clean Architecture & Ports/Adapters):**
   * O **Domínio** (invariantes, modelo de trace, redução) não conhece drivers de banco específicos.
   * Os **Adaptadores** (SQLite, Postgres) apenas implementam a interface `DatabaseDriver`.
   * A **CLI e Relatórios** consumem resultados sem duplicar lógica de execução.
5. **Ciclo de Regressão Fechado:**
   Qualquer bug relatado deve virar uma fixture em `fixtures/` e um teste em `tests/` antes de ser considerado corrigido.

---

## 2. Superfície do Harness

| Componente | Localização | Responsabilidade |
| :--- | :--- | :--- |
| **Regras Operacionais** | `AGENTS.md` | Este contrato de trabalho para agentes e contribuidores |
| **Limites & Design** | `ARCHITECTURE.md` | Definição formal de camadas, portas e garantias de estado |
| **Decisões Arquiteturais** | `docs/adrs/` | Registros imutáveis de decisão técnica (ADRs) |
| **Especificações Executáveis** | `specs/` | Requisitos formais por capacidade do sistema |
| **Fixtures & Casos Adversariais** | `fixtures/` | Schemas, seeds e cenários de anomalias (lost update, write skew, etc.) |
| **Critérios de Qualidade (Evals)** | `evals/` | Benchmarks de taxa de redução do shrinker e tempo de convergência |
| **Gate Local de Validação** | `make verify` | Comando único que executa lint, tipos, testes e integridade do harness |

---

## 3. Fluxo de Trabalho do Agente

1. **Nunca quebre o `make verify`:** Nenhuma alteração é aceita se o lint (`ruff`), checagem de tipos (`mypy`) ou suíte de testes (`pytest`) falharem.
2. **Adicione documentação viva:** Se alterar o comportamento do fuzzer ou do shrinker, atualize a respectiva spec em `specs/`.
3. **Mantenha a Cobertura Mínima:** A cobertura de código não deve ficar abaixo de **85%**.

---

## 4. Comandos Essenciais

```bash
make bootstrap   # Cria virtualenv e instala todas as dependências
make test        # Executa testes unitários e de integração
make lint        # Executa ruff e mypy
make verify      # Executa o gate completo de qualidade
make demo        # Executa os cenários demonstrativos (Lost Update, Oversell, Write Skew)
```
