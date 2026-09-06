const test = require('node:test');
const assert = require('node:assert');
const path = require('node:path');
const fs = require('node:fs');
const os = require('node:os');
const { ChaosHarness } = require('../dist/index');

test('ChaosSQL TypeScript SDK (@chaossql/test)', async (t) => {
  const binPath = path.resolve(__dirname, '../../../bin/chaossql');

  await t.test('passing invariant scenario succeeds', async () => {
    const harness = new ChaosHarness({ driver: 'sqlite', binPath });
    const result = await harness
      .withSchema('CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);')
      .withSeed('INSERT INTO accounts VALUES (1, 500);')
      .withInvariant('positive_balance', 'SELECT balance FROM accounts WHERE id = 1;', 'balance == 500')
      .addOperation('read_op', ['SELECT balance FROM accounts WHERE id = 1'])
      .assertNoAnomalies({ workers: 2, iterations: 5, seed: 42 });

    assert.strictEqual(result.success, true);
    assert.strictEqual(result.violationDetected, false);
    assert.strictEqual(result.anomalyDetected, false);
  });

  await t.test('detects write skew (A5B) in hospital on-call roster', async () => {
    const harness = new ChaosHarness({ driver: 'sqlite', binPath });

    const result = await harness
      .withSchema(`
        CREATE TABLE doctors (id INT PRIMARY KEY, name TEXT, on_call INT);
        INSERT INTO doctors VALUES (1, 'Alice', 1), (2, 'Bob', 1);
      `)
      .withInvariant(
        'at_least_one_doctor_on_call',
        'SELECT count(*) as active FROM doctors WHERE on_call = 1;',
        'active >= 1'
      )
      .addOperation('alice_leaves', [
        'SELECT count(*) as cnt FROM doctors WHERE on_call = 1 -> active',
        'UPDATE doctors SET on_call = 0 WHERE id = 1 AND {active > 1}',
      ])
      .addOperation('bob_leaves', [
        'SELECT count(*) as cnt FROM doctors WHERE on_call = 1 -> active',
        'UPDATE doctors SET on_call = 0 WHERE id = 2 AND {active > 1}',
      ])
      .runAndShrink({ workers: 2, iterations: 20, seed: 1337 });

    assert.strictEqual(result.anomalyDetected, true);
    assert.strictEqual(result.anomalyCode, 'A5B');
    assert.strictEqual(result.minimalOperations.length, 2);
    assert(result.mermaid && result.mermaid.includes('sequenceDiagram'));
  });

  await t.test('assertNoAnomalies throws when anomaly occurs', async () => {
    const harness = new ChaosHarness({ driver: 'sqlite', binPath });

    await assert.rejects(
      async () => {
        await harness
          .withSchema('CREATE TABLE acc (id INT PRIMARY KEY, val INT); INSERT INTO acc VALUES (1, 100);')
          .withInvariant('val_consistent', 'SELECT val FROM acc WHERE id = 1;', 'val == 80')
          .addOperation('sub_10', [
            'SELECT val FROM acc WHERE id = 1 -> cur',
            'UPDATE acc SET val = {cur - 10} WHERE id = 1',
          ])
          .assertNoAnomalies({ workers: 2, iterations: 2, seed: 42 });
      },
      (err) => {
        assert(err instanceof Error);
        assert(err.message.includes('ChaosSQL Isolation Anomaly Detected'));
        return true;
      }
    );
  });

  await t.test('exportStandaloneRepro creates valid standalone reproduction file', async () => {
    const harness = new ChaosHarness({ driver: 'sqlite', binPath });

    const res = await harness
      .withSchema('CREATE TABLE acc (id INT PRIMARY KEY, val INT); INSERT INTO acc VALUES (1, 100);')
      .withInvariant('val_consistent', 'SELECT val FROM acc WHERE id = 1;', 'val == 80')
      .addOperation('sub_10', [
        'SELECT val FROM acc WHERE id = 1 -> cur',
        'UPDATE acc SET val = {cur - 10} WHERE id = 1',
      ])
      .run({ workers: 2, iterations: 2, seed: 42 });

    assert.strictEqual(res.anomalyDetected, true);

    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'chaossql-ts-repro-'));
    const reproFile = path.join(tmpDir, 'generated_repro_test.js');
    const outPath = await res.exportStandaloneRepro(reproFile);

    assert.strictEqual(fs.existsSync(outPath), true);
    const content = fs.readFileSync(outPath, 'utf8');
    assert(content.includes('ChaosSQL (github.com/bregaldahq/chaossql)'));
    assert(content.includes('MINIMAL_OPS'));

    fs.rmSync(tmpDir, { recursive: true, force: true });
  });
});
