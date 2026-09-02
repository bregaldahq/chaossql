package reporter

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
)

type JUnitTestSuites struct {
	XMLName   xml.Name         `xml:"testsuites"`
	TestSuite []JUnitTestSuite `xml:"testsuite"`
}

type JUnitTestSuite struct {
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Time      float64         `xml:"time,attr"`
	TestCases []JUnitTestCase `xml:"testcase"`
}

type JUnitTestCase struct {
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      float64       `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
}

type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// GenerateJUnitXML converts chaos test execution results to standard JUnit XML format.
func GenerateJUnitXML(spec domain.Spec, res domain.ExecutionResult, anomaly domain.AnomalyType) string {
	ts := JUnitTestSuite{
		Name:  fmt.Sprintf("chaossql.%s", spec.Name),
		Tests: 1,
		Time:  res.Duration.Seconds(),
	}

	tc := JUnitTestCase{
		Name:      fmt.Sprintf("isolation_invariants_%s", spec.Name),
		Classname: fmt.Sprintf("chaossql.drivers.%s", spec.Database.Driver),
		Time:      res.Duration.Seconds(),
	}

	if res.ViolationDetected {
		ts.Failures = 1
		msg := "Isolation Anomaly Detected"
		if res.FailingInvariant != nil {
			msg = res.FailingInvariant.String()
		}
		tc.Failure = &JUnitFailure{
			Message: msg,
			Type:    string(anomaly),
			Content: fmt.Sprintf("Scenario: %s\nDriver: %s\nAnomaly: %s\nDetails: %s", spec.Name, spec.Database.Driver, anomaly, msg),
		}
	}

	ts.TestCases = append(ts.TestCases, tc)
	suites := JUnitTestSuites{
		TestSuite: []JUnitTestSuite{ts},
	}

	output, err := xml.MarshalIndent(suites, "", "  ")
	if err != nil {
		return fmt.Sprintf("<!-- Failed to marshal XML: %v -->", err)
	}

	return xml.Header + string(output)
}

// GenerateGitHubSummaryMarkdown creates a rich GitHub Actions Step Summary markdown report.
func GenerateGitHubSummaryMarkdown(spec domain.Spec, res domain.ExecutionResult, shrink *domain.ShrinkResult, anomaly domain.AnomalyType) string {
	statusEmoji := "✅"
	statusBadge := "INVARIANTS SATISFIED"
	if res.ViolationDetected {
		statusEmoji = "❌"
		statusBadge = fmt.Sprintf("ISOLATION ANOMALY DETECTED [%s]", anomaly)
	}

	md := fmt.Sprintf("# %s ChaosSQL Execution Summary: `%s`\n\n", statusEmoji, spec.Name)
	md += "| Attribute | Value |\n"
	md += "| :--- | :--- |\n"
	md += fmt.Sprintf("| **Status** | `%s` |\n", statusBadge)
	md += fmt.Sprintf("| **Database Driver** | `%s` |\n", spec.Database.Driver)
	md += fmt.Sprintf("| **Workers / Ops** | %d workers / %d ops |\n", spec.Engine.Workers, spec.Engine.Iterations)
	md += fmt.Sprintf("| **Seed** | `%d` |\n", spec.Engine.Seed)
	md += fmt.Sprintf("| **Duration** | `%v` |\n\n", res.Duration.Round(time.Millisecond))

	if shrink != nil && res.ViolationDetected {
		md += "## 🔬 Causal Delta-Debugging Reduction ($ddmin$)\n\n"
		md += fmt.Sprintf("- **Noise Reduction:** %d ops ──► **%d minimal ops** (**%.1f%% reduction**)\n",
			shrink.OriginalSize, shrink.ReducedSize, shrink.ReductionRatio)
		md += fmt.Sprintf("- **Iterations / Cost:** %d iterations in %v\n\n", shrink.Iterations, shrink.Duration.Round(time.Millisecond))
	}

	if res.FailingInvariant != nil {
		md += "## ⚠️ Invariant Violation Details\n\n"
		md += fmt.Sprintf("```\n%s\n```\n", res.FailingInvariant.String())
	}

	return md
}
