package version

import (
	"regexp"
	"testing"
)

func TestVersionIsSemantic(t *testing.T) {
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(Version) {
		t.Fatalf("version %q is not MAJOR.MINOR.PATCH", Version)
	}
}
