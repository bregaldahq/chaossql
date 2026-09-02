package domain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSpec_ValidExamples(t *testing.T) {
	examples := []string{
		"banking_lost_update",
		"inventory_oversell",
		"hospital_write_skew",
	}

	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			path := filepath.Join("..", "..", "examples", example, "chaos.yaml")
			spec, err := LoadSpec(path)
			if err != nil {
				t.Fatalf("Failed to load %s: %v", example, err)
			}
			if spec.Version == "" {
				t.Errorf("Missing version in %s", example)
			}
			if spec.Name != example {
				t.Errorf("Expected name %s, got %s", example, spec.Name)
			}
			if spec.Database.Driver == "" {
				t.Errorf("Missing driver in %s", example)
			}
			if len(spec.Invariants) == 0 {
				t.Errorf("No invariants loaded in %s", example)
			}
			if len(spec.Operations) == 0 {
				t.Errorf("No operations loaded in %s", example)
			}
			if spec.Database.Schema == "" && spec.Database.Driver != "" {
				// The examples contain schema files
				t.Errorf("Schema file content not loaded in %s", example)
			}
			if spec.Database.Seed == "" && spec.Database.Driver != "" {
				// The examples contain seed files
				t.Errorf("Seed file content not loaded in %s", example)
			}
		})
	}
}

func TestLoadSpec_Errors(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		yamlContent string
		setupFiles  func(string)
		expectError bool
		errorContains string
	}{
		{
			name: "Invalid YAML",
			yamlContent: "version: '1.0'\n  invalid: \n- unindented",
			expectError: true,
			errorContains: "failed to parse yaml",
		},
		{
			name: "Missing Version",
			yamlContent: `
name: test
database:
  driver: sqlite
invariants:
  - name: inv
operations:
  - name: op
`,
			expectError: true,
			errorContains: "missing or empty 'version'",
		},
		{
			name: "Missing Name",
			yamlContent: `
version: "1.0"
database:
  driver: sqlite
invariants:
  - name: inv
operations:
  - name: op
`,
			expectError: true,
			errorContains: "missing or empty 'name'",
		},
		{
			name: "Missing Driver",
			yamlContent: `
version: "1.0"
name: test
invariants:
  - name: inv
operations:
  - name: op
`,
			expectError: true,
			errorContains: "missing or empty 'database.driver'",
		},
		{
			name: "Missing Invariants",
			yamlContent: `
version: "1.0"
name: test
database:
  driver: sqlite
operations:
  - name: op
`,
			expectError: true,
			errorContains: "'invariants' must have at least one entry",
		},
		{
			name: "Missing Operations",
			yamlContent: `
version: "1.0"
name: test
database:
  driver: sqlite
invariants:
  - name: inv
`,
			expectError: true,
			errorContains: "'operations' must have at least one entry",
		},
		{
			name: "Missing Schema File",
			yamlContent: `
version: "1.0"
name: test
database:
  driver: sqlite
  schema: non_existent.sql
invariants:
  - name: inv
operations:
  - name: op
`,
			expectError: true,
			errorContains: "failed to read schema file",
		},
		{
			name: "Missing Seed File",
			yamlContent: `
version: "1.0"
name: test
database:
  driver: sqlite
  seed: non_existent.sql
invariants:
  - name: inv
operations:
  - name: op
`,
			expectError: true,
			errorContains: "failed to read seed file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specPath := filepath.Join(tmpDir, "chaos.yaml")
			err := os.WriteFile(specPath, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("failed to write temp spec file: %v", err)
			}

			if tt.setupFiles != nil {
				tt.setupFiles(tmpDir)
			}

			_, err = LoadSpec(specPath)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errorContains)
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	// simple implementation to avoid importing strings
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
