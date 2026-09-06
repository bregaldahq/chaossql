#!/usr/bin/env node
// tools/test_english_purity.js
// Automated English Purity Verification Gate & Regression Audit for ChaosSQL
// Guards against unintentional Portuguese commits outside the bilingual web portal.

const fs = require('fs');
const path = require('path');

const ROOT_DIR = path.resolve(__dirname, '..');

// Target directories to audit
const TARGET_DIRS = [
  'docs',
  'specs',
  'evals',
  'examples',
  'cmd',
  'internal',
  'pkg',
  'tools',
  'internal_docs'
];

// Target root-level files to audit
const ROOT_FILES = [
  'Makefile',
  'ARCHITECTURE.md',
  'AGENTS.md',
  'CONTRIBUTING.md',
  'README.md',
  'SECURITY.md',
  'action.yml'
];

// Normalized exclusion rules
const EXCLUDED_PATTERNS = [
  /\.git([\\/]|$)/,
  /\.superpowers([\\/]|$)/,
  /(?:^|[\\/])docs[\\/]superpowers([\\/]|$)/, // Historical planning and design ledgers
  /node_modules([\\/]|$)/,
  /^site([\\/]|$)/,
  /[\\/]site([\\/]|$)/,
  /tools[\\/]test_playground_ui\.js$/,
  /tools[\\/]audit_i18n\.js$/,
  /tools[\\/]test_english_purity\.js$/
];

// Recognized non-Portuguese citations and legitimate accented names
const ALLOWED_CITATIONS = [
  'Röhm',
  'Rohm',
  'Bézier',
  'Bezier',
  'Poincaré',
  'Poincare',
  'Erdős',
  'Erdos',
  'René',
  'Rene'
];

// Portuguese accented characters
const ACCENT_REGEX = /[áàâãéêíóôõúçÁÀÂÃÉÊÍÓÔÕÚÇ]/u;

// Common Portuguese words (both accented and unaccented variations)
const PORTUGUESE_KEYWORDS = [
  // Scenario / Invariant / Anomaly / Report / Execution
  'cenário', 'cenários', 'cenario', 'cenarios',
  'execução', 'execuções', 'execucao', 'execucoes', 'executado', 'executada', 'executados', 'executadas', 'executando',
  'transação', 'transações', 'transacao', 'transacoes',
  'invariante', 'invariantes',
  'anomalia', 'anomalias',
  'detectado', 'detectada', 'detectados', 'detectadas',
  'falha', 'falhas',
  'relatório', 'relatórios', 'relatorio', 'relatorios',
  'visão', 'visões', 'visao', 'visoes',
  'tabela', 'tabelas',
  'banco', 'bancos',
  // Domain / Database / Concurrency vocabulary
  'concorrência', 'concorrencia',
  'isolamento', 'isolamentos',
  'leitura', 'leituras',
  'escrita', 'escritas',
  'registro', 'registros',
  'validação', 'validacao',
  'verificação', 'verificacao',
  'demonstrando', 'demonstracao', 'demonstração',
  'compilando', 'iniciando', 'baixando', 'baixa'
];

// Unicode-aware word boundary pattern for each Portuguese keyword
const KEYWORD_REGEX = new RegExp(
  '(?:^|[^\\p{L}\\p{N}_])(' + PORTUGUESE_KEYWORDS.join('|') + ')(?:[^\\p{L}\\p{N}_]|$)',
  'iu'
);

function isExcluded(relativePath) {
  const normalized = relativePath.replace(/\\/g, '/');
  return EXCLUDED_PATTERNS.some((pattern) => pattern.test(normalized));
}

function collectFiles(dirPath, relativeBase) {
  let results = [];
  if (!fs.existsSync(dirPath)) return results;

  const entries = fs.readdirSync(dirPath, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = path.join(dirPath, entry.name);
    const relPath = path.join(relativeBase, entry.name);

    if (isExcluded(relPath)) continue;

    if (entry.isDirectory()) {
      results = results.concat(collectFiles(fullPath, relPath));
    } else if (entry.isFile()) {
      const ext = path.extname(entry.name).toLowerCase();
      // Skip binaries and images
      if (['.wasm', '.png', '.jpg', '.jpeg', '.gif', '.ico', '.exe', '.db', '.sqlite', '.bin', '.diff'].includes(ext)) {
        continue;
      }
      results.push({ fullPath, relPath });
    }
  }
  return results;
}

