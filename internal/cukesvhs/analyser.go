package cukesvhs

import "fmt"

// AnalyseScenarios checks each scenario's SetupSteps and DemoSteps for translatability.
//
// Expected: scenarios is a slice of parsed ScenarioIR with StepIR.Translatable already set.
// Returns: one AnalysisResult per scenario. Business untranslatable → warning; VHS-only → error.
// Side effects: None.
func AnalyseScenarios(scenarios []ScenarioIR) []AnalysisResult {
	results := make([]AnalysisResult, 0, len(scenarios))

	for i := range scenarios {
		s := &scenarios[i]
		result := AnalysisResult{
			ScenarioID:   BuildScenarioID(s.Source, s.Feature, s.Name),
			ScenarioName: s.Name,
			Feature:      s.Feature,
			Source:       s.Source,
			Translatable: true,
		}
		analyseSteps(s.SetupSteps, s.Source, &result)
		analyseSteps(s.DemoSteps, s.Source, &result)

		results = append(results, result)
	}

	return results
}

func analyseSteps(steps []StepIR, source SourceType, result *AnalysisResult) {
	for _, step := range steps {
		if step.Translatable {
			continue
		}

		result.Translatable = false
		result.UntranslatableSteps = append(result.UntranslatableSteps, step)

		switch source {
		case SourceBusiness:
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Step '%s' is not translatable: %s", step.Text, step.UntranslatableReason))
		case SourceVHSOnly:
			result.Errors = append(result.Errors,
				fmt.Sprintf("Step '%s' not found in mapping. Run `cukes-vhs list --steps` to see available steps.", step.Text))
		}
	}
}
