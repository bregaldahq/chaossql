# Full English Unification & Localization Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Translate and unify all non-website files across the ChaosSQL project into professional English, ensuring that Portuguese exists strictly and exclusively as an optional localized view in the web portal (`site/`).

**Architecture:** Audit every non-site directory (`docs/`, `specs/`, `evals/`, `examples/`, `internal_docs/`, `tools/`), translate all Portuguese academic theory, architecture decision records, specifications, scenario documentation, and terminal tool outputs into high-grade academic and technical English. Implement an automated localization gate in `tools/` that asserts zero Portuguese remnants exist outside of `site/`, preserving full test suite passing status.

**Tech Stack:** Go 1.23, Markdown, Node.js, Bash, Makefile, GitHub Actions.

**Spec:** Requirement from user: "o projeto está em portugues ou ingles? os retornos dos fluxos? Deve ser tudo em ingles, apenas portugues no site que temos a opção de selecionar. Caso esteja em portugues faça um planejamento usando planiling os superpower e execute em seguida no modo looping faça a execução do fluxo usando subagentes."

## Global Constraints

- **Zero CGO:** `CGO_ENABLED=0` remains mandatory across all Go builds and tests.
- **Verification Gate:** `make verify && make demo` must remain 100% GREEN without regressions.
- **Strict English Purity:** All CLI outputs, terminal banners, code comments, error messages, documentation files, ADRs, specs, evals, and example READMEs outside `site/` must be 100% English.
- **Preserve Site Bilingual Capabilities:** The web portal (`site/`) retains its bilingual toggle (`EN` / `PT`) using `site/app.js` and `site/docs-data.js`.
- **TDD & Evidence:** Every task must end with verified tests and a git commit.

---

### Task 1: Academic & Architectural Documentation Translation

**Files:**
- Modify: `docs/ACADEMIC_FOUNDATIONS.md`
- Modify: `docs/THEORY.md`
- Modify: `docs/SCENARIO_ACADEMIC_AUDIT.md`
- Modify: `docs/adrs/0001-deterministic-prng-and-replay.md`
- Modify: `docs/adrs/0002-async-step-interleaving.md`
- Modify: `docs/adrs/0003-causal-delta-debugging-shrinker.md`
- Modify: `docs/adrs/0004-pure-go-sqlite-vs-cgo.md`
- Modify: `docs/adrs/0005-repro-test-standalone-synthesis.md`
- Modify: `docs/adrs/0006-bubbletea-terminal-ux.md`

**Interfaces:**
- Consumes: Existing Portuguese academic papers, theorems, and ADRs.
- Produces: Professional English markdown files maintaining exact math formulas, citations, and section headings.

- [ ] **Step 1: Translate `docs/ACADEMIC_FOUNDATIONS.md` to English**
Translate title to "ChaosSQL Academic Foundations & Formal State of the Art", sections on PCT-SQL, Elle Dependency Inference, Hermitage Taxonomy, NoREC Metamorphic Fuzzing, Delta-Debugging, and Client-Side In-Browser Formal Verification (WASM Architecture).

- [ ] **Step 2: Translate `docs/THEORY.md` and `docs/SCENARIO_ACADEMIC_AUDIT.md` to English**
Translate transaction history models, Bernstein conflict conditions, CSR theorem, Zeller 1-minimality proofs, and scenario academic alignment tables into English.

- [x] **Step 3: Translate ADRs 0001 through 0006 to English**
Update headers ("Status: Accepted", "Context", "Decision", "Consequences") and body text.

- [x] **Step 4: Verify markdown formatting and commit**
Run: `git diff docs/` to verify clean translation.
Commit: `git commit -m "docs: translate academic foundations, formal theory, and ADRs to English"`

---

### Task 2: Standardize Specs, Evals, and Internal Documentation to English

**Files:**
- Modify: `specs/01_invariant_evaluation.md`
- Modify: `specs/02_concurrency_interleaving.md`
- Modify: `specs/03_delta_debugging_shrinker.md`
- Modify: `specs/04_evidence_synthesis.md`
- Modify: `evals/01_shrinking_ratio.md`
- Modify: `evals/02_false_positive_rate.md`
- Modify: `internal_docs/README.md`
- Modify: `internal_docs/01_wasm_in_browser_playground.md`
- Modify: `internal_docs/02_transparent_database_proxy.md`
- Modify: `internal_docs/03_multilanguage_sdks.md`
- Modify: `internal_docs/04_github_pr_commenter_bot.md`
- Modify: `internal_docs/05_automated_sql_fix_synthesizer.md`

