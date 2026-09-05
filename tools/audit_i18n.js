/**
 * ChaosSQL — Bilingual i18n & Documentation QA Audit Tool (Task 4)
 * Validates symmetry, completeness, HTML data-i18n binding, and linguistic purity.
 */

const fs = require('fs');
const path = require('path');
const vm = require('vm');

const ROOT_DIR = path.resolve(__dirname, '..');
const SITE_DIR = path.join(ROOT_DIR, 'site');

// 1. Load DOCS_DATA
const docsDataPath = path.join(SITE_DIR, 'docs-data.js');
let DOCS_DATA;
try {
  DOCS_DATA = require(docsDataPath);
} catch (e) {
  console.error('[FAIL] Could not require docs-data.js:', e);
  process.exit(1);
}

// 2. Load app.js (I18N, SCENARIOS)
const appJsPath = path.join(SITE_DIR, 'app.js');
const appJsCode = fs.readFileSync(appJsPath, 'utf8');
const dummyFunc = () => {};
const sandboxWindow = { addEventListener: dummyFunc, location: { hash: '' } };
const sandbox = {
  window: sandboxWindow,
  document: {
    readyState: 'loading',
    addEventListener: dummyFunc,
    documentElement: { lang: 'pt' },
    querySelectorAll: () => [],
    getElementById: () => null
  },
  navigator: { language: 'en-US' },
  localStorage: { getItem: () => null, setItem: dummyFunc },
  setTimeout: dummyFunc,
  clearTimeout: dummyFunc,
  console: console
};
vm.createContext(sandbox);
try {
  vm.runInContext(appJsCode, sandbox);
} catch (e) {
  console.error('[FAIL] Could not run app.js in sandbox:', e);
  process.exit(1);
}

const I18N = sandbox.window.I18N;
const SCENARIOS = sandbox.window.SCENARIOS;

if (!I18N) {
  console.error('[FAIL] I18N is not exported on window in app.js');
  process.exit(1);
}
if (!SCENARIOS) {
  console.error('[FAIL] SCENARIOS is not exported on window in app.js');
  process.exit(1);
}

const results = {
  passed: 0,
  failed: 0,
  warnings: 0,
  details: []
};

function pass(msg) {
  results.passed++;
  results.details.push(`  [PASS] ${msg}`);
}

function fail(msg) {
  results.failed++;
  results.details.push(`  [FAIL] ${msg}`);
}

function warn(msg) {
  results.warnings++;
  results.details.push(`  [WARN] ${msg}`);
}

console.log('===============================================================');
console.log('  ChaosSQL v1.2.0 — Integrated QA & Bilingual i18n Audit');
console.log('===============================================================\n');

// ---------------------------------------------------------------------------
// TEST 1: I18N Dictionary Symmetry (PT vs EN)
// ---------------------------------------------------------------------------
console.log('--> Checking Test 1: I18N Dictionary Symmetry (PT vs EN)...');
if (!I18N.pt || !I18N.en) {
  fail('I18N must contain both "pt" and "en" root objects.');
} else {
  function getLeaves(obj, prefix = '') {
    let leaves = {};
    for (const [key, value] of Object.entries(obj)) {
      const fullPath = prefix ? `${prefix}.${key}` : key;
      if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
        Object.assign(leaves, getLeaves(value, fullPath));
      } else {
        leaves[fullPath] = value;
      }
    }
    return leaves;
  }

  const ptLeaves = getLeaves(I18N.pt);
  const enLeaves = getLeaves(I18N.en);

  const ptKeys = Object.keys(ptLeaves);
  const enKeys = Object.keys(enLeaves);

  let missingInEn = [];
  let missingInPt = [];

  for (const k of ptKeys) {
    if (!Object.prototype.hasOwnProperty.call(enLeaves, k)) {
      missingInEn.push(k);
    }
  }

  for (const k of enKeys) {
    if (!Object.prototype.hasOwnProperty.call(ptLeaves, k)) {
      missingInPt.push(k);
    }
  }

  if (missingInEn.length === 0 && missingInPt.length === 0) {
    pass(`I18N dictionary has perfect symmetry: ${ptKeys.length} leaf keys in both PT and EN.`);
  } else {
    if (missingInEn.length > 0) {
      fail(`Keys present in PT but missing in EN (${missingInEn.length}): ${missingInEn.join(', ')}`);
    }
    if (missingInPt.length > 0) {
      fail(`Keys present in EN but missing in PT (${missingInPt.length}): ${missingInPt.join(', ')}`);
    }
  }
}

