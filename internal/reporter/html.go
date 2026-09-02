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

// GenerateStandaloneHTMLReport renders a comprehensive, standalone dark-mode HTML report
// visualizing concurrency traces, Adya serialization graphs, invariant audit tables,
// and Delta-Debugging minimal schedules.
func GenerateStandaloneHTMLReport(
	trace domain.ExecutionTrace,
	spec domain.Spec,
	graph *analyzer.AdyaGraph,
	shrinkResult *domain.ShrinkResult,
	invResults []domain.InvariantResult,
) string {
	if graph == nil && len(trace) > 0 {
		graph = analyzer.BuildGraph(trace)
	}
	if graph == nil {
		graph = analyzer.NewAdyaGraph()
	}

	cycles := analyzer.FindCycles(graph)

	// Determine overall anomaly status
	hasAnomaly := false
	anomalyType := domain.AnomalyUnknown

	for _, inv := range invResults {
		if !inv.Passed || inv.Error != nil {
			hasAnomaly = true
			break
		}
	}

	if len(cycles) > 0 {
		hasAnomaly = true
		for _, c := range cycles {
			cls := analyzer.ClassifyCycle(c)
			if cls == domain.AnomalyG0DirtyWrite {
				anomalyType = domain.AnomalyG0DirtyWrite
				break
			}
			if cls == domain.AnomalyWriteSkew {
				anomalyType = domain.AnomalyWriteSkew
				break
			}
			if cls == domain.AnomalyA5AReadSkew {
				anomalyType = domain.AnomalyA5AReadSkew
				break
			}
			if cls == domain.AnomalyLostUpdate {
				anomalyType = domain.AnomalyLostUpdate
			}
		}
		if anomalyType == domain.AnomalyUnknown && len(cycles) > 0 {
			anomalyType = analyzer.ClassifyCycle(cycles[0])
		}
	}

	var sb strings.Builder

	// 1. HTML Header & Document Setup
	sb.WriteString("<!DOCTYPE html>\n")
	sb.WriteString("<html lang=\"en\">\n<head>\n")
	sb.WriteString("  <meta charset=\"UTF-8\" />\n")
	sb.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\" />\n")
	sb.WriteString(fmt.Sprintf("  <title>ChaosSQL Report: %s</title>\n", html.EscapeString(spec.Name)))
	sb.WriteString("  <link rel=\"preconnect\" href=\"https://fonts.googleapis.com\">\n")
	sb.WriteString("  <link rel=\"preconnect\" href=\"https://fonts.gstatic.com\" crossorigin>\n")
	sb.WriteString("  <link href=\"https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600;700&display=swap\" rel=\"stylesheet\">\n")

	// 2. Embedded Dark Mode CSS with Purple/Cyan Neon Accents
	sb.WriteString("  <style>\n")
	sb.WriteString(embeddedCSS())
	sb.WriteString("  </style>\n")
	sb.WriteString("</head>\n<body>\n")

	// 3. Main Container
	sb.WriteString("<div class=\"app-container\">\n")

	// 3.1 Header / Hero Banner
	sb.WriteString("  <header class=\"hero-header\">\n")
	sb.WriteString("    <div class=\"hero-brand\">\n")
	sb.WriteString("      <div class=\"brand-logo\">\n")
	sb.WriteString("        <span class=\"brand-icon\">⚡</span>\n")
	sb.WriteString("        <span class=\"brand-name\">ChaosSQL</span>\n")
	sb.WriteString("      </div>\n")
	sb.WriteString("      <div class=\"brand-tagline\">Causal Concurrency Stress Testing & Isolation Anomaly Synthesis</div>\n")
	sb.WriteString("    </div>\n")
	sb.WriteString(fmt.Sprintf("    <div class=\"report-meta\">Generated at %s</div>\n", time.Now().UTC().Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString("  </header>\n\n")

	// 3.2 Execution Summary Banner
	sb.WriteString(renderHTMLExecutionSummary(spec, trace, hasAnomaly, anomalyType, len(cycles)))

	// 3.3 Adya Serialization Graph Visualization (SVG)
	sb.WriteString(renderHTMLSerializationGraph(graph, cycles))

	// 3.4 Invariant Integrity Audit Table
	if len(invResults) > 0 || len(spec.Invariants) > 0 {
		sb.WriteString(renderHTMLInvariantAuditTable(invResults, spec.Invariants))
	}

	// 3.5 Delta-Debugging (ddmin) Shrink Summary Card
	if shrinkResult != nil {
		sb.WriteString(renderHTMLShrinkSummary(shrinkResult))
	}

	// 3.6 Chronological Event Timeline Swimlane
	sb.WriteString(renderHTMLTimelineSwimlane(trace))

	// 3.7 Footer
	sb.WriteString("  <footer class=\"report-footer\">\n")
	sb.WriteString("    <div>ChaosSQL • Built for deterministic concurrency verification and isolation guarantees.</div>\n")
	sb.WriteString("  </footer>\n")

	sb.WriteString("</div>\n") // end app-container
	sb.WriteString("</body>\n</html>\n")

	return sb.String()
}

func embeddedCSS() string {
	return `
    :root {
      --bg-body: #0a0910;
      --bg-surface: #12101d;
      --bg-card: #181528;
      --bg-card-hover: #1f1b33;
      --bg-code: #0f0d1a;
      --border-base: #282340;
      --border-hover: #3d3560;
      
      --accent-purple: #7d56f4;
      --accent-purple-light: #a855f7;
      --accent-purple-glow: rgba(125, 86, 244, 0.35);
      
      --accent-cyan: #00d7ff;
      --accent-cyan-light: #38bdf8;
      --accent-cyan-glow: rgba(0, 215, 255, 0.3);
      
      --accent-red: #ff3366;
      --accent-red-glow: rgba(255, 51, 102, 0.4);
      
      --accent-green: #10b981;
      --accent-green-glow: rgba(16, 185, 129, 0.35);
      
      --accent-yellow: #fbbf24;
      --accent-yellow-glow: rgba(251, 191, 36, 0.3);
      
      --text-main: #f8fafc;
      --text-secondary: #94a3b8;
      --text-muted: #64748b;
      
      --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      --font-mono: 'JetBrains Mono', 'Fira Code', Menlo, Consolas, monospace;
    }

    * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }

    body {
      background-color: var(--bg-body);
      color: var(--text-main);
      font-family: var(--font-sans);
      line-height: 1.5;
      -webkit-font-smoothing: antialiased;
      padding: 32px 20px;
    }

    .app-container {
      max-width: 1200px;
      margin: 0 auto;
      display: flex;
      flex-direction: column;
      gap: 24px;
    }

    /* Hero Header */
    .hero-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 20px 28px;
      background: linear-gradient(135deg, rgba(125, 86, 244, 0.12), rgba(0, 215, 255, 0.08));
      border: 1px solid var(--border-base);
      border-radius: 16px;
      box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
      backdrop-filter: blur(12px);
    }

    .hero-brand {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .brand-logo {
      display: flex;
      align-items: center;
      gap: 10px;
    }

    .brand-icon {
      font-size: 24px;
      filter: drop-shadow(0 0 8px var(--accent-cyan));
    }

    .brand-name {
      font-size: 26px;
      font-weight: 800;
      background: linear-gradient(90deg, #ffffff, var(--accent-cyan), var(--accent-purple-light));
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      letter-spacing: -0.5px;
    }

    .brand-tagline {
      font-size: 13px;
      color: var(--text-secondary);
      font-weight: 500;
    }

    .report-meta {
      font-size: 12px;
      color: var(--text-muted);
      font-family: var(--font-mono);
      background: rgba(0, 0, 0, 0.3);
      padding: 6px 12px;
      border-radius: 8px;
      border: 1px solid var(--border-base);
    }

    /* Cards */
    .card {
      background: var(--bg-card);
      border: 1px solid var(--border-base);
      border-radius: 16px;
      padding: 24px;
      box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
      transition: border-color 0.2s ease, box-shadow 0.2s ease;
    }

    .card:hover {
      border-color: var(--border-hover);
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;
      padding-bottom: 12px;
      border-bottom: 1px solid var(--border-base);
    }

    .card-title {
      font-size: 18px;
      font-weight: 700;
      color: var(--accent-cyan);
      display: flex;
      align-items: center;
      gap: 10px;
      letter-spacing: -0.3px;
    }

    .card-subtitle {
      font-size: 13px;
      color: var(--text-muted);
    }

    /* Summary Grid */
    .summary-banner {
      display: flex;
      flex-direction: column;
      gap: 20px;
    }

    .status-badge-container {
      display: flex;
      align-items: center;
      gap: 16px;
      flex-wrap: wrap;
    }

    .status-badge {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      padding: 10px 20px;
      border-radius: 12px;
      font-size: 16px;
      font-weight: 800;
      letter-spacing: 0.5px;
      text-transform: uppercase;
    }

    .status-badge.anomaly {
      background: rgba(255, 51, 102, 0.15);
      border: 2px solid var(--accent-red);
      color: var(--accent-red);
      box-shadow: 0 0 20px var(--accent-red-glow);
    }

    .status-badge.satisfied {
      background: rgba(16, 185, 129, 0.15);
      border: 2px solid var(--accent-green);
      color: var(--accent-green);
      box-shadow: 0 0 20px var(--accent-green-glow);
    }

    .anomaly-type-pill {
      font-family: var(--font-mono);
      font-size: 14px;
      font-weight: 700;
      padding: 8px 16px;
      border-radius: 8px;
      background: rgba(125, 86, 244, 0.2);
      border: 1px solid var(--accent-purple);
      color: var(--accent-purple-light);
    }

    .metrics-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
      gap: 16px;
    }

    .metric-box {
      background: var(--bg-surface);
      border: 1px solid var(--border-base);
      border-radius: 12px;
      padding: 16px;
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .metric-label {
      font-size: 12px;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.5px;
      color: var(--text-muted);
    }

    .metric-value {
      font-size: 16px;
      font-weight: 700;
      color: var(--text-main);
      font-family: var(--font-mono);
    }

    .metric-desc {
      font-size: 12px;
      color: var(--text-secondary);
    }

    /* SVG Serialization Graph */
    .graph-container {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }

    .graph-canvas-wrapper {
      background: var(--bg-surface);
      border: 1px solid var(--border-base);
      border-radius: 12px;
      padding: 16px;
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 380px;
      overflow: auto;
    }

    .graph-canvas {
      max-width: 100%;
      height: auto;
    }

    .graph-legend {
      display: flex;
      flex-wrap: wrap;
      gap: 16px;
      padding: 12px 16px;
      background: var(--bg-surface);
      border-radius: 10px;
      border: 1px solid var(--border-base);
    }

    .legend-item {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 13px;
      color: var(--text-secondary);
    }

    .legend-color {
      width: 14px;
      height: 14px;
      border-radius: 4px;
    }

    .legend-color.wr { background: var(--accent-cyan); }
    .legend-color.ww { background: var(--accent-yellow); }
    .legend-color.rw { background: var(--accent-purple-light); }
    .legend-color.cycle {
      background: var(--accent-red);
      box-shadow: 0 0 8px var(--accent-red);
    }

    .cycle-edge {
      stroke: var(--accent-red) !important;
      stroke-width: 3px !important;
      filter: drop-shadow(0 0 6px var(--accent-red));
      animation: pulse-glow 2s infinite ease-in-out;
    }

    .cycle-node {
      stroke: var(--accent-red) !important;
      stroke-width: 2.5px !important;
      filter: drop-shadow(0 0 8px var(--accent-red-glow));
    }

    @keyframes pulse-glow {
      0%, 100% { opacity: 1; stroke-width: 3px; }
      50% { opacity: 0.7; stroke-width: 4px; }
    }

    /* Table Styles */
    .table-wrapper {
      overflow-x: auto;
      border: 1px solid var(--border-base);
      border-radius: 12px;
      background: var(--bg-surface);
    }

    .styled-table {
      width: 100%;
      border-collapse: collapse;
      text-align: left;
      font-size: 14px;
    }

    .styled-table th {
      background: rgba(125, 86, 244, 0.1);
      color: var(--accent-cyan);
      font-weight: 700;
      padding: 14px 18px;
      border-bottom: 1px solid var(--border-base);
      text-transform: uppercase;
      font-size: 12px;
      letter-spacing: 0.5px;
    }

    .styled-table td {
      padding: 14px 18px;
      border-bottom: 1px solid var(--border-base);
      color: var(--text-main);
    }

    .styled-table tr:last-child td {
      border-bottom: none;
    }

    .styled-table tr:hover td {
      background: rgba(255, 255, 255, 0.02);
    }

    .badge {
      display: inline-flex;
      align-items: center;
      padding: 4px 10px;
      border-radius: 6px;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.5px;
    }

    .badge-pass {
      background: rgba(16, 185, 129, 0.15);
      border: 1px solid var(--accent-green);
      color: var(--accent-green);
    }

    .badge-fail {
      background: rgba(255, 51, 102, 0.15);
      border: 1px solid var(--accent-red);
      color: var(--accent-red);
    }

    .code-pill {
      font-family: var(--font-mono);
      font-size: 13px;
      background: var(--bg-code);
      padding: 4px 8px;
      border-radius: 6px;
      border: 1px solid var(--border-base);
      color: #e2e8f0;
      display: inline-block;
    }

    /* Shrink Card */
    .shrink-stats {
      display: flex;
      flex-wrap: wrap;
      gap: 16px;
      margin-bottom: 20px;
    }

    .shrink-highlight {
      background: linear-gradient(135deg, rgba(125, 86, 244, 0.15), rgba(0, 215, 255, 0.1));
      border: 1px solid var(--accent-purple);
      border-radius: 12px;
      padding: 16px 24px;
      display: flex;
      align-items: center;
      gap: 16px;
      flex: 1;
      min-width: 280px;
    }

    .shrink-reduction-pill {
      font-size: 20px;
      font-weight: 800;
      color: var(--accent-yellow);
      background: rgba(251, 191, 36, 0.15);
      border: 1px solid var(--accent-yellow);
      padding: 6px 14px;
      border-radius: 8px;
      font-family: var(--font-mono);
    }

    .minimal-ops-list {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .op-card {
      background: var(--bg-surface);
      border: 1px solid var(--border-base);
      border-radius: 10px;
      padding: 14px 18px;
    }

    .op-header {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 8px;
      font-weight: 600;
    }

    .op-title {
      color: var(--accent-cyan);
      font-family: var(--font-mono);
      font-size: 14px;
    }

    .op-params {
      font-size: 12px;
      color: var(--text-muted);
      font-family: var(--font-mono);
    }

    .step-list {
      display: flex;
      flex-direction: column;
      gap: 6px;
      padding-left: 12px;
      border-left: 2px solid var(--border-base);
    }

    .step-item {
      font-size: 13px;
      font-family: var(--font-mono);
      color: #cbd5e1;
      display: flex;
      gap: 8px;
      align-items: baseline;
    }

    .step-num {
      color: var(--text-muted);
      font-size: 11px;
    }

    .step-sql {
      color: #e2e8f0;
      word-break: break-all;
    }

    .step-capture {
      color: var(--accent-yellow);
      font-size: 12px;
    }

    /* Timeline Swimlane */
    .timeline-container {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .timeline-events-list {
      display: flex;
      flex-direction: column;
      gap: 8px;
      max-height: 520px;
      overflow-y: auto;
      padding-right: 6px;
    }

    .timeline-events-list::-webkit-scrollbar {
      width: 6px;
    }
    .timeline-events-list::-webkit-scrollbar-thumb {
      background: var(--border-base);
      border-radius: 4px;
    }

    .timeline-row {
      display: grid;
      grid-template-columns: 80px 100px 90px 1fr;
      align-items: center;
      gap: 12px;
      background: var(--bg-surface);
      border: 1px solid var(--border-base);
      border-radius: 8px;
      padding: 10px 14px;
      font-size: 13px;
      transition: background 0.15s ease;
    }

    .timeline-row:hover {
      background: var(--bg-card-hover);
    }

    .time-stamp {
      font-family: var(--font-mono);
      font-size: 12px;
      color: var(--text-muted);
    }

    .worker-pill {
      font-family: var(--font-mono);
      font-size: 12px;
      font-weight: 600;
      padding: 2px 8px;
      border-radius: 6px;
      border: 1px solid var(--border-base);
      background: rgba(0, 0, 0, 0.2);
      color: var(--text-secondary);
      text-align: center;
    }

    .worker-w1 { border-color: var(--accent-cyan); color: var(--accent-cyan); }
    .worker-w2 { border-color: var(--accent-purple-light); color: var(--accent-purple-light); }
    .worker-w3 { border-color: var(--accent-yellow); color: var(--accent-yellow); }
    .worker-w4 { border-color: var(--accent-green); color: var(--accent-green); }

    .ev-type {
      font-family: var(--font-mono);
      font-size: 11px;
      font-weight: 700;
      padding: 3px 8px;
      border-radius: 5px;
      text-align: center;
      text-transform: uppercase;
    }

    .ev-begin { background: rgba(129, 140, 248, 0.15); color: #818cf8; border: 1px solid #818cf8; }
    .ev-exec { background: rgba(0, 215, 255, 0.15); color: var(--accent-cyan); border: 1px solid var(--accent-cyan); }
    .ev-commit { background: rgba(16, 185, 129, 0.15); color: var(--accent-green); border: 1px solid var(--accent-green); }
    .ev-rollback { background: rgba(245, 158, 11, 0.15); color: var(--accent-yellow); border: 1px solid var(--accent-yellow); }
    .ev-error { background: rgba(255, 51, 102, 0.2); color: var(--accent-red); border: 1px solid var(--accent-red); }

    .ev-content {
      font-family: var(--font-mono);
      font-size: 12px;
      color: #cbd5e1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .ev-op-tag {
      color: var(--accent-purple-light);
      margin-right: 6px;
      font-weight: 600;
    }

    /* Footer */
    .report-footer {
      text-align: center;
      font-size: 12px;
      color: var(--text-muted);
      padding: 16px 0;
      border-top: 1px solid var(--border-base);
      margin-top: 12px;
    }

    @media (max-width: 768px) {
      .hero-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 12px;
      }
      .timeline-row {
        grid-template-columns: 70px 80px 70px 1fr;
        font-size: 11px;
      }
    }
`
}

func renderHTMLExecutionSummary(spec domain.Spec, trace domain.ExecutionTrace, hasAnomaly bool, anomalyType domain.AnomalyType, cycleCount int) string {
	var sb strings.Builder

	sb.WriteString("  <section class=\"card summary-banner\">\n")
	sb.WriteString("    <div class=\"status-badge-container\">\n")

	if hasAnomaly {
		sb.WriteString("      <div class=\"status-badge anomaly\">\n")
		sb.WriteString("        <span>✘</span>\n")
		sb.WriteString("        <span>ISOLATION ANOMALY DETECTED</span>\n")
		sb.WriteString("      </div>\n")

		anomalyStr := string(anomalyType)
		if anomalyStr == "" || anomalyStr == string(domain.AnomalyUnknown) {
			anomalyStr = "INVARIANT_VIOLATION"
		}
		sb.WriteString(fmt.Sprintf("      <div class=\"anomaly-type-pill\">%s</div>\n", html.EscapeString(anomalyStr)))
	} else {
		sb.WriteString("      <div class=\"status-badge satisfied\">\n")
		sb.WriteString("        <span>✔</span>\n")
		sb.WriteString("        <span>INVARIANTS SATISFIED</span>\n")
		sb.WriteString("      </div>\n")
	}

	sb.WriteString("    </div>\n\n")

	// Metrics Grid
	sb.WriteString("    <div class=\"metrics-grid\">\n")

	// 1. Scenario
	sb.WriteString("      <div class=\"metric-box\">\n")
	sb.WriteString("        <div class=\"metric-label\">Scenario</div>\n")
	sb.WriteString(fmt.Sprintf("        <div class=\"metric-value\">%s</div>\n", html.EscapeString(spec.Name)))
	if spec.Description != "" {
		sb.WriteString(fmt.Sprintf("        <div class=\"metric-desc\">%s</div>\n", html.EscapeString(spec.Description)))
	}
	sb.WriteString("      </div>\n")

	// 2. Database Driver
	driverName := spec.Database.Driver
	if driverName == "" {
		driverName = "sqlite"
	}
	sb.WriteString("      <div class=\"metric-box\">\n")
	sb.WriteString("        <div class=\"metric-label\">Database Driver</div>\n")
	sb.WriteString(fmt.Sprintf("        <div class=\"metric-value\">%s</div>\n", html.EscapeString(driverName)))
	sb.WriteString("        <div class=\"metric-desc\">Engine Dialect & Driver</div>\n")
	sb.WriteString("      </div>\n")

	// 3. Engine Parameters
	workers := spec.Engine.Workers
	if workers <= 0 {
		workers = 4
	}
	iterations := spec.Engine.Iterations
	if iterations <= 0 {
		iterations = len(trace)
	}
	sb.WriteString("      <div class=\"metric-box\">\n")
	sb.WriteString("        <div class=\"metric-label\">Concurrency Parameters</div>\n")
	sb.WriteString(fmt.Sprintf("        <div class=\"metric-value\">%d workers | %d iter</div>\n", workers, iterations))
	sb.WriteString(fmt.Sprintf("        <div class=\"metric-desc\">Seed: %d</div>\n", spec.Engine.Seed))
	sb.WriteString("      </div>\n")

	// 4. Trace & Cycles Stats
	sb.WriteString("      <div class=\"metric-box\">\n")
	sb.WriteString("        <div class=\"metric-label\">Analysis Stats</div>\n")
	sb.WriteString(fmt.Sprintf("        <div class=\"metric-value\">%d events | %d cycles</div>\n", len(trace), cycleCount))
	sb.WriteString("        <div class=\"metric-desc\">Adya Serialization Graph</div>\n")
	sb.WriteString("      </div>\n")

	sb.WriteString("    </div>\n")
	sb.WriteString("  </section>\n\n")

	return sb.String()
}

type graphPoint struct {
	x float64
	y float64
}

func renderHTMLSerializationGraph(graph *analyzer.AdyaGraph, cycles []analyzer.Cycle) string {
	var sb strings.Builder

	sb.WriteString("  <section class=\"card graph-container\">\n")
	sb.WriteString("    <div class=\"card-header\">\n")
	sb.WriteString("      <div class=\"card-title\">\n")
	sb.WriteString("        <span>☊</span>\n")
	sb.WriteString("        <span>Adya Serialization Graph Visualization</span>\n")
	sb.WriteString("      </div>\n")
	sb.WriteString("      <div class=\"card-subtitle\">Directed Dependency Conflict Graph (WR, WW, RW)</div>\n")
	sb.WriteString("    </div>\n\n")

	// Legend
	sb.WriteString("    <div class=\"graph-legend\">\n")
	sb.WriteString("      <div class=\"legend-item\"><div class=\"legend-color wr\"></div><span>WR: Write-Read (Data Flow)</span></div>\n")
	sb.WriteString("      <div class=\"legend-item\"><div class=\"legend-color ww\"></div><span>WW: Write-Write (Overwrite)</span></div>\n")
	sb.WriteString("      <div class=\"legend-item\"><div class=\"legend-color rw\"></div><span>RW: Read-Write (Anti-Dependency)</span></div>\n")
	sb.WriteString("      <div class=\"legend-item\"><div class=\"legend-color cycle\"></div><span>Anomaly Cycle (Red Glow)</span></div>\n")
	sb.WriteString("    </div>\n\n")

	// Check if graph has nodes
	if graph == nil || len(graph.Nodes) == 0 {
		sb.WriteString("    <div class=\"graph-canvas-wrapper\">\n")
		sb.WriteString("      <div style=\"color: var(--text-muted); font-size: 14px;\">No dependency conflict cycles or serialized transactions recorded.</div>\n")
		sb.WriteString("    </div>\n")
		sb.WriteString("  </section>\n\n")
		return sb.String()
	}

	// Build cycle maps for rapid lookup
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

	// Sort node keys deterministically
	var sortedNodes []string
	for n := range graph.Nodes {
		sortedNodes = append(sortedNodes, n)
	}
	sort.Strings(sortedNodes)

	// Layout nodes in a circle
	canvasWidth := 740.0
	canvasHeight := 440.0
	cx := canvasWidth / 2.0
	cy := canvasHeight / 2.0
	radius := 145.0
	if len(sortedNodes) > 6 {
		radius = 165.0
	}

	nodePositions := make(map[string]graphPoint)
	numNodes := len(sortedNodes)

	for i, node := range sortedNodes {
		if numNodes == 1 {
			nodePositions[node] = graphPoint{x: cx, y: cy}
		} else {
			angle := (2.0 * math.Pi * float64(i) / float64(numNodes)) - (math.Pi / 2.0)
			nodePositions[node] = graphPoint{
				x: cx + radius*math.Cos(angle),
				y: cy + radius*math.Sin(angle),
			}
		}
	}

	sb.WriteString("    <div class=\"graph-canvas-wrapper\">\n")
	sb.WriteString(fmt.Sprintf("      <svg class=\"graph-canvas\" viewBox=\"0 0 %.0f %.0f\" width=\"100%%\" height=\"440\" xmlns=\"http://www.w3.org/2000/svg\">\n", canvasWidth, canvasHeight))

	// SVG Markers & Filters
	sb.WriteString(`        <defs>
          <filter id="glow-red" x="-30%" y="-30%" width="160%" height="160%">
            <feGaussianBlur stdDeviation="4" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
          <marker id="arrow-cyan" viewBox="0 0 10 10" refX="10" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
            <path d="M 0 1 L 10 5 L 0 9 z" fill="#00d7ff" />
          </marker>
          <marker id="arrow-amber" viewBox="0 0 10 10" refX="10" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
            <path d="M 0 1 L 10 5 L 0 9 z" fill="#fbbf24" />
          </marker>
          <marker id="arrow-purple" viewBox="0 0 10 10" refX="10" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
            <path d="M 0 1 L 10 5 L 0 9 z" fill="#c084fc" />
          </marker>
          <marker id="arrow-red" viewBox="0 0 10 10" refX="10" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
            <path d="M 0 1 L 10 5 L 0 9 z" fill="#ff3366" />
          </marker>
        </defs>
`)

	// Render Edges
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

			strokeColor := "#00d7ff"
			markerId := "arrow-cyan"
			switch edge.Type {
			case analyzer.DepWW:
				strokeColor = "#fbbf24"
				markerId = "arrow-amber"
			case analyzer.DepRW:
				strokeColor = "#c084fc"
				markerId = "arrow-purple"
			}

			edgeClass := ""
			strokeWidth := "1.8"
			if isCycle {
				strokeColor = "#ff3366"
				markerId = "arrow-red"
				edgeClass = "class=\"cycle-edge\""
				strokeWidth = "3"
			}

			// Render curved path
			if edge.From == edge.To {
				// Self loop
				sx := p1.x
				sy := p1.y - 20
				ex := p1.x + 30
				ey := p1.y - 18
				sb.WriteString(fmt.Sprintf("        <path d=\"M %.1f %.1f C %.1f %.1f, %.1f %.1f, %.1f %.1f\" fill=\"none\" stroke=\"%s\" stroke-width=\"%s\" marker-end=\"url(#%s)\" %s />\n",
					sx, sy, sx-25, sy-45, ex+25, ey-45, ex, ey, strokeColor, strokeWidth, markerId, edgeClass))
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

				// Boundary offsets for start and end
				nodeRadius := 38.0
				sx := p1.x + ux*nodeRadius
				sy := p1.y + uy*nodeRadius
				ex := p2.x - ux*(nodeRadius+8)
				ey := p2.y - uy*(nodeRadius+8)

				// Curvature control point
				mx := (sx + ex) / 2.0
				my := (sy + ey) / 2.0
				curveOffset := 26.0
				qx := mx + px*curveOffset
				qy := my + py*curveOffset

				sb.WriteString(fmt.Sprintf("        <path d=\"M %.1f %.1f Q %.1f %.1f %.1f %.1f\" fill=\"none\" stroke=\"%s\" stroke-width=\"%s\" marker-end=\"url(#%s)\" %s />\n",
					sx, sy, qx, qy, ex, ey, strokeColor, strokeWidth, markerId, edgeClass))

				// Edge Label Pill
				labelStr := string(edge.Type)
				if edge.Item != "" {
					labelStr += " " + edge.Item
				}
				labelWidth := float64(len(labelStr))*6.2 + 12.0
				lx := qx - labelWidth/2.0
				ly := qy - 9.0

				sb.WriteString(fmt.Sprintf("        <rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"18\" rx=\"4\" fill=\"#0f0d1a\" stroke=\"%s\" stroke-width=\"1\" opacity=\"0.95\" />\n",
					lx, ly, labelWidth, strokeColor))
				sb.WriteString(fmt.Sprintf("        <text x=\"%.1f\" y=\"%.1f\" fill=\"%s\" font-family=\"var(--font-mono)\" font-size=\"10\" font-weight=\"600\" text-anchor=\"middle\">%s</text>\n",
					qx, qy+4.0, strokeColor, html.EscapeString(labelStr)))
			}
		}
	}

	// Render Nodes
	for _, node := range sortedNodes {
		pos := nodePositions[node]
		isCycleNode := cycleNodeMap[node]

		nodeStroke := "#7d56f4"
		nodeClass := ""
		nodeFill := "#181528"
		if isCycleNode {
			nodeStroke = "#ff3366"
			nodeClass = "class=\"cycle-node\""
			nodeFill = "#241322"
		}

		boxW := 76.0
		boxH := 36.0
		bx := pos.x - boxW/2.0
		by := pos.y - boxH/2.0

		sb.WriteString(fmt.Sprintf("        <g %s>\n", nodeClass))
		sb.WriteString(fmt.Sprintf("          <rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\" rx=\"8\" fill=\"%s\" stroke=\"%s\" stroke-width=\"2\" />\n",
			bx, by, boxW, boxH, nodeFill, nodeStroke))
		sb.WriteString(fmt.Sprintf("          <text x=\"%.1f\" y=\"%.1f\" fill=\"#f8fafc\" font-family=\"var(--font-mono)\" font-size=\"13\" font-weight=\"700\" text-anchor=\"middle\">%s</text>\n",
			pos.x, pos.y+4.5, html.EscapeString(node)))
		sb.WriteString("        </g>\n")
	}

	sb.WriteString("      </svg>\n")
	sb.WriteString("    </div>\n")
	sb.WriteString("  </section>\n\n")

	return sb.String()
}

