# ChaosSQL

> **Deterministic Concurrency & Invariant Fuzzer for SQL Databases**
> *Finding isolation anomalies (Lost Update, Write Skew, Phantom Reads) and shrinking chaotic traces to 1-minimal reproductions.*

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![Zero CGO](https://img.shields.io/badge/CGO-Disabled%20(Pure%20Go)-success)](https://modernc.org/sqlite)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Harness Engineering](https://img.shields.io/badge/Harness-Verified-blueviolet)](file:///root/chaossql/AGENTS.md)

---

## Theoretical & Academic Foundations

ChaosSQL bridges empirical chaos engineering with seminal academic database theory:

- **PCT-SQL (Probabilistic Concurrency Testing):** Inspired by Burckhardt & Musuvathi (ASPLOS 2010), bounding bug-detection depth with randomized priority schedules.
- **Adya & Elle Cycle Inference:** Formalized by Atul Adya (MIT PhD 1999) and Kingsbury & Alvaro (VLDB 2020), inferring directed dependency graphs over client-observed histories.
- **Causal Delta-Debugging (ddmin):** Derived from Andreas Zeller (IEEE TSE 2002), pruning non-essential noise transactions down to a 1-minimal failing sequence with foreign-key closure guarantees.

---

## 3 Flagship Demonstration Scenarios

1. **Banking Lost Update (P4):** Concurrent bank balance withdrawals under READ COMMITTED losing financial updates.
2. **Inventory Oversell (A3):** Flash sale inventory depletion where concurrent shoppers deplete remaining items.
3. **Hospital Write Skew (A5B):** Doctors going off-call concurrently under Snapshot Isolation, leaving 0 active doctors on duty.

---

## Quickstart



---

## License

MIT License (c) 2026 Ricardo Bregalda (@bregaldahq).