function auditFile(fileObj) {
  const content = fs.readFileSync(fileObj.fullPath, 'utf8');
  const lines = content.split(/\r?\n/);
  const violations = [];

  lines.forEach((line, idx) => {
    const lineNum = idx + 1;

    // Check for common Portuguese keywords
    const keywordMatch = line.match(KEYWORD_REGEX);
    if (keywordMatch) {
      violations.push({
        line: lineNum,
        type: 'KEYWORD',
        token: keywordMatch[1],
        snippet: line.trim()
      });
      return;
    }

    // Scrub allowed citations case-insensitively before checking for accents
    let scrubbedLine = line;
    for (const citation of ALLOWED_CITATIONS) {
      const citeRegex = new RegExp(citation, 'gi');
      scrubbedLine = scrubbedLine.replace(citeRegex, '');
    }

    // Check for Portuguese accented characters
    const accentMatch = scrubbedLine.match(ACCENT_REGEX);
    if (accentMatch) {
      violations.push({
        line: lineNum,
        type: 'ACCENT',
        token: accentMatch[0],
        snippet: line.trim()
      });
    }
  });

  return { file: fileObj.relPath, violations, lineCount: lines.length };
}

function main() {
  console.log('===============================================================');
  console.log('  ChaosSQL — Automated English Purity Verification Gate');
  console.log('===============================================================');

  let allFiles = [];

  // 1. Audit root-level files
  for (const rootFile of ROOT_FILES) {
    const fullPath = path.join(ROOT_DIR, rootFile);
    if (fs.existsSync(fullPath) && !isExcluded(rootFile)) {
      allFiles.push({ fullPath, relPath: rootFile });
    }
  }

  // 2. Audit directories
  for (const dir of TARGET_DIRS) {
    const dirPath = path.join(ROOT_DIR, dir);
    allFiles = allFiles.concat(collectFiles(dirPath, dir));
  }

  console.log(`Auditing ${allFiles.length} files across target directories & repository roots:`);
  TARGET_DIRS.forEach((d) => console.log(`  - ${d}/`));
  ROOT_FILES.forEach((f) => console.log(`  - ${f}`));
  console.log('\nExclusions applied:');
  console.log('  - .git/');
  console.log('  - .superpowers/ & docs/superpowers/ (historical plans)');
  console.log('  - node_modules/');
  console.log('  - site/ (legitimate Portuguese i18n dictionaries)');
  console.log('  - tools/test_playground_ui.js (bilingual UI toggle test)');
  console.log('  - tools/audit_i18n.js (bilingual documentation audit test)');
  console.log('  - tools/test_english_purity.js (self)\n');

  let totalViolations = 0;
  let totalLines = 0;
  const violationReports = [];

  for (const fileObj of allFiles) {
    const result = auditFile(fileObj);
    totalLines += result.lineCount;
    if (result.violations.length > 0) {
      totalViolations += result.violations.length;
      violationReports.push(result);
    }
  }

  if (totalViolations > 0) {
    console.error(`❌ FAILED: Found ${totalViolations} English purity violation(s):\n`);
    for (const report of violationReports) {
      console.error(`  📄 ${report.file}:`);
      for (const v of report.violations) {
        console.error(`     Line ${v.line} [${v.type} '${v.token}']: ${v.snippet.slice(0, 100)}`);
      }
    }
    console.error('\nPlease translate or harmonize the above Portuguese remnants to English.\n');
    process.exit(1);
  }

  console.log(`✔ SUCCESS: ${allFiles.length} files (${totalLines} lines) audited with 0 violations.`);
  console.log('✔ English purity verified: No Portuguese tokens or unauthorized accents found outside site/ portal.\n');
  process.exit(0);
}

main();
