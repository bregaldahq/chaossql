package reporter

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
)

// GenerateEmbeddedTraceViewerHTML renders a calm, clean, editorial, and sophisticated
// single-page application for inspecting concurrency traces and Adya dependency graphs.
// It complies with the Studio Bregalda design system using canonical tokens:
// Deep Ink (#120E1F), Warm Cream (#FCFBF8), Bregalda Purple (#4B2E83), Signal Yellow (#F5C400),
// and Ship Green (#22C55E), with Inter and JetBrains Mono typography and zero external dependencies.
func GenerateEmbeddedTraceViewerHTML(
	trace domain.ExecutionTrace,
	spec domain.Spec,
	graph *analyzer.AdyaGraph,
	shrink *domain.ShrinkResult,
	invResults []domain.InvariantResult,
) string {
	if graph == nil && len(trace) > 0 {
		graph = analyzer.BuildGraph(trace)
	}
	if graph == nil {
		graph = analyzer.NewAdyaGraph()
	}

	cycles := analyzer.FindCycles(graph)

	hasAnomaly := false
	anomalyType := domain.AnomalyUnknown

	for _, inv := range invResults {
		if !inv.Passed || inv.Error != nil {
			hasAnomaly = true
			break
		}
	}

	for _, c := range cycles {
		cls := analyzer.ClassifyCycle(c)
		if cls != domain.AnomalyUnknown {
			anomalyType = cls
			hasAnomaly = true
			break
		}
	}
	if anomalyType == domain.AnomalyUnknown && len(cycles) > 0 {
		anomalyType = analyzer.ClassifyCycle(cycles[0])
		hasAnomaly = true
	}

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n")
	sb.WriteString("<html lang=\"en\">\n")
	sb.WriteString("<head>\n")
	sb.WriteString("  <meta charset=\"UTF-8\">\n")
	sb.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	sb.WriteString(fmt.Sprintf("  <title>ChaosSQL — %s — Trace Viewer</title>\n", html.EscapeString(spec.Name)))
	sb.WriteString("  <style>\n")
	sb.WriteString(embeddedUICentralCSS())
	sb.WriteString("  </style>\n")
	sb.WriteString("</head>\n")
	sb.WriteString("<body>\n")

	sb.WriteString("  <div class=\"app-container\">\n")

	// 1. Header & Scenario Summary
	sb.WriteString(renderUIHeader(spec, trace, hasAnomaly, anomalyType, len(cycles), invResults))

	// 2. Timeline Swimlane Container
	sb.WriteString(renderUITimelineSwimlane(trace))

	// 3. Grid for Adya Graph & Delta Debugging Card
	sb.WriteString("    <div class=\"two-column-grid\">\n")
	sb.WriteString(renderUIAdyaGraph(graph, cycles))
	sb.WriteString(renderUIDeltaDebuggingCard(shrink))
	sb.WriteString("    </div>\n\n")

	// 4. Statement Detail Inspector Table
	sb.WriteString(renderUIStatementInspector(trace))

	// Footer
	sb.WriteString("    <footer class=\"ui-footer\">\n")
	sb.WriteString("      <div class=\"footer-brand\">ChaosSQL Trace Visualizer • Studio Bregalda Design System</div>\n")
	sb.WriteString("      <div class=\"footer-meta\">Calm, deterministic systems engineering verification • Zero-external-dependency SPA</div>\n")
	sb.WriteString("    </footer>\n")

	sb.WriteString("  </div>\n")

	sb.WriteString("  <script>\n")
	sb.WriteString(embeddedUIClientJS(trace))
	sb.WriteString("  </script>\n")

	sb.WriteString("</body>\n")
	sb.WriteString("</html>\n")

	return sb.String()
}

