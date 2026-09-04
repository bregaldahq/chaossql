package reporter

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
)

// OASIS SARIF 2.1.0 Constants
const (
	SarifSchemaURI = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"
	SarifVersion   = "2.1.0"
	ToolName       = "ChaosSQL"
	ToolVersion    = "1.2.0"
	ToolInfoURI    = "https://chaossql.bregalda.com"
)

// SarifReport represents the top-level OASIS SARIF 2.1.0 log container.
type SarifReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SarifRun `json:"runs"`
}

// SarifRun represents a single static or dynamic analysis execution run.
type SarifRun struct {
	Tool      SarifTool       `json:"tool"`
	Artifacts []SarifArtifact `json:"artifacts,omitempty"`
	Results   []SarifResult   `json:"results"`
}

// SarifTool encapsulates the analysis tool metadata and driver.
type SarifTool struct {
	Driver SarifDriver `json:"driver"`
}

// SarifDriver describes the analysis engine and its standardized rule catalog.
type SarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version,omitempty"`
	InformationURI string      `json:"informationUri,omitempty"`
	Rules          []SarifRule `json:"rules"`
}

// SarifRule represents a rule (ReportingDescriptor) in the driver's catalog.
type SarifRule struct {
	ID                   string                  `json:"id"`
	Name                 string                  `json:"name"`
	ShortDescription     SarifMultiformatMessage `json:"shortDescription"`
	FullDescription      SarifMultiformatMessage `json:"fullDescription"`
	DefaultConfiguration SarifConfiguration      `json:"defaultConfiguration"`
	Help                 *SarifHelpMessage       `json:"help,omitempty"`
	Properties           map[string]interface{}  `json:"properties,omitempty"`
}

// SarifMultiformatMessage contains formatted message text.
type SarifMultiformatMessage struct {
	Text string `json:"text"`
}

// SarifHelpMessage provides plain-text and markdown documentation for a rule.
type SarifHelpMessage struct {
	Text     string `json:"text,omitempty"`
	Markdown string `json:"markdown,omitempty"`
}

// SarifConfiguration defines default level and settings for a rule.
type SarifConfiguration struct {
	Level string `json:"level"` // "error", "warning", "note", "none"
}

// SarifResult represents an individual finding/violation.
type SarifResult struct {
	RuleID    string          `json:"ruleId"`
	RuleIndex int             `json:"ruleIndex"`
	Level     string          `json:"level"`
	Message   SarifMessage    `json:"message"`
	Locations []SarifLocation `json:"locations"`
}

// SarifMessage represents a result message with plain text and rich markdown.
type SarifMessage struct {
	Text     string `json:"text"`
	Markdown string `json:"markdown,omitempty"`
}

// SarifLocation indicates the location of a finding.
type SarifLocation struct {
	PhysicalLocation SarifPhysicalLocation `json:"physicalLocation"`
}

// SarifPhysicalLocation specifies the file artifact and line region.
type SarifPhysicalLocation struct {
	ArtifactLocation SarifArtifactLocation `json:"artifactLocation"`
	Region           SarifRegion           `json:"region"`
}

// SarifArtifactLocation points to the file URI.
type SarifArtifactLocation struct {
	URI       string `json:"uri"`
	URIBaseID string `json:"uriBaseId,omitempty"`
}

// SarifRegion specifies the start and end coordinates.
type SarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn,omitempty"`
}

// SarifArtifact defines an artifact tracked in the run.
type SarifArtifact struct {
	Location SarifArtifactLocation `json:"location"`
}

