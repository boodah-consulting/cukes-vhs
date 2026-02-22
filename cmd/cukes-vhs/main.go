// Package main provides the vhsgen CLI for listing and generating VHS tapes.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/baphled/cukes-vhs/internal/vhsgen"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, out io.Writer, errOut io.Writer) int {
	if len(args) == 0 {
		printUsageTo(out)
		return 0
	}

	subcommand := args[0]
	rest := args[1:]

	switch subcommand {
	case "list":
		return runList(rest, out, errOut)
	case "generate":
		return runGenerate(rest, out, errOut)
	case "run":
		return runPipeline(rest, out, errOut)
	case "update-baseline":
		return runUpdateBaseline(rest, out, errOut)
	case "--help", "-h", "help":
		printUsageTo(out)
		return 0
	default:
		fmt.Fprintf(errOut, "Error: unknown subcommand %q\n\n", subcommand)
		printUsageTo(errOut)
		return 1
	}
}

// runList implements the `list` subcommand.
func runList(args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	featuresDir := fs.String("features", "features/", "Directory containing .feature files")
	scenariosDir := fs.String("scenarios-dir", "demos/vhs/scenarios/", "Directory containing VHS-only .feature files")
	asJSON := fs.Bool("json", false, "Output as JSON")
	showCount := fs.Bool("count", false, "Show counts broken down by source")
	showSteps := fs.Bool("steps", false, "Show translatable step patterns")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(errOut, "Error parsing flags: %v\n", err)
		return 1
	}

	if *showSteps {
		return runListSteps(*asJSON, out)
	}

	fmt.Fprintf(out, "Parsing...\n")

	if _, err := os.Stat(*featuresDir); err != nil {
		fmt.Fprintf(errOut, "Error parsing features dir %q: %v\n", *featuresDir, err)
		return 1
	}

	businessScenarios, err := vhsgen.ParseFeatureDir(*featuresDir, vhsgen.SourceBusiness)
	if err != nil {
		fmt.Fprintf(errOut, "Error parsing features dir %q: %v\n", *featuresDir, err)
		return 1
	}

	vhsOnlyScenarios, err := vhsgen.ParseFeatureDir(*scenariosDir, vhsgen.SourceVHSOnly)
	if err != nil {
		fmt.Fprintf(errOut, "Error parsing scenarios dir %q: %v\n", *scenariosDir, err)
		return 1
	}

	allScenarios := make([]vhsgen.ScenarioIR, 0, len(businessScenarios)+len(vhsOnlyScenarios))
	allScenarios = append(allScenarios, businessScenarios...)
	allScenarios = append(allScenarios, vhsOnlyScenarios...)
	results := vhsgen.AnalyseScenarios(allScenarios)

	if *showCount {
		return runListCount(results, out)
	}

	if *asJSON {
		return runListJSON(results, out, errOut)
	}

	return runListTable(results, out)
}

// runListSteps outputs the translatable step patterns.
func runListSteps(asJSON bool, out io.Writer) int {
	patterns := vhsgen.ListTranslatablePatterns()

	if asJSON {
		type jsonPattern struct {
			Pattern  string                            `json:"pattern"`
			Type     string                            `json:"type"`
			Category string                            `json:"category"`
			Params   map[string]vhsgen.ParamConstraint `json:"params,omitempty"`
			Example  string                            `json:"example"`
		}

		output := make([]jsonPattern, 0, len(patterns))
		for _, p := range patterns {
			output = append(output, jsonPattern{
				Pattern:  p.Pattern,
				Type:     p.Type,
				Category: p.Category,
				Params:   p.Params,
				Example:  p.Example,
			})
		}

		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(output); err != nil {
			return 1
		}
		return 0
	}

	colPattern := 50
	colType := 10
	colCategory := 14

	header := fmt.Sprintf("%-*s  %-*s  %-*s  %s",
		colPattern, "Pattern",
		colType, "Type",
		colCategory, "Category",
		"Example",
	)
	separator := strings.Repeat("-", len(header)+10)

	fmt.Fprintln(out, header)
	fmt.Fprintln(out, separator)

	for _, p := range patterns {
		fmt.Fprintf(out, "%-*s  %-*s  %-*s  %s\n",
			colPattern, truncate(p.Pattern, colPattern),
			colType, truncate(p.Type, colType),
			colCategory, truncate(p.Category, colCategory),
			p.Example,
		)
	}

	return 0
}

