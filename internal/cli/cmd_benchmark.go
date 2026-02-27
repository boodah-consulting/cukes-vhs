package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

// newBenchmarkCmd creates the benchmark command.
func newBenchmarkCmd() *cobra.Command {
	var benchmarkAll bool
	var featuresDir string
	var scenariosDir string
	var runs int
	var outputFile string

	cmd := &cobra.Command{
		Use:   "benchmark",
		Short: "Benchmark commands extracted from feature scenarios",
		Long: `Benchmark commands extracted from Gherkin feature scenarios.

Parses feature files, extracts CLI commands from When steps, benchmarks
them by running each command multiple times, and outputs timing results
as JSON.`,
		Example: `  cukes-vhs benchmark --all
  cukes-vhs benchmark --all --runs 5
  cukes-vhs benchmark --all --features features/ --output results.json`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "all" {
				benchmarkAll = true
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !benchmarkAll {
				return errors.New("--all is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			allScenarios, err := parseAllScenarios(featuresDir, scenariosDir, errOut)
			if err != nil {
				return err
			}

			results := cukesvhs.AnalyseScenarios(allScenarios)
			filtered := filterResults(results, allScenarios, benchmarkAll, "", "")
			translatable := filterTranslatable(filtered)

			allTimings := make(map[string]map[string]cukesvhs.CommandTiming)
			var totalCommands, totalScenarios int

			for i := range translatable {
				scenario := translatable[i].scenario
				timings, benchErr := cukesvhs.BenchmarkScenario(scenario, runs)
				if benchErr != nil {
					fmt.Fprintf(errOut, "Error benchmarking %q: %v\n", scenario.Name, benchErr)
					continue
				}
				if len(timings) > 0 {
					allTimings[scenario.Name] = timings
					totalCommands += len(timings)
					totalScenarios++
				}
			}

			jsonData, err := json.MarshalIndent(allTimings, "", "  ")
			if err != nil {
				return fmt.Errorf("encoding benchmark results: %w", err)
			}

			if outputFile != "" {
				if writeErr := os.WriteFile(outputFile, jsonData, 0o600); writeErr != nil {
					return fmt.Errorf("writing benchmark output to %q: %w", outputFile, writeErr)
				}
			} else {
				fmt.Fprintf(out, "%s\n", jsonData)
			}

			fmt.Fprintf(errOut, "Benchmarked %d commands across %d scenarios\n", totalCommands, totalScenarios)
			return nil
		},
	}

	cmd.Flags().BoolVar(&benchmarkAll, "all", false, "Benchmark all translatable scenarios")
	cmd.Flags().StringVarP(&featuresDir, "features", "f", "features/", "Directory containing .feature files")
	cmd.Flags().StringVar(&scenariosDir, "scenarios-dir", "demos/scenarios/", "Directory containing VHS-only .feature files")
	cmd.Flags().IntVarP(&runs, "runs", "n", 3, "Number of benchmark iterations per command")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "JSON output file path (default: stdout)")

	return cmd
}