// StandardRulesCatalog returns the standardized OASIS SARIF rules for ChaosSQL.
func StandardRulesCatalog() []SarifRule {
	tags := []string{"security", "concurrency", "isolation-anomaly", "database"}
	return []SarifRule{
		{
			ID:   "chaossql/P4-lost-update",
			Name: "LostUpdateAnomaly",
			ShortDescription: SarifMultiformatMessage{
				Text: "Lost Update Concurrency Isolation Anomaly (P4)",
			},
			FullDescription: SarifMultiformatMessage{
				Text: "A transaction reads a state and overwrites it based on that state, but a concurrent transaction also overwrites it, causing the first transaction's update to be silently lost.",
			},
			DefaultConfiguration: SarifConfiguration{Level: "error"},
			Help: &SarifHelpMessage{
				Text:     "Remediation: Upgrade isolation level to SERIALIZABLE or use explicit row locking (SELECT ... FOR UPDATE).",
				Markdown: "### Remediation\n- Upgrade transaction isolation to `SERIALIZABLE`\n- Use pessimistic locking: `SELECT ... FOR UPDATE`\n- Implement optimistic concurrency control via version checking (`WHERE version = :expected_version`)",
			},
			Properties: map[string]interface{}{"tags": tags},
		},
		{
			ID:   "chaossql/A5B-write-skew",
			Name: "WriteSkewAnomaly",
			ShortDescription: SarifMultiformatMessage{
				Text: "Write Skew Concurrency Isolation Anomaly (A5B)",
			},
			FullDescription: SarifMultiformatMessage{
				Text: "Two concurrent transactions read overlapping data sets satisfying an invariant, and each makes disjoint writes that together violate the global integrity constraint.",
			},
			DefaultConfiguration: SarifConfiguration{Level: "error"},
			Help: &SarifHelpMessage{
				Text:     "Remediation: Use SERIALIZABLE isolation or lock all records evaluated in the integrity constraint.",
				Markdown: "### Remediation\n- Enforce `SERIALIZABLE` isolation\n- Use predicate locking or materialise conflicting sets",
			},
			Properties: map[string]interface{}{"tags": tags},
		},
		{
			ID:   "chaossql/A5A-read-skew",
			Name: "ReadSkewAnomaly",
			ShortDescription: SarifMultiformatMessage{
				Text: "Read Skew Concurrency Isolation Anomaly (A5A)",
			},
			FullDescription: SarifMultiformatMessage{
				Text: "A transaction reads item X, a concurrent transaction updates X and Y, and the first transaction subsequently reads item Y, observing an inconsistent database snapshot.",
			},
			DefaultConfiguration: SarifConfiguration{Level: "warning"},
			Help: &SarifHelpMessage{
				Text:     "Remediation: Upgrade to REPEATABLE READ or SNAPSHOT ISOLATION.",
				Markdown: "### Remediation\n- Upgrade to `REPEATABLE READ` or `SNAPSHOT ISOLATION` to guarantee point-in-time snapshot reads",
			},
			Properties: map[string]interface{}{"tags": tags},
		},
		{
			ID:   "chaossql/G0-dirty-write",
			Name: "DirtyWriteAnomaly",
			ShortDescription: SarifMultiformatMessage{
				Text: "Dirty Write Concurrency Isolation Anomaly (G0)",
			},
			FullDescription: SarifMultiformatMessage{
				Text: "A transaction modifies a data item that has already been modified by another concurrent uncommitted transaction, violating the basic isolation guarantee.",
			},
			DefaultConfiguration: SarifConfiguration{Level: "error"},
			Help: &SarifHelpMessage{
				Text:     "Remediation: Enforce exclusive write locks on all modified tuples until transaction commit.",
				Markdown: "### Remediation\n- Databases must hold long-duration exclusive write locks until commit (strict 2PL)",
			},
			Properties: map[string]interface{}{"tags": tags},
		},
		{
			ID:   "chaossql/G1a-dirty-read",
			Name: "DirtyReadAnomaly",
			ShortDescription: SarifMultiformatMessage{
				Text: "Dirty Read / Aborted Read Anomaly (G1a)",
			},
			FullDescription: SarifMultiformatMessage{
				Text: "A transaction reads uncommitted changes made by another transaction that subsequently aborts, corrupting transaction state.",
			},
			DefaultConfiguration: SarifConfiguration{Level: "error"},
			Help: &SarifHelpMessage{
				Text:     "Remediation: Upgrade isolation level to READ COMMITTED or higher.",
				Markdown: "### Remediation\n- Upgrade from `READ UNCOMMITTED` to `READ COMMITTED` or higher",
			},
			Properties: map[string]interface{}{"tags": tags},
		},
		{
			ID:   "chaossql/G1b-intermediate-read",
			Name: "IntermediateReadAnomaly",
			ShortDescription: SarifMultiformatMessage{
				Text: "Intermediate Read Anomaly (G1b)",
			},
			FullDescription: SarifMultiformatMessage{
				Text: "A transaction reads an intermediate value produced by another transaction that subsequently updates that value again before committing.",
			},
			DefaultConfiguration: SarifConfiguration{Level: "error"},
			Help: &SarifHelpMessage{
				Text:     "Remediation: Upgrade isolation level to prevent observing intermediate uncommitted states.",
				Markdown: "### Remediation\n- Enforce atomic multi-version visibility barriers or strict 2PL",
			},
			Properties: map[string]interface{}{"tags": tags},
		},
		{
			ID:   "chaossql/G1c-circular-info",
			Name: "CircularInformationFlowAnomaly",
			ShortDescription: SarifMultiformatMessage{
				Text: "Circular Information Flow Anomaly (G1c)",
			},
			FullDescription: SarifMultiformatMessage{
				Text: "A cycle of write-to-read dependencies exists between transactions, indicating that neither transaction preceded the other in serialization order.",
			},
			DefaultConfiguration: SarifConfiguration{Level: "error"},
			Help: &SarifHelpMessage{
				Text:     "Remediation: Upgrade to SERIALIZABLE isolation to enforce acyclic serialization graphs.",
				Markdown: "### Remediation\n- Upgrade to `SERIALIZABLE` isolation",
			},
			Properties: map[string]interface{}{"tags": tags},
		},
		{
			ID:   "chaossql/G2-anti-dependency",
			Name: "AntiDependencyCycleAnomaly",
			ShortDescription: SarifMultiformatMessage{
				Text: "Anti-Dependency Cycle Anomaly (G2)",
			},
			FullDescription: SarifMultiformatMessage{
				Text: "A cycle containing anti-dependency edges (read-write conflicts) exists across transactions, violating strict serializability.",
			},
			DefaultConfiguration: SarifConfiguration{Level: "error"},
			Help: &SarifHelpMessage{
				Text:     "Remediation: Use SERIALIZABLE isolation with SSI (Serializable Snapshot Isolation) or strict 2PL.",
				Markdown: "### Remediation\n- Enable Serializable Snapshot Isolation (SSI) or strict two-phase locking",
			},
			Properties: map[string]interface{}{"tags": tags},
		},
		{
			ID:   "chaossql/G-DL-deadlock",
			Name: "DeadlockCycleAnomaly",
			ShortDescription: SarifMultiformatMessage{
				Text: "Deadlock Cycle and Lock Contention (G-DL)",
			},
			FullDescription: SarifMultiformatMessage{
				Text: "Concurrent transactions encountered mutual lock dependencies causing transaction deadlocks and aborted recovery.",
			},
			DefaultConfiguration: SarifConfiguration{Level: "warning"},
			Help: &SarifHelpMessage{
				Text:     "Remediation: Order table/row updates deterministically across all transactions to prevent circular wait conditions.",
				Markdown: "### Remediation\n- Enforce a strict global lock acquisition order across all transactions\n- Keep transaction scopes small and minimize lock duration",
			},
			Properties: map[string]interface{}{"tags": tags},
		},
		{
			ID:   "chaossql/A3-phantom-read",
			Name: "PhantomReadAnomaly",
			ShortDescription: SarifMultiformatMessage{
				Text: "Phantom Read Concurrency Isolation Anomaly (A3)",
			},
			FullDescription: SarifMultiformatMessage{
				Text: "A transaction executes a range query and a concurrent transaction inserts or deletes tuples satisfying the predicate.",
			},
			DefaultConfiguration: SarifConfiguration{Level: "error"},
			Help: &SarifHelpMessage{
				Text:     "Remediation: Enforce predicate locking or SERIALIZABLE isolation.",
				Markdown: "### Remediation\n- Use range/predicate locks or `SERIALIZABLE` isolation",
			},
			Properties: map[string]interface{}{"tags": tags},
		},
		{
			ID:   "chaossql/unknown-invariant-violation",
			Name: "InvariantViolationAnomaly",
			ShortDescription: SarifMultiformatMessage{
				Text: "Unknown Concurrency Invariant Violation",
			},
			FullDescription: SarifMultiformatMessage{
				Text: "Database state invariant was violated under concurrent interleavings.",
			},
			DefaultConfiguration: SarifConfiguration{Level: "error"},
			Help: &SarifHelpMessage{
				Text:     "Remediation: Inspect failing invariant assertion and interleaving trace.",
				Markdown: "### Remediation\n- Investigate concurrent transaction interleaving trace",
			},
			Properties: map[string]interface{}{"tags": tags},
		},
	}
}

