package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/bregaldahq/chaossql/pkg/mutator"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type VariantSummary struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	File  string `json:"file"`
}

type MutateSummary struct {
	OriginalScenario string           `json:"original_scenario"`
	VariantsCount    int              `json:"variants_count"`
	OutputDir        string           `json:"output_dir"`
	Seed             uint64           `json:"seed"`
	Variants         []VariantSummary `json:"variants"`
}

func newMutateCmd() *cobra.Command {
	var (
		variantsFlag  int
		outputDirFlag string
		seedFlagVal   uint64
		jsonOutput    bool
	)

	cmd := &cobra.Command{
		Use:   "mutate <scenario.yaml>",
		Short: "Generate adversarial and stochastic mutations of a chaos testing specification",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scenarioPath := args[0]

			data, err := os.ReadFile(scenarioPath)
			if err != nil {
				return fmt.Errorf("failed to read scenario file %q: %w", scenarioPath, err)
			}

			spec, err := domain.ParseSpecBytes(data)
			if err != nil {
				return fmt.Errorf("failed to parse scenario specification: %w", err)
			}

			opts := mutator.MutationOptions{
				NumVariants:              variantsFlag,
				Seed:                     seedFlagVal,
				EnableJitterInjection:    true,
				EnableSavepointInjection: true,
				EnableStepShuffle:        true,
				EnableLockInversion:      true,
			}

			variants, err := mutator.MutateScenario(*spec, opts)
			if err != nil {
				return fmt.Errorf("scenario mutation failed: %w", err)
			}

			if err := os.MkdirAll(outputDirFlag, 0755); err != nil {
				return fmt.Errorf("failed to create output directory %s: %w", outputDirFlag, err)
			}

			// Copy schema.sql and seed.sql if referenced as external files
			srcDir := filepath.Dir(scenarioPath)
			if spec.Database.Schema != "" && strings.HasSuffix(strings.TrimSpace(spec.Database.Schema), ".sql") {
				schemaSrc := filepath.Join(srcDir, spec.Database.Schema)
				if content, err := os.ReadFile(schemaSrc); err == nil {
					_ = os.WriteFile(filepath.Join(outputDirFlag, spec.Database.Schema), content, 0644)
				}
			}
			if spec.Database.Seed != "" && strings.HasSuffix(strings.TrimSpace(spec.Database.Seed), ".sql") {
				seedSrc := filepath.Join(srcDir, spec.Database.Seed)
				if content, err := os.ReadFile(seedSrc); err == nil {
					_ = os.WriteFile(filepath.Join(outputDirFlag, spec.Database.Seed), content, 0644)
				}
			}

			summary := MutateSummary{
				OriginalScenario: spec.Name,
				VariantsCount:    len(variants),
				OutputDir:        outputDirFlag,
				Seed:             seedFlagVal,
				Variants:         make([]VariantSummary, 0, len(variants)),
			}

			for i, v := range variants {
				yamlBytes, err := yaml.Marshal(v)
				if err != nil {
					return fmt.Errorf("failed to encode variant %d to YAML: %w", i, err)
				}

				outFile := filepath.Join(outputDirFlag, fmt.Sprintf("variant_%d.yaml", i))
				if err := os.WriteFile(outFile, yamlBytes, 0644); err != nil {
					return fmt.Errorf("failed to write variant file %s: %w", outFile, err)
				}

				summary.Variants = append(summary.Variants, VariantSummary{
					Index: i,
					Name:  v.Name,
					File:  outFile,
				})
			}

			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(summary)
			}

			cmd.Println(reporter.RenderBanner())
			renderMutateTerminal(cmd, summary)

			return nil
		},
	}

	cmd.Flags().IntVar(&variantsFlag, "variants", 5, "Number of mutated scenario variants to generate")
	cmd.Flags().StringVar(&outputDirFlag, "output-dir", "./mutated", "Output directory path for generated YAML variants")
	cmd.Flags().Uint64Var(&seedFlagVal, "seed", 42, "Deterministic PRNG seed for mutation reproducibility")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output summary payload as structured JSON")

	return cmd
}

func renderMutateTerminal(cmd *cobra.Command, summary MutateSummary) {
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	cmd.Println(headerStyle.Render(fmt.Sprintf("ADVERSARIAL SCENARIO MUTATOR -- %s", summary.OriginalScenario)))

	var lines string
	lines += fmt.Sprintf("  * Original Scenario: %s\n", summary.OriginalScenario)
	lines += fmt.Sprintf("  * Seed: %d\n", summary.Seed)
	lines += fmt.Sprintf("  * Variants Generated: %d\n", summary.VariantsCount)
	lines += fmt.Sprintf("  * Output Directory: %s\n\n", summary.OutputDir)

	statusStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
	lines += statusStyle.Render("  GENERATED MUTATIONS:\n")
	lines += "  ----------------------------------------------------------------------------\n"

	for _, v := range summary.Variants {
		lines += fmt.Sprintf("  + [%02d] %-35s -> %s\n", v.Index, v.Name, v.File)
	}

	cmd.Println(cardStyle.Render(lines))
}