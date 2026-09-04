package reporter

import (
	"encoding/json"
	"fmt"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func embeddedUICentralCSS() string {
	return `
    :root {
      /* Studio Bregalda Canonical Design Tokens */
      --canvas: #120E1F;
      --surface: #181328;
      --surface-elevated: #1F1934;
      --surface-hover: #2A2140;
      --cream: #FCFBF8;
      --cream-dim: rgba(252, 251, 248, 0.75);
      --cream-muted: rgba(252, 251, 248, 0.45);
      --border-subtle: rgba(252, 251, 248, 0.08);
      --border-medium: rgba(252, 251, 248, 0.16);
      --border-active: rgba(245, 196, 0, 0.4);

      --purple: #4B2E83;
      --purple-light: #6A44B0;
      --yellow: #F5C400;
      --green: #22C55E;
      --red: #EF4444;

      --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      --font-mono: 'JetBrains Mono', 'Fira Code', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    }

    *, *::before, *::after {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }

    html, body {
      background-color: var(--canvas);
      color: var(--cream);
      font-family: var(--font-sans);
      font-size: 14px;
      line-height: 1.5;
      -webkit-font-smoothing: antialiased;
      -moz-osx-font-smoothing: grayscale;
    }

    .font-mono {
      font-family: var(--font-mono);
    }

    .app-container {
      max-width: 1440px;
      margin: 0 auto;
      padding: 28px 32px 48px;
    }

    .card {
      background: var(--surface);
      border: 1px solid var(--border-medium);
      border-radius: 8px;
      padding: 20px 24px;
      margin-bottom: 24px;
    }

    .card-header-bar {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 18px;
      flex-wrap: wrap;
      gap: 12px;
    }

    .card-title {
      font-size: 16px;
      font-weight: 600;
      letter-spacing: -0.01em;
      color: var(--cream);
    }

    .card-subtitle {
      font-size: 12px;
      color: var(--cream-muted);
      margin-top: 2px;
    }

    .card-tag-pill {
      font-family: var(--font-mono);
      font-size: 11px;
      font-weight: 600;
      padding: 3px 8px;
      background: var(--surface-elevated);
      color: var(--yellow);
      border: 1px solid var(--border-subtle);
      border-radius: 4px;
    }

    .scenario-summary {
      background: var(--surface);
      border: 1px solid var(--border-medium);
      border-radius: 8px;
      padding: 24px 28px;
      margin-bottom: 24px;
    }

    .brand-bar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 22px;
      flex-wrap: wrap;
      gap: 16px;
    }

    .brand-title-group {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .brand-symbol {
      color: var(--yellow);
      font-size: 20px;
    }

    .brand-name {
      font-size: 22px;
      font-weight: 700;
      letter-spacing: -0.02em;
      color: var(--cream);
    }

    .brand-tag {
      font-family: var(--font-mono);
      font-size: 10px;
      font-weight: 700;
      letter-spacing: 0.08em;
      padding: 2px 7px;
      background: var(--purple);
      color: var(--cream);
      border-radius: 3px;
    }

    .brand-version {
      font-family: var(--font-mono);
      font-size: 11px;
      color: var(--cream-muted);
    }

    .status-badge-group {
      display: flex;
      align-items: center;
      gap: 10px;
    }

    .badge {
      display: inline-flex;
      align-items: center;
      gap: 7px;
      font-family: var(--font-mono);
      font-size: 11px;
      font-weight: 700;
      letter-spacing: 0.04em;
      padding: 6px 12px;
      border-radius: 4px;
    }

    .badge-success {
      background: rgba(34, 197, 94, 0.12);
      color: var(--green);
      border: 1px solid rgba(34, 197, 94, 0.3);
    }

    .badge-danger {
      background: rgba(239, 68, 68, 0.12);
      color: var(--red);
      border: 1px solid rgba(239, 68, 68, 0.35);
    }

    .badge-warning {
      background: rgba(245, 196, 0, 0.12);
      color: var(--yellow);
      border: 1px solid rgba(245, 196, 0, 0.35);
    }

    .badge-dot {
      width: 7px;
      height: 7px;
      border-radius: 50%;
    }

    .dot-success { background: var(--green); }
    .dot-danger { background: var(--red); }

    .meta-cards-row {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 14px;
    }

    .meta-card {
      background: var(--surface-elevated);
      border: 1px solid var(--border-subtle);
      border-radius: 6px;
      padding: 12px 16px;
    }

    .meta-card-label {
      font-size: 10px;
      font-weight: 600;
      color: var(--cream-muted);
      letter-spacing: 0.06em;
      margin-bottom: 4px;
    }

    .meta-card-value {
      font-size: 15px;
      font-weight: 600;
      color: var(--cream);
    }

    .invariant-alert-strip {
      margin-top: 18px;
      background: rgba(239, 68, 68, 0.08);
      border: 1px solid rgba(239, 68, 68, 0.25);
      border-radius: 6px;
      padding: 12px 16px;
      display: flex;
      align-items: flex-start;
      gap: 12px;
      font-size: 13px;
    }

    .alert-icon {
      color: var(--yellow);
      font-size: 16px;
    }

    .alert-text code {
      font-family: var(--font-mono);
      background: var(--surface-elevated);
      padding: 2px 6px;
      border-radius: 3px;
      color: var(--yellow);
    }

    .timeline-legend {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }

    .legend-pill {
      font-family: var(--font-mono);
      font-size: 10px;
      font-weight: 600;
      padding: 3px 8px;
      border-radius: 3px;
    }

    .pill-begin { background: var(--purple); color: var(--cream); }
    .pill-exec { background: var(--surface-elevated); color: var(--cream-dim); border: 1px solid var(--border-medium); }
    .pill-commit { background: rgba(34, 197, 94, 0.15); color: var(--green); border: 1px solid rgba(34, 197, 94, 0.3); }
    .pill-rollback { background: rgba(239, 68, 68, 0.15); color: var(--red); border: 1px solid rgba(239, 68, 68, 0.3); }
    .pill-savepoint { background: rgba(245, 196, 0, 0.15); color: var(--yellow); border: 1px solid rgba(245, 196, 0, 0.3); }

    .swimlane-board {
      display: flex;
      flex-direction: column;
      gap: 12px;
      overflow-x: auto;
      padding-bottom: 8px;
    }

    .swimlane-row {
      display: flex;
      align-items: center;
      background: var(--surface-elevated);
      border: 1px solid var(--border-subtle);
      border-radius: 6px;
      padding: 8px 12px;
      min-width: 800px;
    }

    .swimlane-worker-label {
      width: 110px;
      flex-shrink: 0;
      border-right: 1px solid var(--border-subtle);
      padding-right: 12px;
    }

    .worker-badge {
      font-size: 12px;
      font-weight: 600;
      color: var(--cream);
    }

    .worker-event-count {
      font-size: 11px;
      color: var(--cream-muted);
    }

    .swimlane-track {
      display: flex;
      align-items: center;
      gap: 8px;
      padding-left: 14px;
      overflow-x: auto;
      flex-grow: 1;
    }

    .timeline-node {
      flex-shrink: 0;
      display: flex;
      align-items: center;
      gap: 6px;
      padding: 6px 10px;
      border-radius: 4px;
      cursor: pointer;
      user-select: none;
      transition: background-color 0.15s ease, border-color 0.15s ease;
      font-size: 11px;
    }

    .timeline-node:hover {
      border-color: var(--yellow) !important;
    }

    .timeline-node.selected {
      outline: 2px solid var(--yellow);
    }

    .node-tx {
      font-weight: 700;
    }

    .node-type {
      opacity: 0.8;
      font-size: 10px;
      text-transform: uppercase;
    }

    .node-time {
      font-size: 10px;
      opacity: 0.65;
    }

    .node-sql {
      font-size: 10px;
      opacity: 0.85;
      max-width: 180px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .two-column-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 24px;
      margin-bottom: 24px;
    }

    @media (max-width: 1024px) {
      .two-column-grid {
        grid-template-columns: 1fr;
      }
    }

    .graph-legend-compact {
      display: flex;
      align-items: center;
      gap: 12px;
      font-family: var(--font-mono);
      font-size: 11px;
    }

    .legend-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      display: inline-block;
    }

    .dot-wr { background: var(--yellow); }
    .dot-ww { background: var(--purple-light); }
    .dot-rw { background: var(--cream); }
    .dot-cycle { background: var(--yellow); box-shadow: 0 0 4px var(--yellow); }

    .graph-canvas-container {
      background: var(--surface-elevated);
      border: 1px solid var(--border-subtle);
      border-radius: 6px;
      overflow: hidden;
      display: flex;
      justify-content: center;
      align-items: center;
    }

    .adya-svg {
      display: block;
    }

    .cycle-edge {
      stroke: var(--yellow) !important;
    }

    .cycle-node {
      stroke: var(--yellow) !important;
    }

    .shrink-metrics-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 12px;
      margin-bottom: 18px;
    }

    .shrink-metric-box {
      background: var(--surface-elevated);
      border: 1px solid var(--border-subtle);
      border-radius: 6px;
      padding: 12px 14px;
    }

    .metric-box-highlight {
      border-color: rgba(245, 196, 0, 0.3);
      background: rgba(245, 196, 0, 0.05);
    }

    .metric-label {
      font-size: 10px;
      font-weight: 600;
      color: var(--cream-muted);
      letter-spacing: 0.06em;
      margin-bottom: 4px;
    }

    .metric-val {
      font-size: 16px;
      font-weight: 700;
      color: var(--cream);
    }

    .val-reduced {
      color: var(--green);
    }

    .val-highlight {
      color: var(--yellow);
    }

    .minimal-ops-section {
      background: var(--surface-elevated);
      border: 1px solid var(--border-subtle);
      border-radius: 6px;
      padding: 14px;
      max-height: 250px;
      overflow-y: auto;
    }

    .section-label-bar {
      font-size: 11px;
      font-weight: 600;
      color: var(--cream-muted);
      margin-bottom: 10px;
    }

    .minimal-ops-list {
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    .minimal-op-card {
      background: var(--surface);
      border: 1px solid var(--border-subtle);
      border-radius: 4px;
      padding: 8px 12px;
    }

    .op-header {
      display: flex;
      justify-content: space-between;
      margin-bottom: 6px;
      font-size: 11px;
    }

    .op-badge {
      color: var(--yellow);
      font-weight: 600;
    }

    .op-name {
      color: var(--cream-dim);
    }

    .op-step-line {
      font-size: 11px;
      color: var(--cream);
      padding: 2px 0;
      white-space: pre-wrap;
      word-break: break-all;
    }

    .step-num {
      color: var(--cream-muted);
    }

    .search-controls {
      display: flex;
      gap: 10px;
      align-items: center;
    }

    .search-input {
      background: var(--surface-elevated);
      border: 1px solid var(--border-medium);
      color: var(--cream);
      padding: 7px 12px;
      border-radius: 4px;
      font-size: 12px;
      width: 260px;
      outline: none;
    }

    .search-input:focus {
      border-color: var(--yellow);
    }

    .filter-select {
      background: var(--surface-elevated);
      border: 1px solid var(--border-medium);
      color: var(--cream);
      padding: 7px 10px;
      border-radius: 4px;
      font-size: 12px;
      outline: none;
    }

    .table-responsive-wrapper {
      max-height: 480px;
      overflow-y: auto;
      border: 1px solid var(--border-subtle);
      border-radius: 6px;
      background: var(--surface-elevated);
    }

    .inspector-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 12px;
      text-align: left;
    }

    .inspector-table th {
      background: var(--surface);
      position: sticky;
      top: 0;
      padding: 10px 14px;
      font-size: 10px;
      font-weight: 700;
      letter-spacing: 0.05em;
      color: var(--cream-muted);
      border-bottom: 1px solid var(--border-medium);
      z-index: 2;
    }

    .inspector-table td {
      padding: 8px 14px;
      border-bottom: 1px solid var(--border-subtle);
      color: var(--cream-dim);
    }

    .statement-row {
      cursor: pointer;
      transition: background-color 0.1s ease;
    }

    .statement-row:hover {
      background-color: var(--surface-hover);
    }

    .statement-row.active {
      background-color: rgba(75, 46, 131, 0.3) !important;
      outline: 1px solid var(--yellow);
    }

    .col-idx { color: var(--cream-muted); }
    .col-time { color: var(--cream-dim); }
    .col-worker .worker-tag {
      background: var(--surface);
      border: 1px solid var(--border-subtle);
      padding: 2px 6px;
      border-radius: 3px;
      color: var(--cream);
      font-size: 11px;
    }
    .col-tx { color: var(--cream); font-weight: 600; }
    .col-sql {
      color: var(--cream);
      white-space: pre-wrap;
      word-break: break-all;
    }

    .type-chip {
      display: inline-block;
      font-size: 10px;
      font-weight: 700;
      padding: 2px 6px;
      border-radius: 3px;
    }

    .type-begin { background: var(--purple); color: var(--cream); }
    .type-exec { background: var(--surface); color: var(--cream-dim); border: 1px solid var(--border-subtle); }
    .type-commit { background: rgba(34, 197, 94, 0.15); color: var(--green); }
    .type-rollback { background: rgba(239, 68, 68, 0.15); color: var(--red); }
    .type-savepoint { background: rgba(245, 196, 0, 0.15); color: var(--yellow); }

    .status-tag {
      font-size: 10px;
      font-weight: 700;
      padding: 2px 6px;
      border-radius: 3px;
    }
    .status-ok { color: var(--green); }
    .status-abort { color: var(--yellow); }
    .status-err { color: var(--red); }

    .inspector-detail-drawer {
      margin-top: 14px;
      background: var(--surface-elevated);
      border: 1px solid var(--border-subtle);
      border-radius: 6px;
      padding: 14px 18px;
    }

    .drawer-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 10px;
    }

    .drawer-title {
      font-size: 12px;
      font-weight: 600;
      color: var(--cream-muted);
    }

    .drawer-badge {
      font-size: 11px;
      color: var(--yellow);
    }

    .drawer-sql-view {
      font-size: 12px;
      color: var(--cream);
      background: var(--surface);
      border: 1px solid var(--border-subtle);
      border-radius: 4px;
      padding: 12px 14px;
      overflow-x: auto;
      white-space: pre-wrap;
      word-break: break-all;
    }

    .empty-state-box {
      text-align: center;
      padding: 36px 20px;
      color: var(--cream-muted);
    }

    .empty-state-icon {
      font-size: 28px;
      margin-bottom: 8px;
      opacity: 0.6;
    }

    .empty-state-title {
      font-size: 14px;
      font-weight: 600;
      color: var(--cream);
      margin-bottom: 4px;
    }

    .empty-state-desc {
      font-size: 12px;
    }

    .ui-footer {
      text-align: center;
      padding: 20px 0;
      border-top: 1px solid var(--border-subtle);
      margin-top: 20px;
    }

    .footer-brand {
      font-size: 12px;
      font-weight: 600;
      color: var(--cream-dim);
      margin-bottom: 4px;
    }

    .footer-meta {
      font-size: 11px;
      color: var(--cream-muted);
    }
`
}

func embeddedUIClientJS(trace domain.ExecutionTrace) string {
	type jsTraceEvent struct {
		Index     int    `json:"idx"`
		Timestamp string `json:"time"`
		WorkerID  int    `json:"worker"`
		TxID      string `json:"tx"`
		Type      string `json:"type"`
		SQL       string `json:"sql"`
		Error     string `json:"error,omitempty"`
	}

	jsEvents := make([]jsTraceEvent, len(trace))
	for i, ev := range trace {
		tUs := ev.Timestamp.Microseconds()
		tStr := fmt.Sprintf("+%dµs", tUs)
		if tUs >= 1000 {
			tStr = fmt.Sprintf("+%.2fms", float64(tUs)/1000.0)
		}
		jsEvents[i] = jsTraceEvent{
			Index:     i + 1,
			Timestamp: tStr,
			WorkerID:  ev.WorkerID,
			TxID:      fmt.Sprintf("T%d-%d", ev.WorkerID, ev.OpIndex),
			Type:      string(ev.Type),
			SQL:       ev.SQL,
			Error:     ev.Error,
		}
	}

	eventsJSON, _ := json.Marshal(jsEvents)

	return fmt.Sprintf(`
    const traceEvents = %s;

    function filterStatements() {
      const query = (document.getElementById('statement-search').value || '').toLowerCase();
      const worker = document.getElementById('worker-filter').value;
      const rows = document.querySelectorAll('.statement-row');

      rows.forEach(row => {
        const text = row.textContent.toLowerCase();
        const rowWorker = row.getAttribute('data-worker');
        const matchesQuery = !query || text.includes(query);
        const matchesWorker = !worker || rowWorker === worker;

        if (matchesQuery && matchesWorker) {
          row.style.display = '';
        } else {
          row.style.display = 'none';
        }
      });
    }

    function inspectEvent(idx) {
      document.querySelectorAll('.statement-row').forEach(r => r.classList.remove('active'));
      const activeRow = document.getElementById('row-' + idx);
      if (activeRow) {
        activeRow.classList.add('active');
        activeRow.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
      }

      document.querySelectorAll('.timeline-node').forEach(n => n.classList.remove('selected'));
      const activeNode = document.querySelector('.timeline-node[data-idx="' + idx + '"]');
      if (activeNode) {
        activeNode.classList.add('selected');
      }

      const ev = traceEvents[idx];
      if (ev) {
        const badge = document.getElementById('drawer-event-id');
        const sqlView = document.getElementById('drawer-sql');
        if (badge && sqlView) {
          badge.textContent = 'Event #' + ev.idx + ' • ' + ev.tx + ' • ' + ev.time + ' • ' + ev.type;
          let content = ev.sql;
          if (ev.error) {
            content += '\n\n-- ERROR ENCOUNTERED --\n' + ev.error;
          }
          sqlView.textContent = content;
        }
      }
    }
`, string(eventsJSON))
}
