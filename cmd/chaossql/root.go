package main

import "github.com/spf13/cobra"

// newRootCmd constructs the primary ChaosSQL CLI command tree.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "chaossql",
		Short: "ChaosSQL: Deterministic Concurrency & Invariant Fuzzer for SQL Databases",
		Long: `ChaosSQL bridges chaos engineering with formal academic database research (PCT, Elle, Hermitage).
It injects stochastic interleavings across database worker threads to provoke subtle isolation anomalies
(such as Lost Updates, Write Skew, and Phantom Reads) and applies causal Delta-Debugging to shrink
noisy execution traces to minimal, deterministic reproductions.`,
	}

	runCmd := newRunCmd()
	demoCmd := newDemoCmd()
	benchCmd := newBenchCmd()
	diffCmd := newDiffCmd()
	matrixCmd := newMatrixCmd()
	replayCmd := newReplayCmd()
	initCmd := newInitCmd()
	validateCmd := newValidateCmd()
	uiCmd := newUICmd()
	mutateCmd := newMutateCmd()
	swarmCmd := newSwarmCmd()

	rootCmd.AddCommand(
		runCmd,
		demoCmd,
		benchCmd,
		diffCmd,
		matrixCmd,
		replayCmd,
		initCmd,
		validateCmd,
		uiCmd,
		mutateCmd,
		swarmCmd,
	)

	return rootCmd
}