// runListCount outputs counts by source.
func runListCount(results []vhsgen.AnalysisResult, out io.Writer) int {
	var (
		businessTotal        int
		businessTranslatable int
		vhsOnlyTotal         int
		vhsOnlyTranslatable  int
	)

	for i := range results {
		r := &results[i]
		switch r.Source {
		case vhsgen.SourceBusiness:
			businessTotal++
			if r.Translatable {
				businessTranslatable++
			}
		case vhsgen.SourceVHSOnly:
			vhsOnlyTotal++
			if r.Translatable {
				vhsOnlyTranslatable++
			}
		}
	}

	fmt.Fprintf(out, "Business: %d/%d translatable | VHS-only: %d/%d translatable\n",
		businessTranslatable, businessTotal,
		vhsOnlyTranslatable, vhsOnlyTotal,
	)

	return 0
}

// runListJSON outputs the analysis results as JSON.
func runListJSON(results []vhsgen.AnalysisResult, out io.Writer, errOut io.Writer) int {
	type jsonScenario struct {
		ScenarioName string `json:"scenario_name"`
		Feature      string `json:"feature"`
		Source       string `json:"source"`
		Translatable bool   `json:"translatable"`
		Reason       string `json:"reason,omitempty"`
	}

	scenarios := make([]jsonScenario, 0, len(results))
	for i := range results {
		r := &results[i]
		var reason string
		if !r.Translatable && len(r.UntranslatableSteps) > 0 {
			reasons := make([]string, 0, len(r.UntranslatableSteps))
			for _, s := range r.UntranslatableSteps {
				reasons = append(reasons, s.UntranslatableReason)
			}
			reason = strings.Join(reasons, "; ")
		}

		scenarios = append(scenarios, jsonScenario{
			ScenarioName: r.ScenarioName,
			Feature:      r.Feature,
			Source:       string(r.Source),
			Translatable: r.Translatable,
			Reason:       reason,
		})
	}

	payload := map[string]interface{}{
		"scenarios": scenarios,
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		fmt.Fprintf(errOut, "Error encoding JSON: %v\n", err)
		return 1
	}

	return 0
}

// runListTable outputs the analysis results as a formatted table.
func runListTable(results []vhsgen.AnalysisResult, out io.Writer) int {
	colScenario := 40
	colFeature := 25
	colSource := 10
	colTranslatable := 12
	colReason := 40

	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %s",
		colScenario, "Scenario",
		colFeature, "Feature",
		colSource, "Source",
		colTranslatable, "Translatable",
		"Reason",
	)
	separator := strings.Repeat("-", colScenario+colFeature+colSource+colTranslatable+colReason+8)

	fmt.Fprintln(out, header)
	fmt.Fprintln(out, separator)

	for i := range results {
		r := &results[i]
		translatable := "yes"
		var reason string
		if !r.Translatable {
			translatable = "no"
			if len(r.UntranslatableSteps) > 0 {
				reasons := make([]string, 0, len(r.UntranslatableSteps))
				for _, s := range r.UntranslatableSteps {
					reasons = append(reasons, s.UntranslatableReason)
				}
				reason = strings.Join(reasons, "; ")
			}
		}

		fmt.Fprintf(out, "%-*s  %-*s  %-*s  %-*s  %s\n",
			colScenario, truncate(r.ScenarioName, colScenario),
			colFeature, truncate(r.Feature, colFeature),
			colSource, truncate(string(r.Source), colSource),
			colTranslatable, translatable,
			truncate(reason, colReason),
		)
	}

	return 0
}