func renderHTMLInvariantAuditTable(results []domain.InvariantResult, configs []domain.InvariantConfig) string {
	var sb strings.Builder

	sb.WriteString("  <section class=\"card\">\n")
	sb.WriteString("    <div class=\"card-header\">\n")
	sb.WriteString("      <div class=\"card-title\">\n")
	sb.WriteString("        <span>⚖</span>\n")
	sb.WriteString("        <span>Invariant Integrity Audit Table</span>\n")
	sb.WriteString("      </div>\n")
	sb.WriteString("      <div class=\"card-subtitle\">Formal safety & consistency assertions evaluated on final database state</div>\n")
	sb.WriteString("    </div>\n\n")

	sb.WriteString("    <div class=\"table-wrapper\">\n")
	sb.WriteString("      <table class=\"styled-table\">\n")
	sb.WriteString("        <thead>\n")
	sb.WriteString("          <tr>\n")
	sb.WriteString("            <th>Invariant Name</th>\n")
	sb.WriteString("            <th>Status</th>\n")
	sb.WriteString("            <th>Assertion Expression</th>\n")
	sb.WriteString("            <th>Actual Database State</th>\n")
	sb.WriteString("          </tr>\n")
	sb.WriteString("        </thead>\n")
	sb.WriteString("        <tbody>\n")

	// If results provided, render them
	if len(results) > 0 {
		for _, inv := range results {
			statusBadge := "<span class=\"badge badge-pass\">PASS</span>"
			if !inv.Passed || inv.Error != nil {
				statusBadge = "<span class=\"badge badge-fail\">FAIL</span>"
			}

			stateStr := ""
			if inv.Error != nil {
				stateStr = fmt.Sprintf("<span style=\"color: var(--accent-red);\">Error: %s</span>", html.EscapeString(inv.Error.Error()))
			} else {
				jsonBytes, err := json.Marshal(inv.ActualValues)
				if err == nil {
					stateStr = fmt.Sprintf("<span class=\"code-pill\">%s</span>", html.EscapeString(string(jsonBytes)))
				} else {
					stateStr = fmt.Sprintf("<span class=\"code-pill\">%v</span>", html.EscapeString(fmt.Sprintf("%v", inv.ActualValues)))
				}
			}

			sb.WriteString("          <tr>\n")
			sb.WriteString(fmt.Sprintf("            <td style=\"font-weight: 600;\">%s</td>\n", html.EscapeString(inv.Name)))
			sb.WriteString(fmt.Sprintf("            <td>%s</td>\n", statusBadge))
			sb.WriteString(fmt.Sprintf("            <td><span class=\"code-pill\">%s</span></td>\n", html.EscapeString(inv.Expression)))
			sb.WriteString(fmt.Sprintf("            <td>%s</td>\n", stateStr))
			sb.WriteString("          </tr>\n")
		}
	} else {
		// Fallback to configs if results empty
		for _, cfg := range configs {
			sb.WriteString("          <tr>\n")
			sb.WriteString(fmt.Sprintf("            <td style=\"font-weight: 600;\">%s</td>\n", html.EscapeString(cfg.Name)))
			sb.WriteString("            <td><span class=\"badge badge-pass\">PASS</span></td>\n")
			sb.WriteString(fmt.Sprintf("            <td><span class=\"code-pill\">%s</span></td>\n", html.EscapeString(cfg.Assert)))
			sb.WriteString("            <td><span class=\"code-pill\">{}</span></td>\n")
			sb.WriteString("          </tr>\n")
		}
	}

	sb.WriteString("        </tbody>\n")
	sb.WriteString("      </table>\n")
	sb.WriteString("    </div>\n")
	sb.WriteString("  </section>\n\n")

	return sb.String()
}

