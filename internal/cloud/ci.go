package cloud

import (
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var prRefRegex = regexp.MustCompile(`refs/pull/([0-9]+)/(merge|head)`)

// DetectCIContext extracts CI environment metadata or falls back to local git
func DetectCIContext() *CIContext {
	// 1. GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		ref := os.Getenv("GITHUB_REF")
		eventPath := os.Getenv("GITHUB_EVENT_PATH")
		prNumber := ExtractPRNumber(ref, eventPath)

		branch := os.Getenv("GITHUB_HEAD_REF")
		if branch == "" {
			branch = os.Getenv("GITHUB_REF_NAME")
		}

		return &CIContext{
			Provider:          "github-actions",
			Repository:        os.Getenv("GITHUB_REPOSITORY"),
			CommitSHA:         os.Getenv("GITHUB_SHA"),
			Branch:            branch,
			BaseBranch:        os.Getenv("GITHUB_BASE_REF"),
			PullRequestNumber: prNumber,
			RunID:             os.Getenv("GITHUB_RUN_ID"),
			Actor:             os.Getenv("GITHUB_ACTOR"),
		}
	}

	// 2. GitLab CI
	if os.Getenv("GITLAB_CI") == "true" {
		prNum, _ := strconv.Atoi(os.Getenv("CI_MERGE_REQUEST_IID"))
		return &CIContext{
			Provider:          "gitlab-ci",
			Repository:        os.Getenv("CI_PROJECT_PATH"),
			CommitSHA:         os.Getenv("CI_COMMIT_SHA"),
			Branch:            os.Getenv("CI_COMMIT_REF_NAME"),
			BaseBranch:        os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME"),
			PullRequestNumber: prNum,
			RunID:             os.Getenv("CI_PIPELINE_ID"),
			Actor:             os.Getenv("GITLAB_USER_LOGIN"),
		}
	}

	// 3. Local Git Fallback
	commitSHA, branch := detectLocalGit()
	return &CIContext{
		Provider:   "local",
		Repository: detectLocalRepo(),
		CommitSHA:  commitSHA,
		Branch:     branch,
		Actor:      os.Getenv("USER"),
	}
}

// ExtractPRNumber attempts to parse the PR number from GITHUB_REF or GITHUB_EVENT_PATH
func ExtractPRNumber(ref, eventPath string) int {
	if ref != "" {
		matches := prRefRegex.FindStringSubmatch(ref)
		if len(matches) >= 2 {
			if n, err := strconv.Atoi(matches[1]); err == nil && n > 0 {
				return n
			}
		}
	}

	if eventPath != "" {
		data, err := os.ReadFile(eventPath)
		if err == nil {
			var event struct {
				Number      int `json:"number"`
				PullRequest struct {
					Number int `json:"number"`
				} `json:"pull_request"`
			}
			if err := json.Unmarshal(data, &event); err == nil {
				if event.PullRequest.Number > 0 {
					return event.PullRequest.Number
				}
				if event.Number > 0 {
					return event.Number
				}
			}
		}
	}

	return 0
}

func detectLocalGit() (string, string) {
	shaCmd := exec.Command("git", "rev-parse", "HEAD")
	shaOut, err := shaCmd.Output()
	sha := ""
	if err == nil {
		sha = strings.TrimSpace(string(shaOut))
	}

	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	branch := ""
	if err == nil {
		branch = strings.TrimSpace(string(branchOut))
	}

	return sha, branch
}

func detectLocalRepo() string {
	remoteCmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := remoteCmd.Output()
	if err != nil {
		return "local/repository"
	}
	raw := strings.TrimSpace(string(out))
	raw = strings.TrimSuffix(raw, ".git")

	if strings.HasPrefix(raw, "git@github.com:") {
		return strings.TrimPrefix(raw, "git@github.com:")
	}
	if strings.Contains(raw, "github.com/") {
		parts := strings.Split(raw, "github.com/")
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return raw
}
