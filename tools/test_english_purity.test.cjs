const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const { test } = require('node:test');

test('private local planning does not hide a shipped-document language violation', (t) => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'chaossql-language-'));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  fs.mkdirSync(path.join(root, 'tools'));
  fs.mkdirSync(path.join(root, 'internal_docs'));
  fs.mkdirSync(path.join(root, 'docs'));
  const checker = path.join(root, 'tools', 'test_english_purity.js');
  fs.copyFileSync(path.join(__dirname, 'test_english_purity.js'), checker);
  const nonEnglish = String.fromCharCode(97, 231, 227, 111);
  fs.writeFileSync(path.join(root, 'internal_docs', 'local-plan.md'), nonEnglish);
  const local = spawnSync(process.execPath, [checker], { encoding: 'utf8' });
  assert.equal(local.status, 0, local.stderr);
  fs.writeFileSync(path.join(root, 'docs', 'public-guide.md'), nonEnglish);
  const shipped = spawnSync(process.execPath, [checker], { encoding: 'utf8' });
  assert.equal(shipped.status, 1, 'the same text in public documentation must still fail');
  assert.match(shipped.stderr, /public-guide\.md/);
});