func renderHTMLShrinkSummary(shrink *domain.ShrinkResult) string {
	if shrink == nil {
		return ""
	}

	var sb strings.Builder

	pctReduction := shrink.ReductionRatio
	if pctReduction <= 1.0 {
		pctReduction *= 100.0
	}

	sb.WriteString("  <section class=\"card\">\n")
	sb.WriteString("    <div class=\"card-header\">\n")
	sb.WriteString("      <div class=\"card-title\">\n")
	sb.WriteString("        <span>✂</span>\n")
	sb.WriteString("        <span>Delta-Debugging Causal Reduction (ddmin)</span>\n")
	sb.WriteString("      </div>\n")
	sb.WriteString("      <div class=\"card-subtitle\">Synthesized minimal failing transaction interleaving schedule</div>\n")
	sb.WriteString("    </div>\n\n")

	sb.WriteString("    <div class=\"shrink-stats\">\n")
	sb.WriteString("      <div class=\"shrink-highlight\">\n")
	sb.WriteString(fmt.Sprintf("        <div class=\"shrink-reduction-pill\">%.1f%%</div>\n", pctReduction))
	sb.WriteString("        <div>\n")
	sb.WriteString(fmt.Sprintf("          <div style=\"font-size: 15px; font-weight: 700; color: #fff;\">%d ops ──► %d ops</div>\n", shrink.OriginalSize, shrink.ReducedSize))
	sb.WriteString("          <div style=\"font-size: 12px; color: var(--text-muted);\">Schedule Search Reduction</div>\n")
	sb.WriteString("        </div>\n")
	sb.WriteString("      </div>\n")

	sb.WriteString("      <div class=\"metric-box\" style=\"flex: 1; min-width: 200px;\">\n")
	sb.WriteString("        <div class=\"metric-label\">Algorithm Search Cost</div>\n")
	sb.WriteString(fmt.Sprintf("        <div class=\"metric-value\">%d iterations</div>\n", shrink.Iterations))
	sb.WriteString(fmt.Sprintf("        <div class=\"metric-desc\">Reduction Duration: %s</div>\n", shrink.Duration.Round(100*time.Microsecond)))
	sb.WriteString("      </div>\n")
	sb.WriteString("    </div>\n\n")

	if len(shrink.MinimalOps) > 0 {
		sb.WriteString("    <div class=\"minimal-ops-list\">\n")
		for _, op := range shrink.MinimalOps {
			var paramPairs []string
			for k, v := range op.Params {
				paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", k, v))
			}
			sort.Strings(paramPairs)
			paramsStr := ""
			if len(paramPairs) > 0 {
				paramsStr = "{" + strings.Join(paramPairs, ", ") + "}"
			}

			sb.WriteString("      <div class=\"op-card\">\n")
			sb.WriteString("        <div class=\"op-header\">\n")
			sb.WriteString(fmt.Sprintf("          <span class=\"op-title\">[%s #%d]</span>\n", html.EscapeString(op.Name), op.ID))
			if paramsStr != "" {
				sb.WriteString(fmt.Sprintf("          <span class=\"op-params\">%s</span>\n", html.EscapeString(paramsStr)))
			}
			sb.WriteString("        </div>\n")

			sb.WriteString("        <div class=\"step-list\">\n")
			for sIdx, step := range op.Steps {
				capStr := ""
				if step.Capture != "" {
					capStr = fmt.Sprintf("<span class=\"step-capture\">-&gt; capture(%s)</span>", html.EscapeString(step.Capture))
				}
				cleanSQL := sanitizeSQLForMermaid(step.SQL)
				sb.WriteString("          <div class=\"step-item\">\n")
				sb.WriteString(fmt.Sprintf("            <span class=\"step-num\">step %d:</span>\n", sIdx+1))
				sb.WriteString(fmt.Sprintf("            <span class=\"step-sql\">%s</span>\n", html.EscapeString(cleanSQL)))
				if capStr != "" {
					sb.WriteString(fmt.Sprintf("            %s\n", capStr))
				}
				sb.WriteString("          </div>\n")
			}
			sb.WriteString("        </div>\n")
			sb.WriteString("      </div>\n")
		}
		sb.WriteString("    </div>\n")
	}

	sb.WriteString("  </section>\n\n")

	return sb.String()
}

