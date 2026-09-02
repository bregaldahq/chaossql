# ChaosSQL Official Documentation Portal & Cloudflare Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build and deploy a world-class documentation portal and interactive landing website for ChaosSQL (`chaossql.bregalda.com`) incorporating official Bregalda brand assets, interactive scenario explorers, Go SDK guides, and Cloudflare Pages configuration.

**Architecture:**
- Create `site/` static web application directory with semantic HTML5, responsive CSS3 adhering to the Bregalda Design System (`#4B2E83`, `#F5C400`, `#22C55E`), and vanilla JavaScript controllers.
- Embed official Bregalda brand SVGs (`bregalda_monogram.svg`, `bregalda_wordmark.svg`, `icone_bregalda.svg`) into `site/assets/`.
- Configure `site/wrangler.toml` for Cloudflare Pages deployment to `chaossql.bregalda.com`.
- Update `tools/harness_check.go` for 33 artifacts.

**Tech Stack:** HTML5, CSS3 (Modern Glassmorphism & Bregalda Brand Palette), Vanilla JS, Prism.js syntax highlighting, Cloudflare Pages (`wrangler`).

**Spec:** `specs/12_documentation_portal_and_branding_site.md`

## Global Constraints
- Faithful implementation of Bregalda visual language (`#4B2E83`, `#F5C400`, `#22C55E`, `#0E0B16`).
- Zero external tracking or heavy bloated dependencies (fast $< 100\text{ms}$ loading).
- Responsive mobile & desktop layout with interactive tabs, copy-to-clipboard, and live simulation widgets.
- Commit and push to Git after each completed task.

---

### Task 1: Brand Asset Ingestion & Design System Architecture (`site/assets/`)

**Files:**
- Create: `site/assets/bregalda_monogram.svg`
- Create: `site/assets/bregalda_wordmark.svg`
- Create: `site/assets/icone_bregalda.svg`
- Create: `site/assets/style.css`

**Interfaces:**
- Produces: Complete CSS custom properties (`--bregalda-purple: #4B2E83`, `--bregalda-gold: #F5C400`, `--bregalda-green: #22C55E`, `--bg-dark: #0E0B16`), typography rules, card glassmorphism, buttons, and responsive grid layouts.

- [ ] **Step 1: Copy official SVG vector assets from `C:\Users\ricar\Documents\bregalda_brand` to `site/assets/`**
- [ ] **Step 2: Implement `site/assets/style.css` with dark theme, Bregalda neon glows, syntax highlighting styles, and responsive navigation**
- [ ] **Step 3: Commit and push**

---

### Task 2: Interactive Documentation & Landing Portal (`site/index.html` & `site/app.js`)

**Files:**
- Create: `site/index.html`
- Create: `site/app.js`

**Interfaces:**
- Produces:
  - Sticky navbar with Bregalda brand lockup, GitHub stars badge, and navigation links.
  - Hero section with live animated terminal ASCII interface and quickstart install command.
  - Interactive **9 Flagship Scenarios Explorer** with tabbed SQL schema, seed, chaos spec, and $ddmin$ reduction diagram.
  - Interactive **Adya Isolation Anomaly Theory** visualizer with transaction graph diagrams.
  - Interactive **Go Testing SDK (`pkg/chaostest`) Playground** with tabbed code examples.
  - **Hermitage Empirical Isolation Matrix** table.
  - **CLI Command Reference** (`run`, `demo`, `init`, `validate`, `bench`, `matrix`, `diff`, `replay`).
  - Interactive copy-to-clipboard buttons on all code snippets.

- [ ] **Step 1: Implement `site/index.html`**
- [ ] **Step 2: Implement `site/app.js` with interactive tab switching, scenario selector, and clipboard handling**
- [ ] **Step 3: Test rendering locally in browser / curl**
- [ ] **Step 4: Commit and push**

---

### Task 3: Cloudflare Pages Deployment Configuration (`site/wrangler.toml`)

**Files:**
- Create: `site/wrangler.toml`
- Create: `site/_headers`
- Create: `site/_redirects`

**Interfaces:**
- Produces: Cloudflare Pages configuration pointing to `chaossql.bregalda.com` with optimal caching headers and security policies.

- [ ] **Step 1: Create `site/wrangler.toml` specifying project name `chaossql` and compatibility date**
- [ ] **Step 2: Create `site/_headers` with security headers (CSP, X-Frame-Options, Cache-Control)**
- [ ] **Step 3: Commit and push**

---

### Task 4: Quality Gate & Harness Verification

**Files:**
- Modify: `tools/harness_check.go`
- Modify: `README.md`
- Modify: `Makefile` (add `make site` / `make serve-site`)

- [ ] **Step 1: Update `tools/harness_check.go` for 33 artifacts**
- [ ] **Step 2: Add `make serve-site` target in `Makefile`**
- [ ] **Step 3: Run `make verify` and commit/push to GitHub**
