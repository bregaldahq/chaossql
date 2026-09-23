package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/bregaldahq/chaossql/internal/version"
)

func TestVersionFlagReportsReleaseVersion(t *testing.T) {
	cmd := newRootCmd()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), version.Version) {
		t.Fatalf("--version output %q does not contain %s", out.String(), version.Version)
	}
	if reporter.ToolVersion != version.Version {
		t.Fatalf("SARIF tool version %q differs from release %s", reporter.ToolVersion, version.Version)
	}
}
