package cukesvhs_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

var _ = Describe("GenerateTape", func() {
	Describe("expected directives in output", func() {
		var result string

		BeforeEach(func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Happy path registration",
				Feature: "User Registration",
				SetupSteps: []cukesvhs.StepIR{
					{
						Text:         "the database is empty",
						StepType:     "Given",
						Translatable: true,
						Commands:     nil,
					},
				},
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         `I select "capture_event" from the menu`,
						StepType:     "When",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}},
					},
					{
						Text:         `I enter event description "Built API"`,
						StepType:     "When",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Type, Args: []string{"100ms", "Built API"}}},
					},
				},
				Translatable: true,
			}

			config := cukesvhs.GeneratorConfig{
				OutputDir:        "demos/vhs/generated",
				ConfigSourcePath: "config/config.tape",
				SleepDuration:    "2s",
			}

			var err error
			result, err = cukesvhs.GenerateTape(scenario, config)
			Expect(err).NotTo(HaveOccurred())
		})

		It("references an external config source", func() {
			Expect(result).To(ContainSubstring("Source"))
			Expect(result).To(ContainSubstring(".tape"))
		})

		It("specifies GIF output location", func() {
			Expect(result).To(ContainSubstring("Output"))
			Expect(result).To(ContainSubstring(".gif"))
		})

		It("does not generate ASCII output by default", func() {
			Expect(result).NotTo(ContainSubstring(".ascii"))
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

		It("generates typing commands with the specified text", func() {
			Expect(result).To(ContainSubstring("Type"))
			Expect(result).To(ContainSubstring("Built API"))
		})

		It("includes pauses between steps for readability", func() {
			Expect(result).To(ContainSubstring("Sleep"))
		})

		It("contains the feature comment", func() {
			Expect(result).To(ContainSubstring("# Feature: User Registration"))
		})

		It("contains the scenario comment", func() {
			Expect(result).To(ContainSubstring("# Scenario: Happy path registration"))
		})
	})

	Describe("GIF-only Output directive", func() {
		It("generates exactly one GIF output and no ASCII output", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Test scenario",
				Feature: "Test Feature",
				DemoSteps: []cukesvhs.StepIR{
					{Text: "do something", StepType: "When", Translatable: true, Commands: []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}}},
				},
			}

			result, err := cukesvhs.GenerateTape(scenario, cukesvhs.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())

			gifCount := strings.Count(result, ".gif")
			Expect(gifCount).To(Equal(1))
			Expect(result).NotTo(ContainSubstring(".ascii"))
		})
	})

	Describe("untranslatable step", func() {
		It("marks untranslatable steps as requiring manual action", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Form submission",
				Feature: "Event Capture",
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:                 "I submit the event",
						StepType:             "When",
						Translatable:         false,
						UntranslatableReason: "form-bypass: use keyboard navigation instead",
					},
				},
			}

			result, err := cukesvhs.GenerateTape(scenario, cukesvhs.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(ContainSubstring("Manual step needed"))
			Expect(result).To(ContainSubstring("I submit the event"))
			Expect(result).To(ContainSubstring("form-bypass"))
		})
	})

	Describe("no cleanup commands", func() {
		It("does not include rm -rf, DELETE, or DROP in the output", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Test scenario",
				Feature: "Test Feature",
				DemoSteps: []cukesvhs.StepIR{
					{Text: "do something", StepType: "When", Translatable: true, Commands: []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}}},
				},
			}

			result, err := cukesvhs.GenerateTape(scenario, cukesvhs.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).NotTo(ContainSubstring("rm -rf"))
			Expect(result).NotTo(ContainSubstring("DELETE"))
			Expect(result).NotTo(ContainSubstring("DROP"))
		})
	})

	Describe("forbidden patterns in steps", func() {
		DescribeTable("forbidden patterns",
			func(name, command string) {
				scenario := cukesvhs.ScenarioIR{
					Name:    "Dangerous",
					Feature: "Danger",
					DemoSteps: []cukesvhs.StepIR{
						{
							Text:         "clean up",
							StepType:     "When",
							Translatable: true,
							Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Type, Args: []string{command}}},
						},
					},
				}

				_, err := cukesvhs.GenerateTape(scenario, cukesvhs.GeneratorConfig{OutputDir: "out"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("forbidden pattern"))
			},
			Entry("rm -rf in type arg", "rm -rf in type arg", "rm -rf /tmp/data"),
		)

		It("rejects 'rm -rf' as a security guard", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Dangerous",
				Feature: "Security",
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         "delete everything",
						StepType:     "When",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Type, Args: []string{"rm -rf /tmp/data"}}},
					},
				},
			}

			_, err := cukesvhs.GenerateTape(scenario, cukesvhs.GeneratorConfig{OutputDir: "out"})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("default config values", func() {
		It("applies sensible defaults when config is not explicitly set", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Defaults",
				Feature: "Config",
				DemoSteps: []cukesvhs.StepIR{
					{Text: "step one", StepType: "When", Translatable: true, Commands: []cukesvhs.VHSCommand{{Type: cukesvhs.Down}}},
					{Text: "step two", StepType: "When", Translatable: true, Commands: []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}}},
				},
			}

			result, err := cukesvhs.GenerateTape(scenario, cukesvhs.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(ContainSubstring("Source"))
			Expect(result).To(ContainSubstring("Sleep"))
		})
	})

	Describe("Sleep between steps", func() {
		It("inserts configured pause duration between demo steps", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Sleep test",
				Feature: "Sleep",
				DemoSteps: []cukesvhs.StepIR{
					{Text: "step one", StepType: "When", Translatable: true, Commands: []cukesvhs.VHSCommand{{Type: cukesvhs.Down}, {Type: cukesvhs.Down}, {Type: cukesvhs.Enter}}},
					{Text: "step two", StepType: "When", Translatable: true, Commands: []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}}},
				},
			}

			result, err := cukesvhs.GenerateTape(scenario, cukesvhs.GeneratorConfig{OutputDir: "out", SleepDuration: "3s"})
			Expect(err).NotTo(HaveOccurred())

			sleepCount := strings.Count(result, "Sleep 3s")
			Expect(sleepCount).To(Equal(2))
		})
	})

	Describe("setup steps with no commands", func() {
		It("produces an empty Hide/Show block", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Setup only",
				Feature: "Setup",
				SetupSteps: []cukesvhs.StepIR{
					{Text: "the database is empty", StepType: "Given", Translatable: true, Commands: nil},
					{Text: "I am on the main menu", StepType: "Given", Translatable: true, Commands: nil},
				},
				DemoSteps: []cukesvhs.StepIR{
					{Text: "do something", StepType: "When", Translatable: true, Commands: []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}}},
				},
			}

			result, err := cukesvhs.GenerateTape(scenario, cukesvhs.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())

			hideIdx := strings.Index(result, "Hide")
			showIdx := strings.Index(result, "Show")
			Expect(hideIdx).To(BeNumerically(">=", 0))
			Expect(showIdx).To(BeNumerically(">", hideIdx))

			setupBlock := result[hideIdx+4 : showIdx]
			Expect(strings.TrimSpace(setupBlock)).To(BeEmpty())
		})
	})

	Describe("generates VHS directives for all command types", func() {
		var result string

		BeforeEach(func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Command variants",
				Feature: "Commands",
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         "various commands",
						StepType:     "When",
						Translatable: true,
						Commands: []cukesvhs.VHSCommand{
							{Type: cukesvhs.Down},
							{Type: cukesvhs.Up},
							{Type: cukesvhs.Enter},
							{Type: cukesvhs.Escape},
							{Type: cukesvhs.Tab},
							{Type: cukesvhs.Type, Args: []string{"a"}},
							{Type: cukesvhs.Type, Args: []string{"100ms", "hello world"}},
							{Type: cukesvhs.CtrlC},
							{Type: cukesvhs.CtrlE},
							{Type: cukesvhs.CtrlS},
						},
					},
				},
			}

			var err error
			result, err = cukesvhs.GenerateTape(scenario, cukesvhs.GeneratorConfig{OutputDir: "out"})
			Expect(err).NotTo(HaveOccurred())
		})

		It("includes navigation commands", func() {
			Expect(result).To(ContainSubstring("Down"))
			Expect(result).To(ContainSubstring("Up"))
			Expect(result).To(ContainSubstring("Enter"))
			Expect(result).To(ContainSubstring("Escape"))
			Expect(result).To(ContainSubstring("Tab"))
		})

		It("includes text input commands", func() {
			Expect(result).To(ContainSubstring("Type"))
			Expect(result).To(ContainSubstring("a"))
			Expect(result).To(ContainSubstring("hello world"))
		})

		It("includes control key commands", func() {
			Expect(result).To(ContainSubstring("Ctrl+E"))
			Expect(result).To(ContainSubstring("Ctrl+S"))
		})
	})
})

