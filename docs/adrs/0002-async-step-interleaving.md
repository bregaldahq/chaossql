# ADR 0002: Async Step Interleaving e Jitter Injection

* **Status:** Aceito
* **Data:** 2026-09-01

## Contexto
Simplesmente disparar transações em paralelo nem sempre aciona a janela de condição de corrida (race condition) se as queries rodarem rápido demais.

## Decisão
Dividir cada operação em passos (`steps`) e injetar jitter (`asyncio.sleep(delay_ms / 1000.0)`) entre os passos dentro da mesma transação.

## Consequências
* Aumenta drasticamente a probabilidade de colisão entre leituras e escritas concorrentes.
* Permite testar cenários de latência alta no banco.