func renderHTMLTimelineSwimlane(trace domain.ExecutionTrace) string {
	var sb strings.Builder

	sb.WriteString("  <section class=\"card timeline-container\">\n")
	sb.WriteString("    <div class=\"card-header\">\n")
	sb.WriteString("      <div class=\"card-title\">\n")
	sb.WriteString("        <span>⏱</span>\n")
	sb.WriteString("        <span>Chronological Execution Timeline & Swimlane</span>\n")
	sb.WriteString("      </div>\n")
	sb.WriteString(fmt.Sprintf("      <div class=\"card-subtitle\">%d ordered concurrency trace events</div>\n", len(trace)))
	sb.WriteString("    </div>\n\n")

	if len(trace) == 0 {
		sb.WriteString("    <div style=\"color: var(--text-muted); font-size: 13px; padding: 12px 0;\">No execution trace events captured.</div>\n")
		sb.WriteString("  </section>\n\n")
		return sb.String()
	}

	sb.WriteString("    <div class=\"timeline-events-list\">\n")

	for _, ev := range trace {
		timeStr := fmt.Sprintf("+%.2fms", float64(ev.Timestamp.Microseconds())/1000.0)

		workerID := ev.WorkerID
		if workerID <= 0 {
			workerID = 1
		}
		workerClass := fmt.Sprintf("worker-w%d", (workerID%4)+1)
		workerLabel := fmt.Sprintf("Worker %d", workerID)

		evTypeClass := "ev-exec"
		switch ev.Type {
		case domain.EventBegin:
			evTypeClass = "ev-begin"
		case domain.EventCommit:
			evTypeClass = "ev-commit"
		case domain.EventRollback:
			evTypeClass = "ev-rollback"
		case domain.EventError:
			evTypeClass = "ev-error"
		}

		opTag := ""
		if ev.OpName != "" {
			opTag = fmt.Sprintf("[%s #%d] ", ev.OpName, ev.OpIndex)
		}

		content := ev.SQL
		if ev.Type == domain.EventError && ev.Error != "" {
			content = fmt.Sprintf("ERROR: %s (query: %s)", ev.Error, ev.SQL)
		}

		sb.WriteString("      <div class=\"timeline-row\">\n")
		sb.WriteString(fmt.Sprintf("        <div class=\"time-stamp\">%s</div>\n", timeStr))
		sb.WriteString(fmt.Sprintf("        <div class=\"worker-pill %s\">%s</div>\n", workerClass, workerLabel))
		sb.WriteString(fmt.Sprintf("        <div class=\"ev-type %s\">%s</div>\n", evTypeClass, ev.Type))
		sb.WriteString("        <div class=\"ev-content\">\n")
		if opTag != "" {
			sb.WriteString(fmt.Sprintf("          <span class=\"ev-op-tag\">%s</span>\n", html.EscapeString(opTag)))
		}
		sb.WriteString(fmt.Sprintf("          <span>%s</span>\n", html.EscapeString(content)))
		sb.WriteString("        </div>\n")
		sb.WriteString("      </div>\n")
	}

	sb.WriteString("    </div>\n")
	sb.WriteString("  </section>\n\n")

	return sb.String()
}
