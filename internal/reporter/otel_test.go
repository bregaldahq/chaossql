package reporter_test

import (
	"encoding/hex"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
)

func TestGenerateOTLPTraceJSON_ValidHierarchy(t *testing.T) {
	spec := domain.Spec{
		Version:     "1.0",
		Name:        "banking_lost_update",
		Description: "Detects Lost Update (P4)",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
		},
		Engine: domain.EngineConfig{
			Workers: 2,
		},
	}

	trace := domain.ExecutionTrace{
		{
			Timestamp: 10 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "withdraw_vulnerable",
			StepIndex: 0,
			Type:      domain.EventBegin,
			SQL:       "BEGIN",
		},
		{
			Timestamp: 12 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "withdraw_vulnerable",
			StepIndex: 0,
			Type:      domain.EventBegin,
			SQL:       "BEGIN",
		},
		{
			Timestamp: 20 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "withdraw_vulnerable",
			StepIndex: 1,
			Type:      domain.EventExec,
			SQL:       "SELECT balance FROM accounts WHERE id = 1;",
		},
		{
			Timestamp: 25 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "withdraw_vulnerable",
			StepIndex: 1,
			Type:      domain.EventExec,
			SQL:       "SELECT balance FROM accounts WHERE id = 1;",
		},
		{
			Timestamp: 30 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "withdraw_vulnerable",
			StepIndex: 2,
			Type:      domain.EventExec,
			SQL:       "UPDATE accounts SET balance = 900 WHERE id = 1;",
		},
		{
			Timestamp: 35 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "withdraw_vulnerable",
			StepIndex: 3,
			Type:      domain.EventCommit,
			SQL:       "COMMIT",
		},
		{
			Timestamp: 40 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "withdraw_vulnerable",
			StepIndex: 2,
			Type:      domain.EventExec,
			SQL:       "UPDATE accounts SET balance = 900 WHERE id = 1;",
		},
		{
			Timestamp: 45 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "withdraw_vulnerable",
			StepIndex: 3,
			Type:      domain.EventCommit,
			SQL:       "COMMIT",
		},
	}

	jsonStr, err := reporter.GenerateOTLPTraceJSON(trace, spec)
	if err != nil {
		t.Fatalf("unexpected error generating OTLP trace JSON: %v", err)
	}
	if jsonStr == "" {
		t.Fatalf("expected non-empty JSON string")
	}

	var root map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &root); err != nil {
		t.Fatalf("output is not valid JSON: %v\nJSON:\n%s", err, jsonStr)
	}

	// 1. Verify resourceSpans
	resourceSpans, ok := root["resourceSpans"].([]interface{})
	if !ok || len(resourceSpans) == 0 {
		t.Fatalf("expected non-empty resourceSpans array")
	}

	rs0 := resourceSpans[0].(map[string]interface{})

	// 2. Verify Resource Attributes
	resource, ok := rs0["resource"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected resource object in resourceSpans[0]")
	}
	resAttrs, ok := resource["attributes"].([]interface{})
	if !ok {
		t.Fatalf("expected attributes array in resource")
	}

	findAttr := func(attrs []interface{}, key string) (interface{}, bool) {
		for _, a := range attrs {
			am := a.(map[string]interface{})
			if am["key"] == key {
				valMap, _ := am["value"].(map[string]interface{})
				if s, ok := valMap["stringValue"]; ok {
					return s, true
				}
				if i, ok := valMap["intValue"]; ok {
					return i, true
				}
				if b, ok := valMap["boolValue"]; ok {
					return b, true
				}
			}
		}
		return nil, false
	}

	if val, ok := findAttr(resAttrs, "service.name"); !ok || val != "chaossql" {
		t.Errorf("expected resource attribute service.name='chaossql', got %v", val)
	}
	if val, ok := findAttr(resAttrs, "db.system"); !ok || val != "sqlite" {
		t.Errorf("expected resource attribute db.system='sqlite', got %v", val)
	}
	if val, ok := findAttr(resAttrs, "scenario.name"); !ok || val != "banking_lost_update" {
		t.Errorf("expected resource attribute scenario.name='banking_lost_update', got %v", val)
	}

	// 3. Verify Scope
	scopeSpans, ok := rs0["scopeSpans"].([]interface{})
	if !ok || len(scopeSpans) == 0 {
		t.Fatalf("expected non-empty scopeSpans array")
	}
	ss0 := scopeSpans[0].(map[string]interface{})
	scope, ok := ss0["scope"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected scope object in scopeSpans[0]")
	}
	if scope["name"] != "github.com/bregaldahq/chaossql" {
		t.Errorf("expected scope name 'github.com/bregaldahq/chaossql', got %v", scope["name"])
	}
	if scope["version"] != "1.0.0" {
		t.Errorf("expected scope version '1.0.0', got %v", scope["version"])
	}

	// 4. Verify Spans Hierarchy
	spans, ok := ss0["spans"].([]interface{})
	if !ok || len(spans) == 0 {
		t.Fatalf("expected non-empty spans array")
	}

	// Root span should be first
	rootSpan := spans[0].(map[string]interface{})
	rootTraceID, _ := rootSpan["traceId"].(string)
	rootSpanID, _ := rootSpan["spanId"].(string)

	if len(rootTraceID) != 32 {
		t.Errorf("expected 32 hex chars for root traceId, got %q", rootTraceID)
	}
	if len(rootSpanID) != 16 {
		t.Errorf("expected 16 hex chars for root spanId, got %q", rootSpanID)
	}
	if parent, ok := rootSpan["parentSpanId"].(string); ok && parent != "" {
		t.Errorf("expected root span parentSpanId to be empty, got %q", parent)
	}

	// Verify all spans share rootTraceID and have valid IDs
	spanIDs := make(map[string]bool)
	spanMap := make(map[string]map[string]interface{})

	for i, s := range spans {
		sm := s.(map[string]interface{})
		tid, _ := sm["traceId"].(string)
		sid, _ := sm["spanId"].(string)

		if tid != rootTraceID {
			t.Errorf("span %d has mismatched traceId: got %q, expected %q", i, tid, rootTraceID)
		}
		if spanIDs[sid] {
			t.Errorf("duplicate spanId found: %q", sid)
		}
		spanIDs[sid] = true
		spanMap[sid] = sm
	}

	// Find transaction spans (children of rootSpan)
	txSpans := make(map[string]map[string]interface{})
	stmtSpans := make(map[string]map[string]interface{})

	for sid, sm := range spanMap {
		if sid == rootSpanID {
			continue
		}
		parent, _ := sm["parentSpanId"].(string)
		if parent == rootSpanID {
			txSpans[sid] = sm
		} else {
			stmtSpans[sid] = sm
		}
	}

	if len(txSpans) != 2 {
		t.Errorf("expected 2 transaction spans (child of root span), got %d", len(txSpans))
	}

	// Verify statement spans are children of transaction spans
	for sid, stmt := range stmtSpans {
		parent, _ := stmt["parentSpanId"].(string)
		if _, isChildOfTx := txSpans[parent]; !isChildOfTx {
			t.Errorf("statement span %s has parent %s which is not a known transaction span", sid, parent)
		}

		// Verify statement span attributes: db.statement, db.event_type, latency_us
		stmtAttrs, ok := stmt["attributes"].([]interface{})
		if !ok {
			t.Errorf("statement span %s missing attributes array", sid)
			continue
		}

		if stmtSQL, ok := findAttr(stmtAttrs, "db.statement"); !ok || stmtSQL == "" {
			t.Errorf("statement span %s missing or empty db.statement attribute", sid)
		}
		if evType, ok := findAttr(stmtAttrs, "db.event_type"); !ok || evType == "" {
			t.Errorf("statement span %s missing or empty db.event_type attribute", sid)
		}
		if lat, ok := findAttr(stmtAttrs, "latency_us"); !ok {
			t.Errorf("statement span %s missing latency_us attribute", sid)
		} else if latNum, ok := lat.(float64); !ok || latNum < 0 {
			t.Errorf("statement span %s latency_us should be >= 0, got %v", sid, lat)
		}
	}
}