func renderUIHeader(
	spec domain.Spec,
	trace domain.ExecutionTrace,
	hasAnomaly bool,
	anomalyType domain.AnomalyType,
	cycleCount int,
	invResults []domain.InvariantResult,
) string {
	var sb strings.Builder
	sb.WriteString("    <header class=\"scenario-summary header\">\n")
	sb.WriteString("      <div class=\"brand-bar\">\n")
	sb.WriteString("        <div class=\"brand-title-group\">\n")
	sb.WriteString("          <span class=\"brand-symbol\">❖</span>\n")
	sb.WriteString("          <h1 class=\"brand-name\">ChaosSQL</h1>\n")
	sb.WriteString("          <span class=\"brand-tag\">STUDIO BREGALDA</span>\n")
	sb.WriteString("          <span class=\"brand-version\">v1.2</span>\n")
	sb.WriteString("        </div>\n")

	// Status badge
	sb.WriteString("        <div class=\"status-badge-group\">\n")
	if hasAnomaly {
		sb.WriteString("          <div class=\"badge badge-danger\">")
		sb.WriteString("<span class=\"badge-dot dot-danger\"></span>")
		sb.WriteString("<span>ISOLATION ANOMALY DETECTED</span>")
		sb.WriteString("</div>\n")
		if anomalyType != domain.AnomalyUnknown && anomalyType != "" {
			sb.WriteString(fmt.Sprintf("          <div class=\"badge badge-warning\">%s</div>\n", html.EscapeString(string(anomalyType))))
		}
	} else {
		sb.WriteString("          <div class=\"badge badge-success\">")
		sb.WriteString("<span class=\"badge-dot dot-success\"></span>")
		sb.WriteString("<span>INVARIANTS SATISFIED</span>")
		sb.WriteString("</div>\n")
	}
	sb.WriteString("        </div>\n")
	sb.WriteString("      </div>\n\n")

	specName := spec.Name
	if specName == "" {
		specName = "trace_playback"
	}
	driver := spec.Database.Driver
	if driver == "" {
		driver = "sqlite"
	}

	var totalDuration time.Duration
	if len(trace) > 0 {
		totalDuration = trace[len(trace)-1].Timestamp - trace[0].Timestamp
	}

	workersCount := spec.Engine.Workers
	if workersCount == 0 {
		maxW := 0
		for _, e := range trace {
			if e.WorkerID > maxW {
				maxW = e.WorkerID
			}
		}
		workersCount = maxW
		if workersCount == 0 {
			workersCount = 1
		}
	}

	sb.WriteString("      <div class=\"meta-cards-row\">\n")
	sb.WriteString("        <div class=\"meta-card\">\n")
	sb.WriteString("          <div class=\"meta-card-label\">SCENARIO SPEC</div>\n")
	sb.WriteString(fmt.Sprintf("          <div class=\"meta-card-value font-mono\">%s</div>\n", html.EscapeString(specName)))
	sb.WriteString("        </div>\n")

	sb.WriteString("        <div class=\"meta-card\">\n")
	sb.WriteString("          <div class=\"meta-card-label\">DATABASE ENGINE</div>\n")
	sb.WriteString(fmt.Sprintf("          <div class=\"meta-card-value font-mono\">%s</div>\n", html.EscapeString(driver)))
	sb.WriteString("        </div>\n")

	sb.WriteString("        <div class=\"meta-card\">\n")
	sb.WriteString("          <div class=\"meta-card-label\">CONCURRENCY</div>\n")
	sb.WriteString(fmt.Sprintf("          <div class=\"meta-card-value font-mono\">%d Workers</div>\n", workersCount))
	sb.WriteString("        </div>\n")

	sb.WriteString("        <div class=\"meta-card\">\n")
	sb.WriteString("          <div class=\"meta-card-label\">TRACE EVENTS</div>\n")
	sb.WriteString(fmt.Sprintf("          <div class=\"meta-card-value font-mono\">%d Ops</div>\n", len(trace)))
	sb.WriteString("        </div>\n")

	sb.WriteString("        <div class=\"meta-card\">\n")
	sb.WriteString("          <div class=\"meta-card-label\">CYCLE COUNT</div>\n")
	sb.WriteString(fmt.Sprintf("          <div class=\"meta-card-value font-mono\">%d Detected</div>\n", cycleCount))
	sb.WriteString("        </div>\n")

	sb.WriteString("        <div class=\"meta-card\">\n")
	sb.WriteString("          <div class=\"meta-card-label\">TRACE SPAN</div>\n")
	sb.WriteString(fmt.Sprintf("          <div class=\"meta-card-value font-mono\">%s</div>\n", totalDuration.Round(time.Microsecond)))
	sb.WriteString("        </div>\n")
	sb.WriteString("      </div>\n")

	if len(invResults) > 0 {
		var violated []domain.InvariantResult
		for _, inv := range invResults {
			if !inv.Passed || inv.Error != nil {
				violated = append(violated, inv)
			}
		}
		if len(violated) > 0 {
			sb.WriteString("      <div class=\"invariant-alert-strip\">\n")
			sb.WriteString("        <span class=\"alert-icon\">⚠</span>\n")
			sb.WriteString("        <div class=\"alert-text\">\n")
			for _, v := range violated {
				valJSON, _ := json.Marshal(v.ActualValues)
				sb.WriteString(fmt.Sprintf("<div><strong>%s:</strong> Violated expression <code>%s</code> (Actual state: <code>%s</code>)</div>",
					html.EscapeString(v.Name), html.EscapeString(v.Expression), html.EscapeString(string(valJSON))))
			}
			sb.WriteString("        </div>\n")
			sb.WriteString("      </div>\n")
		}
	}

	sb.WriteString("    </header>\n\n")
	return sb.String()
}

