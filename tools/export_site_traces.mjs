#!/usr/bin/env node
// Regenerates the recorded runs the portal animates (site/src/data/traces/).
//
// Every file is produced by the real engine: each example runs on SQLite at
// READ_UNCOMMITTED (the only SQLite level where these anomalies are observable,
// see the chaossql-example-scenarios skill) with its own seed, and the result
// is projected to a small JSON document: the full logical schedule, the
// delta-debugged minimal trace, the failing invariant and the shrink numbers.
//
// Usage: node tools/export_site_traces.mjs [path/to/chaossql]
// Without an argument the CLI is built into a temporary directory first.

import { execFileSync } from 'node:child_process';
import { cpSync, mkdtempSync, readFileSync, rmSync, writeFileSync, mkdirSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const OUT_DIR = join(ROOT, 'site/src/data/traces');
const ISOLATION = 'READ_UNCOMMITTED';

// Examples with a stable, self-explanatory minimal trace on SQLite
// READ_UNCOMMITTED. Deliberately excluded from marketing material:
// - read_skew_financial_audit: its dominant label flips between P4 and A5A
//   across runs (the full trace depends on physical timing);
// - ticket_booking_anti_dependency: its invariant also fails on serial
//   histories (see evals/02_false_positive_rate.md).
const SCENARIOS = ['banking_lost_update', 'inventory_oversell', 'hospital_write_skew'];

const work = mkdtempSync(join(tmpdir(), 'chaossql-site-traces-'));
try {
  let cli = process.argv[2] ? resolve(process.argv[2]) : join(work, 'chaossql');
  if (!process.argv[2]) {
    execFileSync('go', ['build', '-o', cli, './cmd/chaossql'], {
      cwd: ROOT,
      stdio: 'inherit',
      env: { ...process.env, CGO_ENABLED: '0' },
    });
  }
  const version = execFileSync(cli, ['--version'], { encoding: 'utf8' }).trim().split(/\s+/).pop();

  mkdirSync(OUT_DIR, { recursive: true });
  const index = [];
  for (const name of SCENARIOS) {
    const doc = exportScenario(cli, name, version);
    writeFileSync(join(OUT_DIR, `${name}.json`), JSON.stringify(doc, null, 2) + '\n');
    index.push(name);
    console.log(
      `${name}: ${doc.status} ${doc.anomalyType} ` +
        `${doc.shrink.originalOps} -> ${doc.shrink.minimalOps} ops (${doc.shrink.reductionPct}%)`
    );
  }
} finally {
  rmSync(work, { recursive: true, force: true });
}

function exportScenario(cli, name, version) {
  const dir = join(work, name);
  cpSync(join(ROOT, 'examples', name), dir, { recursive: true });
  const specPath = join(dir, 'chaos.yaml');
  const spec = readFileSync(specPath, 'utf8');
  // The featured examples pin READ_UNCOMMITTED themselves; pin it when missing
  // and refuse any other level, which would record a different run.
  const pinned = spec.match(/^\s+isolation:\s*"?([A-Z_]+)"?/m);
  if (pinned && pinned[1] !== ISOLATION) throw new Error(`${name}: example pins ${pinned[1]}, expected ${ISOLATION}`);
  if (!pinned) writeFileSync(specPath, spec.replace(/^(database:\s*\n)/m, `$1  isolation: "${ISOLATION}"\n`));

  const artifactPath = join(dir, 'result.json');
  const out = JSON.parse(
    // --fail-on never: the violation is what we record, not an error.
    execFileSync(cli, ['run', specPath, '--json', '--export-result', artifactPath, '--fail-on', 'never'], {
      cwd: dir,
      encoding: 'utf8',
      maxBuffer: 64 * 1024 * 1024,
    })
  );
  if (out.status !== 'violation' || !out.shrink) {
    throw new Error(`${name}: expected a shrunk violation, got status=${out.status}`);
  }
  const artifact = JSON.parse(readFileSync(artifactPath, 'utf8'));

  // One entry per operation of the full run, in schedule order.
  const rawOps = [];
  const seen = new Set();
  for (const d of out.schedule.decisions) {
    if (seen.has(d.operation_id)) continue;
    seen.add(d.operation_id);
    rawOps.push({ id: d.operation_id, worker: d.worker_id });
  }

  const inv = out.failing_invariant ?? {};
  return {
    scenario: name,
    source: `examples/${name}/chaos.yaml`,
    engineVersion: version,
    driver: out.spec.driver,
    isolation: out.isolation,
    seed: out.seed,
    status: out.status,
    anomalyType: out.anomaly_type,
    invariant: {
      name: inv.name ?? null,
      expression: inv.expression ?? null,
      actual: inv.actual_values ?? {},
    },
    shrink: {
      originalOps: out.shrink.original_size,
      minimalOps: out.shrink.reduced_size,
      reductionPct: Math.round(out.shrink.reduction_ratio * 10) / 10,
      trials: out.shrink.trials,
      durationMs: Math.round(out.shrink.duration / 1e6),
    },
    rawOps,
    minimalOps: out.shrink.minimal_ops.map((op) => ({
      id: op.id,
      name: op.name,
      params: op.params ?? {},
    })),
    // Physical order of the minimal replay. Timestamps are dropped on purpose:
    // they measure this machine, not the bug.
    minimalTrace: artifact.trace.map((e) => ({
      worker: e.worker_id,
      op: e.op_index,
      type: e.type,
      sql: e.sql,
    })),
  };
}