// runGenerate implements the `generate` subcommand.
func runGenerate(args []string, out io.Writer, errOut io.Writer) int {
	opts, err := parseGenerateFlags(args, errOut)
	if err != nil {
		return 1
	}

	fmt.Fprintf(out, "Parsing...\n")

	allScenarios, err := parseAllScenarios(*opts.featuresDir, *opts.scenariosDir, errOut)
	if err != nil {
		return 1
	}

	results := vhsgen.AnalyseScenarios(allScenarios)
	filtered := filterResults(results, allScenarios, *opts.generateAll, *opts.featureFilter, *opts.scenarioFilter)

	fmt.Fprintf(out, "Generating...\n")

	cfg := generateConfig{
		outputDir:    *opts.outputDir,
		configSource: *opts.configSource,
		verbose:      *opts.verbose,
		out:          out,
		errOut:       errOut,
	}
	stats := generateTapes(filtered, cfg)

	fmt.Fprintf(out, "Generated %d tapes (%d from features, %d from scenarios, %d warnings)\n",
		stats.total, stats.fromBusiness, stats.fromVHSOnly, stats.warnings)

	return 0
}

type generateOptions struct {
	generateAll    *bool
	featureFilter  *string
	scenarioFilter *string
	featuresDir    *string
	scenariosDir   *string
	outputDir      *string
	configSource   *string
	verbose        *bool
}

func parseGenerateFlags(args []string, errOut io.Writer) (*generateOptions, error) {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	opts := &generateOptions{
		generateAll:    fs.Bool("all", false, "Generate for all translatable scenarios"),
		featureFilter:  fs.String("feature", "", "Filter by feature name"),
		scenarioFilter: fs.String("scenario", "", "Filter by scenario name"),
		featuresDir:    fs.String("features", "features/", "Directory containing .feature files"),
		scenariosDir:   fs.String("scenarios-dir", "demos/vhs/scenarios/", "Directory containing VHS-only .feature files"),
		outputDir:      fs.String("output", "", "Output directory (required)"),
		configSource:   fs.String("config-source", "demos/vhs/config.tape", "Path to config tape file"),
		verbose:        fs.Bool("verbose", false, "Verbose output"),
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(errOut, "Error parsing flags: %v\n", err)
		return nil, err
	}

	if *opts.outputDir == "" {
		fmt.Fprintf(errOut, "Error: --output is required\n")
		return nil, errors.New("output required")
	}

	if !*opts.generateAll && *opts.featureFilter == "" && *opts.scenarioFilter == "" {
		fmt.Fprintf(errOut, "Error: one of --all, --feature, or --scenario is required\n")
		return nil, errors.New("filter required")
	}

	return opts, nil
}

func parseAllScenarios(featuresDir, scenariosDir string, errOut io.Writer) ([]vhsgen.ScenarioIR, error) {
	if _, err := os.Stat(featuresDir); err != nil {
		fmt.Fprintf(errOut, "Error parsing features dir %q: %v\n", featuresDir, err)
		return nil, err
	}

	businessScenarios, err := vhsgen.ParseFeatureDir(featuresDir, vhsgen.SourceBusiness)
	if err != nil {
		fmt.Fprintf(errOut, "Error parsing features dir %q: %v\n", featuresDir, err)
		return nil, err
	}

	vhsOnlyScenarios, err := vhsgen.ParseFeatureDir(scenariosDir, vhsgen.SourceVHSOnly)
	if err != nil {
		fmt.Fprintf(errOut, "Error parsing scenarios dir %q: %v\n", scenariosDir, err)
		return nil, err
	}

	allScenarios := make([]vhsgen.ScenarioIR, 0, len(businessScenarios)+len(vhsOnlyScenarios))
	allScenarios = append(allScenarios, businessScenarios...)
	allScenarios = append(allScenarios, vhsOnlyScenarios...)
	return allScenarios, nil
}

