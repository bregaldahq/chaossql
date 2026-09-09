package cloud

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractPRNumber(t *testing.T) {
	tests := []struct {
		name      string
		ref       string
		eventJSON string
		expected  int
	}{
		{
			name:     "from ref merge",
			ref:      "refs/pull/42/merge",
			expected: 42,
		},
		{
			name:     "from ref head",
			ref:      "refs/pull/108/head",
			expected: 108,
		},
		{
			name:      "from event json",
			ref:       "refs/heads/feature",
			eventJSON: `{"pull_request": {"number": 382}}`,
			expected:  382,
		},
		{
			name:     "no pr",
			ref:      "refs/heads/main",
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var eventFile string
			if tc.eventJSON != "" {
				tmp := filepath.Join(t.TempDir(), "event.json")
				if err := os.WriteFile(tmp, []byte(tc.eventJSON), 0644); err != nil {
					t.Fatalf("failed to write tmp event file: %v", err)
				}
				eventFile = tmp
			}

			actual := ExtractPRNumber(tc.ref, eventFile)
			if actual != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, actual)
			}
		})
	}
}

func TestDetectCIContextGitHub(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_REPOSITORY", "bregaldahq/chaossql")
	t.Setenv("GITHUB_SHA", "e386ca712839")
	t.Setenv("GITHUB_REF", "refs/pull/99/merge")
	t.Setenv("GITHUB_HEAD_REF", "feat/my-pr")
	t.Setenv("GITHUB_BASE_REF", "main")
	t.Setenv("GITHUB_RUN_ID", "192849182")
	t.Setenv("GITHUB_ACTOR", "octocat")

	ci := DetectCIContext()
	if ci == nil {
		t.Fatal("expected non-nil CIContext")
	}

	if ci.Provider != "github-actions" {
		t.Errorf("expected provider github-actions, got %q", ci.Provider)
	}
	if ci.Repository != "bregaldahq/chaossql" {
		t.Errorf("expected repo bregaldahq/chaossql, got %q", ci.Repository)
	}
	if ci.PullRequestNumber != 99 {
		t.Errorf("expected PR number 99, got %d", ci.PullRequestNumber)
	}
	if ci.Branch != "feat/my-pr" {
		t.Errorf("expected branch feat/my-pr, got %q", ci.Branch)
	}
	if ci.BaseBranch != "main" {
		t.Errorf("expected base branch main, got %q", ci.BaseBranch)
	}
}
