#!/usr/bin/env bash
set -e

echo "=========================================================="
echo "  MANUAL VERIFICATION: ChaosSQL Transparent Reverse Proxy "
echo "=========================================================="

UPSTREAM_PORT=5432
PROXY_PORT=5433
UI_PORT=8095
SARIF_OUT="/tmp/manual_proxy_anomalies.sarif"

rm -f /tmp/proxy_stdout.log /tmp/proxy_stderr.log "$SARIF_OUT"

echo "1. Initializing test schema on PostgreSQL (127.0.0.1:$UPSTREAM_PORT)..."
PGPASSWORD=splitsub psql -h 127.0.0.1 -p "$UPSTREAM_PORT" -U splitsub -d splitsub << 'EOF'
DROP TABLE IF EXISTS chaossql_proxy_accounts;
CREATE TABLE chaossql_proxy_accounts (id INT PRIMARY KEY, balance INT);
INSERT INTO chaossql_proxy_accounts VALUES (1, 1000);

DROP TABLE IF EXISTS chaossql_proxy_doctors;
CREATE TABLE chaossql_proxy_doctors (id INT PRIMARY KEY, name TEXT, on_call BOOLEAN);
INSERT INTO chaossql_proxy_doctors VALUES (1, 'Alice', true), (2, 'Bob', true);
EOF
echo "✔ Test schema created successfully."

echo "2. Starting ChaosSQL Transparent Proxy on 127.0.0.1:$PROXY_PORT (Upstream: $UPSTREAM_PORT)..."
/root/chaossql/bin/chaossql proxy \
  --listen 127.0.0.1:"$PROXY_PORT" \
  --upstream 127.0.0.1:"$UPSTREAM_PORT" \
  --protocol postgres \
  --jitter-min 100us \
  --jitter-max 1ms \
  --commit-barrier-min 1ms \
  --commit-barrier-max 5ms \
  --export-sarif "$SARIF_OUT" \
  --ui-port "$UI_PORT" \
  --fail-on-anomaly=false > /tmp/proxy_stdout.log 2> /tmp/proxy_stderr.log &

PROXY_PID=$!
echo "✔ Proxy started in background with PID $PROXY_PID"

sleep 1

echo "3. Testing transparent connectivity through proxy port $PROXY_PORT..."
BALANCE=$(PGPASSWORD=splitsub psql -h 127.0.0.1 -p "$PROXY_PORT" -U splitsub -d splitsub -t -A -c "SELECT balance FROM chaossql_proxy_accounts WHERE id = 1;")
echo "✔ Read balance through proxy: $BALANCE"

echo "4. Executing concurrent transactions: Banking Lost Update (P4)..."
# Worker 1: Reads, pauses, overwrites
PGPASSWORD=splitsub psql -h 127.0.0.1 -p "$PROXY_PORT" -U splitsub -d splitsub << 'EOF' &
BEGIN;
SELECT balance FROM chaossql_proxy_accounts WHERE id = 1;
SELECT pg_sleep(0.08);
UPDATE chaossql_proxy_accounts SET balance = balance - 100 WHERE id = 1;
COMMIT;
EOF
W1_PID=$!

# Worker 2: Overwrites while W1 is uncommitted
sleep 0.02
PGPASSWORD=splitsub psql -h 127.0.0.1 -p "$PROXY_PORT" -U splitsub -d splitsub << 'EOF' &
BEGIN;
SELECT balance FROM chaossql_proxy_accounts WHERE id = 1;
UPDATE chaossql_proxy_accounts SET balance = balance - 50 WHERE id = 1;
COMMIT;
EOF
W2_PID=$!

wait $W1_PID || true
wait $W2_PID || true
echo "✔ Lost Update concurrent runs completed."

echo "5. Executing concurrent transactions: Hospital Write Skew (A5B)..."
# Doctor 1 attempts to leave on-call
PGPASSWORD=splitsub psql -h 127.0.0.1 -p "$PROXY_PORT" -U splitsub -d splitsub << 'EOF' &
BEGIN;
SELECT count(*) FROM chaossql_proxy_doctors WHERE on_call = true;
SELECT pg_sleep(0.08);
UPDATE chaossql_proxy_doctors SET on_call = false WHERE id = 1;
COMMIT;
EOF
D1_PID=$!

# Doctor 2 concurrently attempts to leave on-call
sleep 0.02
PGPASSWORD=splitsub psql -h 127.0.0.1 -p "$PROXY_PORT" -U splitsub -d splitsub << 'EOF' &
BEGIN;
SELECT count(*) FROM chaossql_proxy_doctors WHERE on_call = true;
UPDATE chaossql_proxy_doctors SET on_call = false WHERE id = 2;
COMMIT;
EOF
D2_PID=$!

wait $D1_PID || true
wait $D2_PID || true
echo "✔ Write Skew concurrent runs completed."

sleep 0.5

echo "6. Querying Live Dashboard & Diagnostic Endpoints (127.0.0.1:$UI_PORT)..."
UI_STATUS=$(curl -s "http://127.0.0.1:$UI_PORT/api/status")
echo "  • Live Status: $UI_STATUS"

UI_ANOMALIES=$(curl -s "http://127.0.0.1:$UI_PORT/api/anomalies")
echo "  • Detected Anomalies Count: $(echo "$UI_ANOMALIES" | grep -o '"id":' | wc -l)"

echo "7. Sending SIGINT to proxy for graceful shutdown..."
kill -INT "$PROXY_PID"
wait "$PROXY_PID" || true
echo "✔ Proxy terminated cleanly."

echo "8. Verifying SARIF 2.1.0 output..."
if [ -f "$SARIF_OUT" ]; then
  echo "✔ SARIF file exists: $SARIF_OUT ($(wc -c < "$SARIF_OUT") bytes)"
  head -n 25 "$SARIF_OUT"
else
  echo "❌ SARIF file was not generated!"
  exit 1
fi

echo "9. Proxy Stderr (Anomaly Alerts):"
cat /tmp/proxy_stderr.log

echo "10. Proxy Stdout (Summary Table):"
cat /tmp/proxy_stdout.log

echo "=========================================================="
echo "  MANUAL VERIFICATION COMPLETED SUCCESSFULLY! "
echo "=========================================================="
