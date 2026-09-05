package reporter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bregaldahq/chaossql/internal/swarm"
)

// GenerateSwarmMarkdownSummary synthesizes a GitHub Flavored Markdown summary from
// a DifferentialReport, designed for $GITHUB_STEP_SUMMARY and pull request audits.
func GenerateSwarmMarkdownSummary(report *swarm.DifferentialReport) string {
	if report == nil {
		return "# 🌪️ ChaosSQL Multi-Engine Swarm Differential Audit\n\nNo differential swarm report available.\n"
	}

	var sb strings.Builder

	// 1. Header
	sb.WriteString("# 🌪️ ChaosSQL Multi-Engine Swarm Differential Audit\n\n")

	// Collect unique target drivers
	driverMap := make(map[string]bool)
	for _, sc := range report.Scenarios {
		for d := range sc.Results {
			driverMap[d] = true
		}
	}
	var targetEngines []string
	for d := range driverMap {
		targetEngines = append(targetEngines, d)
	}
	sort.Strings(targetEngines)

	targetEnginesDisplay := "None"
	cliDrivers := "sqlite,mock"
	if len(targetEngines) > 0 {
		var quoted []string
		for _, d := range targetEngines {
			quoted = append(quoted, fmt.Sprintf("`%s`", d))
		}
		targetEnginesDisplay = strings.Join(quoted, ", ")
		cliDrivers = strings.Join(targetEngines, ",")
	}

	// Determine overall status badge
	overallStatus := "✅ CONSISTENT"
	allEnginesFailed := len(report.Scenarios) > 0
	for _, sc := range report.Scenarios {
		hasValid := false
		for _, res := range sc.Results {
			if res.Error == "" {
				hasValid = true
				break
			}
		}
		if hasValid {
			allEnginesFailed = false
			break
		}
	}

	if len(report.Scenarios) == 0 {
		overallStatus = "⚪ EMPTY"
	} else if allEnginesFailed {
		overallStatus = "⚠️ ERROR / ALL ENGINES OFFLINE"
	} else if report.DivergentCount > 0 {
		overallStatus = fmt.Sprintf("❌ DIVERGENT (%d detected)", report.DivergentCount)
	}

	// 2. Executive Metrics Table
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("| :--- | :--- |\n")
	sb.WriteString(fmt.Sprintf("| **Overall Status** | %s |\n", overallStatus))
	sb.WriteString(fmt.Sprintf("| **Total Scenarios** | %d |\n", report.TotalScenarios))
	sb.WriteString(fmt.Sprintf("| **Total Engine Executions** | %d |\n", report.TotalExecutions))
	sb.WriteString(fmt.Sprintf("| **Divergences Detected** | %d |\n", report.DivergentCount))
	sb.WriteString(fmt.Sprintf("| **Target Engines** | %s |\n", targetEnginesDisplay))
	sb.WriteString(fmt.Sprintf("| **Duration** | %dms |\n\n", report.DurationMs))

	// 3. Matrix Results Table
	sb.WriteString("## 📊 Differential Matrix Results\n\n")
	sb.WriteString("| Scenario | Status | Engine Breakdown |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")

	if len(report.Scenarios) == 0 {
		sb.WriteString("| *(None)* | ⚪ EMPTY | No scenarios evaluated |\n\n")
	} else {
		for _, sc := range report.Scenarios {
			// Sort drivers for deterministic rendering
			var scenarioDrivers []string
			for d := range sc.Results {
				scenarioDrivers = append(scenarioDrivers, d)
			}
			sort.Strings(scenarioDrivers)

			validCount := 0
			var breakdownParts []string
			for _, d := range scenarioDrivers {
				res := sc.Results[d]
				if res.Error != "" {
					cleanErr := strings.ReplaceAll(res.Error, "|", "\\|")
					breakdownParts = append(breakdownParts, fmt.Sprintf("**%s**: ⚠️ Error (%s)", d, cleanErr))
				} else {
					validCount++
					if res.ViolationDetected {
						anomalyDesc := res.DetectedAnomaly
						if anomalyDesc == "" {
							anomalyDesc = res.FailingInvariant
						} else if res.FailingInvariant != "" && res.FailingInvariant != anomalyDesc {
							anomalyDesc = fmt.Sprintf("%s, %s", anomalyDesc, res.FailingInvariant)
						}
						if anomalyDesc == "" {
							anomalyDesc = "violation"
						}
						breakdownParts = append(breakdownParts, fmt.Sprintf("**%s**: ❌ Violation (%s)", d, anomalyDesc))
					} else {
						breakdownParts = append(breakdownParts, fmt.Sprintf("**%s**: ✅ Safe", d))
					}
				}
			}

			var statusStr string
			if validCount == 0 {
				statusStr = "⚠️ `ERROR`"
			} else if validCount == 1 {
				statusStr = "ℹ️ `INCONCLUSIVE`"
			} else if sc.Divergent {
				statusStr = "❌ `DIVERGENT`"
			} else {
				statusStr = "✅ `CONSISTENT`"
			}

			breakdownCell := strings.Join(breakdownParts, "<br>")
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", sc.ScenarioName, statusStr, breakdownCell))
		}
		sb.WriteString("\n")
	}

	// 4. Detailed Divergence Section
	sb.WriteString("## 🔬 Divergence Analysis\n\n")
	if report.DivergentCount == 0 {
		sb.WriteString("> [!NOTE]\n")
		sb.WriteString("> No semantic divergences detected across target engines. All evaluated engines exhibited consistent transaction isolation behavior.\n\n")
	} else {
		for _, sc := range report.Scenarios {
			if !sc.Divergent {
				continue
			}

			sb.WriteString(fmt.Sprintf("### Scenario: `%s`\n\n", sc.ScenarioName))
			sb.WriteString("> [!WARNING]\n")
			sb.WriteString(fmt.Sprintf("> **Semantic Divergence Detected**: %s\n>\n", sc.Summary))
			sb.WriteString("> **Engine Differences**:\n")

			var scenarioDrivers []string
			for d := range sc.Results {
				scenarioDrivers = append(scenarioDrivers, d)
			}
			sort.Strings(scenarioDrivers)

			for _, d := range scenarioDrivers {
				res := sc.Results[d]
				if res.Error != "" {
					sb.WriteString(fmt.Sprintf("> - **%s**: ⚠️ Execution error: `%s`\n", d, res.Error))
				} else if res.ViolationDetected {
					desc := res.DetectedAnomaly
					if desc == "" {
						desc = res.FailingInvariant
					}
					invDetail := ""
					if res.FailingInvariant != "" {
						invDetail = fmt.Sprintf(" (invariant: `%s`)", res.FailingInvariant)
					}
					sb.WriteString(fmt.Sprintf("> - **%s**: ❌ Detected anomaly `%s`%s\n", d, desc, invDetail))
				} else {
					sb.WriteString(fmt.Sprintf("> - **%s**: ✅ Invariants satisfied (safe isolation)\n", d))
				}
			}
			sb.WriteString("\n")
		}
	}

	// 5. Reproducibility Section
	sb.WriteString("## 🔁 Local Reproduction\n\n")
	if report.DivergentCount > 0 {
		sb.WriteString("To replicate the detected divergences locally:\n\n")
		sb.WriteString("```bash\n")
		for _, sc := range report.Scenarios {
			if sc.Divergent {
				sb.WriteString(fmt.Sprintf("# Reproduce divergence in %s:\n", sc.ScenarioName))
				sb.WriteString(fmt.Sprintf("./bin/chaossql swarm diff --scenarios-dir ./examples/%s --drivers %s\n\n", sc.ScenarioName, cliDrivers))
			}
		}
		sb.WriteString("# Re-run full differential swarm:\n")
		sb.WriteString(fmt.Sprintf("./bin/chaossql swarm diff --scenarios-dir ./examples --drivers %s\n", cliDrivers))
		sb.WriteString("```\n")
	} else {
		sb.WriteString("To execute this differential swarm suite locally:\n\n")
		sb.WriteString("```bash\n")
		sb.WriteString(fmt.Sprintf("./bin/chaossql swarm diff --scenarios-dir ./examples --drivers %s\n", cliDrivers))
		sb.WriteString("```\n")
	}

	return sb.String()
}