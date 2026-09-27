package swarm

import (
	"context"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestResolveDSN(t *testing.T) {
	cases := []struct {
		name   string
		spec   domain.Spec
		driver string
		env    map[string]string
		want   string
	}{
		{"spec DSN for its own driver", domain.Spec{Database: domain.DatabaseConfig{Driver: "Postgres", DSN: "postgres://spec"}}, "postgres",
			map[string]string{"DATABASE_URL": "postgres://env"}, "postgres://spec"},
		{"spec DSN ignored for other drivers", domain.Spec{Database: domain.DatabaseConfig{Driver: "mysql", DSN: "mysql://spec"}}, "postgres",
			map[string]string{"DATABASE_URL": "postgres://env"}, "postgres://env"},
		{"postgres falls back to POSTGRES_DSN", domain.Spec{}, "postgresql",
			map[string]string{"DATABASE_URL": "mysql://wrong-engine", "POSTGRES_DSN": "postgres://pg"}, "postgres://pg"},
		{"mysql prefers MYSQL_DSN", domain.Spec{}, "mysql",
			map[string]string{"MYSQL_DSN": "root@tcp/db", "DATABASE_URL": "mysql://url"}, "root@tcp/db"},
		{"mariadb falls back to mysql DATABASE_URL", domain.Spec{}, "mariadb",
			map[string]string{"DATABASE_URL": "mysql://url"}, "mysql://url"},
		{"mysql ignores postgres DATABASE_URL", domain.Spec{}, "mysql",
			map[string]string{"DATABASE_URL": "postgres://url"}, ""},
		{"sqlite uses driver default", domain.Spec{}, "sqlite",
			map[string]string{"DATABASE_URL": "postgres://url"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{"DATABASE_URL", "POSTGRES_DSN", "MYSQL_DSN"} {
				t.Setenv(key, tc.env[key])
			}
			if got := resolveDSN(tc.spec, tc.driver); got != tc.want {
				t.Fatalf("resolveDSN = %q, want %q", got, tc.want)
			}
		})
	}
}

func exec(worker, op int, sql string) domain.TraceEvent {
	return domain.TraceEvent{WorkerID: worker, OpIndex: op, Type: domain.EventExec, SQL: sql}
}

func TestDetectAnomalyFromTrace(t *testing.T) {
	cases := []struct {
		name  string
		trace domain.ExecutionTrace
		want  domain.AnomalyType
	}{
		{"dirty write", domain.ExecutionTrace{
			exec(1, 1, "UPDATE t SET v = 1 WHERE id = 1"),
			exec(2, 1, "UPDATE t SET v = 2 WHERE id = 1"),
			exec(2, 1, "UPDATE t SET v = 2 WHERE id = 2"),
			exec(1, 1, "UPDATE t SET v = 1 WHERE id = 2"),
		}, domain.AnomalyG0DirtyWrite},
		{"circular information flow", domain.ExecutionTrace{
			exec(1, 1, "UPDATE t SET v = 1 WHERE id = 1"),
			exec(2, 1, "SELECT v FROM t WHERE id = 1"),
			exec(2, 1, "UPDATE t SET v = 2 WHERE id = 2"),
			exec(1, 1, "SELECT v FROM t WHERE id = 2"),
		}, domain.AnomalyG1cCircularInfo},
		{"dirty read from aborted writer", domain.ExecutionTrace{
			exec(1, 1, "UPDATE t SET v = 1 WHERE id = 1"),
			exec(2, 1, "SELECT v FROM t WHERE id = 1"),
			exec(2, 1, "UPDATE t SET v = 2 WHERE id = 2"),
			exec(1, 1, "SELECT v FROM t WHERE id = 2"),
			{WorkerID: 1, OpIndex: 1, Type: domain.EventRollback},
		}, domain.AnomalyG1aDirtyRead},
		{"write skew", domain.ExecutionTrace{
			exec(1, 1, "SELECT v FROM t WHERE id = 1"),
			exec(2, 1, "SELECT v FROM t WHERE id = 2"),
			exec(1, 1, "UPDATE t SET v = 0 WHERE id = 2"),
			exec(2, 1, "UPDATE t SET v = 0 WHERE id = 1"),
		}, domain.AnomalyWriteSkew},
		{"anti-dependency cycle", domain.ExecutionTrace{
			exec(1, 1, "SELECT v FROM t WHERE id = 1"),
			exec(2, 1, "SELECT v FROM t WHERE id = 2"),
			exec(3, 1, "SELECT v FROM t WHERE id = 3"),
			exec(1, 1, "UPDATE t SET v = 0 WHERE id = 2"),
			exec(2, 1, "UPDATE t SET v = 0 WHERE id = 3"),
			exec(3, 1, "UPDATE t SET v = 0 WHERE id = 1"),
		}, domain.AnomalyG2AntiDependency},
		{"read skew", domain.ExecutionTrace{
			exec(1, 1, "SELECT v FROM t WHERE id = 1"),
			exec(2, 1, "UPDATE t SET v = 0 WHERE id = 1"),
			exec(2, 1, "UPDATE t SET v = 0 WHERE id = 2"),
			exec(1, 1, "SELECT v FROM t WHERE id = 2"),
		}, domain.AnomalyA5AReadSkew},
		{"lost update", domain.ExecutionTrace{
			exec(1, 1, "SELECT v FROM t WHERE id = 1"),
			exec(2, 1, "SELECT v FROM t WHERE id = 1"),
			exec(1, 1, "UPDATE t SET v = 1 WHERE id = 1"),
			exec(2, 1, "UPDATE t SET v = 2 WHERE id = 1"),
		}, domain.AnomalyLostUpdate},
		{"unclassified cycle", domain.ExecutionTrace{
			exec(1, 1, "UPDATE t SET v = 1 WHERE id = 1"),
			exec(2, 1, "UPDATE t SET v = 2 WHERE id = 1"),
			exec(2, 1, "UPDATE t SET v = 2 WHERE id = 2"),
			exec(1, 1, "SELECT v FROM t WHERE id = 2"),
		}, domain.AnomalyUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detectAnomalyFromTrace(tc.trace); got != string(tc.want) {
				t.Fatalf("detectAnomalyFromTrace = %q, want %q", got, tc.want)
			}
		})
	}

	if got := detectAnomalyFromTrace(nil); got != "" {
		t.Fatalf("empty trace = %q, want \"\"", got)
	}
	if got := detectAnomalyFromTrace(domain.ExecutionTrace{exec(1, 1, "SELECT v FROM t WHERE id = 1")}); got != "" {
		t.Fatalf("acyclic trace = %q, want \"\"", got)
	}
}

func TestExecuteDriverRun_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := executeDriverRun(ctx, domain.Spec{}, nil, "sqlite"); err != context.Canceled {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestEvaluateScenarioDivergence_SingleDriverViolation(t *testing.T) {
	diff := EvaluateScenarioDivergence("banking", map[string]DriverExecutionResult{
		"sqlite":   {Driver: "sqlite", ViolationDetected: true, FailingInvariant: "balance_preserved"},
		"postgres": {Driver: "postgres", Error: "connection refused"},
	}, []string{"sqlite", "postgres"})
	want := "Single active driver sqlite: violation detected (balance_preserved)"
	if diff.Divergent || diff.Summary != want {
		t.Fatalf("diff = %+v, want non-divergent summary %q", diff, want)
	}
}