type scenarioWithResult struct {
	scenario vhsgen.ScenarioIR
	result   vhsgen.AnalysisResult
}

type generateStats struct {
	total        int
	fromBusiness int
	fromVHSOnly  int
	warnings     int
}

type generateConfig struct {
	outputDir    string
	configSource string
	verbose      bool
	out          io.Writer
	errOut       io.Writer
}

// generateTapes processes filtered scenarios and writes tape files.
func generateTapes(filtered []scenarioWithResult, cfg generateConfig) generateStats {
	var stats generateStats

	for i := range filtered {
		entry := &filtered[i]
		scenario := entry.scenario
		result := entry.result

		if !result.Translatable {
			if cfg.verbose {
				fmt.Fprintf(cfg.out, "Skipping %q (not translatable)\n", scenario.Name)
			}
			stats.warnings++
			continue
		}

		outPath, tapeErr := writeScenarioTape(scenario, cfg.outputDir, cfg.configSource)
		if tapeErr != nil {
			fmt.Fprintf(cfg.errOut, "Error generating tape for %q: %v\n", scenario.Name, tapeErr)
			continue
		}

		fmt.Fprintf(cfg.out, "Written: %s\n", outPath)

		switch scenario.Source {
		case vhsgen.SourceBusiness:
			stats.fromBusiness++
		case vhsgen.SourceVHSOnly:
			stats.fromVHSOnly++
		}
	}

	stats.total = stats.fromBusiness + stats.fromVHSOnly
	return stats
}

// filterResults selects scenarios to generate based on flags.
func filterResults(
	results []vhsgen.AnalysisResult,
	scenarios []vhsgen.ScenarioIR,
	all bool,
	featureFilter, scenarioFilter string,
) []scenarioWithResult {
	var out []scenarioWithResult

	resultByName := make(map[string]vhsgen.AnalysisResult, len(results))
	for i := range results {
		r := &results[i]
		resultByName[r.ScenarioName] = *r
	}

	for i := range scenarios {
		s := &scenarios[i]
		result, ok := resultByName[s.Name]
		if !ok {
			continue
		}

		if !all {
			if featureFilter != "" && !strings.EqualFold(s.Feature, featureFilter) {
				continue
			}
			if scenarioFilter != "" && !strings.EqualFold(s.Name, scenarioFilter) {
				continue
			}
		}

		out = append(out, scenarioWithResult{scenario: *s, result: result})
	}

	return out
}

// writeScenarioTape generates and writes a tape file with source-aware routing:
// Business tapes → {output}/{feature-slug}/{scenario-slug}.tape.
// VHS-only tapes → {output}/scenarios/{subdirectory}/{scenario-slug}.tape.
func writeScenarioTape(scenario vhsgen.ScenarioIR, outputDir, configSourcePath string) (string, error) {
	featureSlug := slugify(scenario.Feature)
	scenarioSlug := slugify(scenario.Name)

	var tapeDir string
	switch scenario.Source {
	case vhsgen.SourceVHSOnly:
		tapeDir = filepath.Join(outputDir, "scenarios", featureSlug)
	default:
		tapeDir = filepath.Join(outputDir, featureSlug)
	}

	config := vhsgen.GeneratorConfig{
		OutputDir:        outputDir,
		ConfigSourcePath: configSourcePath,
	}

	content, err := vhsgen.GenerateTape(scenario, config)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(tapeDir, 0o755); err != nil {
		return "", fmt.Errorf("creating output directory %q: %w", tapeDir, err)
	}

	outPath := filepath.Join(tapeDir, scenarioSlug+".tape")
	if err := os.WriteFile(outPath, []byte(content), 0o600); err != nil {
		return "", fmt.Errorf("writing tape file %q: %w", outPath, err)
	}

	return outPath, nil
}

