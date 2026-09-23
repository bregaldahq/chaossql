package cloud

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestDetectCICommitTimeUsesExactSHA(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2020-01-02T03:04:05Z", "GIT_COMMITTER_DATE=2020-01-02T03:04:05Z")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git: %s %v", out, err)
		}
		return strings.TrimSpace(string(out))
	}
	git("init")
	git("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "test")
	sha := git("rev-parse", "HEAD")
	t.Chdir(dir)
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_SHA", sha)
	ci := DetectCIContext()
	expected := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if ci.CommitTimestamp == nil || !ci.CommitTimestamp.Equal(expected) {
		t.Fatalf("got %v, want exact commit time", ci.CommitTimestamp)
	}
	t.Setenv("GITHUB_SHA", strings.Repeat("f", 40))
	if got := DetectCIContext().CommitTimestamp; got != nil {
		t.Fatalf("unknown SHA got timestamp %v", got)
	}
	t.Setenv("GITHUB_SHA", "--all")
	if got := DetectCIContext().CommitTimestamp; got != nil {
		t.Fatalf("invalid SHA got timestamp %v", got)
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

func TestSanitizeRepositoryRemoteRemovesCredentialsAndURLDetails(t *testing.T) {
	tests := []struct {
		remote string
		want   string
	}{
		{"https://alice:secret@gitlab.example.com/group/repo.git?token=other#fragment", "gitlab.example.com/group/repo"},
		{"ssh://git:secret@gitlab.example.com/group/repo.git", "gitlab.example.com/group/repo"},
		{"git@github.com:bregaldahq/chaossql.git", "bregaldahq/chaossql"},
		{"https://token@github.com/bregaldahq/chaossql.git", "bregaldahq/chaossql"},
		{"/home/alice/private/repository", "local/repository"},
	}
	for _, test := range tests {
		if got := sanitizeRepositoryRemote(test.remote); got != test.want {
			t.Errorf("sanitizeRepositoryRemote(%q) = %q, want %q", test.remote, got, test.want)
		}
	}
}
