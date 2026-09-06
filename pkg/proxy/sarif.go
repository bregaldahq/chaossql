package proxy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
)

// GenerateProxySARIF builds an OASIS SARIF 2.1.0 report for intercepted proxy anomalies.
func GenerateProxySARIF(anomalies []DetectedAnomaly, sessionName string) (string, error) {
	rules := reporter.StandardRulesCatalog()
	ruleIndexMap := make(map[string]int, len(rules))
	for i, r := range rules {
		ruleIndexMap[r.ID] = i
	}

	sarifResults := make([]reporter.SarifResult, 0, len(anomalies))

	for _, a := range anomalies {
		ruleID := mapProxyAnomalyToRuleID(a.Type)
		ruleIdx, ok := ruleIndexMap[ruleID]
		if !ok {
			ruleIdx = ruleIndexMap["chaossql/P4-lost-update"]
		}
		rule := rules[ruleIdx]

		textMsg := fmt.Sprintf("Live Concurrency Anomaly '%s' detected by ChaosSQL transparent proxy.", a.Type)
		mdMsg := formatProxyAnomalyMarkdown(a, sessionName)

		sarifResults = append(sarifResults, reporter.SarifResult{
			RuleID:    ruleID,
			RuleIndex: ruleIdx,
			Level:     rule.DefaultConfiguration.Level,
			Message: reporter.SarifMessage{
				Text:     textMsg,
				Markdown: mdMsg,
			},
			Locations: []reporter.SarifLocation{
				{
					PhysicalLocation: reporter.SarifPhysicalLocation{
						ArtifactLocation: reporter.SarifArtifactLocation{
							URI:       fmt.Sprintf("proxy://%s", sessionName),
							URIBaseID: "%SRCROOT%",
						},
						Region: reporter.SarifRegion{
							StartLine:   1,
							StartColumn: 1,
						},
					},
				},
			},
		})
	}

	report := reporter.SarifReport{
		Schema:  reporter.SarifSchemaURI,
		Version: reporter.SarifVersion,
		Runs: []reporter.SarifRun{
			{
				Tool: reporter.SarifTool{
					Driver: reporter.SarifDriver{
						Name:           reporter.ToolName + " Proxy",
						Version:        reporter.ToolVersion,
						InformationURI: reporter.ToolInfoURI,
						Rules:          rules,
					},
				},
				Artifacts: []reporter.SarifArtifact{
					{
						Location: reporter.SarifArtifactLocation{
							URI:       fmt.Sprintf("proxy://%s", sessionName),
							URIBaseID: "%SRCROOT%",
						},
					},
				},
				Results: sarifResults,
			},
		},
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal SARIF JSON: %w", err)
	}

	return string(data), nil
}

// ExportSARIFToFile persists the SARIF report to the target file.
func ExportSARIFToFile(anomalies []DetectedAnomaly, sessionName, outputPath string) error {
	sarifJSON, err := GenerateProxySARIF(anomalies, sessionName)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, []byte(sarifJSON), 0644)
}

func mapProxyAnomalyToRuleID(t domain.AnomalyType) string {
	switch t {
	case domain.AnomalyLostUpdate:
		return "chaossql/P4-lost-update"
	case domain.AnomalyWriteSkew:
		return "chaossql/A5B-write-skew"
	case domain.AnomalyA5AReadSkew:
		return "chaossql/A5A-read-skew"
	case domain.AnomalyG0DirtyWrite:
		return "chaossql/G0-dirty-write"
	case domain.AnomalyG1aDirtyRead:
		return "chaossql/G1a-dirty-read"
	case domain.AnomalyG1bIntermediateRead:
		return "chaossql/G1b-intermediate-read"
	case domain.AnomalyG1cCircularInfo:
		return "chaossql/G1c-circular-info"
	case domain.AnomalyG2AntiDependency:
		return "chaossql/G2-anti-dependency"
	case domain.AnomalyPhantom:
		return "chaossql/A3-phantom-read"
	default:
		return "chaossql/P4-lost-update"
	}
}

func formatProxyAnomalyMarkdown(a DetectedAnomaly, sessionName string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### 💥 ChaosSQL Proxy Concurrency Anomaly: `%s`\n\n", a.Type))
	sb.WriteString(fmt.Sprintf("- **Proxy Session:** `%s`\n", sessionName))
	sb.WriteString(fmt.Sprintf("- **Detected At:** `%s`\n", a.DetectedAt.Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString(fmt.Sprintf("- **Transactions Involved:** `%s`\n", strings.Join(a.TxIDs, ", ")))
	sb.WriteString(fmt.Sprintf("- **Accessed Keys:** `%s`\n\n", strings.Join(a.Items, ", ")))
	sb.WriteString(fmt.Sprintf("#### ⚠️ Description\n%s\n\n", a.Description))

	if len(a.Cycle) > 0 {
		sb.WriteString("#### 🔄 Adya Dependency Cycle\n```\n")
		for _, edge := range a.Cycle {
			sb.WriteString(fmt.Sprintf("%s ──(%s on %s)──► %s\n", edge.From, edge.Type, edge.Item, edge.To))
		}
		sb.WriteString("```\n\n")
	}

	sb.WriteString("#### 🛡️ Remediation\n")
	sb.WriteString("- Enforce `SERIALIZABLE` transaction isolation or explicit pessimistic row locking (`SELECT ... FOR UPDATE`).\n")
	return sb.String()
}
