package cukesvhs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukesvhs/internal/cukesvhs"
)

var _ = Describe("AnalyseScenarios", func() {
	Describe("business scenario with all translatable steps", func() {
		var result cukesvhs.AnalysisResult

		BeforeEach(func() {
			scenarios := []cukesvhs.ScenarioIR{
				{
					Name:    "Navigate menu",
					Feature: "Navigation",
					Source:  cukesvhs.SourceBusiness,
					SetupSteps: []cukesvhs.StepIR{
						{Text: "I am on the main menu", StepType: "Given", Translatable: true},
					},
					DemoSteps: []cukesvhs.StepIR{
						{Text: `I select "manage_skills" from the menu`, StepType: "When", Translatable: true},
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
			}

			results := cukesvhs.AnalyseScenarios(scenarios)
			Expect(results).To(HaveLen(1))
			result = results[0]
		})

		It("is translatable", func() {
			Expect(result.Translatable).To(BeTrue())
		})

		It("has no warnings", func() {
			Expect(result.Warnings).To(BeEmpty())
		})

		It("has no errors", func() {
			Expect(result.Errors).To(BeEmpty())
		})

		It("has no untranslatable steps", func() {
			Expect(result.UntranslatableSteps).To(BeEmpty())
		})
	})

	Describe("business scenario with form-bypass step", func() {
		var result cukesvhs.AnalysisResult

		BeforeEach(func() {
			scenarios := []cukesvhs.ScenarioIR{
				{
					Name:    "Submit event",
					Feature: "Capture Event",
					Source:  cukesvhs.SourceBusiness,
					DemoSteps: []cukesvhs.StepIR{
						{Text: `I enter event description "Built API"`, StepType: "When", Translatable: true},
						{Text: "I submit the event", StepType: "When", Translatable: false, UntranslatableReason: "form-bypass: use keyboard navigation instead"},
					},
				},
			}

			results := cukesvhs.AnalyseScenarios(scenarios)
			Expect(results).To(HaveLen(1))
			result = results[0]
		})

		It("is not translatable", func() {
			Expect(result.Translatable).To(BeFalse())
		})

		It("has exactly 1 warning", func() {
			Expect(result.Warnings).To(HaveLen(1))
		})

		It("has no errors", func() {
			Expect(result.Errors).To(BeEmpty())
		})

		It("has exactly 1 untranslatable step", func() {
			Expect(result.UntranslatableSteps).To(HaveLen(1))
		})

		It("records the correct untranslatable step text", func() {
			Expect(result.UntranslatableSteps[0].Text).To(Equal("I submit the event"))
		})
	})

	Describe("VHS-only scenario with untranslatable step", func() {
		var result cukesvhs.AnalysisResult

		BeforeEach(func() {
			scenarios := []cukesvhs.ScenarioIR{
				{
					Name:    "Demo custom flow",
					Feature: "VHS Demo",
					Source:  cukesvhs.SourceVHSOnly,
					DemoSteps: []cukesvhs.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
						{Text: "I do something custom", StepType: "When", Translatable: false, UntranslatableReason: "unknown step: no matching pattern"},
					},
				},
			}

			results := cukesvhs.AnalyseScenarios(scenarios)
			Expect(results).To(HaveLen(1))
			result = results[0]
		})

		It("is not translatable", func() {
			Expect(result.Translatable).To(BeFalse())
		})

		It("has no warnings for VHS-only", func() {
			Expect(result.Warnings).To(BeEmpty())
		})

		It("has exactly 1 error", func() {
			Expect(result.Errors).To(HaveLen(1))
		})

		It("produces a prescriptive error message", func() {
			wantErr := "Step 'I do something custom' not found in mapping. Run `cukes-vhs list --steps` to see available steps."
			Expect(result.Errors[0]).To(Equal(wantErr))
		})
	})

	Describe("untranslatable setup steps affect translatability", func() {
		var result cukesvhs.AnalysisResult

		BeforeEach(func() {
			scenarios := []cukesvhs.ScenarioIR{
				{
					Name:    "Setup blocks translation",
					Feature: "Resilience",
					Source:  cukesvhs.SourceBusiness,
					SetupSteps: []cukesvhs.StepIR{
						{Text: "some untranslatable setup", StepType: "Given", Translatable: false, UntranslatableReason: "unknown"},
						{Text: "some untranslatable assertion", StepType: "Then", Translatable: false, UntranslatableReason: "unknown"},
					},
					DemoSteps: []cukesvhs.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
			}

			results := cukesvhs.AnalyseScenarios(scenarios)
			Expect(results).To(HaveLen(1))
			result = results[0]
		})

		It("is not translatable", func() {
			Expect(result.Translatable).To(BeFalse())
		})

		It("has 2 warnings for business source", func() {
			Expect(result.Warnings).To(HaveLen(2))
		})
		It("has no errors", func() {
			Expect(result.Errors).To(BeEmpty())
		})

		It("records 2 untranslatable steps", func() {
			Expect(result.UntranslatableSteps).To(HaveLen(2))
		})
	})

	Describe("VHS-only scenario with untranslatable setup step", func() {
		var result cukesvhs.AnalysisResult

		BeforeEach(func() {
			scenarios := []cukesvhs.ScenarioIR{
				{
					Name:    "VHS setup fails",
					Feature: "VHS Demo",
					Source:  cukesvhs.SourceVHSOnly,
					SetupSteps: []cukesvhs.StepIR{
						{Text: "unknown setup action", StepType: "Given", Translatable: false, UntranslatableReason: "no matching pattern"},
					},
					DemoSteps: []cukesvhs.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
			}

			results := cukesvhs.AnalyseScenarios(scenarios)
			Expect(results).To(HaveLen(1))
			result = results[0]
		})

		It("is not translatable", func() {
			Expect(result.Translatable).To(BeFalse())
		})
		It("has no warnings", func() {
			Expect(result.Warnings).To(BeEmpty())
		})

		It("has 1 error for VHS-only source", func() {
			Expect(result.Errors).To(HaveLen(1))
		})

		It("records 1 untranslatable step", func() {
			Expect(result.UntranslatableSteps).To(HaveLen(1))
		})

		It("produces a prescriptive error message", func() {
			Expect(result.Errors[0]).To(ContainSubstring("unknown setup action"))
		})
	})

	Describe("Source field propagation", func() {
		var results []cukesvhs.AnalysisResult

		BeforeEach(func() {
			scenarios := []cukesvhs.ScenarioIR{
				{
					Name:   "Business scenario",
					Source: cukesvhs.SourceBusiness,
					DemoSteps: []cukesvhs.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
				{
					Name:   "VHS-only scenario",
					Source: cukesvhs.SourceVHSOnly,
					DemoSteps: []cukesvhs.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
			}

			results = cukesvhs.AnalyseScenarios(scenarios)
		})

		It("returns 2 results", func() {
			Expect(results).To(HaveLen(2))
		})

		It("propagates SourceBusiness to the first result", func() {
			Expect(results[0].Source).To(Equal(cukesvhs.SourceBusiness))
		})

		It("propagates SourceVHSOnly to the second result", func() {
			Expect(results[1].Source).To(Equal(cukesvhs.SourceVHSOnly))
		})
	})

	Describe("ScenarioName and Feature propagation", func() {
		It("copies ScenarioName and Feature from the input", func() {
			scenarios := []cukesvhs.ScenarioIR{
				{
					Name:    "My Scenario",
					Feature: "My Feature",
					Source:  cukesvhs.SourceBusiness,
					DemoSteps: []cukesvhs.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
			}

			results := cukesvhs.AnalyseScenarios(scenarios)
			Expect(results[0].ScenarioName).To(Equal("My Scenario"))
			Expect(results[0].Feature).To(Equal("My Feature"))
		})
	})

	Describe("ScenarioID generation", func() {
		It("populates ScenarioID as a composite of source, feature, and name", func() {
			scenarios := []cukesvhs.ScenarioIR{
				{
					Name:    "User logs in",
					Feature: "Authentication",
					Source:  cukesvhs.SourceBusiness,
					DemoSteps: []cukesvhs.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
			}

			results := cukesvhs.AnalyseScenarios(scenarios)
			Expect(results[0].ScenarioID).To(Equal("business/authentication/user-logs-in"))
		})

		It("produces unique IDs for same-named scenarios in different features", func() {
			scenarios := []cukesvhs.ScenarioIR{
				{
					Name:    "User logs in",
					Feature: "Feature A",
					Source:  cukesvhs.SourceBusiness,
					DemoSteps: []cukesvhs.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
				{
					Name:    "User logs in",
					Feature: "Feature B",
					Source:  cukesvhs.SourceBusiness,
					DemoSteps: []cukesvhs.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
			}

			results := cukesvhs.AnalyseScenarios(scenarios)
			Expect(results).To(HaveLen(2))
			Expect(results[0].ScenarioID).NotTo(Equal(results[1].ScenarioID))
			Expect(results[0].ScenarioID).To(Equal("business/feature-a/user-logs-in"))
			Expect(results[1].ScenarioID).To(Equal("business/feature-b/user-logs-in"))
		})
	})
	Describe("multiple untranslatable When steps", func() {
		var result cukesvhs.AnalysisResult

		BeforeEach(func() {
			scenarios := []cukesvhs.ScenarioIR{
				{
					Name:   "Multiple blockers",
					Source: cukesvhs.SourceVHSOnly,
					DemoSteps: []cukesvhs.StepIR{
						{Text: "I do thing A", StepType: "When", Translatable: false, UntranslatableReason: "unknown"},
						{Text: "I press enter", StepType: "When", Translatable: true},
						{Text: "I do thing B", StepType: "When", Translatable: false, UntranslatableReason: "unknown"},
					},
				},
			}

			results := cukesvhs.AnalyseScenarios(scenarios)
			result = results[0]
		})

		It("is not translatable", func() {
			Expect(result.Translatable).To(BeFalse())
		})

		It("records 2 untranslatable steps", func() {
			Expect(result.UntranslatableSteps).To(HaveLen(2))
		})

		It("records 2 errors for VHS-only", func() {
			Expect(result.Errors).To(HaveLen(2))
		})
	})

	Describe("empty demo steps", func() {
		It("is translatable when DemoSteps is empty", func() {
			scenarios := []cukesvhs.ScenarioIR{
				{
					Name:   "Setup only",
					Source: cukesvhs.SourceBusiness,
					SetupSteps: []cukesvhs.StepIR{
						{Text: "I am on the main menu", StepType: "Given", Translatable: true},
					},
					DemoSteps: []cukesvhs.StepIR{},
				},
			}

			results := cukesvhs.AnalyseScenarios(scenarios)
			Expect(results[0].Translatable).To(BeTrue())
		})
	})
})