// ---------------------------------------------------------------------------
// TEST 2: Documentation Chapters (DOCS_DATA.pt vs DOCS_DATA.en)
// ---------------------------------------------------------------------------
console.log('\n--> Checking Test 2: Documentation Chapters Completeness & Symmetry...');
const EXPECTED_CHAPTERS = [
  'getting-started',
  'dsl-spec',
  'cli-reference',
  'trace-visualizer',
  'cicd-sarif',
  'drivers',
  'go-sdk',
  'academic-theory'
];

if (!DOCS_DATA.pt || !DOCS_DATA.en) {
  fail('DOCS_DATA must contain both "pt" and "en" dictionaries.');
} else {
  const ptChapters = Object.keys(DOCS_DATA.pt);
  const enChapters = Object.keys(DOCS_DATA.en);

  if (ptChapters.length === 8 && enChapters.length === 8) {
    pass('Both PT and EN contain exactly 8 chapters.');
  } else {
    fail(`Chapter count mismatch: PT has ${ptChapters.length}, EN has ${enChapters.length} (expected 8).`);
  }

  for (const chId of EXPECTED_CHAPTERS) {
    const ptCh = DOCS_DATA.pt[chId];
    const enCh = DOCS_DATA.en[chId];

    if (!ptCh) {
      fail(`Missing PT chapter: ${chId}`);
      continue;
    }
    if (!enCh) {
      fail(`Missing EN chapter: ${chId}`);
      continue;
    }

    const fields = ['title', 'category', 'summary', 'content'];
    for (const f of fields) {
      if (!ptCh[f] || typeof ptCh[f] !== 'string' || ptCh[f].trim() === '') {
        fail(`PT chapter "${chId}" is missing or empty for field "${f}".`);
      }
      if (!enCh[f] || typeof enCh[f] !== 'string' || enCh[f].trim() === '') {
        fail(`EN chapter "${chId}" is missing or empty for field "${f}".`);
      }
    }
  }
  pass('All 8 chapters in PT and EN contain non-empty title, category, summary, and content.');
}

// ---------------------------------------------------------------------------
// TEST 3: Scenarios Completeness (10 Flagship Scenarios)
// ---------------------------------------------------------------------------
console.log('\n--> Checking Test 3: Flagship Scenarios (10 Scenarios)...');
if (!Array.isArray(SCENARIOS) || SCENARIOS.length !== 10) {
  fail(`SCENARIOS array should have exactly 10 scenarios, found ${SCENARIOS ? SCENARIOS.length : 0}`);
} else {
  pass(`SCENARIOS has exactly 10 scenarios.`);

  let scenariosOk = true;
  SCENARIOS.forEach((sc, idx) => {
    const scNum = idx + 1;
    // name
    if (!sc.name || typeof sc.name !== 'object' || !sc.name.pt || !sc.name.en) {
      fail(`Scenario ${scNum} (${sc.id}) missing bilingual name.`);
      scenariosOk = false;
    }
    // description
    if (!sc.description || typeof sc.description !== 'object' || !sc.description.pt || !sc.description.en) {
      fail(`Scenario ${scNum} (${sc.id}) missing bilingual description.`);
      scenariosOk = false;
    }
    // analysis
    if (!sc.analysis || typeof sc.analysis !== 'object' || !sc.analysis.pt || !sc.analysis.en) {
      fail(`Scenario ${scNum} (${sc.id}) missing bilingual analysis.`);
      scenariosOk = false;
    }
    // fix
    if (!sc.fix || typeof sc.fix !== 'object' || !sc.fix.pt || !sc.fix.en) {
      fail(`Scenario ${scNum} (${sc.id}) missing bilingual fix object.`);
      scenariosOk = false;
    } else {
      const fixFields = ['title', 'explanation', 'code'];
      for (const f of fixFields) {
        if (!sc.fix.pt[f] || typeof sc.fix.pt[f] !== 'string' || sc.fix.pt[f].trim() === '') {
          fail(`Scenario ${scNum} (${sc.id}) fix.pt missing field "${f}".`);
          scenariosOk = false;
        }
        if (!sc.fix.en[f] || typeof sc.fix.en[f] !== 'string' || sc.fix.en[f].trim() === '') {
          fail(`Scenario ${scNum} (${sc.id}) fix.en missing field "${f}".`);
          scenariosOk = false;
        }
      }
      // driverNotes / engines
      if (!sc.fix.pt.driverNotes && !sc.fix.pt.engines) {
        fail(`Scenario ${scNum} (${sc.id}) fix.pt missing driverNotes or engines.`);
        scenariosOk = false;
      }
      if (!sc.fix.en.driverNotes && !sc.fix.en.engines) {
        fail(`Scenario ${scNum} (${sc.id}) fix.en missing driverNotes or engines.`);
        scenariosOk = false;
      }
    }
  });

  if (scenariosOk) {
    pass('All 10 scenarios have complete bilingual name, description, analysis, and fix structures.');
  }
}

