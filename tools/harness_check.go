package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	skillsDir   = ".claude/skills"
	flowMapPath = "docs/harness/flow-map.md"
	claudeMD    = "CLAUDE.md"
)

var (
	// Skill names referenced in the flow map and CLAUDE.md.
	skillRefPattern = regexp.MustCompile("`(chaossql-[a-z0-9-]+)`")
	// Backticked tokens inside a skill's "## Source map" section.
	backtickPattern = regexp.MustCompile("`([^`]+)`")
)

func main() {
	requiredFiles := []string{
		"AGENTS.md",
		"ARCHITECTURE.md",
		claudeMD,
		flowMapPath,
		"CONTRIBUTING.md",
		"LICENSE",
		"SECURITY.md",
		"Makefile",
		"README.md",
		"action.yml",
		"site/index.html",
		"site/app.js",
		"site/assets/style.css",
		"site/wrangler.toml",
		".github/workflows/ci.yml",
		".github/workflows/deploy-pages.yml",
		".github/workflows/static-pages.yml",
		"docs/THEORY.md",
		"docs/ACADEMIC_FOUNDATIONS.md",
		"docs/SCENARIO_ACADEMIC_AUDIT.md",
		"docs/adrs/0001-deterministic-prng-and-replay.md",
		"docs/adrs/0002-async-step-interleaving.md",
		"docs/adrs/0003-causal-delta-debugging-shrinker.md",
		"docs/adrs/0004-pure-go-sqlite-vs-cgo.md",
		"docs/adrs/0005-repro-test-standalone-synthesis.md",
		"docs/adrs/0006-bubbletea-terminal-ux.md",
		"evals/01_shrinking_ratio.md",
		"evals/02_false_positive_rate.md",
		"evals/03_deterministic_replay.md",
		"specs/01_invariant_evaluation.md",
		"specs/02_concurrency_interleaving.md",
		"specs/03_delta_debugging_shrinker.md",
		"specs/04_evidence_synthesis.md",
		"specs/05_advanced_anomaly_taxonomy.md",
		"specs/06_mysql_savepoints_and_otel.md",
		"specs/07_differential_fuzzing_and_matrix.md",
		"specs/08_fault_injection_and_dirty_reads.md",
		"specs/09_temporal_invariants_and_g2_cycles.md",
		"specs/10_developer_tooling_and_static_validator.md",
		"specs/11_version_1_1_developer_sdk_and_smart_generators.md",
		"specs/12_documentation_portal_and_branding_site.md",
		"specs/13_interactive_visualizer_sarif_and_register_checker.md",
		"specs/14_wasm_in_browser_playground.md",
		"specs/15_multiagent_qa_and_swarm_fuzzing.md",
		"specs/16_transparent_database_proxy.md",
		"specs/17_multi_language_sdks.md",
		"CHANGELOG.md",
		"site/assets/wasm-worker.js",
		"site/assets/wasm_exec.js",
		"examples/foreign_key_cascade_deadlock/chaos.yaml",
		"tools/test_english_purity.js",
	}

	missing := 0
	for _, f := range requiredFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			fmt.Printf("❌ [MISSING] Required Harness artifact: %s\n", f)
			missing++
		}
	}

	if missing > 0 {
		fmt.Printf("\n[ERROR] %d artifacts missing in Harness.\n", missing)
		os.Exit(1)
	}

	skills, problems := checkSkills()
	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Printf("❌ [SKILL] %s\n", p)
		}
		fmt.Printf("\n[ERROR] %d skill harness problem(s). See the chaossql-skill-maintenance skill.\n", len(problems))
		os.Exit(1)
	}

	fmt.Printf("[HARNESS OK] All %d Harness artifacts are present and verified.\n", len(requiredFiles))
	fmt.Printf("[SKILLS OK] %d agent skills are registered, indexed, and their source maps resolve.\n", len(skills))
}

