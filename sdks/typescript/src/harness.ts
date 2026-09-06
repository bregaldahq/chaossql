/**
 * Fluent ChaosHarness for TypeScript/Node.js (@chaossql/test).
 */

import { executeIPC } from './engine';
import { ChaosHarnessOptions, ChaosResult, RunOptions, ScheduledOpStep } from './types';

export class ChaosHarness {
  private driver: string;
  private dsn: string;
  private binPath?: string;
  private schemaSQL: string = '';
  private seedSQL: string = '';
  private invariants: Array<{ name: string; query: string; assert: string }> = [];
  private operations: Array<{
    name: string;
    weight: number;
    params?: Record<string, string>;
    steps: ScheduledOpStep[];
  }> = [];
  private lastResult?: ChaosResult;

  constructor(options: ChaosHarnessOptions = {}) {
    this.driver = options.driver || 'sqlite';
    this.dsn = options.dsn || ':memory:';
    this.binPath = options.binPath;
  }

  public withSchema(schema: string): this {
    this.schemaSQL = schema.trim();
    return this;
  }

  public withSeed(seed: string): this {
    this.seedSQL = seed.trim();
    return this;
  }

  public withInvariant(name: string, query: string, assertion: string): this {
    this.invariants.push({
      name,
      query: query.trim(),
      assert: assertion.trim(),
    });
    return this;
  }

  public addOperation(
    name: string,
    steps: Array<string | ScheduledOpStep>,
    params?: Record<string, string>,
    weight: number = 1.0
  ): this {
    const parsedSteps: ScheduledOpStep[] = steps.map((s) => {
      if (typeof s === 'string') {
        return { sql: s };
      }
      return s;
    });

    this.operations.push({
      name,
      weight,
      params,
      steps: parsedSteps,
    });
    return this;
  }

  private buildPayload(opts: RunOptions = {}): any {
    return {
      driver: this.driver,
      dsn: this.dsn,
      schema: this.schemaSQL,
      seed: this.seedSQL,
      invariants: this.invariants,
      operations: this.operations,
      workers: opts.workers || 2,
      iterations: opts.iterations || 10,
      seed_value: opts.seed !== undefined ? opts.seed : 42,
    };
  }

  public async run(options: RunOptions = {}): Promise<ChaosResult> {
    const payload = this.buildPayload(options);
    const result = await executeIPC(payload, this.binPath);
    this.lastResult = result;
    return result;
  }

  public async runAndShrink(options: RunOptions = {}): Promise<ChaosResult> {
    return this.run(options);
  }

  public async assertNoAnomalies(options: RunOptions = {}): Promise<ChaosResult> {
    const result = await this.run(options);
    if (result.violationDetected) {
      const invName = result.failingInvariant?.name || 'unknown';
      const lines = [
        '',
        `🚨 ChaosSQL Isolation Anomaly Detected: ${result.anomalyType} [${result.anomalyCode}]`,
        `   Failing Invariant: ${invName}`,
      ];

      if (result.failingInvariant?.observed) {
        lines.push(`   Observed DB State: ${JSON.stringify(result.failingInvariant.observed)}`);
      }

      if (result.minimalOperations.length > 0) {
        lines.push(`   Minimal Counterexample (${result.minimalOperations.length} operations isolated):`);
        for (const op of result.minimalOperations) {
          lines.push(`     - Op #${op.id} [${op.name}]`);
          for (const step of op.steps) {
            const cap = step.capture ? ` (capture: ${step.capture})` : '';
            lines.push(`         SQL: ${step.sql}${cap}`);
          }
        }
      }

      if (result.mermaid) {
        lines.push('   Mermaid Sequence Diagram:');
        lines.push('   ```mermaid');
        for (const mLine of result.mermaid.split('\n')) {
          lines.push(`   ${mLine}`);
        }
        lines.push('   ```');
      }

      throw new Error(lines.join('\n'));
    }

    return result;
  }

  public async exportStandaloneRepro(outputPath: string): Promise<string> {
    if (!this.lastResult) {
      await this.run();
    }
    if (!this.lastResult) {
      throw new Error('No execution result available to export.');
    }
    return this.lastResult.exportStandaloneRepro(outputPath);
  }
}
