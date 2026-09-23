package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestOfficialActionPassesInputsAsDataAndPreservesExit(t *testing.T) {
	data, err := os.ReadFile("../../action.yml")
	if err != nil {
		t.Fatal(err)
	}
	var action struct {
		Outputs map[string]struct {
			Value string `yaml:"value"`
		} `yaml:"outputs"`
		Runs struct {
			Steps []struct {
				ID   string            `yaml:"id"`
				Name string            `yaml:"name"`
				Run  string            `yaml:"run"`
				Env  map[string]string `yaml:"env"`
			} `yaml:"steps"`
		} `yaml:"runs"`
	}
	if err := yaml.Unmarshal(data, &action); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"cloud-run-id", "cloud-run-url", "is-regression"} {
		if action.Outputs[key].Value != "${{ steps.execute.outputs."+key+" }}" {
			t.Errorf("missing declared output %s", key)
		}
	}
	script := ""
	var stepEnv map[string]string
	for _, step := range action.Runs.Steps {
		if strings.Contains(step.Run, "${{") {
			t.Errorf("%s embeds expressions in shell source", step.Name)
		}
		if step.ID == "execute" {
			script = step.Run
			stepEnv = step.Env
		}
	}
	if script == "" {
		t.Fatal("no execute step with output ID")
	}
	dir := t.TempDir()
	record := filepath.Join(dir, "argv")
	marker := filepath.Join(dir, "injected")
	mock := "#!/bin/bash\nprintf '%s\\0' \"$@\" > \"$RECORD_ARGS\"\nexit 23\n"
	if err := os.WriteFile(filepath.Join(dir, "chaossql"), []byte(mock), 0700); err != nil {
		t.Fatal(err)
	}
	malicious := "input with spaces; $(touch " + marker + ") ' \" end"
	env := append(os.Environ(), "RUNNER_TEMP="+dir, "RECORD_ARGS="+record, "GITHUB_STEP_SUMMARY=")
	for key, value := range stepEnv {
		switch value {
		case "${{ inputs.spec-path }}", "${{ inputs.export-html }}":
			env = append(env, key+"="+malicious)
		case "${{ inputs.cloud-fail-fast }}":
			env = append(env, key+"=true")
		case "${{ inputs.post-pr-comment }}":
			env = append(env, key+"=false")
		default:
			env = append(env, key+"=")
		}
	}
	cmd := exec.Command("bash", "-e", "-u", "-o", "pipefail", "-c", script)
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 23 {
		t.Fatalf("Action did not preserve CLI exit: %v %s", err, output)
	}
	args, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	want := "run\x00" + malicious + "\x00--export-html\x00" + malicious + "\x00--cloud-fail-fast\x00--pr-comment=false\x00"
	if string(args) != want {
		t.Fatalf("inputs changed or split: %q want %q", args, want)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("input executed as shell code")
	}
}
