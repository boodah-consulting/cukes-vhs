package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

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
		goldenDir: fs.String("golden", "demos/golden/", "Golden baseline directory"),
		outputDir: fs.String("output", "", "Output directory containing rendered .ascii files (required)"),
	}

	if err := fs.Parse(normaliseArgs(args)); err != nil {
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

		if err := cukesvhs.UpdateBaseline(goldenDir, scenario, asciiPath, gifPath); err != nil {
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
// It searches outputDir recursively for matching .ascii files, supporting the
// nested directory layout ({outputDir}/{featureSlug}/{scenarioSlug}.ascii)
// created by the generator.
//
// Expected: scenarios is a non-empty slice of scenario names; outputDir contains corresponding .ascii files.
// Returns: 0 when all named scenarios are updated; 1 on any failure.
// Side effects: creates or overwrites baseline files under goldenDir.
func updateNamedBaselines(goldenDir, outputDir string, scenarios []string, out, errOut io.Writer) int {
	updated := 0
	for _, scenario := range scenarios {
		scenarioSlug := cukesvhs.Slugify(scenario)
		asciiPath, err := findASCIIFileForScenario(outputDir, scenarioSlug)
		if err != nil {
			fmt.Fprintf(errOut, "Error finding baseline for %q: %v\n", scenario, err)
			return 1
		}

		gifPath := deriveGIFPath(asciiPath)
		if err := cukesvhs.UpdateBaseline(goldenDir, scenario, asciiPath, gifPath); err != nil {
			fmt.Fprintf(errOut, "Error updating baseline for %q: %v\n", scenario, err)
			return 1
		}
		fmt.Fprintf(out, "Updated: %s\n", scenario)
		updated++
	}

	fmt.Fprintf(out, "Updated %d baselines.\n", updated)
	return 0
}
