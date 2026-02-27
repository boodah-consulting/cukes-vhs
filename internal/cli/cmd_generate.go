package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

// newGenerateCmd creates the generate command.
func newGenerateCmd() *cobra.Command {
	var generateAll bool
	var featureFilter string
	var scenarioFilter string
	var featuresDir string
	var scenariosDir string
	var outputDir string
	var configSource string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate VHS tape files from scenarios",
		Long: `Generate VHS tape files from Gherkin scenarios.

Parses feature files and generates .tape files for translatable scenarios.
Use --all to generate for all translatable scenarios, or filter by feature
or scenario name.`,
		Example: `  cukes-vhs generate --all --output /tmp/tapes/
  cukes-vhs generate --all --features features/ --scenarios-dir demos/scenarios/ --output /tmp/tapes/
  cukes-vhs generate --feature onboarding --output /tmp/test/
  cukes-vhs generate --scenario "User logs in" --output /tmp/test/`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "all" {
				generateAll = true
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if outputDir == "" {
				return errors.New("--output is required")
			}
			if !generateAll && featureFilter == "" && scenarioFilter == "" {
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
			filtered := filterResults(results, allScenarios, generateAll, featureFilter, scenarioFilter)

			fmt.Fprintf(out, "Generating...\n")

			cfg := generateConfig{
				outputDir:    outputDir,
				configSource: configSource,
				verbose:      verbose,
				out:          out,
				errOut:       errOut,
			}
			stats := generateTapes(filtered, cfg)
			fmt.Fprintf(out, "Generated %d tapes (%d from features, %d from scenarios, %d warnings)\n",
				stats.total, stats.fromBusiness, stats.fromVHSOnly, stats.warnings)
			return nil
		},
	}

	cmd.Flags().BoolVar(&generateAll, "all", false, "Generate for all translatable scenarios")
	cmd.Flags().StringVar(&featureFilter, "feature", "", "Filter by feature name")
	cmd.Flags().StringVar(&scenarioFilter, "scenario", "", "Filter by scenario name")
	cmd.Flags().StringVar(&featuresDir, "features", "features/", "Directory containing .feature files")
	cmd.Flags().StringVar(&scenariosDir, "scenarios-dir", "demos/scenarios/", "Directory containing VHS-only .feature files")
	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory (required)")
	cmd.Flags().StringVar(&configSource, "config-source", "config/config.tape", "Path to config tape file")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")

	return cmd
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
	cmd := newGenerateCmd()
	cmd.SetArgs(normaliseArgs(args))
	cmd.SetOut(io.Discard)
	cmd.SetErr(errOut)

	opts := &generateOptions{
		generateAll:    new(bool),
		featureFilter:  new(string),
		scenarioFilter: new(string),
		featuresDir:    new(string),
		scenariosDir:   new(string),
		outputDir:      new(string),
		configSource:   new(string),
		verbose:        new(bool),
	}

	if err := cmd.ParseFlags(normaliseArgs(args)); err != nil {
		fmt.Fprintf(errOut, "Error parsing flags: %v\n", err)
		return nil, err
	}

	var err error
	*opts.generateAll, err = cmd.Flags().GetBool("all")
	if err != nil {
		return nil, err
	}
	*opts.featureFilter, err = cmd.Flags().GetString("feature")
	if err != nil {
		return nil, err
	}
	*opts.scenarioFilter, err = cmd.Flags().GetString("scenario")
	if err != nil {
		return nil, err
	}
	*opts.featuresDir, err = cmd.Flags().GetString("features")
	if err != nil {
		return nil, err
	}
	*opts.scenariosDir, err = cmd.Flags().GetString("scenarios-dir")
	if err != nil {
		return nil, err
	}
	*opts.outputDir, err = cmd.Flags().GetString("output")
	if err != nil {
		return nil, err
	}
	*opts.configSource, err = cmd.Flags().GetString("config-source")
	if err != nil {
		return nil, err
	}
	*opts.verbose, err = cmd.Flags().GetBool("verbose")
	if err != nil {
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