func TestGenerateOTLPTraceJSON_EmptyTrace(t *testing.T) {
	spec := domain.Spec{
		Name: "empty_scenario",
		Database: domain.DatabaseConfig{
			Driver: "postgres",
		},
	}

	jsonStr, err := reporter.GenerateOTLPTraceJSON(domain.ExecutionTrace{}, spec)
	if err != nil {
		t.Fatalf("unexpected error for empty trace: %v", err)
	}

	var root map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &root); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	rs, ok := root["resourceSpans"].([]interface{})
	if !ok || len(rs) == 0 {
		t.Fatalf("expected resourceSpans array")
	}
}

func TestGenerateOTLPTraceJSON_ErrorSpanStatus(t *testing.T) {
	spec := domain.Spec{
		Name: "error_scenario",
		Database: domain.DatabaseConfig{
			Driver: "mysql",
		},
	}

	trace := domain.ExecutionTrace{
		{
			Timestamp: 10 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "failing_op",
			StepIndex: 0,
			Type:      domain.EventBegin,
			SQL:       "BEGIN",
		},
		{
			Timestamp: 15 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "failing_op",
			StepIndex: 1,
			Type:      domain.EventError,
			SQL:       "INSERT INTO nonexistent VALUES (1);",
			Error:     "table nonexistent does not exist",
		},
		{
			Timestamp: 20 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "failing_op",
			StepIndex: 2,
			Type:      domain.EventRollback,
			SQL:       "ROLLBACK",
		},
	}

	jsonStr, err := reporter.GenerateOTLPTraceJSON(trace, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var root map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &root); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	rs0 := root["resourceSpans"].([]interface{})[0].(map[string]interface{})
	ss0 := rs0["scopeSpans"].([]interface{})[0].(map[string]interface{})
	spans := ss0["spans"].([]interface{})

	// Find the error span
	foundErrorSpan := false
	for _, s := range spans {
		sm := s.(map[string]interface{})
		status, _ := sm["status"].(map[string]interface{})
		if status != nil {
			if code, ok := status["code"].(float64); ok && code == 2 {
				foundErrorSpan = true
				if msg, _ := status["message"].(string); msg != "" {
					if msg != "table nonexistent does not exist" {
						t.Errorf("expected error message in status, got: %s", msg)
					}
				}
			}
		}
	}

	if !foundErrorSpan {
		t.Errorf("expected to find at least one span with status code 2 (ERROR)")
	}
}

