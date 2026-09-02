package reporter

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
)

// OTLPTracesData represents the root OpenTelemetry Tracing payload compliant with OTLP JSON.
type OTLPTracesData struct {
	ResourceSpans []OTLPResourceSpan `json:"resourceSpans"`
}

// OTLPResourceSpan represents the top-level grouping of resources and scopes.
type OTLPResourceSpan struct {
	Resource   OTLPResource   `json:"resource"`
	ScopeSpans []OTLPScopeSpan `json:"scopeSpans"`
}

// OTLPResource describes the entity producing telemetry.
type OTLPResource struct {
	Attributes []OTLPAttribute `json:"attributes"`
}

// OTLPScopeSpan groups spans emitted by an instrumentation scope.
type OTLPScopeSpan struct {
	Scope OTLPInstrumentationScope `json:"scope"`
	Spans []OTLPSpan               `json:"spans"`
}

// OTLPInstrumentationScope identifies the library generating spans.
type OTLPInstrumentationScope struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// OTLPSpan represents a single unit of work in a trace.
type OTLPSpan struct {
	TraceID           string          `json:"traceId"`
	SpanID            string          `json:"spanId"`
	ParentSpanID      string          `json:"parentSpanId,omitempty"`
	Name              string          `json:"name"`
	Kind              int             `json:"kind"` // 1 = INTERNAL, 3 = CLIENT
	StartTimeUnixNano string          `json:"startTimeUnixNano"`
	EndTimeUnixNano   string          `json:"endTimeUnixNano"`
	Attributes        []OTLPAttribute `json:"attributes,omitempty"`
	Status            OTLPStatus      `json:"status"`
}

// OTLPStatus holds the span completion status.
type OTLPStatus struct {
	Code    int    `json:"code"` // 0 = UNSET, 1 = OK, 2 = ERROR
	Message string `json:"message,omitempty"`
}

// OTLPAttribute represents a key-value attribute pair.
type OTLPAttribute struct {
	Key   string       `json:"key"`
	Value OTLPAnyValue `json:"value"`
}

// OTLPAnyValue is a strongly-typed union representing an attribute value.
type OTLPAnyValue struct {
	StringValue *string  `json:"stringValue,omitempty"`
	IntValue    *int64   `json:"intValue,omitempty"`
	BoolValue   *bool    `json:"boolValue,omitempty"`
	DoubleValue *float64 `json:"doubleValue,omitempty"`
}

func otlpStringAttr(k, v string) OTLPAttribute {
	val := v
	return OTLPAttribute{
		Key: k,
		Value: OTLPAnyValue{
			StringValue: &val,
		},
	}
}

func otlpIntAttr(k string, v int64) OTLPAttribute {
	val := v
	return OTLPAttribute{
		Key: k,
		Value: OTLPAnyValue{
			IntValue: &val,
		},
	}
}

type txGroup struct {
	WorkerID int
	OpIndex  int
	OpName   string
	Events   []domain.TraceEvent
}

