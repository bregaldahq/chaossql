# Guia de Contribuição — ChaosSQL

Obrigado por seu interesse em contribuir com o **ChaosSQL**!

## Princípios de Engenharia
1. **Determinismo Estrito:** Qualquer novo recurso deve ser 100% reprodutível sob a mesma seed.
2. **Test-Driven Development (TDD):** Toda funcionalidade ou correção deve conter testes unitários e de concorrência (go test -race).
3. **Zero CGO:** O código deve compilar estaticamente sem dependências externas de C.

## Comandos Úteis
