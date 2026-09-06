# Spec 01: Invariant Evaluation Engine

## Objective
Deterministically and securely evaluate database invariant assertions expressed in SQL.

## Verifiable Requirements
1. **SQL Execution:** The invariant query must return a single row with named columns (e.g., `SELECT SUM(balance) AS total FROM accounts;`).
2. **Safe Eval Expression:** Column names are injected as local variables into the boolean expression (e.g., `total == 10000`).
3. **Security Isolation:** No OS functions or unsafe built-in operations (e.g., `__import__`, `os`, arbitrary system calls) are permitted within the evaluation environment.
4. **Structured Reporting:** If an assertion fails, the result must contain: invariant name, evaluated boolean expression, actual database values observed, and an explainable diagnostic error message.
