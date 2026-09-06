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
  'tools'
];

// Normalized exclusion rules
const EXCLUDED_PATTERNS = [
  /\.git([\\/]|$)/,
  /(\.?)superpowers([\\/]|$)/,
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
  'validação', 'validacao', 'validações', 'validacoes',
  'verificação', 'verificacao', 'verificações', 'verificacoes',
  'operação', 'operações', 'operacao', 'operacoes',
  'configuração', 'configuracao', 'configurações', 'configuracoes',
  'padrão', 'padrao', 'padrões', 'padroes',
  'usuário', 'usuario', 'usuários', 'usuarios',
  'descrição', 'descricao', 'descrições', 'descricoes',
  'solução', 'solucao', 'soluções', 'solucoes',
  'informação', 'informacao', 'informações', 'informacoes',
  'código', 'codigo', 'códigos', 'codigos'
];

const KEYWORD_REGEX = new RegExp(
  '(?:^|[^\\p{L}\\p{N}_])(' + PORTUGUESE_KEYWORDS.join('|') + ')(?:[^\\p{L}\\p{N}_]|$)',
  'iu'
);

function isExcluded(relPath) {
  const normalized = relPath.replace(/\\/g, '/');
  return EXCLUDED_PATTERNS.some((pattern) => pattern.test(normalized));
}

function collectFiles(dirPath, relBase = '') {
  let fileList = [];
  if (!fs.existsSync(dirPath)) return fileList;

  const entries = fs.readdirSync(dirPath, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = path.join(dirPath, entry.name);
    const relPath = path.join(relBase, entry.name);

    if (isExcluded(relPath)) {
      continue;
    }

    if (entry.isDirectory()) {
      fileList = fileList.concat(collectFiles(fullPath, relPath));
    } else if (entry.isFile()) {
      // Skip known binary extensions
      const ext = path.extname(entry.name).toLowerCase();
      if (['.wasm', '.png', '.jpg', '.jpeg', '.gif', '.ico', '.exe', '.db', '.sqlite', '.bin'].includes(ext)) {
        continue;
      }
      fileList.push({ fullPath, relPath });
    }
  }
  return fileList;
}

function auditFile(fileObj) {
  const violations = [];
  const content = fs.readFileSync(fileObj.fullPath, 'utf8');
  const lines = content.split('\n');

  lines.forEach((line, idx) => {
    const lineNum = idx + 1;

    // 1. Check for Portuguese keywords
    const kwMatch = line.match(KEYWORD_REGEX);
    if (kwMatch) {
      violations.push({
        line: lineNum,
        type: 'KEYWORD',
        token: kwMatch[1],
        snippet: line.trim()
      });
      return;
    }

    // 2. Check for accented Portuguese characters (allowing known citations)
    let scrubbedLine = line;
    for (const citation of ALLOWED_CITATIONS) {
      scrubbedLine = scrubbedLine.replaceAll(citation, '');
    }

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
  console.log('===============================================================\n');

  let allFiles = [];
  for (const dir of TARGET_DIRS) {
    const dirPath = path.join(ROOT_DIR, dir);
    allFiles = allFiles.concat(collectFiles(dirPath, dir));
  }

  console.log(`Auditing ${allFiles.length} files across target directories:`);
  TARGET_DIRS.forEach((d) => console.log(`  - ${d}/`));
  console.log('\nExclusions applied:');
  console.log('  - .git/');
  console.log('  - superpowers/ & .superpowers/');
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