var _ = Describe("Slugify", func() {
	DescribeTable("slugify transformations",
		func(input, want string) {
			Expect(cukesvhs.Slugify(input)).To(Equal(want))
		},
		Entry("simple spaces", "Hello World", "hello-world"),
		Entry("already slug", "hello-world", "hello-world"),
		Entry("special chars", "Hello, World! 123", "hello-world-123"),
		Entry("multiple spaces", "hello   world", "hello-world"),
		Entry("leading trailing spaces", " hello world ", "hello-world"),
		Entry("empty string", "", ""),
		Entry("underscores to hyphens", "capture_event", "capture-event"),
		Entry("mixed separators", "hello_world test", "hello-world-test"),
		Entry("consecutive special chars", "a!!b##c", "abc"),
		Entry("only special chars", "!@#$%", ""),
		Entry("numbers", "version 2", "version-2"),
	)
})

var _ = Describe("renderCommand", func() {
	DescribeTable("commands with no args",
		func(cmd cukesvhs.VHSCommand, expected string) {
			Expect(cukesvhs.RenderCommand(cmd)).To(Equal(expected))
		},
		Entry("Down key", cukesvhs.VHSCommand{Type: cukesvhs.Down}, "Down"),
		Entry("Up key", cukesvhs.VHSCommand{Type: cukesvhs.Up}, "Up"),
		Entry("Enter key", cukesvhs.VHSCommand{Type: cukesvhs.Enter}, "Enter"),
		Entry("Escape key", cukesvhs.VHSCommand{Type: cukesvhs.Escape}, "Escape"),
		Entry("Tab key", cukesvhs.VHSCommand{Type: cukesvhs.Tab}, "Tab"),
		Entry("CtrlC", cukesvhs.VHSCommand{Type: cukesvhs.CtrlC}, "Ctrl+C"),
		Entry("CtrlE", cukesvhs.VHSCommand{Type: cukesvhs.CtrlE}, "Ctrl+E"),
		Entry("CtrlS", cukesvhs.VHSCommand{Type: cukesvhs.CtrlS}, "Ctrl+S"),
		Entry("Sleep no args", cukesvhs.VHSCommand{Type: cukesvhs.Sleep}, "Sleep 1s"),
		Entry("Type no args", cukesvhs.VHSCommand{Type: cukesvhs.Type}, "Type"),
	)

	DescribeTable("commands with args",
		func(cmd cukesvhs.VHSCommand, expected string) {
			Expect(cukesvhs.RenderCommand(cmd)).To(Equal(expected))
		},
		Entry("Type with text", cukesvhs.VHSCommand{Type: cukesvhs.Type, Args: []string{"hello"}}, `Type "hello"`),
		Entry("Type with speed and text", cukesvhs.VHSCommand{Type: cukesvhs.Type, Args: []string{"100ms", "world"}}, `Type@100ms "world"`),
		Entry("Sleep with duration", cukesvhs.VHSCommand{Type: cukesvhs.Sleep, Args: []string{"3s"}}, "Sleep 3s"),
		Entry("Generic command with arg", cukesvhs.VHSCommand{Type: "Custom", Args: []string{"arg1"}}, "Custom arg1"),
	)
})

