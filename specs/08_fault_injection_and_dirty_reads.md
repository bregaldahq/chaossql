# Spec 08: Stochastic Fault Injection, Dirty Read (G1a) Detection & Interactive Trace Replay

## 1. Fault Injection Architecture (`internal/faults/`)
- Inject stochastic system faults during transaction execution:
  - `FaultAbort`: Force a real driver rollback at random step $k$ ($0 < k < |\text{steps}|$). No write from the aborted operation may remain committed.
  - `FaultLatencySpike`: Inject sudden $10\text{ms}$ to $50\text{ms}$ delay to widen race windows. The delay must stop when the execution context is canceled.
  - `FaultDisconnect`: Simulate abrupt client connection loss.
- Configurable in YAML spec under `engine.faults`:
  ```yaml
  engine:
    faults:
      abort_probability: 0.2
      latency_spike_ms: [10, 50]
      latency_probability: 0.1
  ```

- Fault handling uses the same transaction opened for the operation. A cancellation or intentional abort cannot switch to an autocommit connection for cleanup.

## 2. G1a Dirty Read / Aborted Read Anomaly
- Formal Definition (Adya 1999 / Berenson 1995):
  $$w_1(x) \dots r_2(x) \dots (a_1 \text{ and } c_2 \text{ in any order})$$
- Anomaly occurs when transaction $T_2$ reads data modified by transaction $T_1$, and $T_1$ subsequently aborts ($a_1$), while $T_2$ commits ($c_2$) or uses the dirty value.
- Anomaly classification: `AnomalyG1aDirtyRead`.

## 3. Interactive Trace Replayer (`chaossql replay`)
- CLI command to replay any saved execution trace step-by-step or with formatted visual timeline.
- Displays:
  - Chronological transaction event stream.
  - Step index, Worker ID, SQL Statement, Captured Values.
  - Conflict markers (WW, WR, RW edges).
