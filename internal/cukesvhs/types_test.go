package cukesvhs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

var _ = Describe("Types", func() {
	It("defines source type constants", func() {
		Expect(cukesvhs.SourceBusiness).To(Equal(cukesvhs.SourceType("business")))
		Expect(cukesvhs.SourceVHSOnly).To(Equal(cukesvhs.SourceType("vhs-only")))
	})

	It("defines VHS command type constants", func() {
		Expect(cukesvhs.Type).To(Equal(cukesvhs.VHSCommandType("Type")))
		Expect(cukesvhs.Down).To(Equal(cukesvhs.VHSCommandType("Down")))
		Expect(cukesvhs.Up).To(Equal(cukesvhs.VHSCommandType("Up")))
		Expect(cukesvhs.Enter).To(Equal(cukesvhs.VHSCommandType("Enter")))
		Expect(cukesvhs.Escape).To(Equal(cukesvhs.VHSCommandType("Escape")))
		Expect(cukesvhs.Tab).To(Equal(cukesvhs.VHSCommandType("Tab")))
		Expect(cukesvhs.Sleep).To(Equal(cukesvhs.VHSCommandType("Sleep")))
		Expect(cukesvhs.Hide).To(Equal(cukesvhs.VHSCommandType("Hide")))
		Expect(cukesvhs.Show).To(Equal(cukesvhs.VHSCommandType("Show")))
		Expect(cukesvhs.Screenshot).To(Equal(cukesvhs.VHSCommandType("Screenshot")))
		Expect(cukesvhs.Source).To(Equal(cukesvhs.VHSCommandType("Source")))
		Expect(cukesvhs.Output).To(Equal(cukesvhs.VHSCommandType("Output")))
		Expect(cukesvhs.CtrlC).To(Equal(cukesvhs.VHSCommandType("Ctrl+C")))
	})

	It("constructs VHSCommand with various configurations", func() {
		cmdWithArgs := cukesvhs.VHSCommand{Type: cukesvhs.Type, Args: []string{"hello"}}
		Expect(cmdWithArgs.Type).To(Equal(cukesvhs.Type))
		Expect(cmdWithArgs.Args).To(Equal([]string{"hello"}))

		cmdSleep := cukesvhs.VHSCommand{Type: cukesvhs.Sleep, Args: []string{"500ms"}}
		Expect(cmdSleep.Type).To(Equal(cukesvhs.Sleep))
		Expect(cmdSleep.Args).To(Equal([]string{"500ms"}))

		cmdScreenshot := cukesvhs.VHSCommand{Type: cukesvhs.Screenshot, Args: []string{"output.png"}}
		Expect(cmdScreenshot.Type).To(Equal(cukesvhs.Screenshot))
		Expect(cmdScreenshot.Args).To(Equal([]string{"output.png"}))

		cmdNoArgs := cukesvhs.VHSCommand{Type: cukesvhs.Enter}
		Expect(cmdNoArgs.Type).To(Equal(cukesvhs.Enter))
		Expect(cmdNoArgs.Args).To(BeNil())
	})

	It("constructs StepIR with different step types", func() {
		givenStep := cukesvhs.StepIR{
			Text:         "a user is logged in",
			StepType:     "Given",
			Translatable: true,
		}
		Expect(givenStep.Text).To(Equal("a user is logged in"))
		Expect(givenStep.StepType).To(Equal("Given"))

		whenStep := cukesvhs.StepIR{
			Text:         "the user clicks the button",
			StepType:     "When",
			Translatable: true,
		}
		Expect(whenStep.Text).To(Equal("the user clicks the button"))
		Expect(whenStep.StepType).To(Equal("When"))

		thenStep := cukesvhs.StepIR{
			Text:         "the page should display success",
			StepType:     "Then",
			Translatable: true,
		}
		Expect(thenStep.Text).To(Equal("the page should display success"))
		Expect(thenStep.StepType).To(Equal("Then"))

		untranslatableStep := cukesvhs.StepIR{
			Text:                 "some complex step",
			StepType:             "When",
			Translatable:         false,
			UntranslatableReason: "no matching pattern",
		}
		Expect(untranslatableStep.Text).To(Equal("some complex step"))
		Expect(untranslatableStep.StepType).To(Equal("When"))
	})

	Describe("StepIR with commands", func() {
		It("stores multiple commands with correct types", func() {
			step := cukesvhs.StepIR{
				Text:     "the user types their name",
				StepType: "When",
				Commands: []cukesvhs.VHSCommand{
					{Type: cukesvhs.Type, Args: []string{"John Doe"}},
					{Type: cukesvhs.Tab},
				},
				Translatable: true,
			}
			Expect(step.Commands).To(HaveLen(2))
			Expect(step.Commands[0].Type).To(Equal(cukesvhs.Type))
			Expect(step.Commands[1].Type).To(Equal(cukesvhs.Tab))
		})
	})

	It("constructs ScenarioIR with all fields", func() {
		scenario := cukesvhs.ScenarioIR{
			Name:    "User login",
			Feature: "Authentication",
			Tags:    []string{"@critical", "@smoke"},
			Source:  cukesvhs.SourceBusiness,
			SetupSteps: []cukesvhs.StepIR{
				{Text: "setup step", StepType: "Given", Translatable: true},
			},
			DemoSteps: []cukesvhs.StepIR{
				{Text: "demo step", StepType: "When", Translatable: true},
			},
			Translatable: true,
		}

		Expect(scenario.Name).To(Equal("User login"))
		Expect(scenario.Feature).To(Equal("Authentication"))
		Expect(scenario.Source).To(Equal(cukesvhs.SourceBusiness))
		Expect(scenario.Tags).To(HaveLen(2))
		Expect(scenario.SetupSteps).To(HaveLen(1))
		Expect(scenario.DemoSteps).To(HaveLen(1))
	})

	It("accepts different source types", func() {
		businessScenario := cukesvhs.ScenarioIR{Name: "Test scenario", Source: cukesvhs.SourceBusiness}
		Expect(businessScenario.Source).To(Equal(cukesvhs.SourceBusiness))

		vhsScenario := cukesvhs.ScenarioIR{Name: "Test scenario", Source: cukesvhs.SourceVHSOnly}
		Expect(vhsScenario.Source).To(Equal(cukesvhs.SourceVHSOnly))
	})

	It("constructs GeneratorConfig with all fields", func() {
		config := cukesvhs.GeneratorConfig{
			OutputDir:        "/tmp/output",
			ConfigSourcePath: "demos/vhs/config.tape",
			SleepDuration:    "500ms",
		}

		Expect(config.OutputDir).To(Equal("/tmp/output"))
		Expect(config.ConfigSourcePath).To(Equal("demos/vhs/config.tape"))
		Expect(config.SleepDuration).To(Equal("500ms"))
	})

	It("constructs AnalysisResult with all fields", func() {
		result := cukesvhs.AnalysisResult{
			ScenarioName: "Login flow",
			Feature:      "Authentication",
			Translatable: true,
			Source:       cukesvhs.SourceBusiness,
			Warnings:     []string{"slow step detected"},
			Errors:       []string{},
		}

		Expect(result.ScenarioName).To(Equal("Login flow"))
		Expect(result.Feature).To(Equal("Authentication"))
		Expect(result.Translatable).To(BeTrue())
		Expect(result.Source).To(Equal(cukesvhs.SourceBusiness))
		Expect(result.Warnings).To(HaveLen(1))
		Expect(result.Errors).To(BeEmpty())
	})

	It("constructs AnalysisResult with untranslatable steps", func() {
		untranslatableStep := cukesvhs.StepIR{
			Text:                 "complex step",
			StepType:             "When",
			Translatable:         false,
			UntranslatableReason: "no pattern match",
		}
		result := cukesvhs.AnalysisResult{
			ScenarioName:        "Complex scenario",
			Feature:             "Advanced",
			Translatable:        false,
			UntranslatableSteps: []cukesvhs.StepIR{untranslatableStep},
			Source:              cukesvhs.SourceVHSOnly,
			Errors:              []string{"cannot translate scenario"},
		}

		Expect(result.Translatable).To(BeFalse())
		Expect(result.UntranslatableSteps).To(HaveLen(1))
		Expect(result.UntranslatableSteps[0].Text).To(Equal("complex step"))
	})

	It("constructs ParamConstraint", func() {
		constraint := cukesvhs.ParamConstraint{
			Type:   "enum",
			Values: []string{"value1", "value2", "value3"},
		}

		Expect(constraint.Type).To(Equal("enum"))
		Expect(constraint.Values).To(HaveLen(3))
		Expect(constraint.Values[0]).To(Equal("value1"))
	})

	It("constructs StepPattern", func() {
		pattern := cukesvhs.StepPattern{
			Pattern:  `the user types "([^"]+)"`,
			Type:     "When",
			Category: "input",
			Params: map[string]cukesvhs.ParamConstraint{
				"text": {Type: "string", Values: nil},
			},
			Example: `the user types "hello"`,
		}

		Expect(pattern.Pattern).To(Equal(`the user types "([^"]+)"`))
		Expect(pattern.Type).To(Equal("When"))
		Expect(pattern.Category).To(Equal("input"))
		Expect(pattern.Params).To(HaveLen(1))
		Expect(pattern.Example).To(Equal(`the user types "hello"`))
	})

	It("has sensible zero values for all types", func() {
		var cmd cukesvhs.VHSCommand
		Expect(string(cmd.Type)).To(BeEmpty())
		Expect(cmd.Args).To(BeNil())

		var step cukesvhs.StepIR
		Expect(step.Text).To(BeEmpty())
		Expect(step.Translatable).To(BeFalse())

		var scenario cukesvhs.ScenarioIR
		Expect(scenario.Name).To(BeEmpty())
		Expect(string(scenario.Source)).To(BeEmpty())

		var config cukesvhs.GeneratorConfig
		Expect(config.OutputDir).To(BeEmpty())

		var result cukesvhs.AnalysisResult
		Expect(result.ScenarioName).To(BeEmpty())
		Expect(result.Translatable).To(BeFalse())

		var constraint cukesvhs.ParamConstraint
		Expect(constraint.Type).To(BeEmpty())

		var pattern cukesvhs.StepPattern
		Expect(pattern.Pattern).To(BeEmpty())
		Expect(pattern.Params).To(BeNil())
	})

	Describe("Empty slices", func() {
		Context("ScenarioIR with empty slices", func() {
			It("has zero tags, setup steps, and demo steps", func() {
				scenario := cukesvhs.ScenarioIR{
					Name:       "Empty scenario",
					Tags:       []string{},
					SetupSteps: []cukesvhs.StepIR{},
					DemoSteps:  []cukesvhs.StepIR{},
				}
				Expect(scenario.Tags).To(BeEmpty())
				Expect(scenario.SetupSteps).To(BeEmpty())
				Expect(scenario.DemoSteps).To(BeEmpty())
			})
		})

		Context("AnalysisResult with empty slices", func() {
			It("has zero untranslatable steps, warnings, and errors", func() {
				result := cukesvhs.AnalysisResult{
					ScenarioName:        "Empty result",
					UntranslatableSteps: []cukesvhs.StepIR{},
					Warnings:            []string{},
					Errors:              []string{},
				}
				Expect(result.UntranslatableSteps).To(BeEmpty())
				Expect(result.Warnings).To(BeEmpty())
				Expect(result.Errors).To(BeEmpty())
			})
		})
	})

	Describe("BuildScenarioID", func() {
		It("combines source, feature, and name with slugified components", func() {
			id := cukesvhs.BuildScenarioID(cukesvhs.SourceBusiness, "My Feature", "User logs in")
			Expect(id).To(Equal("business/my-feature/user-logs-in"))
		})

		It("handles VHS-only source type", func() {
			id := cukesvhs.BuildScenarioID(cukesvhs.SourceVHSOnly, "Demo Feature", "Custom Flow")
			Expect(id).To(Equal("vhs-only/demo-feature/custom-flow"))
		})

		It("produces distinct IDs for same scenario name in different features", func() {
			idA := cukesvhs.BuildScenarioID(cukesvhs.SourceBusiness, "Feature A", "User logs in")
			idB := cukesvhs.BuildScenarioID(cukesvhs.SourceBusiness, "Feature B", "User logs in")
			Expect(idA).NotTo(Equal(idB))
		})

		It("produces distinct IDs for same scenario name with different sources", func() {
			idBusiness := cukesvhs.BuildScenarioID(cukesvhs.SourceBusiness, "Feature A", "User logs in")
			idVHS := cukesvhs.BuildScenarioID(cukesvhs.SourceVHSOnly, "Feature A", "User logs in")
			Expect(idBusiness).NotTo(Equal(idVHS))
		})

		It("handles special characters in names", func() {
			id := cukesvhs.BuildScenarioID(cukesvhs.SourceBusiness, "Feature (v2)", "User's login & logout")
			Expect(id).To(Equal("business/feature-v2/users-login-logout"))
		})
	})
})