var _ = Describe("WriteTape", func() {
	Describe("creating the tape file", func() {
		It("creates the file at the expected path with correct content", func() {
			tmpDir := GinkgoT().TempDir()

			scenario := cukesvhs.ScenarioIR{
				Name:    "Write test",
				Feature: "File Output",
				DemoSteps: []cukesvhs.StepIR{
					{Text: "action", StepType: "When", Translatable: true, Commands: []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}}},
				},
			}

			err := cukesvhs.WriteTape(scenario, cukesvhs.GeneratorConfig{OutputDir: tmpDir, ConfigSourcePath: "config/config.tape"})
			Expect(err).NotTo(HaveOccurred())

			expectedPath := filepath.Join(tmpDir, "file-output", "write-test.tape")
			data, err := os.ReadFile(expectedPath)
			Expect(err).NotTo(HaveOccurred())

			content := string(data)
			Expect(content).To(MatchRegexp(`Source .*/cukesvhs-.*\.tape`))
			Expect(content).To(ContainSubstring("Ctrl+C"))
		})
	})

	Describe("creating parent directories", func() {
		It("creates all required nested directories", func() {
			tmpDir := GinkgoT().TempDir()
			nestedDir := filepath.Join(tmpDir, "deep", "nested")

			scenario := cukesvhs.ScenarioIR{
				Name:    "Nested",
				Feature: "Deep",
				DemoSteps: []cukesvhs.StepIR{
					{Text: "action", StepType: "When", Translatable: true, Commands: []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}}},
				},
			}

			err := cukesvhs.WriteTape(scenario, cukesvhs.GeneratorConfig{OutputDir: nestedDir})
			Expect(err).NotTo(HaveOccurred())

			expectedPath := filepath.Join(nestedDir, "deep", "nested.tape")
			_, statErr := os.Stat(expectedPath)
			Expect(statErr).NotTo(HaveOccurred())
		})
	})

	Describe("invalid output directory", func() {
		It("returns an error when the directory cannot be created", func() {
			skipIfWindows("unix path /invalid/... resolves differently on windows")
			scenario := cukesvhs.ScenarioIR{
				Name:    "Test",
				Feature: "Test",
				DemoSteps: []cukesvhs.StepIR{
					{Text: "action", StepType: "When", Translatable: true, Commands: []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}}},
				},
			}

			err := cukesvhs.WriteTape(scenario, cukesvhs.GeneratorConfig{OutputDir: "/invalid/nonexistent/path/that/cannot/be/created"})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GenerateTape error propagation", func() {
		It("returns error when scenario contains a forbidden pattern", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Forbidden",
				Feature: "Dangerous",
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         "clean up",
						StepType:     "When",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Type, Args: []string{"rm -rf /data"}}},
					},
				},
			}

			err := cukesvhs.WriteTape(scenario, cukesvhs.GeneratorConfig{OutputDir: GinkgoT().TempDir()})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("forbidden pattern"))
		})
	})

	Describe("WriteFile error", func() {
		It("returns error when output file cannot be written to read-only dir", func() {
			skipIfWindows("os.Chmod does not restrict access on windows")
			tmpDir := GinkgoT().TempDir()
			featureDir := filepath.Join(tmpDir, "write-fail")
			err := os.MkdirAll(featureDir, 0o750)
			Expect(err).NotTo(HaveOccurred())

			scenario := cukesvhs.ScenarioIR{
				Name:    "Write Fail",
				Feature: "Write Fail",
				DemoSteps: []cukesvhs.StepIR{
					{Text: "action", StepType: "When", Translatable: true, Commands: []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}}},
				},
			}

			err = os.Chmod(featureDir, 0o000)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Chmod(featureDir, 0o700) }()

			err = cukesvhs.WriteTape(scenario, cukesvhs.GeneratorConfig{OutputDir: tmpDir})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("writing tape file"))
		})
	})
})