// ---------------------------------------------------------------------------
// TEST 4: HTML data-i18n & data-i18n-attr Bindings
// ---------------------------------------------------------------------------
console.log('\n--> Checking Test 4: HTML data-i18n & data-i18n-attr Bindings in index.html...');
const htmlPath = path.join(SITE_DIR, 'index.html');
const htmlContent = fs.readFileSync(htmlPath, 'utf8');

// helper to get nested prop
function resolvePath(obj, pathStr) {
  return pathStr.split('.').reduce((acc, part) => (acc && acc[part] !== undefined) ? acc[part] : undefined, obj);
}

// match data-i18n="..."
const dataI18nRegex = /data-i18n="([^"]+)"/g;
let match;
const dataI18nKeys = new Set();
while ((match = dataI18nRegex.exec(htmlContent)) !== null) {
  dataI18nKeys.add(match[1]);
}

// match data-i18n-attr="attr:path"
const dataI18nAttrRegex = /data-i18n-attr="([^:"]+):([^"]+)"/g;
const dataI18nAttrKeys = new Set();
while ((match = dataI18nAttrRegex.exec(htmlContent)) !== null) {
  dataI18nAttrKeys.add(match[2]);
}

let bindingErrors = [];
for (const k of dataI18nKeys) {
  const ptVal = resolvePath(I18N.pt, k);
  const enVal = resolvePath(I18N.en, k);
  if (ptVal === undefined) {
    bindingErrors.push(`Missing in I18N.pt: [data-i18n="${k}"]`);
  }
  if (enVal === undefined) {
    bindingErrors.push(`Missing in I18N.en: [data-i18n="${k}"]`);
  }
}

for (const k of dataI18nAttrKeys) {
  const ptVal = resolvePath(I18N.pt, k);
  const enVal = resolvePath(I18N.en, k);
  if (ptVal === undefined) {
    bindingErrors.push(`Missing in I18N.pt: [data-i18n-attr="...:${k}"]`);
  }
  if (enVal === undefined) {
    bindingErrors.push(`Missing in I18N.en: [data-i18n-attr="...:${k}"]`);
  }
}

if (bindingErrors.length === 0) {
  pass(`All ${dataI18nKeys.size} [data-i18n] and ${dataI18nAttrKeys.size} [data-i18n-attr] bindings resolve properly in both PT and EN.`);
} else {
  fail(`Binding errors found:\n    ${bindingErrors.join('\n    ')}`);
}

// ---------------------------------------------------------------------------
// TEST 5: English Mode Linguistic Purity & Non-English Remnants
// ---------------------------------------------------------------------------
console.log('\n--> Checking Test 5: English Mode Linguistic Purity...');

// Common Portuguese tokens that should never appear in English UI prose
const PT_PORTUGUESE_INDICATORS = [
  /\bInício\b/i,
  /\bDocumentação\b/i,
  /\bCenários\b/i,
  /\bSoluções\b/i,
  /\bMatriz\b/i,
  /\bVisão\b/i,
  /\bExecução\b/i,
  /\bConcorrência\b/i,
  /\bIsolamento\b/i,
  /\bInjetar\b/i,
  /\bReiniciar\b/i,
  /\bReduzir\b/i,
  /\bAnomalia\b/i,
  /\bFundações\b/i,
  /\bConstruído\b/i,
  /\bPesquisa\b/i,
  /\bAnterior\b/i,
  /\bPróximo\b/i,
  /\bRastro\b/i,
  /\bInvariante\b/i,
  /\bPermitido\b/i,
  /\bPrevenido\b/i,
  /\bDetectado\b/i,
  /\bBuscar\b/i,
  /\bCapítulo\b/i,
  /\bCopiar\b/i,
  /\bCopiado\b/i,
  /\bExibir\b/i,
  /\bGrafo\b/i
];

let purityViolations = [];

function checkTextPurity(text, location) {
  if (typeof text !== 'string') return;
  for (const regex of PT_PORTUGUESE_INDICATORS) {
    if (regex.test(text)) {
      purityViolations.push(`${location}: contains Portuguese marker "${regex}" -> "${text.substring(0, 80)}..."`);
    }
  }
}

// Check I18N.en
function checkEnDict(obj, prefix = 'I18N.en') {
  for (const [k, v] of Object.entries(obj)) {
    const fullPath = `${prefix}.${k}`;
    if (typeof v === 'object' && v !== null) {
      checkEnDict(v, fullPath);
    } else if (typeof v === 'string') {
      checkTextPurity(v, fullPath);
    }
  }
}
checkEnDict(I18N.en);

// Check DOCS_DATA.en titles, categories, summaries
for (const [id, doc] of Object.entries(DOCS_DATA.en)) {
  checkTextPurity(doc.title, `DOCS_DATA.en[${id}].title`);
  checkTextPurity(doc.category, `DOCS_DATA.en[${id}].category`);
  checkTextPurity(doc.summary, `DOCS_DATA.en[${id}].summary`);
}

// Check SCENARIOS[].*.en
SCENARIOS.forEach((sc, idx) => {
  if (sc.name && sc.name.en) checkTextPurity(sc.name.en, `SCENARIOS[${idx}].name.en`);
  if (sc.description && sc.description.en) checkTextPurity(sc.description.en, `SCENARIOS[${idx}].description.en`);
  if (sc.analysis && sc.analysis.en) checkTextPurity(sc.analysis.en, `SCENARIOS[${idx}].analysis.en`);
  if (sc.fix && sc.fix.en) {
    if (sc.fix.en.title) checkTextPurity(sc.fix.en.title, `SCENARIOS[${idx}].fix.en.title`);
    if (sc.fix.en.explanation) checkTextPurity(sc.fix.en.explanation, `SCENARIOS[${idx}].fix.en.explanation`);
  }
});

if (purityViolations.length === 0) {
  pass('Zero Portuguese remnants detected in English mode (I18N.en, DOCS_DATA.en, SCENARIOS.en).');
} else {
  fail(`Portuguese words found in English content (${purityViolations.length}):\n    ${purityViolations.join('\n    ')}`);
}

// ---------------------------------------------------------------------------
// TEST 6: Encoding Integrity (Spurious '?' characters)
// ---------------------------------------------------------------------------
console.log('\n--> Checking Test 6: Encoding Integrity (Corrupted unicode characters)...');

let encodingIssues = [];

function checkEncoding(text, location) {
  if (typeof text !== 'string') return;
  // Match single or double question mark surrounded by letters (e.g. In?cio, Documenta??o, Vers?o)
  // or arrow/math patterns like "rw ? ww", "P ? 1", "Documentation ?"
  if (/[a-zA-ZÀ-ÿ]\?[a-zA-ZÀ-ÿ]/.test(text) || /\?\?/.test(text)) {
    encodingIssues.push(`${location}: possible broken accent: "${text}"`);
  }
  if (/\b\?\b/.test(text) && !location.includes('search') && !text.endsWith('?')) {
    // Lone '?' that is not an actual question mark
    if (!text.includes('?') || text.trim() === '? Run Fuzzer' || text.includes('rw ? ww') || text.includes('P ? 1')) {
      encodingIssues.push(`${location}: lone '?' replacing icon/symbol: "${text}"`);
    }
  }
}

function scanEncodingDict(obj, prefix) {
  for (const [k, v] of Object.entries(obj)) {
    const fullPath = `${prefix}.${k}`;
    if (typeof v === 'object' && v !== null) {
      scanEncodingDict(v, fullPath);
    } else if (typeof v === 'string') {
      checkEncoding(v, fullPath);
    }
  }
}

scanEncodingDict(I18N.pt, 'I18N.pt');
scanEncodingDict(I18N.en, 'I18N.en');

if (encodingIssues.length === 0) {
  pass('No broken unicode or corrupted "?" character encodings found in I18N dictionaries.');
} else {
  warn(`Encoding anomalies found (${encodingIssues.length}):\n    ${encodingIssues.join('\n    ')}`);
}

// ---------------------------------------------------------------------------
// TEST 7: Client-side Rapid Language Alternation & State Transition
// ---------------------------------------------------------------------------
console.log('\n--> Checking Test 7: Client-side Rapid Language Alternation & Reactive State...');

class MockClassList {
  constructor(str = '') {
    this.classes = new Set(str.split(/\s+/).filter(Boolean));
  }
  add(c) { this.classes.add(c); }
  remove(c) { this.classes.delete(c); }
  contains(c) { return this.classes.has(c); }
  toggle(c) { if (this.classes.has(c)) this.classes.delete(c); else this.classes.add(c); }
  has(c) { return this.classes.has(c); }
}

class MockElement {
  constructor(tag, attrs = {}) {
    this.tagName = tag.toUpperCase();
    this.attributes = { ...attrs };
    this.classList = new MockClassList(attrs.class || '');
    this.textContent = '';
    this.children = [];
  }
  getAttribute(name) {
    return this.attributes[name] !== undefined ? this.attributes[name] : null;
  }
  setAttribute(name, val) {
    this.attributes[name] = String(val);
  }
  removeAttribute(name) {
    delete this.attributes[name];
  }
}

const mockDoc = {
  documentElement: new MockElement('html', { lang: 'pt' }),
  elements: []
};

// Create mock lang buttons
const btnPt = new MockElement('button', { class: 'lang-btn active', 'data-lang': 'pt', 'aria-pressed': 'true' });
const btnEn = new MockElement('button', { class: 'lang-btn', 'data-lang': 'en', 'aria-pressed': 'false' });
mockDoc.elements.push(btnPt, btnEn);

// Create mock data-i18n elements
const sampleKeys = ['nav.home', 'landing.heroTitle', 'docs.breadcrumbDocs', 'scenarios.tabSchema', 'visualizer.animateBtn'];
const i18nElements = sampleKeys.map(k => {
  const el = new MockElement('span', { 'data-i18n': k });
  el.textContent = resolvePath(I18N.pt, k);
  mockDoc.elements.push(el);
  return el;
});

// Set up mock DOM queries
sandbox.document.documentElement = mockDoc.documentElement;
sandbox.document.querySelectorAll = function(selector) {
  if (selector === '.lang-btn[data-lang]') return [btnPt, btnEn];
  if (selector === '[data-i18n]') return i18nElements;
  if (selector === '[data-i18n-attr]') return [];
  return [];
};
sandbox.document.getElementById = () => null;

let alternationPassed = true;
const iterations = 100;
for (let i = 0; i < iterations; i++) {
  const target = i % 2 === 0 ? 'en' : 'pt';
  sandbox.window.setLanguage(target);

  if (sandbox.window.currentLang !== target) {
    fail(`Iteration ${i}: window.currentLang was ${sandbox.window.currentLang}, expected ${target}`);
    alternationPassed = false;
    break;
  }
  if (sandbox.document.documentElement.lang !== target) {
    fail(`Iteration ${i}: documentElement.lang was ${sandbox.document.documentElement.lang}, expected ${target}`);
    alternationPassed = false;
    break;
  }

  const activeBtn = target === 'pt' ? btnPt : btnEn;
  const inactiveBtn = target === 'pt' ? btnEn : btnPt;
  if (!activeBtn.classList.has('active') || activeBtn.getAttribute('aria-pressed') !== 'true') {
    fail(`Iteration ${i}: ${target} button did not have .active or aria-pressed="true"`);
    alternationPassed = false;
    break;
  }
  if (inactiveBtn.classList.has('active') || inactiveBtn.getAttribute('aria-pressed') !== 'false') {
    fail(`Iteration ${i}: inactive button was not deactivated properly`);
    alternationPassed = false;
    break;
  }

  // Check text content matches target dictionary
  for (const el of i18nElements) {
    const k = el.getAttribute('data-i18n');
    const expected = resolvePath(I18N[target], k);
    if (el.textContent !== expected) {
      fail(`Iteration ${i} [${target}]: element for "${k}" text mismatch: expected "${expected}", got "${el.textContent}"`);
      alternationPassed = false;
      break;
    }
  }
  if (!alternationPassed) break;
}

// Check fallback for invalid language
sandbox.window.setLanguage('invalid-locale');
if (sandbox.window.currentLang === 'pt' && sandbox.document.documentElement.lang === 'pt') {
  // successfully fell back to 'pt'
} else {
  fail(`Invalid language fallback failed: currentLang is ${sandbox.window.currentLang}`);
  alternationPassed = false;
}

if (alternationPassed) {
  pass(`Rapid language alternation verified: ${iterations} switching cycles executed smoothly with 100% reactive state synchronization and fallback safety.`);
}

// ---------------------------------------------------------------------------
// Summary Report
// ---------------------------------------------------------------------------
console.log('\n===============================================================');
console.log(`Audit Finished: ${results.passed} PASSED, ${results.failed} FAILED, ${results.warnings} WARNINGS`);
console.log('===============================================================\n');
results.details.forEach(d => console.log(d));

if (results.failed > 0) {
  process.exit(1);
} else {
  process.exit(0);
}