func renderUITimelineSwimlane(trace domain.ExecutionTrace) string {
	var sb strings.Builder
	sb.WriteString("    <section class=\"card timeline-container timeline-swimlane\">\n")
	sb.WriteString("      <div class=\"card-header-bar\">\n")
	sb.WriteString("        <div>\n")
	sb.WriteString("          <h2 class=\"card-title\">Transaction Timeline Swimlane</h2>\n")
	sb.WriteString("          <div class=\"card-subtitle\">Gantt timeline mapping concurrent worker goroutines and transaction steps</div>\n")
	sb.WriteString("        </div>\n")
	sb.WriteString("        <div class=\"timeline-legend\">\n")
	sb.WriteString("          <span class=\"legend-pill pill-begin\">BEGIN</span>\n")
	sb.WriteString("          <span class=\"legend-pill pill-exec\">EXEC</span>\n")
	sb.WriteString("          <span class=\"legend-pill pill-commit\">COMMIT</span>\n")
	sb.WriteString("          <span class=\"legend-pill pill-rollback\">ROLLBACK</span>\n")
	sb.WriteString("          <span class=\"legend-pill pill-savepoint\">SAVEPOINT</span>\n")
	sb.WriteString("        </div>\n")
	sb.WriteString("      </div>\n\n")

	if len(trace) == 0 {
		sb.WriteString("      <div class=\"empty-state-box\">\n")
		sb.WriteString("        <div class=\"empty-state-icon\">◫</div>\n")
		sb.WriteString("        <div class=\"empty-state-title\">No Concurrency Events Recorded</div>\n")
		sb.WriteString("        <div class=\"empty-state-desc\">The execution trace is empty. Run a chaos scenario to populate timeline events.</div>\n")
		sb.WriteString("      </div>\n")
		sb.WriteString("    </section>\n\n")
		return sb.String()
	}

	eventsByWorker := make(map[int][]int)
	var workerIDs []int
	for idx, event := range trace {
		w := event.WorkerID
		if len(eventsByWorker[w]) == 0 {
			workerIDs = append(workerIDs, w)
		}
		eventsByWorker[w] = append(eventsByWorker[w], idx)
	}
	sort.Ints(workerIDs)

	sb.WriteString("      <div class=\"swimlane-board\">\n")
	for _, w := range workerIDs {
		indices := eventsByWorker[w]
		sb.WriteString("        <div class=\"swimlane-row\">\n")
		sb.WriteString(fmt.Sprintf("          <div class=\"swimlane-worker-label\">\n            <div class=\"worker-badge\">Worker %d</div>\n            <div class=\"worker-event-count font-mono\">%d ops</div>\n          </div>\n", w, len(indices)))
		sb.WriteString("          <div class=\"swimlane-track\">\n")

		for _, idx := range indices {
			ev := trace[idx]
			pillClass := "pill-exec"
			switch ev.Type {
			case domain.EventBegin:
				pillClass = "pill-begin"
			case domain.EventCommit:
				pillClass = "pill-commit"
			case domain.EventRollback, domain.EventError:
				pillClass = "pill-rollback"
			case domain.EventSavepoint, domain.EventRollbackTo:
				pillClass = "pill-savepoint"
			}

			timeUs := ev.Timestamp.Microseconds()
			timeStr := fmt.Sprintf("+%dµs", timeUs)
			if timeUs >= 1000 {
				timeStr = fmt.Sprintf("+%.1fms", float64(timeUs)/1000.0)
			}

			cleanSQL := strings.TrimSpace(ev.SQL)
			if len(cleanSQL) > 40 {
				cleanSQL = cleanSQL[:37] + "..."
			}

			txLabel := fmt.Sprintf("T%d-%d", ev.WorkerID, ev.OpIndex)

			sb.WriteString(fmt.Sprintf(
				"            <div class=\"timeline-node %s\" data-idx=\"%d\" onclick=\"inspectEvent(%d)\" title=\"%s: %s (%s)\">\n"+
					"              <span class=\"node-tx\">%s</span>\n"+
					"              <span class=\"node-type\">%s</span>\n"+
					"              <span class=\"node-time font-mono\">%s</span>\n"+
					"              <span class=\"node-sql font-mono\">%s</span>\n"+
					"            </div>\n",
				pillClass, idx, idx, html.EscapeString(txLabel), html.EscapeString(ev.SQL), timeStr,
				html.EscapeString(txLabel), html.EscapeString(string(ev.Type)), html.EscapeString(timeStr), html.EscapeString(cleanSQL),
			))
		}

		sb.WriteString("          </div>\n")
		sb.WriteString("        </div>\n")
	}
	sb.WriteString("      </div>\n")
	sb.WriteString("    </section>\n\n")

	return sb.String()
}