// checkSkills validates the agent skill harness: every skill directory has a
// well-formed SKILL.md, is listed in the flow map and CLAUDE.md, and every path
// in its "## Source map" section exists. It returns the skill names found.
func checkSkills() ([]string, []string) {
	var problems []string

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, []string{fmt.Sprintf("cannot read %s: %v", skillsDir, err)}
	}

	var skills []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		skills = append(skills, name)
		problems = append(problems, checkSkillFile(name)...)
	}
	sort.Strings(skills)

	onDisk := make(map[string]bool, len(skills))
	for _, s := range skills {
		onDisk[s] = true
	}

	for _, doc := range []string{flowMapPath, claudeMD} {
		content, err := os.ReadFile(doc)
		if err != nil {
			problems = append(problems, fmt.Sprintf("cannot read %s: %v", doc, err))
			continue
		}
		referenced := make(map[string]bool)
		for _, m := range skillRefPattern.FindAllStringSubmatch(string(content), -1) {
			referenced[m[1]] = true
		}
		for _, s := range skills {
			if !referenced[s] {
				problems = append(problems, fmt.Sprintf("skill %q is not listed in %s", s, doc))
			}
		}
		for ref := range referenced {
			if !onDisk[ref] {
				problems = append(problems, fmt.Sprintf("%s references missing skill %q", doc, ref))
			}
		}
	}

	sort.Strings(problems)
	return skills, problems
}

func checkSkillFile(name string) []string {
	var problems []string
	path := filepath.Join(skillsDir, name, "SKILL.md")
	file, err := os.Open(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: missing SKILL.md", name)}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNo := 0
	inFrontmatter := false
	frontmatterDone := false
	frontName, frontDescription := "", ""
	inSourceMap := false
	sourceMapSeen := false

	for scanner.Scan() {
		line := scanner.Text()
		lineNo++

		if lineNo == 1 {
			if line != "---" {
				problems = append(problems, fmt.Sprintf("%s: SKILL.md must start with YAML frontmatter", name))
				return problems
			}
			inFrontmatter = true
			continue
		}
		if inFrontmatter {
			if line == "---" {
				inFrontmatter = false
				frontmatterDone = true
				continue
			}
			if v, ok := strings.CutPrefix(line, "name:"); ok {
				frontName = strings.TrimSpace(v)
			}
			if v, ok := strings.CutPrefix(line, "description:"); ok {
				frontDescription = strings.TrimSpace(v)
			}
			continue
		}

		if strings.HasPrefix(line, "## ") {
			inSourceMap = strings.TrimSpace(line) == "## Source map"
			if inSourceMap {
				sourceMapSeen = true
			}
			continue
		}
		if !inSourceMap {
			continue
		}
		for _, m := range backtickPattern.FindAllStringSubmatch(line, -1) {
			ref := m[1]
			if !looksLikeRepoPath(ref) {
				continue
			}
			if _, err := os.Stat(filepath.FromSlash(strings.TrimSuffix(ref, "/"))); err != nil {
				problems = append(problems, fmt.Sprintf("%s: source map path %q does not exist (line %d)", name, ref, lineNo))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		problems = append(problems, fmt.Sprintf("%s: read error: %v", name, err))
	}

	if !frontmatterDone {
		problems = append(problems, fmt.Sprintf("%s: unterminated frontmatter", name))
	}
	if frontName != name {
		problems = append(problems, fmt.Sprintf("%s: frontmatter name %q must match the directory name", name, frontName))
	}
	if frontDescription == "" {
		problems = append(problems, fmt.Sprintf("%s: frontmatter description is empty", name))
	}
	if !sourceMapSeen {
		problems = append(problems, fmt.Sprintf("%s: missing a \"## Source map\" section", name))
	}
	return problems
}

// looksLikeRepoPath accepts repository-relative file or directory paths and
// ignores commands, identifiers, and URLs that may also appear in backticks.
func looksLikeRepoPath(ref string) bool {
	if ref == "" || strings.ContainsAny(ref, " *?$<>{}()=:\"'") {
		return false
	}
	if strings.HasPrefix(ref, "/") || strings.HasPrefix(ref, "-") {
		return false
	}
	return strings.Contains(ref, "/") || filepath.Ext(ref) != ""
}
