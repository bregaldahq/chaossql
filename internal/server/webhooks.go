package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	neturl "net/url"
	"strings"
	"time"
)

type RegressionAlert struct {
	RepoFullName   string `json:"repo_full_name"`
	Branch         string `json:"branch"`
	PRNumber       int    `json:"pr_number"`
	CommitSHA      string `json:"commit_sha"`
	AnomalyType    string `json:"anomaly_type"`
	AnomalyName    string `json:"anomaly_name"`
	Driver         string `json:"driver"`
	Isolation      string `json:"isolation"`
	Scenario       string `json:"scenario"`
	Seed           uint64 `json:"seed"`
	DurationMS     int64  `json:"duration_ms"`
	BaselineStatus string `json:"baseline_status"`
	RunURL         string `json:"run_url"`
	IsRegression   bool   `json:"is_regression"`
}

type WebhookDispatcher struct {
	client *http.Client
}

func NewWebhookDispatcher(client *http.Client) *WebhookDispatcher {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &WebhookDispatcher{client: client}
}

func (d *WebhookDispatcher) DispatchAlert(ctx context.Context, wh WebhookRecord, alert *RegressionAlert) error {
	switch strings.ToLower(wh.TargetType) {
	case "discord":
		return d.sendDiscord(ctx, wh.URL, alert)
	case "slack":
		return d.sendSlack(ctx, wh.URL, alert)
	default:
		return d.sendGeneric(ctx, wh.URL, alert)
	}
}