type uiGraphPoint struct {
	x float64
	y float64
}

func renderUIAdyaGraph(graph *analyzer.AdyaGraph, cycles []analyzer.Cycle) string {
	var sb strings.Builder
	sb.WriteString("      <section class=\"card adya-graph-card adya-graph\">\n")
	sb.WriteString("        <div class=\"card-header-bar\">\n")
	sb.WriteString("          <div>\n")
	sb.WriteString("            <h2 class=\"card-title\">Adya Dependency Graph</h2>\n")
	sb.WriteString("            <div class=\"card-subtitle\">Static directed conflict cycle analysis (WR, WW, RW)</div>\n")
	sb.WriteString("          </div>\n")
	sb.WriteString("          <div class=\"graph-legend-compact\">\n")
	sb.WriteString("            <span class=\"legend-dot dot-wr\"></span><span class=\"legend-label\">WR</span>\n")
	sb.WriteString("            <span class=\"legend-dot dot-ww\"></span><span class=\"legend-label\">WW</span>\n")
	sb.WriteString("            <span class=\"legend-dot dot-rw\"></span><span class=\"legend-label\">RW</span>\n")
	sb.WriteString("            <span class=\"legend-dot dot-cycle\"></span><span class=\"legend-label\">Cycle</span>\n")
	sb.WriteString("          </div>\n")
	sb.WriteString("        </div>\n\n")

	if graph == nil || len(graph.Nodes) == 0 {
		sb.WriteString("        <div class=\"empty-state-box\">\n")
		sb.WriteString("          <div class=\"empty-state-icon\">☊</div>\n")
		sb.WriteString("          <div class=\"empty-state-title\">No Inter-Transaction Dependencies</div>\n")
		sb.WriteString("          <div class=\"empty-state-desc\">No conflicting reads or writes were detected across concurrent transactions.</div>\n")
		sb.WriteString("        </div>\n")
		sb.WriteString("      </section>\n")
		return sb.String()
	}

	cycleEdgeMap := make(map[string]bool)
	cycleNodeMap := make(map[string]bool)
	for _, cycle := range cycles {
		for _, e := range cycle {
			key := fmt.Sprintf("%s->%s:%s:%s", e.From, e.To, e.Type, e.Item)
			cycleEdgeMap[key] = true
			cycleNodeMap[e.From] = true
			cycleNodeMap[e.To] = true
		}
	}

	var sortedNodes []string
	for n := range graph.Nodes {
		sortedNodes = append(sortedNodes, n)
	}
	sort.Strings(sortedNodes)

	canvasWidth := 560.0
	canvasHeight := 380.0
	cx := canvasWidth / 2.0
	cy := canvasHeight / 2.0
	radius := 125.0
	if len(sortedNodes) > 4 {
		radius = 140.0
	}

	nodePositions := make(map[string]uiGraphPoint)
	numNodes := len(sortedNodes)
	for i, node := range sortedNodes {
		if numNodes == 1 {
			nodePositions[node] = uiGraphPoint{x: cx, y: cy}
		} else {
			angle := (2.0 * math.Pi * float64(i) / float64(numNodes)) - (math.Pi / 2.0)
			nodePositions[node] = uiGraphPoint{
				x: cx + radius*math.Cos(angle),
				y: cy + radius*math.Sin(angle),
			}
		}
	}

	sb.WriteString("        <div class=\"graph-canvas-container\">\n")
	sb.WriteString(fmt.Sprintf("          <svg class=\"adya-svg\" viewBox=\"0 0 %.0f %.0f\" width=\"100%%\" height=\"380\" xmlns=\"http://www.w3.org/2000/svg\">\n", canvasWidth, canvasHeight))
	sb.WriteString(`            <defs>
              <marker id="ui-arrow-wr" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
                <path d="M 0 1 L 10 5 L 0 9 z" fill="#F5C400" />
              </marker>
              <marker id="ui-arrow-ww" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
                <path d="M 0 1 L 10 5 L 0 9 z" fill="#6A44B0" />
              </marker>
              <marker id="ui-arrow-rw" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
                <path d="M 0 1 L 10 5 L 0 9 z" fill="#FCFBF8" />
              </marker>
              <marker id="ui-arrow-cycle" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
                <path d="M 0 1 L 10 5 L 0 9 z" fill="#F5C400" />
              </marker>
            </defs>
`)

	for _, fromNode := range sortedNodes {
		edges := graph.Edges[fromNode]
		for _, edge := range edges {
			p1, ok1 := nodePositions[edge.From]
			p2, ok2 := nodePositions[edge.To]
			if !ok1 || !ok2 {
				continue
			}

			key := fmt.Sprintf("%s->%s:%s:%s", edge.From, edge.To, edge.Type, edge.Item)
			isCycle := cycleEdgeMap[key]

			strokeColor := "#F5C400"
			markerId := "ui-arrow-wr"
			switch edge.Type {
			case analyzer.DepWW:
				strokeColor = "#6A44B0"
				markerId = "ui-arrow-ww"
			case analyzer.DepRW:
				strokeColor = "#FCFBF8"
				markerId = "ui-arrow-rw"
			}

			strokeWidth := "1.8"
			edgeClass := "graph-edge"
			if isCycle {
				strokeColor = "#F5C400"
				markerId = "ui-arrow-cycle"
				strokeWidth = "2.8"
				edgeClass = "graph-edge cycle cycle-edge"
			}

			if edge.From == edge.To {
				sx := p1.x
				sy := p1.y - 18
				ex := p1.x + 24
				ey := p1.y - 16
				sb.WriteString(fmt.Sprintf("            <path d=\"M %.1f %.1f C %.1f %.1f, %.1f %.1f, %.1f %.1f\" fill=\"none\" stroke=\"%s\" stroke-width=\"%s\" marker-end=\"url(#%s)\" class=\"%s\" />\n",
					sx, sy, sx-20, sy-35, ex+20, ey-35, ex, ey, strokeColor, strokeWidth, markerId, edgeClass))
			} else {
				dx := p2.x - p1.x
				dy := p2.y - p1.y
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist < 1.0 {
					dist = 1.0
				}
				ux := dx / dist
				uy := dy / dist
				px := -uy
				py := ux

				nodeRadius := 34.0
				sx := p1.x + ux*nodeRadius
				sy := p1.y + uy*nodeRadius
				ex := p2.x - ux*(nodeRadius+6)
				ey := p2.y - uy*(nodeRadius+6)

				mx := (sx + ex) / 2.0
				my := (sy + ey) / 2.0
				curveOffset := 22.0
				qx := mx + px*curveOffset
				qy := my + py*curveOffset

				sb.WriteString(fmt.Sprintf("            <path d=\"M %.1f %.1f Q %.1f %.1f %.1f %.1f\" fill=\"none\" stroke=\"%s\" stroke-width=\"%s\" marker-end=\"url(#%s)\" class=\"%s\" />\n",
					sx, sy, qx, qy, ex, ey, strokeColor, strokeWidth, markerId, edgeClass))

				labelStr := string(edge.Type)
				if edge.Item != "" {
					labelStr += " " + edge.Item
				}
				labelWidth := float64(len(labelStr))*5.8 + 10.0
				lx := qx - labelWidth/2.0
				ly := qy - 8.0

				sb.WriteString(fmt.Sprintf("            <rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"16\" rx=\"3\" fill=\"#181328\" stroke=\"%s\" stroke-width=\"1\" opacity=\"0.95\" />\n",
					lx, ly, labelWidth, strokeColor))
				sb.WriteString(fmt.Sprintf("            <text x=\"%.1f\" y=\"%.1f\" fill=\"%s\" font-family=\"var(--font-mono)\" font-size=\"9\" font-weight=\"600\" text-anchor=\"middle\">%s</text>\n",
					qx, qy+3.5, strokeColor, html.EscapeString(labelStr)))
			}
		}
	}

	for _, node := range sortedNodes {
		pos := nodePositions[node]
		isCycleNode := cycleNodeMap[node]

		nodeStroke := "#4B2E83"
		nodeFill := "#181328"
		nodeClass := "graph-node"
		if isCycleNode {
			nodeStroke = "#F5C400"
			nodeFill = "#2A2140"
			nodeClass = "graph-node cycle cycle-node"
		}

		boxW := 68.0
		boxH := 32.0
		x := pos.x - boxW/2.0
		y := pos.y - boxH/2.0

		sb.WriteString(fmt.Sprintf("            <rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\" rx=\"6\" fill=\"%s\" stroke=\"%s\" stroke-width=\"2\" class=\"%s\" />\n",
			x, y, boxW, boxH, nodeFill, nodeStroke, nodeClass))
		sb.WriteString(fmt.Sprintf("            <text x=\"%.1f\" y=\"%.1f\" fill=\"#FCFBF8\" font-family=\"var(--font-mono)\" font-size=\"12\" font-weight=\"700\" text-anchor=\"middle\">%s</text>\n",
			pos.x, pos.y+4.0, html.EscapeString(node)))
	}

	sb.WriteString("          </svg>\n")
	sb.WriteString("        </div>\n")
	sb.WriteString("      </section>\n\n")

	return sb.String()
}

