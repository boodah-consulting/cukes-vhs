package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

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

	results := cukesvhs.AnalyseScenarios(allScenarios)
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

	var renderFailed bool
	if stats.total > 0 {
		fmt.Fprintf(out, "Rendering...\n")

		renderer := cukesvhs.NewRenderer(*opts.binaryPath)
		timeout := pipelineTimeout(*opts.timeoutSec)

		renderResults, renderErr := renderer.RenderAll(*opts.outputDir, timeout)
		if renderErr != nil {
			fmt.Fprintf(errOut, "Error rendering tapes: %v\n", renderErr)
			renderFailed = true
		} else {
			for _, r := range renderResults {
				if !r.Success {
					fmt.Fprintf(errOut, "Render failed for %s: %s\n", r.TapePath, r.Error)
					renderFailed = true
				}
			}
		}
	}

	fmt.Fprintf(out, "Generated %d tapes (%d from features, %d from scenarios, %d warnings)\n",
		stats.total, stats.fromBusiness, stats.fromVHSOnly, stats.warnings)

	if renderFailed {
		return 1
	}

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
	timeoutSec     *int
	binaryPath     *string
}

func parseGenerateFlags(args []string, errOut io.Writer) (*generateOptions, error) {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	opts := &generateOptions{
		generateAll:    fs.Bool("all", false, "Generate for all translatable scenarios"),
		featureFilter:  fs.String("feature", "", "Filter by feature name"),
		scenarioFilter: fs.String("scenario", "", "Filter by scenario name"),
		featuresDir:    fs.String("features", "features/", "Directory containing .feature files"),
		scenariosDir:   fs.String("scenarios-dir", "demos/scenarios/", "Directory containing VHS-only .feature files"),
		outputDir:      fs.String("output", "", "Output directory (required)"),
		configSource:   fs.String("config-source", "config/config.tape", "Path to config tape file"),
		verbose:        fs.Bool("verbose", false, "Verbose output"),
		timeoutSec:     fs.Int("timeout", 120, "Per-tape render timeout in seconds"),
		binaryPath:     fs.String("binary-path", "", "Path to vhs binary (default: vhs in PATH)"),
	}

	if err := fs.Parse(normaliseArgs(args)); err != nil {
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
		case cukesvhs.SourceBusiness:
			stats.fromBusiness++
		case cukesvhs.SourceVHSOnly:
			stats.fromVHSOnly++
		}
	}

	stats.total = stats.fromBusiness + stats.fromVHSOnly
	return stats
}