func TestGenerateOTLPTraceJSON_SavepointsAndRollbackTo(t *testing.T) {
	spec := domain.Spec{
		Name: "savepoint_scenario",
		Database: domain.DatabaseConfig{
			Driver: "postgres",
		},
	}

	trace := domain.ExecutionTrace{
		{
			Timestamp: 5 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "savepoint_op",
			StepIndex: 0,
			Type:      domain.EventBegin,
			SQL:       "BEGIN",
		},
		{
			Timestamp: 10 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "savepoint_op",
			StepIndex: 1,
			Type:      domain.EventSavepoint,
			SQL:       "SAVEPOINT sp1",
		},
		{
			Timestamp: 15 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "savepoint_op",
			StepIndex: 2,
			Type:      domain.EventExec,
			SQL:       "INSERT INTO log (msg) VALUES ('partial');",
		},
		{
			Timestamp: 20 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "savepoint_op",
			StepIndex: 3,
			Type:      domain.EventRollbackTo,
			SQL:       "ROLLBACK TO SAVEPOINT sp1",
		},
		{
			Timestamp: 25 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "savepoint_op",
			StepIndex: 4,
			Type:      domain.EventReleaseSavepoint,
			SQL:       "RELEASE SAVEPOINT sp1",
		},
		{
			Timestamp: 30 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "savepoint_op",
			StepIndex: 5,
			Type:      domain.EventCommit,
			SQL:       "COMMIT",
		},
	}

	jsonStr, err := reporter.GenerateOTLPTraceJSON(trace, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var root map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &root); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	rs0 := root["resourceSpans"].([]interface{})[0].(map[string]interface{})
	ss0 := rs0["scopeSpans"].([]interface{})[0].(map[string]interface{})
	spans := ss0["spans"].([]interface{})

	// 1 root span + 1 tx span + 6 statement spans = 8 spans
	if len(spans) != 8 {
		t.Errorf("expected 8 spans for savepoint scenario, got %d", len(spans))
	}
}

func TestGenerateOTLPTraceJSON_HexIDFormats(t *testing.T) {
	spec := domain.Spec{
		Name: "hex_test",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
		},
	}

	trace := domain.ExecutionTrace{
		{
			Timestamp: 5 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "test_op",
			StepIndex: 0,
			Type:      domain.EventBegin,
			SQL:       "BEGIN",
		},
	}

	jsonStr, err := reporter.GenerateOTLPTraceJSON(trace, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var root map[string]interface{}
	_ = json.Unmarshal([]byte(jsonStr), &root)
	rs0 := root["resourceSpans"].([]interface{})[0].(map[string]interface{})
	ss0 := rs0["scopeSpans"].([]interface{})[0].(map[string]interface{})
	spans := ss0["spans"].([]interface{})

	hex32Regex := regexp.MustCompile(`^[0-9a-f]{32}$`)
	hex16Regex := regexp.MustCompile(`^[0-9a-f]{16}$`)

	for _, s := range spans {
		sm := s.(map[string]interface{})
		traceID, _ := sm["traceId"].(string)
		spanID, _ := sm["spanId"].(string)

		if !hex32Regex.MatchString(traceID) {
			t.Errorf("traceId %q does not match 32 lowercase hex chars", traceID)
		}
		if _, err := hex.DecodeString(traceID); err != nil {
			t.Errorf("traceId %q is not valid hex: %v", traceID, err)
		}

		if !hex16Regex.MatchString(spanID) {
			t.Errorf("spanId %q does not match 16 lowercase hex chars", spanID)
		}
		if _, err := hex.DecodeString(spanID); err != nil {
			t.Errorf("spanId %q is not valid hex: %v", spanID, err)
		}
	}
}
