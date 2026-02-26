package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

// newListCmd creates the list command.
func newListCmd() *cobra.Command {
	var featuresDir string
	var scenariosDir string
	var asJSON bool
	var showCount bool
	var showSteps bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List scenarios and their translatability",
		Long: `List all scenarios from feature files and show their translatability status.

By default, outputs a formatted table showing each scenario with its feature,
source (business or VHS-only), and whether it can be translated to VHS commands.`,
		Example: `  cukes-vhs list
  cukes-vhs list --features features/ --scenarios-dir demos/scenarios/
  cukes-vhs list --json
  cukes-vhs list --count
  cukes-vhs list --steps
  cukes-vhs list --steps --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Context()

			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			if showSteps {
				return runListStepsCmd(asJSON, out)
			}

			fmt.Fprintf(out, "Parsing...\n")

			if _, err := cliFs.Stat(featuresDir); err != nil {
				fmt.Fprintf(errOut, "Error parsing features dir %q: %v\n", featuresDir, err)
				return err
			}

			businessScenarios, err := cukesvhs.ParseFeatureDir(featuresDir, cukesvhs.SourceBusiness)
			if err != nil {
				fmt.Fprintf(errOut, "Error parsing features dir %q: %v\n", featuresDir, err)
				return err
			}

			vhsOnlyScenarios, err := cukesvhs.ParseFeatureDir(scenariosDir, cukesvhs.SourceVHSOnly)
			if err != nil {
				fmt.Fprintf(errOut, "Error parsing scenarios dir %q: %v\n", scenariosDir, err)
				return err
			}

			allScenarios := make([]cukesvhs.ScenarioIR, 0, len(businessScenarios)+len(vhsOnlyScenarios))
			allScenarios = append(allScenarios, businessScenarios...)
			allScenarios = append(allScenarios, vhsOnlyScenarios...)
			results := cukesvhs.AnalyseScenarios(allScenarios)

			if showCount {
				return runListCountCmd(results, out)
			}

			if asJSON {
				return runListJSONCmd(results, out, errOut)
			}

			return runListTableCmd(results, out)
		},
	}

	cmd.Flags().StringVar(&featuresDir, "features", "features/", "Directory containing .feature files")
	cmd.Flags().StringVar(&scenariosDir, "scenarios-dir", "demos/scenarios/", "Directory containing VHS-only .feature files")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&showCount, "count", false, "Show counts broken down by source")
	cmd.Flags().BoolVar(&showSteps, "steps", false, "Show translatable step patterns")

	return cmd
}

// runListStepsCmd outputs the translatable step patterns.
func runListStepsCmd(asJSON bool, out io.Writer) error {
	patterns := cukesvhs.ListTranslatablePatterns()

	if asJSON {
		type jsonPattern struct {
			Pattern  string                              `json:"pattern"`
			Type     string                              `json:"type"`
			Category string                              `json:"category"`
			Params   map[string]cukesvhs.ParamConstraint `json:"params,omitempty"`
			Example  string                              `json:"example"`
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
		return enc.Encode(output)
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

	return nil
}

// runListCountCmd outputs counts by source.
func runListCountCmd(results []cukesvhs.AnalysisResult, out io.Writer) error {
	var (
		businessTotal        int
		businessTranslatable int
		vhsOnlyTotal         int
		vhsOnlyTranslatable  int
	)

	for i := range results {
		r := &results[i]
		switch r.Source {
		case cukesvhs.SourceBusiness:
			businessTotal++
			if r.Translatable {
				businessTranslatable++
			}
		case cukesvhs.SourceVHSOnly:
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

	return nil
}

// runListJSONCmd outputs the analysis results as JSON.
func runListJSONCmd(results []cukesvhs.AnalysisResult, out io.Writer, errOut io.Writer) error {
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
		return err
	}

	return nil
}

// runListTableCmd outputs the analysis results as a formatted table.
func runListTableCmd(results []cukesvhs.AnalysisResult, out io.Writer) error {
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

	return nil
}

// Legacy wrappers for backward compatibility with tests

func runListCount(results []cukesvhs.AnalysisResult, out io.Writer) int {
	if err := runListCountCmd(results, out); err != nil {
		return 1
	}
	return 0
}

func runListJSON(results []cukesvhs.AnalysisResult, out io.Writer, errOut io.Writer) int {
	if err := runListJSONCmd(results, out, errOut); err != nil {
		return 1
	}
	return 0
}
