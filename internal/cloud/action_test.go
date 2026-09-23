package cloud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestActionOutputsRejectLineInjectionWithoutPartialWrite(t *testing.T) {
	for _, malicious := range []string{"run\nextra-output=secret", "run\rextra-output=secret"} {
		path := filepath.Join(t.TempDir(), "outputs")
		if err := WriteActionOutputs(path, &RunIngestResponse{RunID: malicious, URL: "https://example.com"}); err == nil {
			t.Fatal("accepted injected output")
		}
		if data, _ := os.ReadFile(path); len(data) != 0 {
			t.Fatalf("partial outputs: %s", data)
		}
	}
}

func TestActionOutputsAppend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "outputs")
	_ = os.WriteFile(path, []byte("existing=value\n"), 0600)
	if err := WriteActionOutputs(path, &RunIngestResponse{RunID: "run", URL: "https://example.com", IsRegression: true}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(data), "existing=value\ncloud-run-id=run\n") || !strings.Contains(string(data), "is-regression=true\n") {
		t.Fatalf("incorrect outputs %s", data)
	}
}