// GenerateOTLPTraceJSON generates a compliant OpenTelemetry Traces JSON string from a chaos ExecutionTrace.
func GenerateOTLPTraceJSON(trace domain.ExecutionTrace, spec domain.Spec) (string, error) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	traceID := generateOTLPTraceID()
	rootSpanID := generateOTLPSpanID()

	// 1. Group events into transactions
	var txGroups []*txGroup
	txMap := make(map[string]*txGroup)
	nextAutoTx := 1

	for _, ev := range trace {
		var key string
		if ev.OpIndex > 0 {
			key = fmt.Sprintf("op_%d", ev.OpIndex)
		} else {
			key = fmt.Sprintf("w_%d_tx_%d", ev.WorkerID, nextAutoTx)
			if ev.Type == domain.EventCommit || ev.Type == domain.EventRollback {
				nextAutoTx++
			}
		}

		grp, exists := txMap[key]
		if !exists {
			grp = &txGroup{
				WorkerID: ev.WorkerID,
				OpIndex:  ev.OpIndex,
				OpName:   ev.OpName,
			}
			txMap[key] = grp
			txGroups = append(txGroups, grp)
		}
		grp.Events = append(grp.Events, ev)
	}

	// 2. Build Spans List
	var spans []OTLPSpan

	rootStartTime := baseTime
	rootEndTime := baseTime.Add(1 * time.Millisecond)

	if len(trace) > 0 {
		rootStartTime = baseTime.Add(trace[0].Timestamp)
		lastEv := trace[len(trace)-1]
		rootEndTime = baseTime.Add(lastEv.Timestamp).Add(100 * time.Microsecond)
	}

	scenarioName := spec.Name
	if scenarioName == "" {
		scenarioName = "chaossql_scenario"
	}

	rootSpan := OTLPSpan{
		TraceID:           traceID,
		SpanID:            rootSpanID,
		ParentSpanID:      "",
		Name:              scenarioName,
		Kind:              1, // SPAN_KIND_INTERNAL
		StartTimeUnixNano: fmt.Sprintf("%d", rootStartTime.UnixNano()),
		EndTimeUnixNano:   fmt.Sprintf("%d", rootEndTime.UnixNano()),
		Attributes: []OTLPAttribute{
			otlpStringAttr("scenario.name", spec.Name),
			otlpStringAttr("db.system", spec.Database.Driver),
			otlpStringAttr("service.name", "chaossql"),
		},
		Status: OTLPStatus{
			Code: 1, // STATUS_CODE_OK
		},
	}
	spans = append(spans, rootSpan)

	// 3. Create Transaction and Statement Spans
	for _, grp := range txGroups {
		if len(grp.Events) == 0 {
			continue
		}

		txSpanID := generateOTLPSpanID()
		txStartTime := baseTime.Add(grp.Events[0].Timestamp)

		// Calculate statement latencies and end time
		type stmtMeta struct {
			event     domain.TraceEvent
			startTime time.Time
			endTime   time.Time
			latencyUs int64
		}

		stmts := make([]stmtMeta, len(grp.Events))
		var txHasError bool
		var txErrMsg string

		for i, ev := range grp.Events {
			sTime := baseTime.Add(ev.Timestamp)
			var lat time.Duration
			if i+1 < len(grp.Events) {
				lat = grp.Events[i+1].Timestamp - ev.Timestamp
			}
			if lat <= 0 {
				lat = 100 * time.Microsecond
			}
			eTime := sTime.Add(lat)
			latUs := lat.Microseconds()
			if latUs <= 0 {
				latUs = 1
			}

			if ev.Type == domain.EventError || ev.Error != "" {
				txHasError = true
				if ev.Error != "" {
					txErrMsg = ev.Error
				} else {
					txErrMsg = "transaction step error"
				}
			}

			stmts[i] = stmtMeta{
				event:     ev,
				startTime: sTime,
				endTime:   eTime,
				latencyUs: latUs,
			}
		}

		txEndTime := stmts[len(stmts)-1].endTime
		if txEndTime.After(rootEndTime) {
			rootEndTime = txEndTime
			spans[0].EndTimeUnixNano = fmt.Sprintf("%d", rootEndTime.UnixNano())
		}

		txName := fmt.Sprintf("tx:%s #%d", grp.OpName, grp.OpIndex)
		if grp.OpName == "" && grp.OpIndex == 0 {
			txName = fmt.Sprintf("tx:worker_%d", grp.WorkerID)
		} else if grp.OpIndex == 0 {
			txName = fmt.Sprintf("tx:%s", grp.OpName)
		}

		txStatusCode := 1
		txStatusMsg := ""
		if txHasError {
			txStatusCode = 2
			txStatusMsg = txErrMsg
		}

		txSpan := OTLPSpan{
			TraceID:           traceID,
			SpanID:            txSpanID,
			ParentSpanID:      rootSpanID,
			Name:              txName,
			Kind:              1, // SPAN_KIND_INTERNAL
			StartTimeUnixNano: fmt.Sprintf("%d", txStartTime.UnixNano()),
			EndTimeUnixNano:   fmt.Sprintf("%d", txEndTime.UnixNano()),
			Attributes: []OTLPAttribute{
				otlpStringAttr("tx.op_name", grp.OpName),
				otlpIntAttr("tx.op_index", int64(grp.OpIndex)),
				otlpIntAttr("worker.id", int64(grp.WorkerID)),
			},
			Status: OTLPStatus{
				Code:    txStatusCode,
				Message: txStatusMsg,
			},
		}
		spans = append(spans, txSpan)

		// Add Statement Spans
		for _, sm := range stmts {
			stmtSpanID := generateOTLPSpanID()
			stmtName := sm.event.SQL
			if stmtName == "" {
				stmtName = string(sm.event.Type)
			}

			stmtAttrs := []OTLPAttribute{
				otlpStringAttr("db.statement", sm.event.SQL),
				otlpStringAttr("db.event_type", string(sm.event.Type)),
				otlpIntAttr("latency_us", sm.latencyUs),
			}

			stmtStatusCode := 1
			stmtStatusMsg := ""
			if sm.event.Type == domain.EventError || sm.event.Error != "" {
				stmtStatusCode = 2
				stmtStatusMsg = sm.event.Error
				if stmtStatusMsg == "" {
					stmtStatusMsg = "statement error"
				}
				stmtAttrs = append(stmtAttrs, otlpStringAttr("error", stmtStatusMsg))
			}

			stmtSpan := OTLPSpan{
				TraceID:           traceID,
				SpanID:            stmtSpanID,
				ParentSpanID:      txSpanID,
				Name:              stmtName,
				Kind:              3, // SPAN_KIND_CLIENT
				StartTimeUnixNano: fmt.Sprintf("%d", sm.startTime.UnixNano()),
				EndTimeUnixNano:   fmt.Sprintf("%d", sm.endTime.UnixNano()),
				Attributes:        stmtAttrs,
				Status: OTLPStatus{
					Code:    stmtStatusCode,
					Message: stmtStatusMsg,
				},
			}
			spans = append(spans, stmtSpan)
		}
	}

	// 4. Construct OTLP ResourceSpans
	payload := OTLPTracesData{
		ResourceSpans: []OTLPResourceSpan{
			{
				Resource: OTLPResource{
					Attributes: []OTLPAttribute{
						otlpStringAttr("service.name", "chaossql"),
						otlpStringAttr("db.system", spec.Database.Driver),
						otlpStringAttr("scenario.name", spec.Name),
					},
				},
				ScopeSpans: []OTLPScopeSpan{
					{
						Scope: OTLPInstrumentationScope{
							Name:    "github.com/bregaldahq/chaossql",
							Version: "1.0.0",
						},
						Spans: spans,
					},
				},
			},
		},
	}

	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal OTLP trace JSON: %w", err)
	}

	return string(out), nil
}

func generateOTLPTraceID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	if isAllZeroBytes(b[:]) {
		b[0] = 1
	}
	return hex.EncodeToString(b[:])
}

func generateOTLPSpanID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	if isAllZeroBytes(b[:]) {
		b[0] = 1
	}
	return hex.EncodeToString(b[:])
}

func isAllZeroBytes(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