**Interfaces:**
- Consumes: Legacy Portuguese specs (01-04) and roadmap notes in `internal_docs/`.
- Produces: Consistent English specs aligned with specs 05-15.

- [ ] **Step 1: Translate specs 01, 02, 03, 04 to English**
Fix encoding issues (e.g. `mltiplos` -> `multiple`) and translate goals, requirements, and acceptance criteria to English.

- [ ] **Step 2: Translate evals 01 and 02 to English**
Standardize evaluation hypotheses, methodologies, and metrics tables to English.

- [ ] **Step 3: Translate internal_docs roadmap files to English**
Translate `internal_docs/README.md` and all 5 roadmap architecture designs (WASM Playground, Transparent DB Proxy, Multi-language SDKs, GitHub PR Commenter Bot, Automated SQL Fix Synthesizer).

- [ ] **Step 4: Verify specs and commit**
Run: `go run tools/harness_check.go` to confirm artifact integrity.
Commit: `git commit -m "docs(specs,evals): standardize legacy specifications, evals, and internal roadmap to English"`

---

### Task 3: Translate Legacy Example Scenario READMEs to English

**Files:**
- Modify: `examples/banking_lost_update/README.md`
- Modify: `examples/inventory_oversell/README.md`
- Modify: `examples/hospital_write_skew/README.md`
- Modify: `examples/read_skew_financial_audit/README.md`
- Modify: `examples/dirty_write_auction/README.md`

**Interfaces:**
- Consumes: Portuguese READMEs from scenarios 01-05.
- Produces: Professional English scenario descriptions matching scenarios 06-10 (`Business Context`, `Anomaly Breakdown`, `Consistency Invariant`, `Formal Solution`).

- [ ] **Step 1: Translate scenario 01 (Banking Lost Update) and scenario 02 (Inventory Oversell)**
Translate business context, problem narrative, invariant definitions, and mitigations.

- [ ] **Step 2: Translate scenario 03 (Hospital Write Skew), scenario 04 (Read Skew), and scenario 05 (Dirty Write)**
Translate on-call doctors scheduling, financial ledger audit, and concurrent auction bidding narratives.

- [ ] **Step 3: Verify all 10 scenario READMEs are in English**
Run: `head -n 5 examples/*/README.md` to confirm unified English headers.
Commit: `git commit -m "docs(examples): translate legacy scenario READMEs 01-05 to English"`

---

### Task 4: Harmonize Tools Terminal Outputs and SVG Text to English

**Files:**
- Modify: `tools/harness_check.go:60-70`
- Modify: `tools/headless_worker_stress.js:380-385`

**Interfaces:**
- Consumes: Existing tool scripts with Portuguese stdout and SVG strings.
- Produces: 100% English outputs for CLI execution and stress benchmarks.

- [ ] **Step 1: Update `tools/harness_check.go`**
Replace `❌ [FALTANDO] Artefato obrigatorio...` with `❌ [MISSING] Mandatory Harness artifact: %s
`.
Replace `[ERRO] %d artefatos ausentes no Harness.` with `[ERROR] %d artifacts missing in Harness.
`.
Replace `[HARNESS OK] Todos os %d artefatos...` with `[HARNESS OK] All %d Harness artifacts are present and verified.
`.

- [ ] **Step 2: Update `tools/headless_worker_stress.js`**
Replace `CICLO ADYA CLASSIFICADO:` with `CLASSIFIED ADYA CYCLE:`.

- [ ] **Step 3: Verify tools execution**
Run: `go run tools/harness_check.go`
Run: `node tools/test_wasm_worker.js && node tools/test_playground_ui.js`
Commit: `git commit -m "fix(tools): harmonize harness checker and headless stress harness stdout to English"`

---

### Task 5: Automated English Purity Verification Gate & Regression Audit

**Files:**
- Create: `tools/test_english_purity.js`
- Modify: `Makefile` (integrate purity check into `verify` target)

**Interfaces:**
- Consumes: All repository files excluding `site/` (and git history).
- Produces: Automated test suite asserting zero Portuguese keywords in non-site files.

- [ ] **Step 1: Create `tools/test_english_purity.js`**
Implement regex scanning for common Portuguese vocabulary across `docs/`, `specs/`, `evals/`, `examples/`, `cmd/`, `internal/`, `pkg/`, `tools/`. Assert 0 violations found.

- [ ] **Step 2: Integrate into Makefile**
Add `node tools/test_english_purity.js` to `make verify`.

- [ ] **Step 3: Run full verification gate**
Run: `make verify && make demo`
Ensure all Go tests, WASM tests, UI tests, and English purity tests pass with exit code 0.
Commit: `git commit -m "ci(test): implement automated english purity verification gate in make verify"`
