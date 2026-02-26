package cli

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

// newRunCmd creates the run command.
func newRunCmd() *cobra.Command {
	var runAll bool
	var featureFilter string
	var scenarioFilter string
	var featuresDir string
	var scenariosDir string
	var outputDir string
	var goldenDir string
	var timeoutSec int
	var configSource string
	var binaryPath string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Full pipeline: generate → render → validate",
		Long: `Run the full cukes-vhs pipeline: generate → render → validate.

Generates VHS tape files from translatable scenarios, renders them using vhs,
and validates the output against golden baselines.`,
		Example: `  cukes-vhs run --all --output /tmp/tapes/ --golden demos/golden/
  cukes-vhs run --feature onboarding --output /tmp/tapes/
  cukes-vhs run --scenario "User logs in" --output /tmp/tapes/`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "all" {
				runAll = true
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if outputDir == "" {
				return errors.New("--output is required")
			}
			if !runAll && featureFilter == "" && scenarioFilter == "" {
				return errors.New("one of --all, --feature, or --scenario is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Context()
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			fmt.Fprintf(out, "Parsing...\n")

			allScenarios, err := parseAllScenarios(featuresDir, scenariosDir, errOut)
			if err != nil {
				return err
			}

			results := cukesvhs.AnalyseScenarios(allScenarios)
			filtered := filterResults(results, allScenarios, runAll, featureFilter, scenarioFilter)
			translatableFiltered := filterTranslatable(filtered)

			if len(translatableFiltered) == 0 {
				fmt.Fprintf(out, "No translatable scenarios found.\n")
				fmt.Fprintf(out, "Results: 0 PASS, 0 FAIL, 0 NEW\n")
				return nil
			}

			genCfg := generateConfig{
				outputDir:    outputDir,
				configSource: configSource,
				verbose:      false,
				out:          io.Discard,
				errOut:       errOut,
			}

			fmt.Fprintf(out, "Generating...\n")
			stats := generateTapes(translatableFiltered, genCfg)
			fmt.Fprintf(out, "Generated %d tape(s): %d from business features, %d from VHS-only scenarios\n",
				stats.total, stats.fromBusiness, stats.fromVHSOnly)

			exitCode := renderAndValidate(out, errOut, outputDir, goldenDir, timeoutSec, binaryPath)
			if exitCode != 0 {
				return errors.New("pipeline failed")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&runAll, "all", false, "Run pipeline for all translatable scenarios")
	cmd.Flags().StringVar(&featureFilter, "feature", "", "Filter by feature name")
	cmd.Flags().StringVar(&scenarioFilter, "scenario", "", "Filter by scenario name")
	cmd.Flags().StringVar(&featuresDir, "features", "features/", "Directory containing .feature files")
	cmd.Flags().StringVar(&scenariosDir, "scenarios-dir", "demos/scenarios/", "Directory containing VHS-only .feature files")
	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory for generated tapes (required)")
	cmd.Flags().StringVar(&goldenDir, "golden", "demos/golden/", "Golden baseline directory")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 120, "Per-tape render timeout in seconds")
	cmd.Flags().StringVar(&configSource, "config-source", "config/config.tape", "Path to config tape file")
	cmd.Flags().StringVar(&binaryPath, "binary-path", "", "Path to vhs binary (default: vhs in PATH)")

	return cmd
}

// runOptions holds parsed flags for the run subcommand.
type runOptions struct {
	runAll         *bool
	featureFilter  *string
	scenarioFilter *string
	featuresDir    *string
	scenariosDir   *string
	outputDir      *string
	goldenDir      *string
	timeoutSec     *int
	configSource   *string
	binaryPath     *string
}

// parseRunFlags parses the flags for the run subcommand.
func parseRunFlags(args []string, errOut io.Writer) (*runOptions, error) {
	cmd := newRunCmd()
	cmd.SetArgs(normaliseArgs(args))
	cmd.SetOut(io.Discard)
	cmd.SetErr(errOut)

	opts := &runOptions{
		runAll:         new(bool),
		featureFilter:  new(string),
		scenarioFilter: new(string),
		featuresDir:    new(string),
		scenariosDir:   new(string),
		outputDir:      new(string),
		goldenDir:      new(string),
		timeoutSec:     new(int),
		configSource:   new(string),
		binaryPath:     new(string),
	}

	if err := cmd.ParseFlags(normaliseArgs(args)); err != nil {
		fmt.Fprintf(errOut, "Error parsing flags: %v\n", err)
		return nil, err
	}

	*opts.runAll, _ = cmd.Flags().GetBool("all")
	*opts.featureFilter, _ = cmd.Flags().GetString("feature")
	*opts.scenarioFilter, _ = cmd.Flags().GetString("scenario")
	*opts.featuresDir, _ = cmd.Flags().GetString("features")
	*opts.scenariosDir, _ = cmd.Flags().GetString("scenarios-dir")
	*opts.outputDir, _ = cmd.Flags().GetString("output")
	*opts.goldenDir, _ = cmd.Flags().GetString("golden")
	*opts.timeoutSec, _ = cmd.Flags().GetInt("timeout")
	*opts.configSource, _ = cmd.Flags().GetString("config-source")
	*opts.binaryPath, _ = cmd.Flags().GetString("binary-path")

	if *opts.outputDir == "" {
		fmt.Fprintf(errOut, "Error: --output is required\n")
		return nil, errors.New("output required")
	}

	if !*opts.runAll && *opts.featureFilter == "" && *opts.scenarioFilter == "" {
		fmt.Fprintf(errOut, "Error: one of --all, --feature, or --scenario is required\n")
		return nil, errors.New("filter required")
	}

	return opts, nil
}

// filterTranslatable returns only those scenarioWithResult entries where the result is translatable.
func filterTranslatable(filtered []scenarioWithResult) []scenarioWithResult {
	out := make([]scenarioWithResult, 0, len(filtered))
	for i := range filtered {
		if filtered[i].result.Translatable {
			out = append(out, filtered[i])
		}
	}

	return out
}

// renderAndValidate renders all tapes in outputDir and validates them against goldenDir.
func renderAndValidate(out, errOut io.Writer, outputDir, goldenDir string, timeoutSec int, binaryPath string) int {
	fmt.Fprintf(out, "Rendering...\n")

	renderer := cukesvhs.NewRenderer(binaryPath)
	timeout := pipelineTimeout(timeoutSec)

	renderResults, renderErr := renderer.RenderAll(outputDir, timeout)
	if renderErr != nil {
		fmt.Fprintf(errOut, "Error rendering tapes: %v\n", renderErr)
		return 1
	}

	for _, result := range renderResults {
		if !result.Success {
			fmt.Fprintf(errOut, "Error rendering tapes: %s\n", result.Error)
			return 1
		}
	}

	fmt.Fprintf(out, "Validating...\n")

	validationResults, validErr := cukesvhs.ValidateAll(goldenDir, outputDir)
	if validErr != nil {
		fmt.Fprintf(errOut, "Error validating: %v\n", validErr)
		return 1
	}

	return reportPipelineResults(out, validationResults)
}

// pipelineTimeout converts a seconds integer into a time.Duration.
func pipelineTimeout(secs int) time.Duration {
	return time.Duration(secs) * time.Second
}

// reportPipelineResults prints per-scenario status lines and a summary, returning 1 if any FAIL.
func reportPipelineResults(out io.Writer, results []cukesvhs.ValidationResult) int {
	var pass, fail, newCount int

	for i := range results {
		r := &results[i]
		switch r.Status {
		case cukesvhs.ValidationPass:
			fmt.Fprintf(out, "[PASS] %s\n", r.Scenario)
			pass++
		case cukesvhs.ValidationFail:
			fmt.Fprintf(out, "[FAIL] %s\n", r.Scenario)
			fail++
		case cukesvhs.ValidationNew:
			fmt.Fprintf(out, "[NEW]  %s\n", r.Scenario)
			newCount++
		}
	}

	fmt.Fprintf(out, "Results: %d PASS, %d FAIL, %d NEW\n", pass, fail, newCount)

	if fail > 0 {
		return 1
	}

	return 0
}
