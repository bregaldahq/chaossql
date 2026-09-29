package domain

import (
	"fmt"
	"io/fs"
	"os"
	"path"
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
	baseDir := filepath.Dir(filePath)
	return loadSpec(filePath, os.ReadFile, func(name string) string { return filepath.Join(baseDir, name) })
}

// LoadSpecFS is LoadSpec for a file inside fsys (for example the embedded
// examples); schema and seed files resolve next to the spec within fsys.
func LoadSpecFS(fsys fs.FS, name string) (*Spec, error) {
	read := func(p string) ([]byte, error) { return fs.ReadFile(fsys, p) }
	baseDir := path.Dir(name)
	return loadSpec(name, read, func(file string) string { return path.Join(baseDir, file) })
}

func loadSpec(specPath string, read func(string) ([]byte, error), resolve func(string) string) (*Spec, error) {
	data, err := read(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	spec, err := ParseSpecBytes(data)
	if err != nil {
		return nil, err
	}

	if spec.Database.Schema != "" && strings.HasSuffix(strings.TrimSpace(spec.Database.Schema), ".sql") {
		schemaPath := resolve(spec.Database.Schema)
		schemaData, err := read(schemaPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema file: %w", err)
		}
		spec.Database.Schema = string(schemaData)
	}

	if spec.Database.Seed != "" && strings.HasSuffix(strings.TrimSpace(spec.Database.Seed), ".sql") {
		seedPath := resolve(spec.Database.Seed)
		seedData, err := read(seedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read seed file: %w", err)
		}
		spec.Database.Seed = string(seedData)
	}

	return spec, nil
}