func (d *WebhookDispatcher) sendDiscord(ctx context.Context, url string, alert *RegressionAlert) error {
	color := 0xDC2626 // Alert Red
	title := "🚨 REGRESSÃO DE CONCORRÊNCIA DETECTADA!"
	if !alert.IsRegression {
		color = 0xF5C400 // Signal Yellow
		title = "⚠️ ANOMALIA DE ISOLAMENTO DETECTADA"
	}

	prText := "Direto na branch"
	if alert.PRNumber > 0 {
		prText = fmt.Sprintf("#%d", alert.PRNumber)
	}

	commitShort := alert.CommitSHA
	if len(commitShort) > 7 {
		commitShort = commitShort[:7]
	}

	baselineText := fmt.Sprintf("Base: %s ➔ PR: %s", alert.BaselineStatus, alert.AnomalyType)
	if !alert.IsRegression {
		baselineText = fmt.Sprintf("Status: %s", alert.AnomalyType)
	}

	payload := map[string]interface{}{
		"username":   "ChaosSQL Concurrency Sentinel",
		"avatar_url": "https://chaossql.bregalda.com/brand/icone_bregalda.svg",
		"embeds": []map[string]interface{}{
			{
				"title":       title,
				"description": fmt.Sprintf("Uma nova anomalia concorrente foi identificada no repositório **%s**.", alert.RepoFullName),
				"color":       color,
				"fields": []map[string]interface{}{
					{"name": "📦 Repositório", "value": fmt.Sprintf("`%s`", alert.RepoFullName), "inline": true},
					{"name": "🌿 Branch / PR", "value": fmt.Sprintf("`%s` (%s)", alert.Branch, prText), "inline": true},
					{"name": "🔗 Commit", "value": fmt.Sprintf("`%s`", commitShort), "inline": true},
					{"name": "💥 Anomalia Detectada", "value": fmt.Sprintf("**%s** (%s)", alert.AnomalyName, alert.AnomalyType), "inline": true},
					{"name": "🗄️ Motor & Isolamento", "value": fmt.Sprintf("%s • %s", alert.Driver, alert.Isolation), "inline": true},
					{"name": "📊 Comparação de Baseline", "value": fmt.Sprintf("**%s**", baselineText), "inline": true},
					{"name": "🔍 Cenário & Seed", "value": fmt.Sprintf("`%s` (seed: `%d`)", alert.Scenario, alert.Seed), "inline": true},
					{"name": "⏱️ Duração", "value": fmt.Sprintf("%dms", alert.DurationMS), "inline": true},
				},
				"footer": map[string]interface{}{
					"text":     "ChaosSQL Concurrency Observability • Studio Bregalda",
					"icon_url": "https://chaossql.bregalda.com/brand/icone_bregalda.svg",
				},
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	return d.postJSON(ctx, url, payload)
}

func (d *WebhookDispatcher) sendSlack(ctx context.Context, url string, alert *RegressionAlert) error {
	prText := "Branch"
	if alert.PRNumber > 0 {
		prText = fmt.Sprintf("PR #%d", alert.PRNumber)
	}

	headline := "🚨 *ChaosSQL: Concurrency Regression Detected!*"
	if !alert.IsRegression {
		headline = "⚠️ *ChaosSQL: Concurrency Anomaly Detected*"
	}

	payload := map[string]interface{}{
		"text": headline,
		"blocks": []map[string]interface{}{
			{
				"type": "section",
				"text": map[string]interface{}{
					"type": "mrkdwn",
					"text": fmt.Sprintf("%s\n*Repository:* `%s` | *Target:* `%s` (%s)", headline, alert.RepoFullName, alert.Branch, prText),
				},
			},
			{
				"type": "section",
				"fields": []map[string]interface{}{
					{"type": "mrkdwn", "text": fmt.Sprintf("*Anomaly:* %s (%s)", alert.AnomalyName, alert.AnomalyType)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Engine:* %s (%s)", alert.Driver, alert.Isolation)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Baseline:* %s ➔ Broken", alert.BaselineStatus)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Scenario:* `%s` (seed: %d)", alert.Scenario, alert.Seed)},
				},
			},
		},
	}

	return d.postJSON(ctx, url, payload)
}

func (d *WebhookDispatcher) sendGeneric(ctx context.Context, url string, alert *RegressionAlert) error {
	payload := map[string]interface{}{
		"event":     "concurrency.regression",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"alert":     alert,
	}
	return d.postJSON(ctx, url, payload)
}

func (d *WebhookDispatcher) postJSON(ctx context.Context, url string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ChaosSQL-Webhook-Dispatcher/1.5")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook endpoint returned status: %d", resp.StatusCode)
	}
	return nil
}

// ValidateWebhookTarget rejects caller-supplied webhook destinations that would
// turn the dispatcher into an SSRF relay. Callers register arbitrary URLs, and
// the server fetches them with its own network position, so a URL pointing at
// loopback, link-local or RFC1918 space could reach cloud metadata services and
// internal admin endpoints that are otherwise unreachable from the internet.
//
// Public HTTPS endpoints of any provider stay allowed, so custom "generic"
// webhooks keep working.
//
// Note: the hostname is resolved here, at registration time. A hostname that
// resolves to a public address now but to an internal one at dispatch time
// (DNS rebinding) is not caught by this check alone; closing that gap requires
// a guarded dialer on the dispatcher's transport.
func ValidateWebhookTarget(rawURL string) error {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return fmt.Errorf("webhook URL is required")
	}

	parsed, err := neturl.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("malformed webhook URL")
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("webhook URL must use https")
	}
	if parsed.User != nil {
		return fmt.Errorf("webhook URL must not contain credentials")
	}

	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("webhook URL must contain a host")
	}

	// A literal IP is checked directly; a name is resolved so that a hostname
	// pointing at internal space is rejected too.
	var addrs []net.IP
	if ip := net.ParseIP(host); ip != nil {
		addrs = []net.IP{ip}
	} else {
		resolved, err := lookupIP(host)
		if err != nil || len(resolved) == 0 {
			return fmt.Errorf("webhook host could not be resolved")
		}
		addrs = resolved
	}

	for _, ip := range addrs {
		if isInternalIP(ip) {
			return fmt.Errorf("webhook URL must not target an internal address")
		}
	}

	return nil
}

// lookupIP is a package-level indirection so tests can resolve hosts without
// depending on live DNS.
var lookupIP = net.LookupIP

func isInternalIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast() {
		return true
	}
	// 100.64.0.0/10 (CGNAT) is not covered by IsPrivate but is routinely used
	// for internal service meshes.
	if v4 := ip.To4(); v4 != nil && v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
		return true
	}
	return false
}
