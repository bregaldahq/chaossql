# Security Policy & Defensive Architecture

ChaosSQL is designed with a defense-in-depth security posture, ensuring safe execution in automated CI/CD pipelines, containerized clusters, and multi-tenant testing environments.

---

## 🛡️ Security Architecture & Guardrails

### 1. Zero CGO Memory Safety
- **Design:** ChaosSQL uses pure Go database drivers (including `modernc.org/sqlite`).
- **Guarantee:** Eliminates memory corruption risks, use-after-free, and buffer overflow vulnerabilities common in native C/C++ SQLite wrappers.

### 2. Invariant Expression Sandboxing
- **Design:** Invariant expressions (`assert: ...`) are compiled and executed using the isolated `expr-lang/expr` runtime.
- **Guarantee:** Expressions cannot perform system calls, invoke unauthorized reflection, execute shell commands, or access the OS filesystem.

### 3. Credential & DSN Redaction
- **Design:** All database DSNs are sanitized via `drivers.MaskDSN`.
- **Guarantee:** Plaintext passwords, authentication tokens, and private URIs are redacted (`******`) from runtime logs, CLI error outputs, OpenTelemetry traces, JUnit XML exports, and HTML reports.

### 4. Path Traversal Containment
- **Design:** Schema and seed SQL file resolution uses cleaned directory paths bounded to the scenario workspace.
- **Guarantee:** Prevents arbitrary file reads outside the intended scenario directory.

### 5. Deterministic Non-Destructive Fuzzing
- **Design:** Target database schemas and test seeds are strictly scoped to isolated test databases or in-memory SQLite instances (`:memory:`).

---

## 🔒 Reporting a Vulnerability

If you discover a potential security vulnerability in ChaosSQL:
1. **Do NOT open a public issue.**
2. Please privately email the maintainer: **`security@bregalda.io`** or use GitHub's private vulnerability reporting feature on [`https://github.com/bregaldahq/chaossql/security/advisories`](https://github.com/bregaldahq/chaossql/security/advisories).
3. Include detailed steps to reproduce the issue and environment details.

We will acknowledge receipt within 24 hours and provide a remediation timeline.
