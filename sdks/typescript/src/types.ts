/**
 * Strongly-typed definitions for ChaosSQL TypeScript/Node.js SDK (@chaossql/test).
 */

export interface InvariantResult {
  name: string;
  query: string;
  expression: string;
  passed: boolean;
  observed?: Record<string, any>;
  errorMessage?: string;
}

export interface ScheduledOpStep {
  sql: string;
  capture?: string;
}

export interface ScheduledOp {
  id: number;
  name: string;
  params?: Record<string, string>;
  steps: ScheduledOpStep[];
}

export interface ShrinkResult {
  originalSize: number;
  reducedSize: number;
  reductionRatio: number;
  iterations: number;
  minimalOps: ScheduledOp[];
}

export interface ScheduleDecision {
  sequence: number;
  operation_id: number;
  worker_id: number;
  step_index: number;
  jitter_ms: number;
  latency_ms: number;
  abort: boolean;
}

export interface SchedulePlan {
  version: number;
  seed: number;
  workers: number;
  decisions: ScheduleDecision[];
}

export interface ChaosResult {
  status: 'passed' | 'violation' | 'execution_error' | 'inconclusive' | 'canceled' | string;
  isolation?: string;
  seed: number;
  schedule: SchedulePlan;
  operationErrors: Array<{
    operation_id: number;
    operation: string;
    step_index?: number;
    phase: string;
    message: string;
  }>;
  success: boolean;
  violationDetected: boolean;
  anomalyDetected: boolean;
  anomalyType: string;
  anomalyCode: string;
  failingInvariant?: InvariantResult;
  durationMs: number;
  traceEventsCount: number;
  minimalOperations: ScheduledOp[];
  shrink?: ShrinkResult;
  mermaid?: string;
  reproGo?: string;
  reproPython?: string;
  reproTypeScript?: string;
  error?: string;
  exportStandaloneRepro(filePath: string): Promise<string>;
}

export interface ChaosHarnessOptions {
  driver?: 'sqlite' | 'postgres' | 'mysql' | string;
  dsn?: string;
  isolation?: 'READ_UNCOMMITTED' | 'READ_COMMITTED' | 'REPEATABLE_READ' | 'SERIALIZABLE' | string;
  binPath?: string;
}

export interface RunOptions {
  workers?: number;
  iterations?: number;
  seed?: number;
}
