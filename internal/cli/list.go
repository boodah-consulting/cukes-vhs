package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

// runList implements the `list` subcommand.
func runList(args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	featuresDir := fs.String("features", "features/", "Directory containing .feature files")
	scenariosDir := fs.String("scenarios-dir", "demos/scenarios/", "Directory containing VHS-only .feature files")
	asJSON := fs.Bool("json", false, "Output as JSON")
	showCount := fs.Bool("count", false, "Show counts broken down by source")
	showSteps := fs.Bool("steps", false, "Show translatable step patterns")
	_ = fs.Bool("all", false, "List all scenarios (default behaviour)")

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

	businessScenarios, err := cukesvhs.ParseFeatureDir(*featuresDir, cukesvhs.SourceBusiness)
	if err != nil {
		fmt.Fprintf(errOut, "Error parsing features dir %q: %v\n", *featuresDir, err)
		return 1
	}

	vhsOnlyScenarios, err := cukesvhs.ParseFeatureDir(*scenariosDir, cukesvhs.SourceVHSOnly)
	if err != nil {
		fmt.Fprintf(errOut, "Error parsing scenarios dir %q: %v\n", *scenariosDir, err)
		return 1
	}

	allScenarios := make([]cukesvhs.ScenarioIR, 0, len(businessScenarios)+len(vhsOnlyScenarios))
	allScenarios = append(allScenarios, businessScenarios...)
	allScenarios = append(allScenarios, vhsOnlyScenarios...)
	results := cukesvhs.AnalyseScenarios(allScenarios)

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
func runListCount(results []cukesvhs.AnalysisResult, out io.Writer) int {
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

	return 0
}

// runListJSON outputs the analysis results as JSON.
func runListJSON(results []cukesvhs.AnalysisResult, out io.Writer, errOut io.Writer) int {
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
func runListTable(results []cukesvhs.AnalysisResult, out io.Writer) int {
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
