package domain

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadSpec reads a YAML chaos testing specification from the given filePath,
// parses it, validates it, and resolves external SQL files for schema and seed.
func LoadSpec(filePath string) (*Spec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	// Validate mandatory fields
	if spec.Version == "" {
		return nil, fmt.Errorf("%w: missing or empty 'version'", ErrSpecValidationFailed)
	}
	if spec.Name == "" {
		return nil, fmt.Errorf("%w: missing or empty 'name'", ErrSpecValidationFailed)
	}
	if spec.Database.Driver == "" {
		return nil, fmt.Errorf("%w: missing or empty 'database.driver'", ErrSpecValidationFailed)
	}
	if len(spec.Invariants) == 0 {
		return nil, fmt.Errorf("%w: 'invariants' must have at least one entry", ErrSpecValidationFailed)
	}
	if len(spec.Operations) == 0 {
		return nil, fmt.Errorf("%w: 'operations' must have at least one entry", ErrSpecValidationFailed)
	}

	baseDir := filepath.Dir(filePath)

	if spec.Database.Schema != "" {
		schemaPath := filepath.Join(baseDir, spec.Database.Schema)
		schemaData, err := os.ReadFile(schemaPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema file: %w", err)
		}
		spec.Database.Schema = string(schemaData)
	}

	if spec.Database.Seed != "" {
		seedPath := filepath.Join(baseDir, spec.Database.Seed)
		seedData, err := os.ReadFile(seedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read seed file: %w", err)
		}
		spec.Database.Seed = string(seedData)
	}

	return &spec, nil
}
