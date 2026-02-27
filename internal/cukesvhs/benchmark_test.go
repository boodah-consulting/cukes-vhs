package cukesvhs_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

var _ = Describe("CommandTiming", func() {
	It("stores command, duration, and run count", func() {
		timing := cukesvhs.CommandTiming{
			Command:  "echo hello",
			Duration: 50 * time.Millisecond,
			Runs:     3,
		}
		Expect(timing.Command).To(Equal("echo hello"))
		Expect(timing.Duration).To(Equal(50 * time.Millisecond))
		Expect(timing.Runs).To(Equal(3))
	})

	It("has sensible zero values", func() {
		var timing cukesvhs.CommandTiming
		Expect(timing.Command).To(BeEmpty())
		Expect(timing.Duration).To(BeZero())
		Expect(timing.Runs).To(BeZero())
	})
})

var _ = Describe("BenchmarkCommand", func() {
	Context("when running a simple command", func() {
		It("returns a timing result with measured duration", func() {
			timing, err := cukesvhs.BenchmarkCommand("echo hello", 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(timing.Command).To(Equal("echo hello"))
			Expect(timing.Duration).To(BeNumerically(">", 0))
			Expect(timing.Runs).To(Equal(1))
		})
	})

	Context("when running multiple iterations", func() {
		It("averages the duration across all runs", func() {
			timing, err := cukesvhs.BenchmarkCommand("echo hello", 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(timing.Runs).To(Equal(3))
			Expect(timing.Duration).To(BeNumerically(">", 0))
		})
	})

	Context("when the command does not exist", func() {
		It("returns an error", func() {
			_, err := cukesvhs.BenchmarkCommand("nonexistent-binary-xyz", 1)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when runs is zero", func() {
		It("returns an error", func() {
			_, err := cukesvhs.BenchmarkCommand("echo hello", 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("runs must be at least 1"))
		})
	})

	Context("when the command is empty", func() {
		It("returns an error", func() {
			_, err := cukesvhs.BenchmarkCommand("", 1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("command must not be empty"))
		})
	})
})

var _ = Describe("ExtractCommands", func() {
	Context("when steps contain Type commands with CLI invocations", func() {
		It("extracts commands starting with ./", func() {
			steps := []cukesvhs.StepIR{
				{
					Text:     `I type "./cukes-vhs generate --all"`,
					StepType: "When",
					Commands: []cukesvhs.VHSCommand{
						{Type: cukesvhs.Type, Args: []string{"100ms", "./cukes-vhs generate --all"}},
					},
					Translatable: true,
				},
			}
			commands := cukesvhs.ExtractCommands(steps)
			Expect(commands).To(HaveLen(1))
			Expect(commands[0]).To(Equal("./cukes-vhs generate --all"))
		})

		It("extracts multiple commands from multiple steps", func() {
			steps := []cukesvhs.StepIR{
				{
					Text:     `I type "./cukes-vhs list"`,
					StepType: "When",
					Commands: []cukesvhs.VHSCommand{
						{Type: cukesvhs.Type, Args: []string{"100ms", "./cukes-vhs list"}},
					},
					Translatable: true,
				},
				{
					Text:     `I type "./cukes-vhs --help"`,
					StepType: "When",
					Commands: []cukesvhs.VHSCommand{
						{Type: cukesvhs.Type, Args: []string{"100ms", "./cukes-vhs --help"}},
					},
					Translatable: true,
				},
			}
			commands := cukesvhs.ExtractCommands(steps)
			Expect(commands).To(HaveLen(2))
			Expect(commands).To(ContainElement("./cukes-vhs list"))
			Expect(commands).To(ContainElement("./cukes-vhs --help"))
		})
	})

	Context("when steps contain non-CLI Type commands", func() {
		It("excludes plain text input commands", func() {
			steps := []cukesvhs.StepIR{
				{
					Text:     `I enter event description "Built a REST API"`,
					StepType: "When",
					Commands: []cukesvhs.VHSCommand{
						{Type: cukesvhs.Type, Args: []string{"100ms", "Built a REST API"}},
					},
					Translatable: true,
				},
			}
			commands := cukesvhs.ExtractCommands(steps)
			Expect(commands).To(BeEmpty())
		})
	})

	Context("when steps contain non-Type commands", func() {
		It("excludes Enter, Down, and other keystroke commands", func() {
			steps := []cukesvhs.StepIR{
				{
					Text:     "I press enter",
					StepType: "When",
					Commands: []cukesvhs.VHSCommand{
						{Type: cukesvhs.Enter},
					},
					Translatable: true,
				},
				{
					Text:     "I navigate down",
					StepType: "When",
					Commands: []cukesvhs.VHSCommand{
						{Type: cukesvhs.Down},
					},
					Translatable: true,
				},
			}
			commands := cukesvhs.ExtractCommands(steps)
			Expect(commands).To(BeEmpty())
		})
	})

	Context("when steps slice is empty", func() {
		It("returns an empty slice", func() {
			commands := cukesvhs.ExtractCommands([]cukesvhs.StepIR{})
			Expect(commands).To(BeEmpty())
		})
	})

	Context("when steps slice is nil", func() {
		It("returns an empty slice", func() {
			commands := cukesvhs.ExtractCommands(nil)
			Expect(commands).To(BeEmpty())
		})
	})

	Context("when a Type command has no args", func() {
		It("skips the command gracefully", func() {
			steps := []cukesvhs.StepIR{
				{
					Text:     "some step",
					StepType: "When",
					Commands: []cukesvhs.VHSCommand{
						{Type: cukesvhs.Type, Args: []string{}},
					},
					Translatable: true,
				},
			}
			commands := cukesvhs.ExtractCommands(steps)
			Expect(commands).To(BeEmpty())
		})
	})
})

var _ = Describe("BenchmarkScenario", func() {
	Context("when a scenario has benchmarkable When steps", func() {
		It("returns timing results for each command", func() {
			skipIfWindows("command benchmarking requires Unix shell")
			scenario := cukesvhs.ScenarioIR{
				Name:    "Generate all tapes",
				Feature: "Tape Generation",
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:     `I type "/bin/echo benchmark-test"`,
						StepType: "When",
						Commands: []cukesvhs.VHSCommand{
							{Type: cukesvhs.Type, Args: []string{"100ms", "/bin/echo benchmark-test"}},
						},
						Translatable: true,
					},
				},
				Translatable: true,
			}

			results, err := cukesvhs.BenchmarkScenario(scenario, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))
			Expect(results).To(HaveKey("/bin/echo benchmark-test"))
			Expect(results["/bin/echo benchmark-test"].Duration).To(BeNumerically(">", 0))
		})
	})

	Context("when a scenario has no benchmarkable commands", func() {
		It("returns an empty map without error", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Navigate menu",
				Feature: "Navigation",
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:     "I press enter",
						StepType: "When",
						Commands: []cukesvhs.VHSCommand{
							{Type: cukesvhs.Enter},
						},
						Translatable: true,
					},
				},
				Translatable: true,
			}

			results, err := cukesvhs.BenchmarkScenario(scenario, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(BeEmpty())
		})
	})

	Context("when a scenario has setup steps and demo steps", func() {
		It("only extracts commands from demo steps", func() {
			skipIfWindows("command benchmarking requires Unix shell")
			scenario := cukesvhs.ScenarioIR{
				Name:    "Full pipeline",
				Feature: "Pipeline",
				SetupSteps: []cukesvhs.StepIR{
					{
						Text:         "the database is empty",
						StepType:     "Given",
						Translatable: true,
					},
				},
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:     `I type "/bin/echo setup-done"`,
						StepType: "When",
						Commands: []cukesvhs.VHSCommand{
							{Type: cukesvhs.Type, Args: []string{"100ms", "/bin/echo setup-done"}},
						},
						Translatable: true,
					},
				},
				Translatable: true,
			}

			results, err := cukesvhs.BenchmarkScenario(scenario, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))
			Expect(results).To(HaveKey("/bin/echo setup-done"))
		})
	})

	Context("when runs is zero", func() {
		It("returns an error", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Any scenario",
				Feature: "Any Feature",
			}

			_, err := cukesvhs.BenchmarkScenario(scenario, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("runs must be at least 1"))
		})
	})
})

