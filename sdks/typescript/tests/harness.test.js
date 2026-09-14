const test = require('node:test');
const assert = require('node:assert');
const path = require('node:path');
const fs = require('node:fs');
const os = require('node:os');
const { ChaosHarness, executeIPC } = require('../dist/index');

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
    assert.strictEqual(result.status, 'passed');
    assert.strictEqual(result.violationDetected, false);
    assert.strictEqual(result.anomalyDetected, false);
    assert.strictEqual(result.seed, 42);
    assert.strictEqual(result.schedule.version, 1);
    assert.strictEqual(result.schedule.decisions.length, 5);
  });

  await t.test('rejects seeds that cannot be represented exactly by JavaScript', async () => {
    const harness = new ChaosHarness({ driver: 'sqlite', binPath });
    await assert.rejects(
      () => harness.run({ seed: Number.MAX_SAFE_INTEGER + 1 }),
      (err) => err instanceof RangeError && err.message.includes('safe integer')
    );
  });

  await t.test('executeIPC rejects unsafe seeds before spawning the engine', async () => {
    await assert.rejects(
      () => executeIPC({ seed_value: Number.MAX_SAFE_INTEGER + 1 }, '/missing/chaossql'),
      (err) => err instanceof RangeError && err.message.includes('seed_value')
    );
  });

  await t.test('detects and minimizes an invariant violation', async () => {
    const harness = new ChaosHarness({ driver: 'sqlite', binPath });

    const result = await harness
      .withSchema('CREATE TABLE accounts (id INT PRIMARY KEY, balance INT); INSERT INTO accounts VALUES (1, 100);')
      .withInvariant(
        'balance_unchanged',
        'SELECT balance FROM accounts WHERE id = 1;',
        'balance == 100'
      )
      .addOperation('change_balance', ['UPDATE accounts SET balance = 99 WHERE id = 1'])
      .runAndShrink({ workers: 1, iterations: 1, seed: 1337 });

    assert.strictEqual(result.anomalyDetected, true);
    assert.strictEqual(result.status, 'violation');
    assert.strictEqual(result.minimalOperations.length, 1);
    assert(result.mermaid && result.mermaid.includes('sequenceDiagram'));
  });

  await t.test('assertNoAnomalies throws when anomaly occurs', async () => {
    const harness = new ChaosHarness({ driver: 'sqlite', binPath });

    await assert.rejects(
      async () => {
        await harness
          .withSchema('CREATE TABLE acc (id INT PRIMARY KEY, val INT); INSERT INTO acc VALUES (1, 100);')
          .withInvariant('val_consistent', 'SELECT val FROM acc WHERE id = 1;', 'val == 100')
          .addOperation('change_val', ['UPDATE acc SET val = 99 WHERE id = 1'])
          .assertNoAnomalies({ workers: 1, iterations: 1, seed: 42 });
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
      .withInvariant('val_consistent', 'SELECT val FROM acc WHERE id = 1;', 'val == 100')
      .addOperation('change_val', ['UPDATE acc SET val = 99 WHERE id = 1'])
      .run({ workers: 1, iterations: 1, seed: 42 });

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
