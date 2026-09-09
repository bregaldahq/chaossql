package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// FormatPRMarkdown constructs a high-impact GitHub Markdown report
func FormatPRMarkdown(req *RunIngestRequest, resp *RunIngestResponse) string {
	var sb strings.Builder

	isRegression := resp != nil && resp.IsRegression
	isFailed := req.Result.Status == "failed" || !req.Result.Success

	if isRegression {
		sb.WriteString("## ChaosSQL ❌ Concurrency Regression Detected\n\n")
		sb.WriteString("> ⚠️ **This Pull Request introduces a concurrency violation that broke the default branch baseline.**\n\n")
	} else if isFailed {
		sb.WriteString("## ChaosSQL ⚠️ Concurrency Anomaly Detected\n\n")
		sb.WriteString("> Concurrency testing detected an invariant violation under parallel execution.\n\n")
	} else {
		sb.WriteString("## ChaosSQL ✅ Concurrency Verification Passed\n\n")
		sb.WriteString("> All interleaved schedules executed cleanly without concurrency anomalies.\n\n")
	}

	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("| :--- | :--- |\n")

	scenarioName := req.Scenario.Name
	if scenarioName == "" {
		scenarioName = "default"
	}
	sb.WriteString(fmt.Sprintf("| **Scenario** | `%s` |\n", scenarioName))

	if req.Scenario.Driver != "" {
		sb.WriteString(fmt.Sprintf("| **Database Engine** | `%s` |\n", req.Scenario.Driver))
	}

	if isFailed {
		anomaly := req.Result.AnomalyType
		if anomaly == "" {
			anomaly = "Invariant Violation"
		}
		sb.WriteString(fmt.Sprintf("| **Anomaly Type** | **%s** |\n", anomaly))
	}

	if resp != nil && resp.Baseline != nil {
		sb.WriteString(fmt.Sprintf("| **Baseline (`%s`)** | `%s` |\n", resp.Baseline.Branch, strings.ToUpper(resp.Baseline.Status)))
	}

	sb.WriteString(fmt.Sprintf("| **Current PR Status** | `%s` |\n", strings.ToUpper(req.Result.Status)))

	if req.Result.TotalSchedules > 0 {
		sb.WriteString(fmt.Sprintf("| **Schedules Tested** | %d (%d failed) |\n", req.Result.TotalSchedules, req.Result.FailedSchedules))
	}

	if req.Result.DurationMS > 0 {
		sb.WriteString(fmt.Sprintf("| **Execution Time** | %d ms |\n", req.Result.DurationMS))
	}

	sb.WriteString("\n")

	// Invariant violation details
	if req.Result.FailingInvariant != nil {
		sb.WriteString("### 🚨 Failing Invariant\n")
		sb.WriteString(fmt.Sprintf("- **Name:** `%s`\n", req.Result.FailingInvariant.Name))
		if req.Result.FailingInvariant.Assertion != "" {
			sb.WriteString(fmt.Sprintf("- **Assertion:** `%s`\n", req.Result.FailingInvariant.Assertion))
		}
		if req.Result.FailingInvariant.Actual != "" {
			sb.WriteString(fmt.Sprintf("- **Actual Result:** `%s`\n", req.Result.FailingInvariant.Actual))
		}
		sb.WriteString("\n")
	}

	// Minimal trace steps
	if req.Reproduction != nil && len(req.Reproduction.SanitizedMinimalTrace) > 0 {
		sb.WriteString("### 🔬 Minimal Causal Trace\n")
		sb.WriteString("```sql\n")
		for _, op := range req.Reproduction.SanitizedMinimalTrace {
			worker := op.Worker
			if worker == "" {
				worker = "Worker"
			}
			sb.WriteString(fmt.Sprintf("%s: %s\n", worker, op.SQL))
		}
		sb.WriteString("```\n\n")
	}

	// Cloud & Reproduction Action Links
	sb.WriteString("### 🔗 Links & Resources\n")
	if resp != nil && resp.URL != "" {
		sb.WriteString(fmt.Sprintf("- 🔍 [View Full Trace in ChaosSQL Cloud](%s)\n", resp.URL))
	}
	sb.WriteString("- ⚡ Reproduce locally: `chaossql run --seed " + fmt.Sprintf("%d", req.Scenario.Seed) + "`\n")

	return sb.String()
}

// WriteStepSummary appends markdown to GitHub Actions Step Summary if configured
func WriteStepSummary(summaryPath, markdown string) error {
	if summaryPath == "" {
		summaryPath = os.Getenv("GITHUB_STEP_SUMMARY")
	}
	if summaryPath == "" {
		return nil // Not in GitHub Actions or step summary disabled
	}

	f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_STEP_SUMMARY (%s): %w", summaryPath, err)
	}
	defer f.Close()

	if _, err := f.WriteString(markdown + "\n\n"); err != nil {
		return fmt.Errorf("failed to write to GITHUB_STEP_SUMMARY: %w", err)
	}

	return nil
}

// PostPRComment posts a comment to the GitHub PR using GitHub REST API
func PostPRComment(ctx context.Context, httpClient *http.Client, ghToken, repo string, prNumber int, body string) (string, error) {
	if ghToken == "" || repo == "" || prNumber <= 0 {
		return "", nil // Skipping comment: missing PR or token
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/comments", repo, prNumber)

	payload := map[string]string{"body": body}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal pr comment payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+ghToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ChaosSQL-Bot")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to dispatch pr comment request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var ghResp struct {
		ID      int64  `json:"id"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(respBody, &ghResp); err == nil && ghResp.HTMLURL != "" {
		return ghResp.HTMLURL, nil
	}

	return fmt.Sprintf("comment_%d", time.Now().Unix()), nil
}
