# Spec 12: ChaosSQL Official Documentation Portal & Landing Website (`chaossql.bregalda.com`)

## 1. Brand Identity & Visual Language
- Incorporates official Bregalda Brand Assets from `C:\Users\ricar\Documents\bregalda_brand`:
  - **Primary Purple:** `#4B2E83` (Hero gradients, primary CTAs, active states)
  - **Accent Gold / Yellow:** `#F5C400` (Highlight badges, anomaly indicators, star ratings)
  - **Accent Green:** `#22C55E` (Passing tests, safe isolation indicators, terminal success)
  - **Background & Dark Shades:** `#0E0B16`, `#181224`, `#2A2140`, `#FCFBF8` (Light contrast)
  - **Logos & Monogram:** Embedded SVG vector assets (`bregalda_monogram.svg`, `bregalda_wordmark.svg`, `icone_bregalda.svg`).

## 2. Information Architecture & Interactive Features
- **Hero & Value Proposition:** Dynamic interactive isolation anomaly simulator with terminal preview.
- **Academic Foundations:** Interactive visualization of Adya directed dependency graphs ($SG(S)$).
- **Interactive 9 Scenarios Catalog:** Tabbed interactive explorer with SQL schema, seed, chaos workload, and $ddmin$ reduction breakdown for all 9 scenarios:
  1. Banking Lost Update ($P4$)
  2. Inventory Oversell ($A3$)
  3. Hospital Write Skew ($A5B$)
  4. Financial Read Skew ($A5A$)
  5. Auction Dirty Write ($G0$)
  6. Crypto Arbitrage ($G1c$)
  7. Flash Crash Dirty Read ($G1a$)
  8. Ticket Booking Anti-Dependency ($G2$)
  9. Deadlock Cycle & Recovery ($G\text{-DL}$)
- **Go Testing SDK (`pkg/chaostest`) Guide:** Tabbed copy-paste examples for Go backend engineers.
- **Hermitage Isolation Matrix:** Dynamic comparative matrix across SQLite, PostgreSQL, and MySQL.
- **CLI & GitHub Action Hub:** Interactive command palette and GitHub Actions workflow generator.

## 3. Deployment & Cloudflare Hosting
- Directory: `site/` (or `docs_site/`) in repository with:
  - `index.html`: Modern, responsive, semantic HTML5 SPA with Bregalda visual hierarchy.
  - `assets/`: Embedded brand SVGs, favicon, CSS stylesheets, and JS interactive controllers.
  - `wrangler.toml`: Cloudflare Pages configuration for domain `chaossql.bregalda.com`.
