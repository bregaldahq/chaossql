package reporter_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
)

func TestGenerateJUnitXML(t *testing.T) {
	spec := domain.Spec{
		Name: "banking_test",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
		},
	}
	resSuccess := domain.ExecutionResult{
		Success:  true,
		Duration: 100 * time.Millisecond,
	}

	xmlSuccess := reporter.GenerateJUnitXML(spec, resSuccess, "")
	if !strings.Contains(xmlSuccess, `<testsuites>`) || !strings.Contains(xmlSuccess, `tests="1"`) {
		t.Errorf("expected valid JUnit XML for success, got: %s", xmlSuccess)
	}

	resFailure := domain.ExecutionResult{
		Success:           false,
		ViolationDetected: true,
		Duration:          200 * time.Millisecond,
		FailingInvariant: &domain.InvariantResult{
			Name:       "balance_check",
			Expression: "balance == 1000",
		},
	}
	xmlFailure := reporter.GenerateJUnitXML(spec, resFailure, domain.AnomalyLostUpdate)
	if !strings.Contains(xmlFailure, `<failure`) || !strings.Contains(xmlFailure, `P4_LOST_UPDATE`) {
		t.Errorf("expected failure element in JUnit XML, got: %s", xmlFailure)
	}

	resError := domain.ExecutionResult{Status: domain.StatusExecutionError, Error: errors.New("statement failed")}
	xmlError := reporter.GenerateJUnitXML(spec, resError, domain.AnomalyUnknown)
	if !strings.Contains(xmlError, `errors="1"`) || !strings.Contains(xmlError, `<error`) || strings.Contains(xmlError, `<failure`) {
		t.Errorf("expected JUnit error element, got: %s", xmlError)
	}
}

func TestGenerateGitHubSummaryMarkdown(t *testing.T) {
	spec := domain.Spec{
		Name: "hospital_write_skew",
		Database: domain.DatabaseConfig{
			Driver: "postgres",
		},
		Engine: domain.EngineConfig{
			Workers:    4,
			Iterations: 20,
			Seed:       42,
		},
	}
	res := domain.ExecutionResult{
		Success:           false,
		ViolationDetected: true,
		Duration:          300 * time.Millisecond,
		FailingInvariant: &domain.InvariantResult{
			Name:       "active_doctors",
			Expression: "email != 'alice@example.com'",
			ActualValues: map[string]interface{}{
				"email": "alice@example.com",
				"dsn":   "postgres://alice:secret@db/private",
			},
		},
	}
	shrink := &domain.ShrinkResult{
		OriginalSize:   20,
		ReducedSize:    2,
		ReductionRatio: 90.0,
		Iterations:     4,
		Duration:       150 * time.Millisecond,
	}

	md := reporter.GenerateGitHubSummaryMarkdown(spec, res, shrink, domain.AnomalyWriteSkew)
	if !strings.Contains(md, "hospital_write_skew") || !strings.Contains(md, "90.0% reduction") {
		t.Errorf("expected markdown to contain scenario name and reduction, got: %s", md)
	}
	if !strings.Contains(md, "active_doctors") {
		t.Errorf("expected safe invariant name, got: %s", md)
	}
	for _, forbidden := range []string{"alice@example.com", "postgres://alice:secret@db/private", "email !="} {
		if strings.Contains(md, forbidden) {
			t.Errorf("GitHub summary contains forbidden detail %q: %s", forbidden, md)
		}
	}
}