var _ = Describe("resolveConfigPath", func() {
	Context("when custom path points to an existing file", func() {
		It("returns the custom path and empty warning", func() {
			tmpDir := GinkgoT().TempDir()
			customPath := filepath.Join(tmpDir, "custom-config.tape")
			Expect(os.WriteFile(customPath, []byte("custom content"), 0o600)).To(Succeed())

			result, warning, cleanup, err := cukesvhs.ResolveConfigPath(customPath)
			Expect(err).NotTo(HaveOccurred())
			defer cleanup()
			Expect(result).To(Equal(customPath))
			Expect(warning).To(BeEmpty())
		})
	})
	Context("when custom path does not exist", func() {
		It("falls back to writing embedded config to a temp file", func() {
			result, warning, cleanup, err := cukesvhs.ResolveConfigPath("/nonexistent/path/config.tape")
			defer cleanup()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(Equal("/nonexistent/path/config.tape"))
			Expect(warning).NotTo(BeEmpty())
			Expect(warning).To(ContainSubstring("Warning: config file not found"))
		})
		It("returns a valid file path that exists on disc", func() {
			result, _, cleanup, err := cukesvhs.ResolveConfigPath("/nonexistent/path/config.tape")
			defer cleanup()
			Expect(err).NotTo(HaveOccurred())

			_, statErr := os.Stat(result)
			Expect(statErr).NotTo(HaveOccurred())
		})
		It("writes the embedded config content to the fallback path", func() {
			result, _, cleanup, err := cukesvhs.ResolveConfigPath("/nonexistent/path/config.tape")
			defer cleanup()
			Expect(err).NotTo(HaveOccurred())
			data, readErr := os.ReadFile(result)
			Expect(readErr).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("Set Shell"))
		})
	})
	Context("when custom path is empty", func() {
		It("falls back to embedded config", func() {
			result, warning, cleanup, err := cukesvhs.ResolveConfigPath("")
			defer cleanup()
			Expect(err).NotTo(HaveOccurred())

			_, statErr := os.Stat(result)
			Expect(statErr).NotTo(HaveOccurred())
			Expect(warning).To(BeEmpty())
		})

		It("returns a path containing the expected filename", func() {
			result, _, cleanup, err := cukesvhs.ResolveConfigPath("")
			defer cleanup()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring("cukesvhs-"))
			Expect(result).To(HaveSuffix(".tape"))
		})
	})
})

var _ = Describe("DefaultConfig", func() {
	It("returns a non-empty string", func() {
		Expect(cukesvhs.DefaultConfig()).NotTo(BeEmpty())
	})

	It("contains the Set Shell directive", func() {
		Expect(cukesvhs.DefaultConfig()).To(ContainSubstring("Set Shell"))
	})

	It("contains the Set FontSize directive", func() {
		Expect(cukesvhs.DefaultConfig()).To(ContainSubstring("Set FontSize"))
	})

	It("contains the Set Theme directive", func() {
		Expect(cukesvhs.DefaultConfig()).To(ContainSubstring("Set Theme"))
	})

	It("contains the Set Width directive", func() {
		Expect(cukesvhs.DefaultConfig()).To(ContainSubstring("Set Width"))
	})

	It("contains the Set Height directive", func() {
		Expect(cukesvhs.DefaultConfig()).To(ContainSubstring("Set Height"))
	})
})