func renderUIDeltaDebuggingCard(shrink *domain.ShrinkResult) string {
	var sb strings.Builder
	sb.WriteString("      <section class=\"card delta-debugging shrink-card\">\n")
	sb.WriteString("        <div class=\"card-header-bar\">\n")
	sb.WriteString("          <div>\n")
	sb.WriteString("            <h2 class=\"card-title\">Delta-Debugging Minimization</h2>\n")
	sb.WriteString("            <div class=\"card-subtitle\">Causal schedule reduction via Zeller's ddmin algorithm</div>\n")
	sb.WriteString("          </div>\n")
	sb.WriteString("          <span class=\"card-tag-pill\">ddmin</span>\n")
	sb.WriteString("        </div>\n\n")

	if shrink == nil {
		sb.WriteString("        <div class=\"empty-state-box\">\n")
		sb.WriteString("          <div class=\"empty-state-icon\">✂</div>\n")
		sb.WriteString("          <div class=\"empty-state-title\">Full Trace Intact</div>\n")
		sb.WriteString("          <div class=\"empty-state-desc\">Delta-Debugging was not triggered or no invariant violations required causal reduction.</div>\n")
		sb.WriteString("        </div>\n")
		sb.WriteString("      </section>\n")
		return sb.String()
	}

	ratioPercent := shrink.ReductionRatio * 100.0

	sb.WriteString("        <div class=\"shrink-metrics-grid\">\n")
	sb.WriteString("          <div class=\"shrink-metric-box\">\n")
	sb.WriteString("            <div class=\"metric-label\">ORIGINAL SCHEDULE</div>\n")
	sb.WriteString(fmt.Sprintf("            <div class=\"metric-val font-mono\">%d ops</div>\n", shrink.OriginalSize))
	sb.WriteString("          </div>\n")

	sb.WriteString("          <div class=\"shrink-metric-box\">\n")
	sb.WriteString("            <div class=\"metric-label\">1-MINIMAL SUBSET</div>\n")
	sb.WriteString(fmt.Sprintf("            <div class=\"metric-val val-reduced font-mono\">%d ops</div>\n", shrink.ReducedSize))
	sb.WriteString("          </div>\n")

	sb.WriteString("          <div class=\"shrink-metric-box metric-box-highlight\">\n")
	sb.WriteString("            <div class=\"metric-label\">NOISE REDUCTION</div>\n")
	sb.WriteString(fmt.Sprintf("            <div class=\"metric-val val-highlight font-mono\">-%.1f%%</div>\n", ratioPercent))
	sb.WriteString("          </div>\n")

	sb.WriteString("          <div class=\"shrink-metric-box\">\n")
	sb.WriteString("            <div class=\"metric-label\">DDMIN SEARCH</div>\n")
	sb.WriteString(fmt.Sprintf("            <div class=\"metric-val font-mono\">%d iters (%s)</div>\n", shrink.Iterations, shrink.Duration.Round(time.Millisecond)))
	sb.WriteString("          </div>\n")
	sb.WriteString("        </div>\n\n")

	sb.WriteString("        <div class=\"minimal-ops-section\">\n")
	sb.WriteString("          <div class=\"section-label-bar\">Minimal Causal Counterexample Operations:</div>\n")
	sb.WriteString("          <div class=\"minimal-ops-list\">\n")
	for i, op := range shrink.MinimalOps {
		sb.WriteString("            <div class=\"minimal-op-card\">\n")
		sb.WriteString("              <div class=\"op-header\">\n")
		sb.WriteString(fmt.Sprintf("                <span class=\"op-badge font-mono\">#%d • Op %d</span>\n", i+1, op.ID))
		sb.WriteString(fmt.Sprintf("                <span class=\"op-name font-mono\">%s</span>\n", html.EscapeString(op.Name)))
		sb.WriteString("              </div>\n")
		sb.WriteString("              <div class=\"op-steps\">\n")
		for sIdx, st := range op.Steps {
			sb.WriteString(fmt.Sprintf("                <div class=\"op-step-line font-mono\"><span class=\"step-num\">%d.</span> %s</div>\n",
				sIdx+1, html.EscapeString(st.SQL)))
		}
		sb.WriteString("              </div>\n")
		sb.WriteString("            </div>\n")
	}
	sb.WriteString("          </div>\n")
	sb.WriteString("        </div>\n")

	sb.WriteString("      </section>\n\n")
	return sb.String()
}

