# Contributing Guide — ChaosSQL

Thank you for your interest in contributing to **ChaosSQL**!

## Engineering Principles
1. **Strict Determinism:** Any new capability must be 100% reproducible under the same seed.
2. **Test-Driven Development (TDD):** Every feature or bugfix must include unit and integration tests (`go test -race`).
3. **Zero CGO:** Code must compile statically without external C dependencies (`CGO_ENABLED=0`).
4. **Language Purity:** All non-portal codebase files, CLI commands, logs, and documentation must be written in 100% professional English.

## Useful Commands

```bash
make verify      # Run unified quality gate
make demo        # Run 10 interactive demonstration scenarios
make test        # Run test suite
make build       # Compile chaossql binary (Zero CGO)
```
