package proxy

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// StartLiveUI launches the live HTTP telemetry and diagnostic dashboard.
func StartLiveUI(addr string, server *Server) (*http.Server, string, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to bind UI listener on %s: %w", addr, err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		stats := server.Stats()
		anomalies := server.Engine().GetAnomalies()
		html := renderProxyDashboardHTML(server.cfg, stats, anomalies)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	})

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(server.Stats())
	})

	mux.HandleFunc("/api/anomalies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(server.Engine().GetAnomalies())
	})

	mux.HandleFunc("/api/sarif", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		sarifJSON, err := GenerateProxySARIF(server.Engine().GetAnomalies(), "live-proxy")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(sarifJSON))
	})

	mux.HandleFunc("/api/trace", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(server.Engine().GetTrace())
	})

	httpServer := &http.Server{
		Handler: mux,
	}

	go func() {
		_ = httpServer.Serve(ln)
	}()

	actualURL := fmt.Sprintf("http://%s", ln.Addr().String())
	return httpServer, actualURL, nil
}

func renderProxyDashboardHTML(cfg ProxyConfig, stats ProxyStats, anomalies []DetectedAnomaly) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta http-equiv="refresh" content="3">
  <title>ChaosSQL — Transparent Database Reverse Proxy</title>
  <style>
    :root {
      --bg: #0d1117;
      --card-bg: #161b22;
      --border: #30363d;
      --text: #c9d1d9;
      --text-muted: #8b949e;
      --accent: #58a6ff;
      --danger: #f85149;
      --warning: #d29922;
      --success: #3fb950;
    }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
      background: var(--bg);
      color: var(--text);
      margin: 0;
      padding: 24px;
    }
    .container { max-width: 1100px; margin: 0 auto; }
    header {
      border-bottom: 1px solid var(--border);
      padding-bottom: 16px;
      margin-bottom: 24px;
    }
    h1 { margin: 0 0 8px 0; font-size: 24px; color: #fff; }
    .badge {
      display: inline-block;
      padding: 2px 8px;
      border-radius: 12px;
      font-size: 12px;
      font-weight: 600;
      background: var(--border);
    }
    .badge-danger { background: rgba(248, 81, 73, 0.2); color: var(--danger); }
    .badge-success { background: rgba(63, 185, 80, 0.2); color: var(--success); }
    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 16px;
      margin-bottom: 24px;
    }
    .card {
      background: var(--card-bg);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 16px;
    }
    .metric-value { font-size: 28px; font-weight: 700; color: #fff; margin: 4px 0; }
    .metric-label { font-size: 12px; color: var(--text-muted); text-transform: uppercase; }
    .anomaly-card {
      background: var(--card-bg);
      border: 1px solid var(--danger);
      border-radius: 8px;
      padding: 16px;
      margin-bottom: 12px;
    }
    pre {
      background: rgba(0,0,0,0.3);
      padding: 10px;
      border-radius: 6px;
      overflow-x: auto;
      font-size: 13px;
    }
  </style>
</head>
<body>
<div class="container">
  <header>
    <h1>ChaosSQL — Layer-7 Transparent Database Proxy</h1>
    <span class="badge">Protocol: ` + string(cfg.Protocol) + `</span>
    <span class="badge">Listen: ` + cfg.ListenAddr + `</span>
    <span class="badge">Upstream: ` + cfg.UpstreamAddr + `</span>
  </header>

  <div class="grid">
    <div class="card">
      <div class="metric-label">Active Sessions</div>
      <div class="metric-value">` + fmt.Sprintf("%d", stats.ActiveSessions) + `</div>
    </div>
    <div class="card">
      <div class="metric-label">Intercepted Queries</div>
      <div class="metric-value">` + fmt.Sprintf("%d", stats.TotalQueries) + `</div>
    </div>
    <div class="card">
      <div class="metric-label">Jitter Delay Points</div>
      <div class="metric-value">` + fmt.Sprintf("%d", stats.JitterInjectedCount) + `</div>
    </div>
    <div class="card">
      <div class="metric-label">Anomalies Detected</div>
      <div class="metric-value" style="color: ` + getAnomalyColor(stats.AnomaliesDetected) + `">` +
		fmt.Sprintf("%d", stats.AnomaliesDetected) + `</div>
    </div>
  </div>

  <h2>💥 Live Detected Concurrency Anomalies</h2>
`)

	if len(anomalies) == 0 {
		sb.WriteString(`<p style="color: var(--text-muted)">No concurrency anomalies detected yet. Application traffic is serializable.</p>`)
	} else {
		for _, a := range anomalies {
			sb.WriteString(fmt.Sprintf(`
  <div class="anomaly-card">
    <div style="display:flex; justify-content:space-between; align-items:center;">
      <span class="badge badge-danger">%s</span>
      <span style="font-size:12px; color:var(--text-muted)">%s</span>
    </div>
    <p><strong>Description:</strong> %s</p>
    <p><strong>Transactions:</strong> %s</p>
    <p><strong>Items:</strong> %s</p>
`, a.Type, a.DetectedAt.Format(time.RFC3339), a.Description, strings.Join(a.TxIDs, ", "), strings.Join(a.Items, ", ")))

			if len(a.Cycle) > 0 {
				sb.WriteString("<pre>")
				for _, edge := range a.Cycle {
					sb.WriteString(fmt.Sprintf("%s ──(%s on %s)──► %s\n", edge.From, edge.Type, edge.Item, edge.To))
				}
				sb.WriteString("</pre>")
			}
			sb.WriteString("</div>")
		}
	}

	sb.WriteString(`
</div>
</body>
</html>
`)
	return sb.String()
}

func getAnomalyColor(count int) string {
	if count > 0 {
		return "var(--danger)"
	}
	return "var(--success)"
}
