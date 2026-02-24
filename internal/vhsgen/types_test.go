package vhsgen_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukesvhs/internal/vhsgen"
)

var _ = Describe("Types", func() {
	Describe("SourceType constants", func() {
		It("defines SourceBusiness as 'business'", func() {
			Expect(vhsgen.SourceBusiness).To(Equal(vhsgen.SourceType("business")))
		})

		It("defines SourceVHSOnly as 'vhs-only'", func() {
			Expect(vhsgen.SourceVHSOnly).To(Equal(vhsgen.SourceType("vhs-only")))
		})
	})

	Describe("VHSCommandType constants", func() {
		It("defines Type command", func() {
			Expect(vhsgen.Type).To(Equal(vhsgen.VHSCommandType("Type")))
		})

		It("defines Down command", func() {
			Expect(vhsgen.Down).To(Equal(vhsgen.VHSCommandType("Down")))
		})

		It("defines Up command", func() {
			Expect(vhsgen.Up).To(Equal(vhsgen.VHSCommandType("Up")))
		})

		It("defines Enter command", func() {
			Expect(vhsgen.Enter).To(Equal(vhsgen.VHSCommandType("Enter")))
		})

		It("defines Escape command", func() {
			Expect(vhsgen.Escape).To(Equal(vhsgen.VHSCommandType("Escape")))
		})

		It("defines Tab command", func() {
			Expect(vhsgen.Tab).To(Equal(vhsgen.VHSCommandType("Tab")))
		})

		It("defines Sleep command", func() {
			Expect(vhsgen.Sleep).To(Equal(vhsgen.VHSCommandType("Sleep")))
		})

		It("defines Hide command", func() {
			Expect(vhsgen.Hide).To(Equal(vhsgen.VHSCommandType("Hide")))
		})

		It("defines Show command", func() {
			Expect(vhsgen.Show).To(Equal(vhsgen.VHSCommandType("Show")))
		})

		It("defines Screenshot command", func() {
			Expect(vhsgen.Screenshot).To(Equal(vhsgen.VHSCommandType("Screenshot")))
		})

		It("defines Source command", func() {
			Expect(vhsgen.Source).To(Equal(vhsgen.VHSCommandType("Source")))
		})

		It("defines Output command", func() {
			Expect(vhsgen.Output).To(Equal(vhsgen.VHSCommandType("Output")))
		})

		It("defines CtrlC command", func() {
			Expect(vhsgen.CtrlC).To(Equal(vhsgen.VHSCommandType("Ctrl+C")))
		})
	})

	Describe("VHSCommand construction", func() {
		Context("when constructed with Type and text args", func() {
			It("stores the correct type and args", func() {
				cmd := vhsgen.VHSCommand{Type: vhsgen.Type, Args: []string{"hello"}}
				Expect(cmd.Type).To(Equal(vhsgen.Type))
				Expect(cmd.Args).To(Equal([]string{"hello"}))
			})
		})

		Context("when constructed with Sleep and duration", func() {
			It("stores the correct type and args", func() {
				cmd := vhsgen.VHSCommand{Type: vhsgen.Sleep, Args: []string{"500ms"}}
				Expect(cmd.Type).To(Equal(vhsgen.Sleep))
				Expect(cmd.Args).To(Equal([]string{"500ms"}))
			})
		})

		Context("when constructed with Screenshot and path", func() {
			It("stores the correct type and args", func() {
				cmd := vhsgen.VHSCommand{Type: vhsgen.Screenshot, Args: []string{"output.png"}}
				Expect(cmd.Type).To(Equal(vhsgen.Screenshot))
				Expect(cmd.Args).To(Equal([]string{"output.png"}))
			})
		})

		Context("when constructed with no args", func() {
			It("has nil args", func() {
				cmd := vhsgen.VHSCommand{Type: vhsgen.Enter}
				Expect(cmd.Type).To(Equal(vhsgen.Enter))
				Expect(cmd.Args).To(BeNil())
			})
		})
	})

	Describe("StepIR construction", func() {
		Context("Given step", func() {
			It("stores text and step type", func() {
				step := vhsgen.StepIR{
					Text:         "a user is logged in",
					StepType:     "Given",
					Translatable: true,
				}
				Expect(step.Text).To(Equal("a user is logged in"))
				Expect(step.StepType).To(Equal("Given"))
			})
		})

		Context("When step", func() {
			It("stores text and step type", func() {
				step := vhsgen.StepIR{
					Text:         "the user clicks the button",
					StepType:     "When",
					Translatable: true,
				}
				Expect(step.Text).To(Equal("the user clicks the button"))
				Expect(step.StepType).To(Equal("When"))
			})
		})

		Context("Then step", func() {
			It("stores text and step type", func() {
				step := vhsgen.StepIR{
					Text:         "the page should display success",
					StepType:     "Then",
					Translatable: true,
				}
				Expect(step.Text).To(Equal("the page should display success"))
				Expect(step.StepType).To(Equal("Then"))
			})
		})

		Context("untranslatable step", func() {
			It("stores the untranslatable reason", func() {
				step := vhsgen.StepIR{
					Text:                 "some complex step",
					StepType:             "When",
					Translatable:         false,
					UntranslatableReason: "no matching pattern",
				}
				Expect(step.Text).To(Equal("some complex step"))
				Expect(step.StepType).To(Equal("When"))
			})
		})
	})

	Describe("StepIR with commands", func() {
		It("stores multiple commands with correct types", func() {
			step := vhsgen.StepIR{
				Text:     "the user types their name",
				StepType: "When",
				Commands: []vhsgen.VHSCommand{
					{Type: vhsgen.Type, Args: []string{"John Doe"}},
					{Type: vhsgen.Tab},
				},
				Translatable: true,
			}
			Expect(step.Commands).To(HaveLen(2))
			Expect(step.Commands[0].Type).To(Equal(vhsgen.Type))
			Expect(step.Commands[1].Type).To(Equal(vhsgen.Tab))
		})
	})

	Describe("ScenarioIR construction", func() {
		var scenario vhsgen.ScenarioIR

		BeforeEach(func() {
			scenario = vhsgen.ScenarioIR{
				Name:    "User login",
				Feature: "Authentication",
				Tags:    []string{"@critical", "@smoke"},
				Source:  vhsgen.SourceBusiness,
				SetupSteps: []vhsgen.StepIR{
					{Text: "setup step", StepType: "Given", Translatable: true},
				},
				DemoSteps: []vhsgen.StepIR{
					{Text: "demo step", StepType: "When", Translatable: true},
				},
				Translatable: true,
			}
		})

		It("stores the name", func() {
			Expect(scenario.Name).To(Equal("User login"))
		})

		It("stores the feature", func() {
			Expect(scenario.Feature).To(Equal("Authentication"))
		})

		It("stores the source", func() {
			Expect(scenario.Source).To(Equal(vhsgen.SourceBusiness))
		})

		It("stores 2 tags", func() {
			Expect(scenario.Tags).To(HaveLen(2))
		})

		It("stores 1 setup step", func() {
			Expect(scenario.SetupSteps).To(HaveLen(1))
		})

		It("stores 1 demo step", func() {
			Expect(scenario.DemoSteps).To(HaveLen(1))
		})
	})

	Describe("ScenarioIR source types", func() {
		It("accepts SourceBusiness", func() {
			scenario := vhsgen.ScenarioIR{Name: "Test scenario", Source: vhsgen.SourceBusiness}
			Expect(scenario.Source).To(Equal(vhsgen.SourceBusiness))
		})

		It("accepts SourceVHSOnly", func() {
			scenario := vhsgen.ScenarioIR{Name: "Test scenario", Source: vhsgen.SourceVHSOnly}
			Expect(scenario.Source).To(Equal(vhsgen.SourceVHSOnly))
		})
	})

	Describe("GeneratorConfig construction", func() {
		var config vhsgen.GeneratorConfig

		BeforeEach(func() {
			config = vhsgen.GeneratorConfig{
				OutputDir:        "/tmp/output",
				TemplatePath:     "/path/to/template.tape",
				ConfigSourcePath: "demos/vhs/config.tape",
				SleepDuration:    "500ms",
				ScenariosDir:     "features/",
			}
		})

		It("stores OutputDir", func() {
			Expect(config.OutputDir).To(Equal("/tmp/output"))
		})

		It("stores TemplatePath", func() {
			Expect(config.TemplatePath).To(Equal("/path/to/template.tape"))
		})

		It("stores ConfigSourcePath", func() {
			Expect(config.ConfigSourcePath).To(Equal("demos/vhs/config.tape"))
		})

		It("stores SleepDuration", func() {
			Expect(config.SleepDuration).To(Equal("500ms"))
		})

		It("stores ScenariosDir", func() {
			Expect(config.ScenariosDir).To(Equal("features/"))
		})
	})

	Describe("AnalysisResult construction", func() {
		var result vhsgen.AnalysisResult

		BeforeEach(func() {
			result = vhsgen.AnalysisResult{
				ScenarioName: "Login flow",
				Feature:      "Authentication",
				Translatable: true,
				Source:       vhsgen.SourceBusiness,
				Warnings:     []string{"slow step detected"},
				Errors:       []string{},
			}
		})

		It("stores ScenarioName", func() {
			Expect(result.ScenarioName).To(Equal("Login flow"))
		})

		It("stores Feature", func() {
			Expect(result.Feature).To(Equal("Authentication"))
		})

		It("is translatable", func() {
			Expect(result.Translatable).To(BeTrue())
		})

		It("stores Source", func() {
			Expect(result.Source).To(Equal(vhsgen.SourceBusiness))
		})

		It("has 1 warning", func() {
			Expect(result.Warnings).To(HaveLen(1))
		})

		It("has 0 errors", func() {
			Expect(result.Errors).To(BeEmpty())
		})
	})

	Describe("AnalysisResult with untranslatable steps", func() {
		var result vhsgen.AnalysisResult

		BeforeEach(func() {
			untranslatableStep := vhsgen.StepIR{
				Text:                 "complex step",
				StepType:             "When",
				Translatable:         false,
				UntranslatableReason: "no pattern match",
			}
			result = vhsgen.AnalysisResult{
				ScenarioName:        "Complex scenario",
				Feature:             "Advanced",
				Translatable:        false,
				UntranslatableSteps: []vhsgen.StepIR{untranslatableStep},
				Source:              vhsgen.SourceVHSOnly,
				Errors:              []string{"cannot translate scenario"},
			}
		})

		It("is not translatable", func() {
			Expect(result.Translatable).To(BeFalse())
		})

		It("has 1 untranslatable step", func() {
			Expect(result.UntranslatableSteps).To(HaveLen(1))
		})

		It("preserves the step text", func() {
			Expect(result.UntranslatableSteps[0].Text).To(Equal("complex step"))
		})
	})

	Describe("ParamConstraint construction", func() {
		var constraint vhsgen.ParamConstraint

		BeforeEach(func() {
			constraint = vhsgen.ParamConstraint{
				Type:   "enum",
				Values: []string{"value1", "value2", "value3"},
			}
		})

		It("stores the type", func() {
			Expect(constraint.Type).To(Equal("enum"))
		})

		It("stores 3 values", func() {
			Expect(constraint.Values).To(HaveLen(3))
		})

		It("stores the first value correctly", func() {
			Expect(constraint.Values[0]).To(Equal("value1"))
		})
	})

	Describe("StepPattern construction", func() {
		var pattern vhsgen.StepPattern

		BeforeEach(func() {
			pattern = vhsgen.StepPattern{
				Pattern:  `the user types "([^"]+)"`,
				Type:     "When",
				Category: "input",
				Params: map[string]vhsgen.ParamConstraint{
					"text": {Type: "string", Values: nil},
				},
				Example: `the user types "hello"`,
			}
		})

		It("stores Pattern", func() {
			Expect(pattern.Pattern).To(Equal(`the user types "([^"]+)"`))
		})

		It("stores Type", func() {
			Expect(pattern.Type).To(Equal("When"))
		})

		It("stores Category", func() {
			Expect(pattern.Category).To(Equal("input"))
		})

		It("stores 1 param", func() {
			Expect(pattern.Params).To(HaveLen(1))
		})

		It("stores Example", func() {
			Expect(pattern.Example).To(Equal(`the user types "hello"`))
		})
	})

	Describe("Zero values", func() {
		It("VHSCommand zero value has empty type and nil args", func() {
			var cmd vhsgen.VHSCommand
			Expect(string(cmd.Type)).To(BeEmpty())
			Expect(cmd.Args).To(BeNil())
		})

		It("StepIR zero value has empty text and false translatable", func() {
			var step vhsgen.StepIR
			Expect(step.Text).To(BeEmpty())
			Expect(step.Translatable).To(BeFalse())
		})

		It("ScenarioIR zero value has empty name and empty source", func() {
			var scenario vhsgen.ScenarioIR
			Expect(scenario.Name).To(BeEmpty())
			Expect(string(scenario.Source)).To(BeEmpty())
		})

		It("GeneratorConfig zero value has empty OutputDir", func() {
			var config vhsgen.GeneratorConfig
			Expect(config.OutputDir).To(BeEmpty())
		})

		It("AnalysisResult zero value has empty ScenarioName and false translatable", func() {
			var result vhsgen.AnalysisResult
			Expect(result.ScenarioName).To(BeEmpty())
			Expect(result.Translatable).To(BeFalse())
		})

		It("ParamConstraint zero value has empty type", func() {
			var constraint vhsgen.ParamConstraint
			Expect(constraint.Type).To(BeEmpty())
		})

		It("StepPattern zero value has empty pattern and nil params", func() {
			var pattern vhsgen.StepPattern
			Expect(pattern.Pattern).To(BeEmpty())
			Expect(pattern.Params).To(BeNil())
		})
	})

	Describe("Empty slices", func() {
		Context("ScenarioIR with empty slices", func() {
			It("has zero tags, setup steps, and demo steps", func() {
				scenario := vhsgen.ScenarioIR{
					Name:       "Empty scenario",
					Tags:       []string{},
					SetupSteps: []vhsgen.StepIR{},
					DemoSteps:  []vhsgen.StepIR{},
				}
				Expect(scenario.Tags).To(BeEmpty())
				Expect(scenario.SetupSteps).To(BeEmpty())
				Expect(scenario.DemoSteps).To(BeEmpty())
			})
		})

		Context("AnalysisResult with empty slices", func() {
			It("has zero untranslatable steps, warnings, and errors", func() {
				result := vhsgen.AnalysisResult{
					ScenarioName:        "Empty result",
					UntranslatableSteps: []vhsgen.StepIR{},
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
			id := vhsgen.BuildScenarioID(vhsgen.SourceBusiness, "My Feature", "User logs in")
			Expect(id).To(Equal("business/my-feature/user-logs-in"))
		})

		It("handles VHS-only source type", func() {
			id := vhsgen.BuildScenarioID(vhsgen.SourceVHSOnly, "Demo Feature", "Custom Flow")
			Expect(id).To(Equal("vhs-only/demo-feature/custom-flow"))
		})

		It("produces distinct IDs for same scenario name in different features", func() {
			idA := vhsgen.BuildScenarioID(vhsgen.SourceBusiness, "Feature A", "User logs in")
			idB := vhsgen.BuildScenarioID(vhsgen.SourceBusiness, "Feature B", "User logs in")
			Expect(idA).NotTo(Equal(idB))
		})

		It("produces distinct IDs for same scenario name with different sources", func() {
			idBusiness := vhsgen.BuildScenarioID(vhsgen.SourceBusiness, "Feature A", "User logs in")
			idVHS := vhsgen.BuildScenarioID(vhsgen.SourceVHSOnly, "Feature A", "User logs in")
			Expect(idBusiness).NotTo(Equal(idVHS))
		})

		It("handles special characters in names", func() {
			id := vhsgen.BuildScenarioID(vhsgen.SourceBusiness, "Feature (v2)", "User's login & logout")
			Expect(id).To(Equal("business/feature-v2/users-login-logout"))
		})
	})
})
