/**
 * IPC subprocess client communicating with ChaosSQL engine over stdin/stdout.
 */

import { spawn } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import { ChaosResult, InvariantResult, ScheduledOp, ShrinkResult } from './types';

export function findChaosSQLBinary(explicitPath?: string): string {
  if (explicitPath && fs.existsSync(explicitPath)) {
    return path.resolve(explicitPath);
  }

  const envPath = process.env.CHAOSSQL_BIN_PATH;
  if (envPath && fs.existsSync(envPath)) {
    return path.resolve(envPath);
  }

  const binName = process.platform === 'win32' ? 'chaossql.exe' : 'chaossql';

  // Walk up directory tree from current module looking for bin/chaossql
  let currentDir = __dirname;
  for (let i = 0; i < 6; i++) {
    const candidate = path.join(currentDir, 'bin', binName);
    if (fs.existsSync(candidate)) {
      return candidate;
    }
    const candidateNoExt = path.join(currentDir, 'bin', 'chaossql');
    if (fs.existsSync(candidateNoExt)) {
      return candidateNoExt;
    }
    const parent = path.dirname(currentDir);
    if (parent === currentDir) break;
    currentDir = parent;
  }

  // Look in PATH
  const envPaths = (process.env.PATH || '').split(path.delimiter);
  for (const p of envPaths) {
    const candidate = path.join(p, binName);
    if (fs.existsSync(candidate)) {
      return candidate;
    }
    const candidateNoExt = path.join(p, 'chaossql');
    if (fs.existsSync(candidateNoExt)) {
      return candidateNoExt;
    }
  }

  throw new Error(
    'ChaosSQL core engine binary not found. Please compile it using `make build` ' +
      'or configure the CHAOSSQL_BIN_PATH environment variable.'
  );
}

function deriveAnomalyCode(anomalyType: string): string {
  const t = (anomalyType || '').toUpperCase();
  if (t.includes('LOST_UPDATE') || t.includes('P4')) return 'P4';
  if (t.includes('WRITE_SKEW') || t.includes('A5B')) return 'A5B';
  if (t.includes('READ_SKEW') || t.includes('A5A')) return 'A5A';
  if (t.includes('DIRTY_WRITE') || t.includes('G0')) return 'G0';
  if (t.includes('DIRTY_READ') || t.includes('G1A')) return 'G1a';
  if (t.includes('CIRCULAR') || t.includes('G1C')) return 'G1c';
  if (t.includes('ANTI_DEPENDENCY') || t.includes('G2')) return 'G2';
  return t;
}

export function executeIPC(payload: any, binaryPath?: string, timeoutMs: number = 60000): Promise<ChaosResult> {
  return new Promise((resolve, reject) => {
    let bin: string;
    try {
      bin = findChaosSQLBinary(binaryPath);
    } catch (err) {
      return reject(err);
    }

    const child = spawn(bin, ['engine'], {
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    let stdoutData = '';
    let stderrData = '';

    const timer = setTimeout(() => {
      child.kill();
      reject(new Error(`ChaosSQL engine execution timed out after ${timeoutMs}ms`));
    }, timeoutMs);

    child.stdout.on('data', (chunk) => {
      stdoutData += chunk.toString();
    });

    child.stderr.on('data', (chunk) => {
      stderrData += chunk.toString();
    });

    child.on('error', (err) => {
      clearTimeout(timer);
      reject(new Error(`Failed to spawn ChaosSQL engine process (${bin}): ${err.message}`));
    });

    child.on('close', (code) => {
      clearTimeout(timer);
      if (code !== 0 && !stdoutData.trim()) {
        return reject(
          new Error(`ChaosSQL engine exited with code ${code}. Stderr:\n${stderrData.trim()}`)
        );
      }

      try {
        const raw = JSON.parse(stdoutData.trim());
        const anomalyType = raw.anomaly_type || 'UNKNOWN';
        const anomalyCode = deriveAnomalyCode(anomalyType);

        let failingInv: InvariantResult | undefined;
        if (raw.failing_invariant) {
          failingInv = {
            name: raw.failing_invariant.name,
            query: raw.failing_invariant.query,
            expression: raw.failing_invariant.expression,
            passed: raw.failing_invariant.passed,
            observed: raw.failing_invariant.observed,
            errorMessage: raw.failing_invariant.error_message,
          };
        }

        const minimalOps: ScheduledOp[] = (raw.minimal_operations || []).map((op: any) => ({
          id: op.id,
          name: op.name,
          params: op.params,
          steps: (op.steps || []).map((s: any) => ({
            sql: s.sql || s.SQL,
            capture: s.capture || s.Capture,
          })),
        }));

        let shrinkRes: ShrinkResult | undefined;
        if (raw.shrink) {
          shrinkRes = {
            originalSize: raw.shrink.original_size,
            reducedSize: raw.shrink.reduced_size,
            reductionRatio: raw.shrink.reduction_ratio,
            iterations: raw.shrink.iterations,
            minimalOps: (raw.shrink.minimal_ops || []).map((op: any) => ({
              id: op.id,
              name: op.name,
              params: op.params,
              steps: (op.steps || []).map((s: any) => ({
                sql: s.sql || s.SQL,
                capture: s.capture || s.Capture,
              })),
            })),
          };
        }

        const result: ChaosResult = {
          success: Boolean(raw.success),
          violationDetected: Boolean(raw.violation_detected),
          anomalyDetected: Boolean(raw.violation_detected),
          anomalyType,
          anomalyCode,
          failingInvariant: failingInv,
          durationMs: raw.duration_ms || 0,
          traceEventsCount: raw.trace_events_count || 0,
          minimalOperations: minimalOps,
          shrink: shrinkRes,
          mermaid: raw.mermaid || '',
          reproGo: raw.repro_go || '',
          reproPython: raw.repro_python || '',
          reproTypeScript: raw.repro_typescript || '',
          error: raw.error,
          async exportStandaloneRepro(filePath: string): Promise<string> {
            const code = this.reproTypeScript || this.reproPython;
            if (!code) {
              throw new Error('No reproduction code available in execution result.');
            }
            const absPath = path.resolve(filePath);
            await fs.promises.mkdir(path.dirname(absPath), { recursive: true });
            await fs.promises.writeFile(absPath, code, 'utf8');
            return absPath;
          },
        };

        resolve(result);
      } catch (err: any) {
        reject(
          new Error(
            `Failed to parse JSON output from ChaosSQL engine: ${err.message}\nStdout:\n${stdoutData}\nStderr:\n${stderrData}`
          )
        );
      }
    });

    // Write input payload to stdin and close pipe
    child.stdin.write(JSON.stringify(payload));
    child.stdin.end();
  });
}
