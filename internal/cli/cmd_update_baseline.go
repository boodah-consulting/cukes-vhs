package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

// newUpdateBaselineCmd creates the update-baseline command.
func newUpdateBaselineCmd() *cobra.Command {
	var updateAll bool
	var goldenDir string
	var outputDir string

	cmd := &cobra.Command{
		Use:   "update-baseline [scenarios...]",
		Short: "Accept current output as new golden baseline",
		Long: `Accept current rendered output as the new golden baseline.

Use --all to accept all current outputs, or specify scenario names to
update specific baselines.`,
		Example: `  cukes-vhs update-baseline --all --output /tmp/tapes/ --golden demos/golden/
  cukes-vhs update-baseline my-scenario --output /tmp/tapes/
  cukes-vhs update-baseline "User logs in" "User logs out" --output /tmp/tapes/`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "all" {
				updateAll = true
				return nil
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if outputDir == "" {
				return errors.New("--output is required")
			}
			positional := filterPositionalArgs(args, updateAll)
			if !updateAll && len(positional) == 0 {
				return errors.New("--all or a scenario name is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			positional := filterPositionalArgs(args, updateAll)

			if updateAll {
				return runUpdateAllBaselines(goldenDir, outputDir, out, errOut)
			}

			return runUpdateNamedBaselines(goldenDir, outputDir, positional, out, errOut)
		},
	}

	cmd.Flags().BoolVar(&updateAll, "all", false, "Accept all current outputs as golden baselines")
	cmd.Flags().StringVar(&goldenDir, "golden", "demos/golden/", "Golden baseline directory")
	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory containing rendered .ascii files (required)")

	return cmd
}

// filterPositionalArgs filters out "all" if updateAll is true.
func filterPositionalArgs(args []string, updateAll bool) []string {
	if len(args) == 0 {
		return args
	}
	if updateAll && args[0] == "all" {
		return args[1:]
	}
	return args
}

// runUpdateAllBaselines scans outputDir for .ascii files and accepts each as the golden baseline.
func runUpdateAllBaselines(goldenDir, outputDir string, out, errOut io.Writer) error {
	asciiFiles, err := collectOutputASCIIFiles(outputDir)
	if err != nil {
		fmt.Fprintf(errOut, "Error scanning output dir %q: %v\n", outputDir, err)
		return err
	}

	if len(asciiFiles) == 0 {
		fmt.Fprintf(out, "No .ascii files found in %q.\n", outputDir)
		fmt.Fprintf(out, "Updated 0 baselines.\n")
		return nil
	}

	updated := 0

	for _, asciiPath := range asciiFiles {
		scenario := deriveScenarioName(outputDir, asciiPath)
		gifPath := deriveGIFPath(asciiPath)

		if err := cukesvhs.UpdateBaseline(goldenDir, scenario, asciiPath, gifPath); err != nil {
			fmt.Fprintf(errOut, "Error updating baseline for %q: %v\n", scenario, err)
			return err
		}

		fmt.Fprintf(out, "Updated: %s\n", scenario)
		updated++
	}

	fmt.Fprintf(out, "Updated %d baselines.\n", updated)

	return nil
}

// runUpdateNamedBaselines accepts specific named scenarios as golden baselines.
func runUpdateNamedBaselines(goldenDir, outputDir string, scenarios []string, out, errOut io.Writer) error {
	updated := 0
	for _, scenario := range scenarios {
		scenarioSlug := cukesvhs.Slugify(scenario)
		asciiPath, err := findASCIIFileForScenario(outputDir, scenarioSlug)
		if err != nil {
			fmt.Fprintf(errOut, "Error finding baseline for %q: %v\n", scenario, err)
			return err
		}

		gifPath := deriveGIFPath(asciiPath)
		if err := cukesvhs.UpdateBaseline(goldenDir, scenario, asciiPath, gifPath); err != nil {
			fmt.Fprintf(errOut, "Error updating baseline for %q: %v\n", scenario, err)
			return err
		}
		fmt.Fprintf(out, "Updated: %s\n", scenario)
		updated++
	}

	fmt.Fprintf(out, "Updated %d baselines.\n", updated)
	return nil
}

// updateBaselineOptions holds parsed flags for the update-baseline subcommand.
type updateBaselineOptions struct {
	updateAll *bool
	goldenDir *string
	outputDir *string
}

// parseUpdateBaselineFlags parses the flags for the update-baseline subcommand.
func parseUpdateBaselineFlags(args []string, errOut io.Writer) (*updateBaselineOptions, []string, error) {
	cmd := newUpdateBaselineCmd()
	cmd.SetArgs(normaliseArgs(args))
	cmd.SetOut(io.Discard)
	cmd.SetErr(errOut)

	opts := &updateBaselineOptions{
		updateAll: new(bool),
		goldenDir: new(string),
		outputDir: new(string),
	}

	if err := cmd.ParseFlags(normaliseArgs(args)); err != nil {
		fmt.Fprintf(errOut, "Error parsing flags: %v\n", err)
		return nil, nil, err
	}

	*opts.updateAll, _ = cmd.Flags().GetBool("all")
	*opts.goldenDir, _ = cmd.Flags().GetString("golden")
	*opts.outputDir, _ = cmd.Flags().GetString("output")

	positionalArgs := cmd.Flags().Args()
	for _, arg := range args {
		if arg == "all" || arg == "--all" {
			*opts.updateAll = true
			break
		}
	}

	if *opts.outputDir == "" {
		fmt.Fprintf(errOut, "Error: --output is required\n")
		return nil, nil, errors.New("output required")
	}

	return opts, positionalArgs, nil
}

// Legacy wrappers for backward compatibility with tests

func updateAllBaselines(goldenDir, outputDir string, out, errOut io.Writer) int {
	if err := runUpdateAllBaselines(goldenDir, outputDir, out, errOut); err != nil {
		return 1
	}
	return 0
}
