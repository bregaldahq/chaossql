package main

import (
	"errors"
	"fmt"

	"github.com/bregaldahq/chaossql/internal/cloud"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/engine"
)

// Exit codes follow grep and diff: 1 means the tool found what it looks for,
// 2 means it could not reach a trustworthy answer.
const (
	exitOK      = 0
	exitFinding = 1
	exitTrouble = 2
)

// --fail-on values for run.
const (
	failOnViolation  = "violation"
	failOnRegression = "regression"
	failOnNever      = "never"
)

// findingError is a trustworthy result the caller asked to fail on: an
// invariant violation, a Cloud regression, or a divergence between engines.
type findingError struct{ msg string }

func (e *findingError) Error() string { return e.msg }

// exitCode maps a command error to the process exit status. A joined error
// exits with its most serious part, so a finding plus a Cloud failure is 2.
func exitCode(err error) int {
	if err == nil {
		return exitOK
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		code := exitOK
		for _, part := range joined.Unwrap() {
			code = max(code, exitCode(part))
		}
		return code
	}
	var finding *findingError
	if errors.As(err, &finding) {
		return exitFinding
	}
	return exitTrouble
}

// checkFailOn rejects an unknown --fail-on value, and regression without a
// Cloud token, before anything runs.
func checkFailOn(failOn, cloudToken string) error {
	switch failOn {
	case failOnViolation, failOnNever:
		return nil
	case failOnRegression:
		if cloudToken == "" {
			return errors.New("--fail-on regression needs ChaosSQL Cloud to compare against the baseline: set --cloud-token or CHAOSSQL_CLOUD_TOKEN")
		}
		return nil
	default:
		return fmt.Errorf("invalid --fail-on %q: use violation, regression or never", failOn)
	}
}

// runOutcomeError decides how a finished run exits. Unreliable runs always
// exit 2; after that, failOn picks which trustworthy result fails the command.
// cloudResp is the Cloud answer, nil when the run was not recorded.
func runOutcomeError(result *engine.RunResult, anomaly domain.AnomalyType, failOn string, cloudResp *cloud.RunIngestResponse) error {
	if err := unreliableRunError(result); err != nil {
		return err
	}
	switch failOn {
	case failOnNever:
		return nil
	case failOnRegression:
		if cloudResp == nil {
			return errors.New("--fail-on regression could not get a verdict: the run was not recorded by ChaosSQL Cloud")
		}
		if cloudResp.IsRegression {
			return &findingError{msg: "concurrency regression against the default-branch baseline (use --fail-on never to exit 0 on findings)"}
		}
		return nil
	default:
		if result == nil || result.Status != domain.StatusViolation {
			return nil
		}
		invariant := "an invariant"
		if result.FailingInvariant != nil {
			invariant = fmt.Sprintf("invariant %q", result.FailingInvariant.Name)
		}
		return &findingError{msg: fmt.Sprintf("%s violated (%s) (use --fail-on never to exit 0 on findings)", invariant, anomaly)}
	}
}
