package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/boodah-consulting/cukesvhs/internal/vhsgen"
)

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
}

// parseRunFlags parses the flags for the run subcommand.
//
// Expected: args contains the arguments after the "run" subcommand token.
// Returns: populated *runOptions and nil error on success; nil and non-nil error on parse failure.
// Side effects: writes error messages to errOut.
func parseRunFlags(args []string, errOut io.Writer) (*runOptions, error) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	opts := &runOptions{
		runAll:         fs.Bool("all", false, "Run pipeline for all translatable scenarios"),
		featureFilter:  fs.String("feature", "", "Filter by feature name"),
		scenarioFilter: fs.String("scenario", "", "Filter by scenario name"),
		featuresDir:    fs.String("features", "features/", "Directory containing .feature files"),
		scenariosDir:   fs.String("scenarios-dir", "demos/scenarios/", "Directory containing VHS-only .feature files"),
		outputDir:      fs.String("output", "", "Output directory for generated tapes (required)"),
		goldenDir:      fs.String("golden", "demos/golden/", "Golden baseline directory"),
		timeoutSec:     fs.Int("timeout", 120, "Per-tape render timeout in seconds"),
		configSource:   fs.String("config-source", "demos/config.tape", "Path to config tape file"),
	}

	if err := fs.Parse(normaliseArgs(args)); err != nil {
		fmt.Fprintf(errOut, "Error parsing flags: %v\n", err)
		return nil, err
	}

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

// runPipeline implements the `run` subcommand: generate → render → validate → report.
//
// Expected: args are the arguments after the "run" subcommand token; out and errOut are non-nil writers.
// Returns: 0 on success or when all scenarios PASS/NEW; 1 on any FAIL or pipeline error.
// Side effects: writes tape files to outputDir; may create golden baselines; spawns vhs child processes.
func runPipeline(args []string, out io.Writer, errOut io.Writer) int {
	opts, err := parseRunFlags(args, errOut)
	if err != nil {
		return 1
	}

	fmt.Fprintf(out, "Parsing...\n")

	allScenarios, err := parseAllScenarios(*opts.featuresDir, *opts.scenariosDir, errOut)
	if err != nil {
		return 1
	}

	results := vhsgen.AnalyseScenarios(allScenarios)
	filtered := filterResults(results, allScenarios, *opts.runAll, *opts.featureFilter, *opts.scenarioFilter)
	translatableFiltered := filterTranslatable(filtered)

	if len(translatableFiltered) == 0 {
		fmt.Fprintf(out, "No translatable scenarios found.\n")
		fmt.Fprintf(out, "Results: 0 PASS, 0 FAIL, 0 NEW\n")
		return 0
	}

	genCfg := generateConfig{
		outputDir:    *opts.outputDir,
		configSource: *opts.configSource,
		verbose:      false,
		out:          io.Discard,
		errOut:       errOut,
	}

	fmt.Fprintf(out, "Generating...\n")
	generateTapes(translatableFiltered, genCfg)

	return renderAndValidate(out, errOut, *opts.outputDir, *opts.goldenDir, *opts.timeoutSec)
}

// filterTranslatable returns only those scenarioWithResult entries where the result is translatable.
//
// Expected: filtered is the output of filterResults.
// Returns: a new slice containing only translatable entries.
// Side effects: none.
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
//
// Expected: outputDir contains generated .tape files; goldenDir is writable.
// Returns: 0 when all validations PASS or NEW; 1 on any FAIL or error.
// Side effects: spawns vhs child processes; may create golden baselines.
func renderAndValidate(out, errOut io.Writer, outputDir, goldenDir string, timeoutSec int) int {
	fmt.Fprintf(out, "Rendering...\n")

	renderer := vhsgen.NewRenderer()
	timeout := pipelineTimeout(timeoutSec)

	_, renderErr := renderer.RenderAll(outputDir, timeout)
	if renderErr != nil {
		fmt.Fprintf(errOut, "Error rendering tapes: %v\n", renderErr)
		return 1
	}

	fmt.Fprintf(out, "Validating...\n")

	validationResults, validErr := vhsgen.ValidateAll(goldenDir, outputDir)
	if validErr != nil {
		fmt.Fprintf(errOut, "Error validating: %v\n", validErr)
		return 1
	}

	return reportPipelineResults(out, validationResults)
}

// pipelineTimeout converts a seconds integer into a time.Duration.
//
// Expected: secs is a positive integer representing seconds.
// Returns: the equivalent time.Duration.
// Side effects: none.
func pipelineTimeout(secs int) time.Duration {
	return time.Duration(secs) * time.Second
}

// reportPipelineResults prints per-scenario status lines and a summary, returning 1 if any FAIL.
//
// Expected: results is the slice of ValidationResult from ValidateAll.
// Returns: 0 when all results are PASS or NEW; 1 when at least one is FAIL.
// Side effects: writes to out.
func reportPipelineResults(out io.Writer, results []vhsgen.ValidationResult) int {
	var pass, fail, newCount int

	for i := range results {
		r := &results[i]
		switch r.Status {
		case vhsgen.ValidationPass:
			fmt.Fprintf(out, "[PASS] %s\n", r.Scenario)
			pass++
		case vhsgen.ValidationFail:
			fmt.Fprintf(out, "[FAIL] %s\n", r.Scenario)
			fail++
		case vhsgen.ValidationNew:
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
