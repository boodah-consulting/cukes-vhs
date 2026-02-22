package vhsgen

import "fmt"

// AnalyseScenarios checks each scenario's DemoSteps for translatability.
//
// Expected: scenarios is a slice of parsed ScenarioIR with StepIR.Translatable already set.
// Returns: one AnalysisResult per scenario. Business untranslatable → warning; VHS-only → error.
// Side effects: None.
func AnalyseScenarios(scenarios []ScenarioIR) []AnalysisResult {
	results := make([]AnalysisResult, 0, len(scenarios))

	for i := range scenarios {
		s := &scenarios[i]
		result := AnalysisResult{
			ScenarioName: s.Name,
			Feature:      s.Feature,
			Source:       s.Source,
			Translatable: true,
		}

		for _, step := range s.DemoSteps {
			if step.Translatable {
				continue
			}

			result.Translatable = false
			result.UntranslatableSteps = append(result.UntranslatableSteps, step)

			switch s.Source {
			case SourceBusiness:
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Step '%s' is not translatable: %s", step.Text, step.UntranslatableReason))
			case SourceVHSOnly:
				result.Errors = append(result.Errors,
					fmt.Sprintf("Step '%s' not found in mapping. Run `vhsgen list --steps` to see available steps.", step.Text))
			}
		}

		results = append(results, result)
	}

	return results
}
