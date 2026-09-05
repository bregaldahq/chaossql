package domain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseSpecBytes parses raw YAML specification bytes without disk access.
func ParseSpecBytes(data []byte) (*Spec, error) {
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	// Validate mandatory fields
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	return &spec, nil
}

// ParseSpecString parses YAML content with optional inline schema and seed SQL strings.
func ParseSpecString(yamlContent, schemaSQL, seedSQL string) (*Spec, error) {
	spec, err := ParseSpecBytes([]byte(yamlContent))
	if err != nil {
		return nil, err
	}
	if schemaSQL != "" {
		spec.Database.Schema = schemaSQL
	}
	if seedSQL != "" {
		spec.Database.Seed = seedSQL
	}
	return spec, nil
}

// LoadSpec reads a YAML chaos testing specification from the given filePath,
// parses it, validates it, and resolves external SQL files for schema and seed.
func LoadSpec(filePath string) (*Spec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	spec, err := ParseSpecBytes(data)
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(filePath)

	if spec.Database.Schema != "" && strings.HasSuffix(strings.TrimSpace(spec.Database.Schema), ".sql") {
		schemaPath := filepath.Join(baseDir, spec.Database.Schema)
		schemaData, err := os.ReadFile(schemaPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema file: %w", err)
		}
		spec.Database.Schema = string(schemaData)
	}

	if spec.Database.Seed != "" && strings.HasSuffix(strings.TrimSpace(spec.Database.Seed), ".sql") {
		seedPath := filepath.Join(baseDir, spec.Database.Seed)
		seedData, err := os.ReadFile(seedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read seed file: %w", err)
		}
		spec.Database.Seed = string(seedData)
	}

	return spec, nil
}