// GenerateSARIFReport creates a compliant OASIS SARIF 2.1.0 JSON report.
func GenerateSARIFReport(
	spec domain.Spec,
	results []domain.InvariantResult,
	graph *analyzer.AdyaGraph,
	shrink *domain.ShrinkResult,
) (string, error) {
	rules := StandardRulesCatalog()
	ruleIndexMap := make(map[string]int, len(rules))
	for i, r := range rules {
		ruleIndexMap[r.ID] = i
	}

	specURI := resolveSpecURI(spec)

	sarifResults := make([]SarifResult, 0)

	// 1. Detect if Deadlock cycle occurred
	isDeadlock := detectDeadlock(spec, results, graph)

	// 2. Detect isolation anomaly / invariant violation
	hasFailedInvariant := false
	var firstFailedInv *domain.InvariantResult
	for i := range results {
		if !results[i].Passed {
			hasFailedInvariant = true
			if firstFailedInv == nil {
				firstFailedInv = &results[i]
			}
		}
	}

	var cycles []analyzer.Cycle
	if graph != nil {
		cycles = analyzer.FindCycles(graph)
	}

	if isDeadlock {
		ruleID := "chaossql/G-DL-deadlock"
		ruleIdx := ruleIndexMap[ruleID]
		rule := rules[ruleIdx]

		textMsg := fmt.Sprintf("Deadlock cycle detected in scenario '%s': concurrent transactions formed a circular lock dependency.", spec.Name)
		mdMsg := formatDeadlockMarkdown(spec, cycles, shrink)

		sarifResults = append(sarifResults, SarifResult{
			RuleID:    ruleID,
			RuleIndex: ruleIdx,
			Level:     rule.DefaultConfiguration.Level,
			Message: SarifMessage{
				Text:     textMsg,
				Markdown: mdMsg,
			},
			Locations: []SarifLocation{
				{
					PhysicalLocation: SarifPhysicalLocation{
						ArtifactLocation: SarifArtifactLocation{
							URI:       specURI,
							URIBaseID: "%SRCROOT%",
						},
						Region: SarifRegion{
							StartLine:   1,
							StartColumn: 1,
						},
					},
				},
			},
		})
	} else if hasFailedInvariant || len(cycles) > 0 {
		anomaly := classifyAnomaly(spec, firstFailedInv, cycles)
		ruleID := mapAnomalyToRuleID(anomaly, spec)
		ruleIdx, ok := ruleIndexMap[ruleID]
		if !ok {
			ruleIdx = ruleIndexMap["chaossql/P4-lost-update"]
		}
		rule := rules[ruleIdx]

		invDesc := ""
		if firstFailedInv != nil {
			invDesc = fmt.Sprintf(" Invariant '%s' violated (%s).", firstFailedInv.Name, firstFailedInv.Expression)
		}
		textMsg := fmt.Sprintf("Concurrency isolation anomaly '%s' detected in scenario '%s'.%s", ruleID, spec.Name, invDesc)
		mdMsg := formatAnomalyMarkdown(ruleID, spec, firstFailedInv, cycles, shrink)

		sarifResults = append(sarifResults, SarifResult{
			RuleID:    ruleID,
			RuleIndex: ruleIdx,
			Level:     rule.DefaultConfiguration.Level,
			Message: SarifMessage{
				Text:     textMsg,
				Markdown: mdMsg,
			},
			Locations: []SarifLocation{
				{
					PhysicalLocation: SarifPhysicalLocation{
						ArtifactLocation: SarifArtifactLocation{
							URI:       specURI,
							URIBaseID: "%SRCROOT%",
						},
						Region: SarifRegion{
							StartLine:   1,
							StartColumn: 1,
						},
					},
				},
			},
		})
	}

	report := SarifReport{
		Schema:  SarifSchemaURI,
		Version: SarifVersion,
		Runs: []SarifRun{
			{
				Tool: SarifTool{
					Driver: SarifDriver{
						Name:           ToolName,
						Version:        ToolVersion,
						InformationURI: ToolInfoURI,
						Rules:          rules,
					},
				},
				Artifacts: []SarifArtifact{
					{
						Location: SarifArtifactLocation{
							URI:       specURI,
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

func resolveSpecURI(spec domain.Spec) string {
	name := strings.TrimSpace(spec.Name)
	if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
		return name
	}
	if strings.Contains(name, "/") {
		return name + "/chaos.yaml"
	}
	return fmt.Sprintf("examples/%s/chaos.yaml", name)
}

func detectDeadlock(spec domain.Spec, results []domain.InvariantResult, graph *analyzer.AdyaGraph) bool {
	lowerName := strings.ToLower(spec.Name)
	if strings.Contains(lowerName, "deadlock") {
		return true
	}
	for _, r := range results {
		if strings.Contains(strings.ToLower(r.Name), "deadlock") {
			return true
		}
		if r.Error != nil && strings.Contains(strings.ToLower(r.Error.Error()), "deadlock") {
			return true
		}
	}
	return false
}

func classifyAnomaly(spec domain.Spec, failedInv *domain.InvariantResult, cycles []analyzer.Cycle) domain.AnomalyType {
	if len(cycles) > 0 {
		return analyzer.ClassifyCycle(cycles[0])
	}

	lower := strings.ToLower(spec.Name)
	if failedInv != nil {
		lower += " " + strings.ToLower(failedInv.Name) + " " + strings.ToLower(failedInv.Expression)
	}

	switch {
	case strings.Contains(lower, "lost_update") || strings.Contains(lower, "p4"):
		return domain.AnomalyLostUpdate
	case strings.Contains(lower, "oversell") || strings.Contains(lower, "phantom") || strings.Contains(lower, "a3"):
		return domain.AnomalyPhantom
	case strings.Contains(lower, "write_skew") || strings.Contains(lower, "a5b"):
		return domain.AnomalyWriteSkew
	case strings.Contains(lower, "read_skew") || strings.Contains(lower, "a5a"):
		return domain.AnomalyA5AReadSkew
	case strings.Contains(lower, "dirty_write") || strings.Contains(lower, "g0"):
		return domain.AnomalyG0DirtyWrite
	case strings.Contains(lower, "intermediate") || strings.Contains(lower, "g1b"):
		return domain.AnomalyG1bIntermediateRead
	case strings.Contains(lower, "dirty_read") || strings.Contains(lower, "g1a") || strings.Contains(lower, "flash_crash"):
		return domain.AnomalyG1aDirtyRead
	case strings.Contains(lower, "circular") || strings.Contains(lower, "g1c") || strings.Contains(lower, "arbitrage"):
		return domain.AnomalyG1cCircularInfo
	case strings.Contains(lower, "anti_dependency") || strings.Contains(lower, "ticket") || strings.Contains(lower, "g2"):
		return domain.AnomalyG2AntiDependency
	default:
		return domain.AnomalyLostUpdate
	}
}

func mapAnomalyToRuleID(anomaly domain.AnomalyType, spec domain.Spec) string {
	switch anomaly {
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
		lower := strings.ToLower(spec.Name)
		if strings.Contains(lower, "write_skew") {
			return "chaossql/A5B-write-skew"
		}
		if strings.Contains(lower, "read_skew") {
			return "chaossql/A5A-read-skew"
		}
		if strings.Contains(lower, "dirty_write") {
			return "chaossql/G0-dirty-write"
		}
		if strings.Contains(lower, "dirty_read") {
			return "chaossql/G1a-dirty-read"
		}
		if strings.Contains(lower, "circular") {
			return "chaossql/G1c-circular-info"
		}
		if strings.Contains(lower, "anti_dependency") || strings.Contains(lower, "ticket") {
			return "chaossql/G2-anti-dependency"
		}
		return "chaossql/P4-lost-update"
	}
}

func formatAnomalyMarkdown(
	ruleID string,
	spec domain.Spec,
	failedInv *domain.InvariantResult,
	cycles []analyzer.Cycle,
	shrink *domain.ShrinkResult,
) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### 💥 Concurrency Isolation Anomaly: `%s`\n\n", ruleID))
	sb.WriteString(fmt.Sprintf("- **Scenario:** `%s`\n", spec.Name))
	sb.WriteString(fmt.Sprintf("- **Database Driver:** `%s`\n", spec.Database.Driver))
	if spec.Engine.Workers > 0 {
		sb.WriteString(fmt.Sprintf("- **Concurrency:** %d workers, %d iterations (seed: `%d`)\n\n",
			spec.Engine.Workers, spec.Engine.Iterations, spec.Engine.Seed))
	}

	if failedInv != nil {
		sb.WriteString("#### ⚠️ Invariant Violation Details\n")
		sb.WriteString(fmt.Sprintf("- **Name:** `%s`\n", failedInv.Name))
		sb.WriteString(fmt.Sprintf("- **Assert Expression:** `%s`\n", failedInv.Expression))
		if len(failedInv.ActualValues) > 0 {
			sb.WriteString(fmt.Sprintf("- **Observed Database State:** `%v`\n\n", failedInv.ActualValues))
		} else {
			sb.WriteString("\n")
		}
	}

	if shrink != nil && len(shrink.MinimalOps) > 0 {
		sb.WriteString("#### 🔬 Minimal Causal Reproducing Operations ($ddmin$)\n")
		sb.WriteString(fmt.Sprintf("Reduced from **%d** operations down to **%d minimal operations** (**%.1f%% reduction**) in %d iterations (%v):\n\n",
			shrink.OriginalSize, shrink.ReducedSize, shrink.ReductionRatio, shrink.Iterations, shrink.Duration.Round(time.Millisecond)))

		for i, op := range shrink.MinimalOps {
			sb.WriteString(fmt.Sprintf("%d. **Op %d** (`%s`):\n", i+1, op.ID, op.Name))
			for _, step := range op.Steps {
				sb.WriteString(fmt.Sprintf("   ```sql\n   %s\n   ```\n", strings.TrimSpace(step.SQL)))
			}
		}
		sb.WriteString("\n")
	}

	if len(cycles) > 0 {
		sb.WriteString("#### 🔄 Adya Dependency Cycle\n```\n")
		for _, edge := range cycles[0] {
			sb.WriteString(fmt.Sprintf("%s ──(%s on %s)──► %s\n", edge.From, edge.Type, edge.Item, edge.To))
		}
		sb.WriteString("```\n\n")
	}

	sb.WriteString("#### 🛡️ Remediation\n")
	sb.WriteString("- Enforce `SERIALIZABLE` transaction isolation or appropriate row/predicate locks.\n")

	return sb.String()
}

func formatDeadlockMarkdown(spec domain.Spec, cycles []analyzer.Cycle, shrink *domain.ShrinkResult) string {
	var sb strings.Builder
	sb.WriteString("### ⚠️ Deadlock Cycle & Lock Contention: `chaossql/G-DL-deadlock`\n\n")
	sb.WriteString(fmt.Sprintf("- **Scenario:** `%s`\n", spec.Name))
	sb.WriteString(fmt.Sprintf("- **Database Driver:** `%s`\n\n", spec.Database.Driver))
	sb.WriteString("Concurrent transactions attempted to acquire mutually conflicting locks, producing a circular wait condition that triggered transaction abort and rollback.\n\n")

	if len(cycles) > 0 {
		sb.WriteString("#### 🔄 Circular Wait Graph\n```\n")
		for _, edge := range cycles[0] {
			sb.WriteString(fmt.Sprintf("%s ──(locks %s)──► %s\n", edge.From, edge.Item, edge.To))
		}
		sb.WriteString("```\n\n")
	}

	if shrink != nil && len(shrink.MinimalOps) > 0 {
		sb.WriteString("#### 🔬 Minimal Reproducing Interleaving Operations\n")
		for i, op := range shrink.MinimalOps {
			sb.WriteString(fmt.Sprintf("%d. **Op %d** (`%s`):\n", i+1, op.ID, op.Name))
			for _, step := range op.Steps {
				sb.WriteString(fmt.Sprintf("   ```sql\n   %s\n   ```\n", strings.TrimSpace(step.SQL)))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("#### 🛡️ Remediation\n")
	sb.WriteString("- Ensure deterministic lock acquisition ordering across all transactions (e.g. sort records by primary key before updating).\n")
	sb.WriteString("- Keep transaction boundaries as short as possible.\n")

	return sb.String()
}
