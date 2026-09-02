package reporter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
)

// GenerateMermaidSequence generates a valid Mermaid sequenceDiagram representing
// the chronological interleaving of database operations across workers, including conflict notes.
func GenerateMermaidSequence(trace domain.ExecutionTrace) string {
	var sb strings.Builder
	sb.WriteString("sequenceDiagram\n")
	sb.WriteString("    autonumber\n")

	// Collect unique workers
	workerMap := make(map[int]bool)
	for _, ev := range trace {
		if ev.WorkerID > 0 {
			workerMap[ev.WorkerID] = true
		}
	}

	var workerIDs []int
	for id := range workerMap {
		workerIDs = append(workerIDs, id)
	}
	sort.Ints(workerIDs)

	if len(workerIDs) == 0 {
		sb.WriteString("    participant DB as Database\n")
		sb.WriteString("    Note over DB: No trace events recorded\n")
		return sb.String()
	}

	for _, wid := range workerIDs {
		sb.WriteString(fmt.Sprintf("    participant W%d as Worker %d\n", wid, wid))
	}
	sb.WriteString("    participant DB as Database\n\n")

	// Write events
	for _, ev := range trace {
		sender := fmt.Sprintf("W%d", ev.WorkerID)
		if ev.WorkerID <= 0 {
			sender = "W1"
		}

		opTag := ""
		if ev.OpName != "" || ev.OpIndex > 0 {
			opTag = fmt.Sprintf("[%s #%d] ", ev.OpName, ev.OpIndex)
		}

		switch ev.Type {
		case domain.EventBegin:
			sb.WriteString(fmt.Sprintf("    %s->>DB: %sBEGIN\n", sender, opTag))
		case domain.EventExec, domain.EventSavepoint, domain.EventRollbackTo, domain.EventReleaseSavepoint:
			cleanSQL := sanitizeSQLForMermaid(ev.SQL)
			sb.WriteString(fmt.Sprintf("    %s->>DB: %s%s\n", sender, opTag, cleanSQL))
		case domain.EventCommit:
			sb.WriteString(fmt.Sprintf("    %s->>DB: %sCOMMIT\n", sender, opTag))
		case domain.EventRollback:
			sb.WriteString(fmt.Sprintf("    %s->>DB: %sROLLBACK\n", sender, opTag))
		case domain.EventError:
			cleanErr := sanitizeSQLForMermaid(ev.Error)
			sb.WriteString(fmt.Sprintf("    %s--xDB: %sERROR: %s\n", sender, opTag, cleanErr))
			sb.WriteString(fmt.Sprintf("    Note over %s,DB: Error: %s\n", sender, cleanErr))
		}
	}

	// Adya Conflict / Anomaly Analysis for Conflict Notes
	if len(trace) > 0 {
		graph := analyzer.BuildGraph(trace)
		cycles := analyzer.FindCycles(graph)
		if len(cycles) > 0 {
			sb.WriteString("\n    %% Detected Dependency Conflicts and Anomalies\n")
			for i, cycle := range cycles {
				anomaly := analyzer.ClassifyCycle(cycle)
				var edgeStrs []string
				for _, edge := range cycle {
					edgeStrs = append(edgeStrs, fmt.Sprintf("%s -[%s on %s]-> %s", edge.From, edge.Type, edge.Item, edge.To))
				}
				cycleDesc := strings.Join(edgeStrs, ", ")
				sb.WriteString(fmt.Sprintf("    Note over DB: [Anomaly #%d] %s Detected!\n", i+1, anomaly))
				sb.WriteString(fmt.Sprintf("    Note over DB: Cycle: %s\n", cycleDesc))
			}
		}
	}

	return sb.String()
}

func sanitizeSQLForMermaid(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	s = strings.TrimSpace(s)
	return s
}