var (
	slugStripRe    = regexp.MustCompile(`[^a-z0-9-]`)
	slugCollapseRe = regexp.MustCompile(`-{2,}`)
)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	s = slugStripRe.ReplaceAllString(s, "")
	s = slugCollapseRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
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
		scenariosDir:   fs.String("scenarios-dir", "demos/vhs/scenarios/", "Directory containing VHS-only .feature files"),
		outputDir:      fs.String("output", "", "Output directory for generated tapes (required)"),
		goldenDir:      fs.String("golden", "demos/vhs/golden/", "Golden baseline directory"),
		timeoutSec:     fs.Int("timeout", 120, "Per-tape render timeout in seconds"),
		configSource:   fs.String("config-source", "demos/vhs/config.tape", "Path to config tape file"),
	}

	if err := fs.Parse(args); err != nil {
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

// updateBaselineOptions holds parsed flags for the update-baseline subcommand.
type updateBaselineOptions struct {
	updateAll *bool
	goldenDir *string
	outputDir *string
}

// parseUpdateBaselineFlags parses the flags for the update-baseline subcommand.
//
// Expected: args contains the arguments after the "update-baseline" subcommand token.
// Returns: populated *updateBaselineOptions and nil error on success; nil and non-nil error on failure.
// Side effects: writes error messages to errOut.
func parseUpdateBaselineFlags(args []string, errOut io.Writer) (*updateBaselineOptions, []string, error) {
	fs := flag.NewFlagSet("update-baseline", flag.ContinueOnError)
	opts := &updateBaselineOptions{
		updateAll: fs.Bool("all", false, "Accept all current outputs as golden baselines"),
		goldenDir: fs.String("golden", "demos/vhs/golden/", "Golden baseline directory"),
		outputDir: fs.String("output", "", "Output directory containing rendered .ascii files (required)"),
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(errOut, "Error parsing flags: %v\n", err)
		return nil, nil, err
	}

	if *opts.outputDir == "" {
		fmt.Fprintf(errOut, "Error: --output is required\n")
		return nil, nil, errors.New("output required")
	}

	return opts, fs.Args(), nil
}

// runUpdateBaseline implements the `update-baseline` subcommand.
//
// Expected: args are the arguments after the "update-baseline" subcommand token.
// Returns: 0 on success; 1 on flag parse error or update failure.
// Side effects: overwrites golden baseline files in goldenDir.
func runUpdateBaseline(args []string, out io.Writer, errOut io.Writer) int {
	opts, positional, err := parseUpdateBaselineFlags(args, errOut)
	if err != nil {
		return 1
	}

	if !*opts.updateAll && len(positional) == 0 {
		fmt.Fprintf(errOut, "Error: --all or a scenario name is required\n")
		return 1
	}

	if *opts.updateAll {
		return updateAllBaselines(*opts.goldenDir, *opts.outputDir, out, errOut)
	}

	return updateNamedBaselines(*opts.goldenDir, *opts.outputDir, positional, out, errOut)
}

// updateAllBaselines scans outputDir for .ascii files and accepts each as the golden baseline.
//
// Expected: outputDir contains rendered .ascii files; goldenDir is writable.
// Returns: 0 when all updates succeed; 1 on any update failure.
// Side effects: creates or overwrites baseline files under goldenDir.
func updateAllBaselines(goldenDir, outputDir string, out, errOut io.Writer) int {
	asciiFiles, err := collectOutputASCIIFiles(outputDir)
	if err != nil {
		fmt.Fprintf(errOut, "Error scanning output dir %q: %v\n", outputDir, err)
		return 1
	}

	if len(asciiFiles) == 0 {
		fmt.Fprintf(out, "No .ascii files found in %q.\n", outputDir)
		fmt.Fprintf(out, "Updated 0 baselines.\n")
		return 0
	}

	updated := 0

	for _, asciiPath := range asciiFiles {
		scenario := deriveScenarioName(outputDir, asciiPath)
		gifPath := deriveGIFPath(asciiPath)

		if err := vhsgen.UpdateBaseline(goldenDir, scenario, asciiPath, gifPath); err != nil {
			fmt.Fprintf(errOut, "Error updating baseline for %q: %v\n", scenario, err)
			return 1
		}

		fmt.Fprintf(out, "Updated: %s\n", scenario)
		updated++
	}

	fmt.Fprintf(out, "Updated %d baselines.\n", updated)

	return 0
}

// updateNamedBaselines accepts specific named scenarios as golden baselines.
//
// Expected: scenarios is a non-empty slice of scenario names; outputDir contains corresponding .ascii files.
// Returns: 0 when all named scenarios are updated; 1 on any failure.
// Side effects: creates or overwrites baseline files under goldenDir.
func updateNamedBaselines(goldenDir, outputDir string, scenarios []string, out, errOut io.Writer) int {
	updated := 0

	for _, scenario := range scenarios {
		asciiPath := filepath.Join(outputDir, slugify(scenario)+".ascii")
		gifPath := deriveGIFPath(asciiPath)

		if err := vhsgen.UpdateBaseline(goldenDir, scenario, asciiPath, gifPath); err != nil {
			fmt.Fprintf(errOut, "Error updating baseline for %q: %v\n", scenario, err)
			return 1
		}

		fmt.Fprintf(out, "Updated: %s\n", scenario)
		updated++
	}

	fmt.Fprintf(out, "Updated %d baselines.\n", updated)

	return 0
}

// collectOutputASCIIFiles returns all .ascii files found recursively under dir.
//
// Expected: dir is a readable directory (or non-existent, returning empty slice).
// Returns: slice of absolute .ascii file paths; empty slice when dir is empty or missing; non-nil error on walk failure.
// Side effects: none.
func collectOutputASCIIFiles(dir string) ([]string, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []string{}, nil
	}

	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !d.IsDir() && strings.HasSuffix(path, ".ascii") {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if files == nil {
		return []string{}, nil
	}

	return files, nil
}

// deriveScenarioName converts an ASCII output path to a scenario slug by stripping the outputDir prefix and .ascii suffix.
//
// Expected: outputDir is the root output directory; asciiPath is an absolute path under outputDir.
// Returns: a slugified scenario name derived from the relative path.
// Side effects: none.
func deriveScenarioName(outputDir, asciiPath string) string {
	rel := asciiPath

	if r, err := filepath.Rel(outputDir, asciiPath); err == nil {
		rel = r
	}

	rel = strings.TrimSuffix(rel, ".ascii")
	rel = strings.ReplaceAll(rel, string(filepath.Separator), "-")

	return slugify(rel)
}

// deriveGIFPath converts an .ascii path to the corresponding .gif path by replacing the extension.
//
// Expected: asciiPath ends in ".ascii".
// Returns: the same path with ".ascii" replaced by ".gif".
// Side effects: none.
func deriveGIFPath(asciiPath string) string {
	return strings.TrimSuffix(asciiPath, ".ascii") + ".gif"
}

func printUsageTo(out io.Writer) {
	fmt.Fprintln(out, "vhsgen — VHS tape generator for KaRiya")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  vhsgen list [flags]              List scenarios and their translatability")
	fmt.Fprintln(out, "  vhsgen generate [flags]          Generate VHS tape files from scenarios")
	fmt.Fprintln(out, "  vhsgen run [flags]               Full pipeline: generate → render → validate")
	fmt.Fprintln(out, "  vhsgen update-baseline [flags]   Accept current output as new golden baseline")
	fmt.Fprintln(out, "")
	printListFlags(out)
	printGenerateFlags(out)
	printRunFlags(out)
	printUpdateBaselineFlags(out)
	printExamples(out)
}

func printListFlags(out io.Writer) {
	fmt.Fprintln(out, "list flags:")
	fmt.Fprintln(out, "  --features DIR       Directory with .feature files (default: features/)")
	fmt.Fprintln(out, "  --scenarios-dir DIR  Directory with VHS-only .feature files (default: demos/vhs/scenarios/)")
	fmt.Fprintln(out, "  --json               Output as JSON")
	fmt.Fprintln(out, "  --count              Show counts broken down by source")
	fmt.Fprintln(out, "  --steps              Show translatable step patterns")
	fmt.Fprintln(out, "")
}

func printGenerateFlags(out io.Writer) {
	fmt.Fprintln(out, "generate flags:")
	fmt.Fprintln(out, "  --all                Generate for all translatable scenarios")
	fmt.Fprintln(out, "  --feature NAME       Filter by feature name")
	fmt.Fprintln(out, "  --scenario NAME      Filter by scenario name")
	fmt.Fprintln(out, "  --features DIR       Directory with .feature files (default: features/)")
	fmt.Fprintln(out, "  --scenarios-dir DIR  Directory with VHS-only .feature files (default: demos/vhs/scenarios/)")
	fmt.Fprintln(out, "  --output DIR         Output directory (required)")
	fmt.Fprintln(out, "  --config-source PATH Path to config tape file (default: demos/vhs/config.tape)")
	fmt.Fprintln(out, "  --verbose            Verbose output")
	fmt.Fprintln(out, "")
}

func printRunFlags(out io.Writer) {
	fmt.Fprintln(out, "run flags:")
	fmt.Fprintln(out, "  --all                Run for all translatable scenarios")
	fmt.Fprintln(out, "  --feature NAME       Filter by feature name")
	fmt.Fprintln(out, "  --scenario NAME      Filter by scenario name")
	fmt.Fprintln(out, "  --features DIR       Directory with .feature files (default: features/)")
	fmt.Fprintln(out, "  --scenarios-dir DIR  Directory with VHS-only .feature files (default: demos/vhs/scenarios/)")
	fmt.Fprintln(out, "  --output DIR         Output directory (required)")
	fmt.Fprintln(out, "  --golden DIR         Golden baseline directory (default: demos/vhs/golden/)")
	fmt.Fprintln(out, "  --timeout N          Per-tape render timeout in seconds (default: 120)")
	fmt.Fprintln(out, "  --config-source PATH Path to config tape file (default: demos/vhs/config.tape)")
	fmt.Fprintln(out, "")
}

func printUpdateBaselineFlags(out io.Writer) {
	fmt.Fprintln(out, "update-baseline flags:")
	fmt.Fprintln(out, "  --all                Accept all current outputs as golden baselines")
	fmt.Fprintln(out, "  --golden DIR         Golden baseline directory (default: demos/vhs/golden/)")
	fmt.Fprintln(out, "  --output DIR         Output directory containing rendered .ascii files (required)")
	fmt.Fprintln(out, "")
}

func printExamples(out io.Writer) {
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  vhsgen list --features features/ --scenarios-dir demos/vhs/scenarios/")
	fmt.Fprintln(out, "  vhsgen list --json")
	fmt.Fprintln(out, "  vhsgen list --count")
	fmt.Fprintln(out, "  vhsgen list --steps")
	fmt.Fprintln(out, "  vhsgen list --steps --json")
	fmt.Fprintln(out, "  vhsgen generate --all --features features/ --scenarios-dir demos/vhs/scenarios/ --output /tmp/tapes/")
	fmt.Fprintln(out, "  vhsgen generate --feature onboarding --output /tmp/test/")
	fmt.Fprintln(out, "  vhsgen run --all --output /tmp/tapes/ --golden demos/vhs/golden/")
	fmt.Fprintln(out, "  vhsgen run --feature onboarding --output /tmp/tapes/")
	fmt.Fprintln(out, "  vhsgen update-baseline --all --output /tmp/tapes/ --golden demos/vhs/golden/")
	fmt.Fprintln(out, "  vhsgen update-baseline my-scenario --output /tmp/tapes/")
}