func renderUIStatementInspector(trace domain.ExecutionTrace) string {
	var sb strings.Builder
	sb.WriteString("    <section class=\"card statement-inspector\">\n")
	sb.WriteString("      <div class=\"card-header-bar\">\n")
	sb.WriteString("        <div>\n")
	sb.WriteString("          <h2 class=\"card-title\">Statement Detail Inspector</h2>\n")
	sb.WriteString("          <div class=\"card-subtitle\">Live forensic statement analysis with filtering and execution inspection</div>\n")
	sb.WriteString("        </div>\n")
	sb.WriteString("        <div class=\"search-controls\">\n")
	sb.WriteString("          <input type=\"text\" id=\"statement-search\" class=\"search-input font-mono\" placeholder=\"Filter SQL, worker, type...\" oninput=\"filterStatements()\">\n")
	sb.WriteString("          <select id=\"worker-filter\" class=\"filter-select font-mono\" onchange=\"filterStatements()\">\n")
	sb.WriteString("            <option value=\"\">All Workers</option>\n")

	workerMap := make(map[int]bool)
	for _, ev := range trace {
		workerMap[ev.WorkerID] = true
	}
	var workers []int
	for w := range workerMap {
		workers = append(workers, w)
	}
	sort.Ints(workers)
	for _, w := range workers {
		sb.WriteString(fmt.Sprintf("            <option value=\"W%d\">Worker %d</option>\n", w, w))
	}
	sb.WriteString("          </select>\n")
	sb.WriteString("        </div>\n")
	sb.WriteString("      </div>\n\n")

	if len(trace) == 0 {
		sb.WriteString("      <div class=\"empty-state-box\">\n")
		sb.WriteString("        <div class=\"empty-state-title\">No Statements Executed</div>\n")
		sb.WriteString("        <div class=\"empty-state-desc\">Trace does not contain any logged database statements.</div>\n")
		sb.WriteString("      </div>\n")
		sb.WriteString("    </section>\n\n")
		return sb.String()
	}

	sb.WriteString("      <div class=\"table-responsive-wrapper\">\n")
	sb.WriteString("        <table class=\"inspector-table font-mono\" id=\"statements-table\">\n")
	sb.WriteString("          <thead>\n")
	sb.WriteString("            <tr>\n")
	sb.WriteString("              <th style=\"width: 60px;\">#</th>\n")
	sb.WriteString("              <th style=\"width: 110px;\">TIME</th>\n")
	sb.WriteString("              <th style=\"width: 90px;\">WORKER</th>\n")
	sb.WriteString("              <th style=\"width: 80px;\">TX ID</th>\n")
	sb.WriteString("              <th style=\"width: 100px;\">TYPE</th>\n")
	sb.WriteString("              <th>SQL STATEMENT / INSTRUCTION</th>\n")
	sb.WriteString("              <th style=\"width: 120px;\">STATUS</th>\n")
	sb.WriteString("            </tr>\n")
	sb.WriteString("          </thead>\n")
	sb.WriteString("          <tbody>\n")

	for idx, ev := range trace {
		timeUs := ev.Timestamp.Microseconds()
		timeStr := fmt.Sprintf("+%dµs", timeUs)
		if timeUs >= 1000 {
			timeStr = fmt.Sprintf("+%.2fms", float64(timeUs)/1000.0)
		}

		typeBadgeClass := "type-exec"
		switch ev.Type {
		case domain.EventBegin:
			typeBadgeClass = "type-begin"
		case domain.EventCommit:
			typeBadgeClass = "type-commit"
		case domain.EventRollback:
			typeBadgeClass = "type-rollback"
		case domain.EventSavepoint, domain.EventRollbackTo:
			typeBadgeClass = "type-savepoint"
		}

		statusText := "OK"
		statusClass := "status-ok"
		if ev.Error != "" {
			statusText = "ERROR"
			statusClass = "status-err"
		} else if ev.Type == domain.EventRollback {
			statusText = "ABORT"
			statusClass = "status-abort"
		}

		txID := fmt.Sprintf("T%d-%d", ev.WorkerID, ev.OpIndex)

		sb.WriteString(fmt.Sprintf(
			"            <tr class=\"statement-row\" id=\"row-%d\" onclick=\"inspectEvent(%d)\" data-worker=\"W%d\" data-type=\"%s\">\n"+
				"              <td class=\"col-idx\">%d</td>\n"+
				"              <td class=\"col-time\">%s</td>\n"+
				"              <td class=\"col-worker\"><span class=\"worker-tag\">W%d</span></td>\n"+
				"              <td class=\"col-tx\">%s</td>\n"+
				"              <td class=\"col-type\"><span class=\"type-chip %s\">%s</span></td>\n"+
				"              <td class=\"col-sql\">%s</td>\n"+
				"              <td class=\"col-status\"><span class=\"status-tag %s\">%s</span></td>\n"+
				"            </tr>\n",
			idx, idx, ev.WorkerID, html.EscapeString(string(ev.Type)),
			idx+1, html.EscapeString(timeStr), ev.WorkerID, html.EscapeString(txID),
			typeBadgeClass, html.EscapeString(string(ev.Type)), html.EscapeString(ev.SQL),
			statusClass, html.EscapeString(statusText),
		))
	}

	sb.WriteString("          </tbody>\n")
	sb.WriteString("        </table>\n")
	sb.WriteString("      </div>\n\n")

	sb.WriteString("      <div id=\"statement-inspector-drawer\" class=\"inspector-detail-drawer\">\n")
	sb.WriteString("        <div class=\"drawer-header\">\n")
	sb.WriteString("          <span class=\"drawer-title font-mono\">Selected Event Inspector</span>\n")
	sb.WriteString("          <span id=\"drawer-event-id\" class=\"drawer-badge font-mono\">Click any event or row above</span>\n")
	sb.WriteString("        </div>\n")
	sb.WriteString("        <pre id=\"drawer-sql\" class=\"drawer-sql-view font-mono\">-- Full SQL statement and execution attributes will be displayed here --</pre>\n")
	sb.WriteString("      </div>\n")

	sb.WriteString("    </section>\n\n")
	return sb.String()
}