var _ = Describe("StepIR Duration field", func() {
	It("stores a timing duration on a step", func() {
		step := cukesvhs.StepIR{
			Text:     `I type "./cukes-vhs list"`,
			StepType: "When",
			Duration: 150 * time.Millisecond,
		}
		Expect(step.Duration).To(Equal(150 * time.Millisecond))
	})

	It("defaults to zero duration", func() {
		var step cukesvhs.StepIR
		Expect(step.Duration).To(BeZero())
	})
})

var _ = Describe("ApplyBenchmarkTimings", func() {
	Context("when timings match When step commands", func() {
		It("populates Duration on When steps with matching benchmark timings", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Benchmark test",
				Feature: "Benchmark",
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         `I type "./cukes-vhs generate --all"`,
						StepType:     "When",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Type, Args: []string{"100ms", "./cukes-vhs generate --all"}}},
					},
				},
			}
			timings := map[string]cukesvhs.CommandTiming{
				"./cukes-vhs generate --all": {Command: "./cukes-vhs generate --all", Duration: 2 * time.Second, Runs: 3},
			}
			cukesvhs.ApplyBenchmarkTimings(&scenario, timings)
			Expect(scenario.DemoSteps[0].Duration).To(BeNumerically(">", 0))
		})
	})

	Context("when step is a Then step", func() {
		It("does not modify Then steps", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Then test",
				Feature: "Benchmark",
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         "I should see the output",
						StepType:     "Then",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Sleep, Args: []string{"2s"}}},
					},
				},
			}
			timings := map[string]cukesvhs.CommandTiming{
				"./cukes-vhs generate --all": {Command: "./cukes-vhs generate --all", Duration: 2 * time.Second, Runs: 3},
			}
			cukesvhs.ApplyBenchmarkTimings(&scenario, timings)
			Expect(scenario.DemoSteps[0].Duration).To(BeZero())
		})
	})

	Context("when scenario has SetupSteps", func() {
		It("does not modify SetupSteps", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Setup test",
				Feature: "Benchmark",
				SetupSteps: []cukesvhs.StepIR{
					{
						Text:         "the app is running",
						StepType:     "Given",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Type, Args: []string{"100ms", "./cukes-vhs list"}}},
					},
				},
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         "I press enter",
						StepType:     "When",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}},
					},
				},
			}
			timings := map[string]cukesvhs.CommandTiming{
				"./cukes-vhs list": {Command: "./cukes-vhs list", Duration: 1 * time.Second, Runs: 3},
			}
			cukesvhs.ApplyBenchmarkTimings(&scenario, timings)
			Expect(scenario.SetupSteps[0].Duration).To(BeZero())
		})
	})

	Context("when a When step has multiple matching commands", func() {
		It("sums durations for multi-command When steps", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Multi-command test",
				Feature: "Benchmark",
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         "I run two commands",
						StepType:     "When",
						Translatable: true,
						Commands: []cukesvhs.VHSCommand{
							{Type: cukesvhs.Type, Args: []string{"100ms", "./cukes-vhs list"}},
							{Type: cukesvhs.Enter},
							{Type: cukesvhs.Type, Args: []string{"100ms", "./cukes-vhs generate --all"}},
						},
					},
				},
			}
			timings := map[string]cukesvhs.CommandTiming{
				"./cukes-vhs list":           {Command: "./cukes-vhs list", Duration: 1 * time.Second, Runs: 3},
				"./cukes-vhs generate --all": {Command: "./cukes-vhs generate --all", Duration: 2 * time.Second, Runs: 3},
			}
			cukesvhs.ApplyBenchmarkTimings(&scenario, timings)
			Expect(scenario.DemoSteps[0].Duration).To(BeNumerically(">", 3*time.Second))
		})
	})

	Context("when no timings match step commands", func() {
		It("leaves Duration zero for steps with no matching timings", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "No match test",
				Feature: "Benchmark",
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         "I press enter",
						StepType:     "When",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}},
					},
				},
			}
			timings := map[string]cukesvhs.CommandTiming{
				"./cukes-vhs list": {Command: "./cukes-vhs list", Duration: 1 * time.Second, Runs: 3},
			}
			cukesvhs.ApplyBenchmarkTimings(&scenario, timings)
			Expect(scenario.DemoSteps[0].Duration).To(BeZero())
		})
	})

	Context("buffer calculation", func() {
		It("applies buffer to benchmark duration", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:    "Buffer test",
				Feature: "Benchmark",
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         `I type "./cukes-vhs list"`,
						StepType:     "When",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Type, Args: []string{"100ms", "./cukes-vhs list"}}},
					},
				},
			}
			timings := map[string]cukesvhs.CommandTiming{
				"./cukes-vhs list": {Command: "./cukes-vhs list", Duration: 1 * time.Second, Runs: 3},
			}
			cukesvhs.ApplyBenchmarkTimings(&scenario, timings)
			expected := time.Duration(float64(1*time.Second)*1.2) + 500*time.Millisecond
			Expect(scenario.DemoSteps[0].Duration).To(Equal(expected))
		})
	})
})
