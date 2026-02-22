package vhsgen_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukes-vhs/internal/vhsgen"
)

var _ = Describe("GenerateTape", func() {
	Describe("expected directives in output", func() {
		var result string

		BeforeEach(func() {
			scenario := vhsgen.ScenarioIR{
				Name:    "Happy path registration",
				Feature: "User Registration",
				SetupSteps: []vhsgen.StepIR{
					{
						Text:         "the database is empty",
						StepType:     "Given",
						Translatable: true,
						Commands:     nil,
					},
				},
				DemoSteps: []vhsgen.StepIR{
					{
						Text:         `I select "capture_event" from the menu`,
						StepType:     "When",
						Translatable: true,
						Commands:     []vhsgen.VHSCommand{{Type: vhsgen.Enter}},
					},
					{
						Text:         `I enter event description "Built API"`,
						StepType:     "When",
						Translatable: true,
						Commands:     []vhsgen.VHSCommand{{Type: vhsgen.Type, Args: []string{"100ms", "Built API"}}},
					},
				},
				Translatable: true,
			}

			config := vhsgen.GeneratorConfig{
				OutputDir:        "demos/vhs/generated",
				ConfigSourcePath: "config/config.tape",
				SleepDuration:    "2s",
			}

			var err error
			result, err = vhsgen.GenerateTape(scenario, config)
			Expect(err).NotTo(HaveOccurred())
		})

		It("contains the Source directive", func() {
			Expect(result).To(ContainSubstring("Source /tmp/vhsgen-config.tape"))
		})

		It("contains the GIF output directive", func() {
			Expect(result).To(ContainSubstring("Output demos/vhs/generated/user-registration/happy-path-registration.gif"))
		})

		It("contains the ASCII output directive", func() {
			Expect(result).To(ContainSubstring("Output demos/vhs/generated/user-registration/happy-path-registration.ascii"))
		})

		It("contains a Hide block", func() {
			Expect(result).To(ContainSubstring("Hide"))
		})

		It("contains a Show block", func() {
			Expect(result).To(ContainSubstring("Show"))
		})

		It("ends with Ctrl+C", func() {
			Expect(result).To(ContainSubstring("Ctrl+C"))
		})

		It("contains the Enter command", func() {
			Expect(result).To(ContainSubstring("Enter"))
		})

		It("contains the Type command with speed", func() {
			Expect(result).To(ContainSubstring(`Type@100ms "Built API"`))
		})

		It("contains a Sleep between steps", func() {
			Expect(result).To(ContainSubstring("Sleep 2s"))
		})

		It("contains the feature comment", func() {
			Expect(result).To(ContainSubstring("# Feature: User Registration"))
		})

		It("contains the scenario comment", func() {
			Expect(result).To(ContainSubstring("# Scenario: Happy path registration"))
		})
	})

	Describe("dual Output directives", func() {
		It("contains exactly 1 GIF and 1 ASCII Output directive", func() {
			scenario := vhsgen.ScenarioIR{
				Name:    "Test scenario",
				Feature: "Test Feature",
				DemoSteps: []vhsgen.StepIR{
					{Text: "do something", StepType: "When", Translatable: true, Commands: []vhsgen.VHSCommand{{Type: vhsgen.Enter}}},
				},
			}

			result, err := vhsgen.GenerateTape(scenario, vhsgen.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())

			Expect(strings.Count(result, "Output out/test-feature/test-scenario.gif")).To(Equal(1))
			Expect(strings.Count(result, "Output out/test-feature/test-scenario.ascii")).To(Equal(1))
		})
	})

	Describe("untranslatable step", func() {
		It("produces a TODO comment for the step", func() {
			scenario := vhsgen.ScenarioIR{
				Name:    "Form submission",
				Feature: "Event Capture",
				DemoSteps: []vhsgen.StepIR{
					{
						Text:                 "I submit the event",
						StepType:             "When",
						Translatable:         false,
						UntranslatableReason: "form-bypass: use keyboard navigation instead",
					},
				},
			}

			result, err := vhsgen.GenerateTape(scenario, vhsgen.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())

			expected := "# [Manual step needed] — I submit the event (form-bypass: use keyboard navigation instead)"
			Expect(result).To(ContainSubstring(expected))
		})
	})

	Describe("no cleanup commands", func() {
		It("does not include rm -rf, DELETE, or DROP in the output", func() {
			scenario := vhsgen.ScenarioIR{
				Name:    "Test scenario",
				Feature: "Test Feature",
				DemoSteps: []vhsgen.StepIR{
					{Text: "do something", StepType: "When", Translatable: true, Commands: []vhsgen.VHSCommand{{Type: vhsgen.Enter}}},
				},
			}

			result, err := vhsgen.GenerateTape(scenario, vhsgen.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).NotTo(ContainSubstring("rm -rf"))
			Expect(result).NotTo(ContainSubstring("DELETE"))
			Expect(result).NotTo(ContainSubstring("DROP"))
		})
	})

	Describe("forbidden patterns in steps", func() {
		forbiddenCases := []struct {
			name    string
			command string
		}{
			{"rm -rf in type arg", "rm -rf /tmp/data"},
		}

		for _, tc := range forbiddenCases {
			Context("when a step contains "+tc.name, func() {
				It("returns an error mentioning forbidden pattern", func() {
					scenario := vhsgen.ScenarioIR{
						Name:    "Dangerous",
						Feature: "Danger",
						DemoSteps: []vhsgen.StepIR{
							{
								Text:         "clean up",
								StepType:     "When",
								Translatable: true,
								Commands:     []vhsgen.VHSCommand{{Type: vhsgen.Type, Args: []string{tc.command}}},
							},
						},
					}

					_, err := vhsgen.GenerateTape(scenario, vhsgen.GeneratorConfig{OutputDir: "out"})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("forbidden pattern"))
				})
			})
		}

		It("rejects 'rm -rf' as a security guard", func() {
			scenario := vhsgen.ScenarioIR{
				Name:    "Dangerous",
				Feature: "Security",
				DemoSteps: []vhsgen.StepIR{
					{
						Text:         "delete everything",
						StepType:     "When",
						Translatable: true,
						Commands:     []vhsgen.VHSCommand{{Type: vhsgen.Type, Args: []string{"rm -rf /tmp/data"}}},
					},
				},
			}

			_, err := vhsgen.GenerateTape(scenario, vhsgen.GeneratorConfig{OutputDir: "out"})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("default config values", func() {
		It("uses default ConfigSourcePath and SleepDuration when not set", func() {
			scenario := vhsgen.ScenarioIR{
				Name:    "Defaults",
				Feature: "Config",
				DemoSteps: []vhsgen.StepIR{
					{Text: "step one", StepType: "When", Translatable: true, Commands: []vhsgen.VHSCommand{{Type: vhsgen.Down}}},
					{Text: "step two", StepType: "When", Translatable: true, Commands: []vhsgen.VHSCommand{{Type: vhsgen.Enter}}},
				},
			}

			result, err := vhsgen.GenerateTape(scenario, vhsgen.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(ContainSubstring("Source /tmp/vhsgen-config.tape"))
			Expect(result).To(ContainSubstring("Sleep 2s"))
		})
	})

	Describe("Sleep between steps", func() {
		It("inserts Sleep exactly once between 2 demo steps", func() {
			scenario := vhsgen.ScenarioIR{
				Name:    "Sleep test",
				Feature: "Sleep",
				DemoSteps: []vhsgen.StepIR{
					{Text: "step one", StepType: "When", Translatable: true, Commands: []vhsgen.VHSCommand{{Type: vhsgen.Down}, {Type: vhsgen.Down}, {Type: vhsgen.Enter}}},
					{Text: "step two", StepType: "When", Translatable: true, Commands: []vhsgen.VHSCommand{{Type: vhsgen.Enter}}},
				},
			}

			result, err := vhsgen.GenerateTape(scenario, vhsgen.GeneratorConfig{OutputDir: "out", SleepDuration: "3s"})
			Expect(err).NotTo(HaveOccurred())

			sleepCount := strings.Count(result, "Sleep 3s")
			Expect(sleepCount).To(Equal(1))
		})
	})

	Describe("setup steps with no commands", func() {
		It("produces an empty Hide/Show block", func() {
			scenario := vhsgen.ScenarioIR{
				Name:    "Setup only",
				Feature: "Setup",
				SetupSteps: []vhsgen.StepIR{
					{Text: "the database is empty", StepType: "Given", Translatable: true, Commands: nil},
					{Text: "I am on the main menu", StepType: "Given", Translatable: true, Commands: nil},
				},
				DemoSteps: []vhsgen.StepIR{
					{Text: "do something", StepType: "When", Translatable: true, Commands: []vhsgen.VHSCommand{{Type: vhsgen.Enter}}},
				},
			}

			result, err := vhsgen.GenerateTape(scenario, vhsgen.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())

			hideIdx := strings.Index(result, "Hide")
			showIdx := strings.Index(result, "Show")
			Expect(hideIdx).To(BeNumerically(">=", 0))
			Expect(showIdx).To(BeNumerically(">", hideIdx))

			setupBlock := result[hideIdx+4 : showIdx]
			Expect(strings.TrimSpace(setupBlock)).To(BeEmpty())
		})
	})

	Describe("command rendering variants", func() {
		var result string

		BeforeEach(func() {
			scenario := vhsgen.ScenarioIR{
				Name:    "Command variants",
				Feature: "Commands",
				DemoSteps: []vhsgen.StepIR{
					{
						Text:         "various commands",
						StepType:     "When",
						Translatable: true,
						Commands: []vhsgen.VHSCommand{
							{Type: vhsgen.Down},
							{Type: vhsgen.Up},
							{Type: vhsgen.Enter},
							{Type: vhsgen.Escape},
							{Type: vhsgen.Tab},
							{Type: vhsgen.Type, Args: []string{"a"}},
							{Type: vhsgen.Type, Args: []string{"100ms", "hello world"}},
							{Type: vhsgen.CtrlC},
							{Type: vhsgen.CtrlE},
							{Type: vhsgen.CtrlS},
						},
					},
				},
			}

			var err error
			result, err = vhsgen.GenerateTape(scenario, vhsgen.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())
		})

		It("renders Down key", func() { Expect(result).To(ContainSubstring("\nDown\n")) })
		It("renders Up key", func() { Expect(result).To(ContainSubstring("\nUp\n")) })
		It("renders Enter key", func() { Expect(result).To(ContainSubstring("\nEnter\n")) })
		It("renders Escape key", func() { Expect(result).To(ContainSubstring("\nEscape\n")) })
		It("renders Tab key", func() { Expect(result).To(ContainSubstring("\nTab\n")) })
		It("renders Type char", func() { Expect(result).To(ContainSubstring(`Type "a"`)) })
		It("renders Type with speed", func() { Expect(result).To(ContainSubstring(`Type@100ms "hello world"`)) })
		It("renders Ctrl+E", func() { Expect(result).To(ContainSubstring("Ctrl+E")) })
		It("renders Ctrl+S", func() { Expect(result).To(ContainSubstring("Ctrl+S")) })
	})
})

var _ = Describe("slugify", func() {
	slugifyCases := []struct {
		name  string
		input string
		want  string
	}{
		{"simple spaces", "Hello World", "hello-world"},
		{"already slug", "hello-world", "hello-world"},
		{"special chars", "Hello, World! 123", "hello-world-123"},
		{"multiple spaces", "hello   world", "hello-world"},
		{"leading trailing spaces", " hello world ", "hello-world"},
		{"empty string", "", ""},
		{"underscores to hyphens", "capture_event", "capture-event"},
		{"mixed separators", "hello_world test", "hello-world-test"},
		{"consecutive special chars", "a!!b##c", "abc"},
		{"only special chars", "!@#$%", ""},
		{"numbers", "version 2", "version-2"},
	}

	for _, tc := range slugifyCases {
		Context(tc.name, func() {
			It("produces the expected slug", func() {
				Expect(vhsgen.Slugify(tc.input)).To(Equal(tc.want))
			})
		})
	}
})

var _ = Describe("renderCommand", func() {
	Describe("commands with no args", func() {
		noArgCases := []struct {
			name     string
			cmd      vhsgen.VHSCommand
			expected string
		}{
			{"Down key", vhsgen.VHSCommand{Type: vhsgen.Down}, "Down"},
			{"Up key", vhsgen.VHSCommand{Type: vhsgen.Up}, "Up"},
			{"Enter key", vhsgen.VHSCommand{Type: vhsgen.Enter}, "Enter"},
			{"Escape key", vhsgen.VHSCommand{Type: vhsgen.Escape}, "Escape"},
			{"Tab key", vhsgen.VHSCommand{Type: vhsgen.Tab}, "Tab"},
			{"CtrlC", vhsgen.VHSCommand{Type: vhsgen.CtrlC}, "Ctrl+C"},
			{"CtrlE", vhsgen.VHSCommand{Type: vhsgen.CtrlE}, "Ctrl+E"},
			{"CtrlS", vhsgen.VHSCommand{Type: vhsgen.CtrlS}, "Ctrl+S"},
			{"Sleep no args", vhsgen.VHSCommand{Type: vhsgen.Sleep}, "Sleep 1s"},
			{"Type no args", vhsgen.VHSCommand{Type: vhsgen.Type}, "Type"},
		}

		for _, tc := range noArgCases {
			Context(tc.name, func() {
				It("renders the expected output", func() {
					Expect(vhsgen.RenderCommand(tc.cmd)).To(Equal(tc.expected))
				})
			})
		}
	})

	Describe("commands with args", func() {
		withArgCases := []struct {
			name     string
			cmd      vhsgen.VHSCommand
			expected string
		}{
			{"Type with text", vhsgen.VHSCommand{Type: vhsgen.Type, Args: []string{"hello"}}, `Type "hello"`},
			{"Type with speed and text", vhsgen.VHSCommand{Type: vhsgen.Type, Args: []string{"100ms", "world"}}, `Type@100ms "world"`},
			{"Sleep with duration", vhsgen.VHSCommand{Type: vhsgen.Sleep, Args: []string{"3s"}}, "Sleep 3s"},
			{"Generic command with arg", vhsgen.VHSCommand{Type: "Custom", Args: []string{"arg1"}}, "Custom arg1"},
		}

		for _, tc := range withArgCases {
			Context(tc.name, func() {
				It("renders the expected output", func() {
					Expect(vhsgen.RenderCommand(tc.cmd)).To(Equal(tc.expected))
				})
			})
		}
	})
})

var _ = Describe("WriteTape", func() {
	Describe("creating the tape file", func() {
		It("creates the file at the expected path with correct content", func() {
			tmpDir := GinkgoT().TempDir()

			scenario := vhsgen.ScenarioIR{
				Name:    "Write test",
				Feature: "File Output",
				DemoSteps: []vhsgen.StepIR{
					{Text: "action", StepType: "When", Translatable: true, Commands: []vhsgen.VHSCommand{{Type: vhsgen.Enter}}},
				},
			}

			err := vhsgen.WriteTape(scenario, vhsgen.GeneratorConfig{OutputDir: tmpDir, ConfigSourcePath: "config/config.tape"})
			Expect(err).NotTo(HaveOccurred())

			expectedPath := filepath.Join(tmpDir, "file-output", "write-test.tape")
			data, err := os.ReadFile(expectedPath)
			Expect(err).NotTo(HaveOccurred())

			content := string(data)
			Expect(content).To(ContainSubstring("Source /tmp/vhsgen-config.tape"))
			Expect(content).To(ContainSubstring("Ctrl+C"))
		})
	})

	Describe("creating parent directories", func() {
		It("creates all required nested directories", func() {
			tmpDir := GinkgoT().TempDir()
			nestedDir := filepath.Join(tmpDir, "deep", "nested")

			scenario := vhsgen.ScenarioIR{
				Name:    "Nested",
				Feature: "Deep",
				DemoSteps: []vhsgen.StepIR{
					{Text: "action", StepType: "When", Translatable: true, Commands: []vhsgen.VHSCommand{{Type: vhsgen.Enter}}},
				},
			}

			err := vhsgen.WriteTape(scenario, vhsgen.GeneratorConfig{OutputDir: nestedDir})
			Expect(err).NotTo(HaveOccurred())

			expectedPath := filepath.Join(nestedDir, "deep", "nested.tape")
			_, statErr := os.Stat(expectedPath)
			Expect(statErr).NotTo(HaveOccurred())
		})
	})

	Describe("invalid output directory", func() {
		It("returns an error when the directory cannot be created", func() {
			scenario := vhsgen.ScenarioIR{
				Name:    "Test",
				Feature: "Test",
				DemoSteps: []vhsgen.StepIR{
					{Text: "action", StepType: "When", Translatable: true, Commands: []vhsgen.VHSCommand{{Type: vhsgen.Enter}}},
				},
			}

			err := vhsgen.WriteTape(scenario, vhsgen.GeneratorConfig{OutputDir: "/invalid/nonexistent/path/that/cannot/be/created"})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GenerateTape error propagation", func() {
		It("returns error when scenario contains a forbidden pattern", func() {
			scenario := vhsgen.ScenarioIR{
				Name:    "Forbidden",
				Feature: "Dangerous",
				DemoSteps: []vhsgen.StepIR{
					{
						Text:         "clean up",
						StepType:     "When",
						Translatable: true,
						Commands:     []vhsgen.VHSCommand{{Type: vhsgen.Type, Args: []string{"rm -rf /data"}}},
					},
				},
			}

			err := vhsgen.WriteTape(scenario, vhsgen.GeneratorConfig{OutputDir: GinkgoT().TempDir()})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("forbidden pattern"))
		})
	})

	Describe("WriteFile error", func() {
		It("returns error when output file cannot be written to read-only dir", func() {
			tmpDir := GinkgoT().TempDir()
			featureDir := filepath.Join(tmpDir, "write-fail")
			err := os.MkdirAll(featureDir, 0o755)
			Expect(err).NotTo(HaveOccurred())

			scenario := vhsgen.ScenarioIR{
				Name:    "Write Fail",
				Feature: "Write Fail",
				DemoSteps: []vhsgen.StepIR{
					{Text: "action", StepType: "When", Translatable: true, Commands: []vhsgen.VHSCommand{{Type: vhsgen.Enter}}},
				},
			}

			err = os.Chmod(featureDir, 0o000)
			Expect(err).NotTo(HaveOccurred())
			defer os.Chmod(featureDir, 0o755) //nolint:errcheck

			err = vhsgen.WriteTape(scenario, vhsgen.GeneratorConfig{OutputDir: tmpDir})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("writing tape file"))
		})
	})
})
