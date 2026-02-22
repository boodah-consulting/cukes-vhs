package vhsgen_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/baphled/cukes-vhs/internal/vhsgen"
)

var _ = Describe("AnalyseScenarios", func() {
	Describe("business scenario with all translatable steps", func() {
		var result vhsgen.AnalysisResult

		BeforeEach(func() {
			scenarios := []vhsgen.ScenarioIR{
				{
					Name:    "Navigate menu",
					Feature: "Navigation",
					Source:  vhsgen.SourceBusiness,
					SetupSteps: []vhsgen.StepIR{
						{Text: "I am on the main menu", StepType: "Given", Translatable: true},
					},
					DemoSteps: []vhsgen.StepIR{
						{Text: `I select "manage_skills" from the menu`, StepType: "When", Translatable: true},
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
			}

			results := vhsgen.AnalyseScenarios(scenarios)
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
		var result vhsgen.AnalysisResult

		BeforeEach(func() {
			scenarios := []vhsgen.ScenarioIR{
				{
					Name:    "Submit event",
					Feature: "Capture Event",
					Source:  vhsgen.SourceBusiness,
					DemoSteps: []vhsgen.StepIR{
						{Text: `I enter event description "Built API"`, StepType: "When", Translatable: true},
						{Text: "I submit the event", StepType: "When", Translatable: false, UntranslatableReason: "form-bypass: use keyboard navigation instead"},
					},
				},
			}

			results := vhsgen.AnalyseScenarios(scenarios)
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
		var result vhsgen.AnalysisResult

		BeforeEach(func() {
			scenarios := []vhsgen.ScenarioIR{
				{
					Name:    "Demo custom flow",
					Feature: "VHS Demo",
					Source:  vhsgen.SourceVHSOnly,
					DemoSteps: []vhsgen.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
						{Text: "I do something custom", StepType: "When", Translatable: false, UntranslatableReason: "unknown step: no matching pattern"},
					},
				},
			}

			results := vhsgen.AnalyseScenarios(scenarios)
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
			wantErr := "Step 'I do something custom' not found in mapping. Run `vhsgen list --steps` to see available steps."
			Expect(result.Errors[0]).To(Equal(wantErr))
		})
	})

	Describe("Given/Then steps do not affect translatability", func() {
		var result vhsgen.AnalysisResult

		BeforeEach(func() {
			scenarios := []vhsgen.ScenarioIR{
				{
					Name:    "Setup does not block",
					Feature: "Resilience",
					Source:  vhsgen.SourceBusiness,
					SetupSteps: []vhsgen.StepIR{
						{Text: "some untranslatable setup", StepType: "Given", Translatable: false, UntranslatableReason: "unknown"},
						{Text: "some untranslatable assertion", StepType: "Then", Translatable: false, UntranslatableReason: "unknown"},
					},
					DemoSteps: []vhsgen.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
			}

			results := vhsgen.AnalyseScenarios(scenarios)
			Expect(results).To(HaveLen(1))
			result = results[0]
		})

		It("is translatable because only When steps matter", func() {
			Expect(result.Translatable).To(BeTrue())
		})

		It("has no warnings", func() {
			Expect(result.Warnings).To(BeEmpty())
		})

		It("has no errors", func() {
			Expect(result.Errors).To(BeEmpty())
		})
	})

	Describe("Source field propagation", func() {
		var results []vhsgen.AnalysisResult

		BeforeEach(func() {
			scenarios := []vhsgen.ScenarioIR{
				{
					Name:   "Business scenario",
					Source: vhsgen.SourceBusiness,
					DemoSteps: []vhsgen.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
				{
					Name:   "VHS-only scenario",
					Source: vhsgen.SourceVHSOnly,
					DemoSteps: []vhsgen.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
			}

			results = vhsgen.AnalyseScenarios(scenarios)
		})

		It("returns 2 results", func() {
			Expect(results).To(HaveLen(2))
		})

		It("propagates SourceBusiness to the first result", func() {
			Expect(results[0].Source).To(Equal(vhsgen.SourceBusiness))
		})

		It("propagates SourceVHSOnly to the second result", func() {
			Expect(results[1].Source).To(Equal(vhsgen.SourceVHSOnly))
		})
	})

	Describe("ScenarioName and Feature propagation", func() {
		It("copies ScenarioName and Feature from the input", func() {
			scenarios := []vhsgen.ScenarioIR{
				{
					Name:    "My Scenario",
					Feature: "My Feature",
					Source:  vhsgen.SourceBusiness,
					DemoSteps: []vhsgen.StepIR{
						{Text: "I press enter", StepType: "When", Translatable: true},
					},
				},
			}

			results := vhsgen.AnalyseScenarios(scenarios)
			Expect(results[0].ScenarioName).To(Equal("My Scenario"))
			Expect(results[0].Feature).To(Equal("My Feature"))
		})
	})

	Describe("multiple untranslatable When steps", func() {
		var result vhsgen.AnalysisResult

		BeforeEach(func() {
			scenarios := []vhsgen.ScenarioIR{
				{
					Name:   "Multiple blockers",
					Source: vhsgen.SourceVHSOnly,
					DemoSteps: []vhsgen.StepIR{
						{Text: "I do thing A", StepType: "When", Translatable: false, UntranslatableReason: "unknown"},
						{Text: "I press enter", StepType: "When", Translatable: true},
						{Text: "I do thing B", StepType: "When", Translatable: false, UntranslatableReason: "unknown"},
					},
				},
			}

			results := vhsgen.AnalyseScenarios(scenarios)
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
			scenarios := []vhsgen.ScenarioIR{
				{
					Name:   "Setup only",
					Source: vhsgen.SourceBusiness,
					SetupSteps: []vhsgen.StepIR{
						{Text: "I am on the main menu", StepType: "Given", Translatable: true},
					},
					DemoSteps: []vhsgen.StepIR{},
				},
			}

			results := vhsgen.AnalyseScenarios(scenarios)
			Expect(results[0].Translatable).To(BeTrue())
		})
	})
})
