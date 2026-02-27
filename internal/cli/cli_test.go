package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

func skipIfNoVHS() {
	if _, err := exec.LookPath("vhs"); err != nil {
		Skip("vhs binary not available")
	}
}

func skipIfWindows(reason string) {
	if runtime.GOOS == "windows" {
		Skip(reason)
	}
}

var _ = Describe("cukesvhs CLI", func() {
	Context("no subcommand", func() {
		It("prints usage and returns 0", func() {
			var out, errOut bytes.Buffer
			code := Run([]string{}, &out, &errOut)

			Expect(code).To(Equal(0))
			Expect(out.String()).To(ContainSubstring("cukes-vhs"))
			Expect(out.String()).To(ContainSubstring("list"))
			Expect(out.String()).To(ContainSubstring("generate"))
		})
	})

	Context("--help flag", func() {
		It("prints usage and returns 0", func() {
			var out, errOut bytes.Buffer
			code := Run([]string{"--help"}, &out, &errOut)

			Expect(code).To(Equal(0))
			Expect(out.String()).To(ContainSubstring("cukes-vhs"))
		})
	})

	Context("unknown subcommand", func() {
		It("returns exit code 1 with error message", func() {
			var out, errOut bytes.Buffer
			code := Run([]string{"foobar"}, &out, &errOut)

			Expect(code).To(Equal(1))
			Expect(errOut.String()).To(ContainSubstring("unknown subcommand"))
		})
	})

	Describe("list subcommand", func() {
		Context("with a valid features directory", func() {
			It("prints a table with Scenario, Feature, Source, Translatable, Reason columns", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"list", "--features", "testdata/features/", "--scenarios-dir", "testdata/scenarios/"}, &out, &errOut)

				Expect(code).To(Equal(0))
				output := out.String()
				Expect(output).To(ContainSubstring("Parsing..."))
				Expect(output).To(ContainSubstring("Scenario"))
				Expect(output).To(ContainSubstring("Feature"))
				Expect(output).To(ContainSubstring("Source"))
				Expect(output).To(ContainSubstring("Translatable"))
			})
		})

		Context("--json flag", func() {
			It("produces valid JSON with a scenarios array containing source field", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"list", "--features", "testdata/features/", "--scenarios-dir", "testdata/scenarios/", "--json"}, &out, &errOut)

				Expect(code).To(Equal(0))

				var payload map[string]interface{}
				outputBytes := out.Bytes()
				jsonStart := bytes.Index(outputBytes, []byte("{"))
				Expect(jsonStart).To(BeNumerically(">=", 0), "no JSON found in output")

				err := json.Unmarshal(outputBytes[jsonStart:], &payload)
				Expect(err).NotTo(HaveOccurred())
				Expect(payload).To(HaveKey("scenarios"))

				scenarios, ok := payload["scenarios"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(scenarios).NotTo(BeEmpty())

				first := scenarios[0].(map[string]interface{})
				Expect(first).To(HaveKey("source"))
				Expect(first).To(HaveKey("scenario_name"))
				Expect(first).To(HaveKey("feature"))
				Expect(first).To(HaveKey("translatable"))
			})
		})

		Context("--count flag", func() {
			It("shows counts broken down by source", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"list", "--features", "testdata/features/", "--scenarios-dir", "testdata/scenarios/", "--count"}, &out, &errOut)

				Expect(code).To(Equal(0))
				Expect(out.String()).To(MatchRegexp(`Business: \d+/\d+ translatable \| VHS-only: \d+/\d+ translatable`))
			})
		})

		Context("--steps flag", func() {
			It("outputs a readable table of translatable step patterns", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"list", "--steps"}, &out, &errOut)

				Expect(code).To(Equal(0))
				output := out.String()
				Expect(output).To(ContainSubstring("Pattern"))
				Expect(output).To(ContainSubstring("Type"))
				Expect(output).To(ContainSubstring("Category"))
				Expect(output).To(ContainSubstring("Example"))
			})

			It("includes navigation patterns", func() {
				var out, errOut bytes.Buffer
				Run([]string{"list", "--steps"}, &out, &errOut)

				Expect(out.String()).To(ContainSubstring("navigation"))
			})

			It("includes input patterns", func() {
				var out, errOut bytes.Buffer
				Run([]string{"list", "--steps"}, &out, &errOut)

				Expect(out.String()).To(ContainSubstring("input"))
			})
		})

		Context("--steps --json flag", func() {
			It("outputs JSON array with pattern, type, category, params, example fields", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"list", "--steps", "--json"}, &out, &errOut)

				Expect(code).To(Equal(0))

				var patterns []map[string]interface{}
				err := json.Unmarshal(out.Bytes(), &patterns)
				Expect(err).NotTo(HaveOccurred())
				Expect(patterns).NotTo(BeEmpty())

				first := patterns[0]
				Expect(first).To(HaveKey("pattern"))
				Expect(first).To(HaveKey("type"))
				Expect(first).To(HaveKey("category"))
				Expect(first).To(HaveKey("example"))
			})
		})

		Context("with non-existent scenarios-dir", func() {
			It("handles missing scenarios dir gracefully (no error)", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"list", "--features", "testdata/features/", "--scenarios-dir", "/nonexistent/path/"}, &out, &errOut)

				Expect(code).To(Equal(0))
				Expect(errOut.String()).To(BeEmpty())
			})
		})

		Context("with non-existent features dir", func() {
			It("returns exit code 1 with error message", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"list", "--features", "/nonexistent/features/", "--scenarios-dir", "/nonexistent/scenarios/"}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("Error parsing features dir"))
			})
		})

		Context("with unknown flag", func() {
			It("returns exit code 1 with error message", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"list", "--unknown-flag-xyz"}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("unknown flag"))
			})
		})
	})

	Describe("generate subcommand", func() {
		var tmpDir string

		BeforeEach(func() {
			tmpDir = GinkgoT().TempDir()
		})

		Context("--output missing", func() {
			It("returns exit code 1 with error message", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"generate", "--all", "--features", "testdata/features/"}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("--output is required"))
			})
		})

		Context("no filter flags", func() {
			It("returns exit code 1 requiring --all, --feature, or --scenario", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"generate", "--output", tmpDir}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("--all"))
			})
		})

		Context("with non-existent features directory", func() {
			It("returns exit code 1 when features dir does not exist", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{
					"generate",
					"--all",
					"--features", "/nonexistent/features/",
					"--scenarios-dir", "/nonexistent/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(code).To(Equal(1))
			})
		})

		Context("--all flag", func() {
			It("generates tape files and shows summary", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{
					"generate",
					"--all",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				output := out.String()
				Expect(output).To(ContainSubstring("Parsing..."))
				Expect(output).To(ContainSubstring("Generating..."))
				Expect(output).NotTo(ContainSubstring("Rendering..."))
				Expect(output).To(MatchRegexp(`Generated \d+ tapes`))
			})

			It("does not invoke VHS rendering", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{
					"generate",
					"--all",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				output := out.String()
				Expect(output).NotTo(ContainSubstring("Rendering..."))
				Expect(errOut.String()).NotTo(ContainSubstring("render"))
			})

			It("outputs only parsing, generating, written, and summary lines", func() {
				var out, errOut bytes.Buffer
				Run([]string{
					"generate",
					"--all",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				lines := strings.Split(strings.TrimSpace(out.String()), "\n")
				for _, line := range lines {
					Expect(line).To(SatisfyAny(
						Equal("Parsing..."),
						Equal("Generating..."),
						HavePrefix("Written:"),
						MatchRegexp(`^Generated \d+ tapes`),
						HavePrefix("Skipping"),
					))
				}
			})

			It("runs without error even when all scenarios are untranslatable", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{
					"generate",
					"--all",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(code).To(Equal(0))
			})

			It("reports 'Written: <path>' for each tape file written", func() {
				var out, errOut bytes.Buffer
				Run([]string{
					"generate",
					"--all",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				output := out.String()
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "Written:") {
						path := strings.TrimSpace(strings.TrimPrefix(line, "Written:"))
						Expect(path).To(HaveSuffix(".tape"))
					}
				}
			})
		})

		Context("bare 'all' positional argument", func() {
			It("treats 'all' as --all and generates tapes", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{
					"generate",
					"all",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				Expect(out.String()).To(MatchRegexp(`Generated \d+ tapes`))
			})
		})

		Context("--feature flag", func() {
			It("filters by feature name (case-insensitive) and shows summary", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{
					"generate",
					"--feature", "Docker Management",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				Expect(out.String()).To(MatchRegexp(`Generated \d+ tapes`))
			})
		})

		Context("--scenario flag", func() {
			It("filters by scenario name", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{
					"generate",
					"--scenario", "List running containers",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				Expect(out.String()).To(MatchRegexp(`Generated \d+ tapes`))
			})
		})

		Context("output summary format", func() {
			It("shows summary with from features, from scenarios, and warnings counts", func() {
				var out, errOut bytes.Buffer
				Run([]string{
					"generate",
					"--all",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(out.String()).To(MatchRegexp(
					`Generated \d+ tapes \(\d+ from features, \d+ from scenarios, \d+ warnings\)`,
				))
			})
		})

		Context("rendering-only flags", func() {
			It("does not accept --binary-path flag", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{
					"generate",
					"--all",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
					"--binary-path", "/usr/bin/vhs",
				}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("unknown flag"))
			})

			It("does not accept --timeout flag", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{
					"generate",
					"--all",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
					"--timeout", "60",
				}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("unknown flag"))
			})
		})
	})

	Describe("slugify helper", func() {
		DescribeTable("converts names to URL-safe slugs",
			func(input, expected string) {
				Expect(cukesvhs.Slugify(input)).To(Equal(expected))
			},
			Entry("lowercase words", "hello world", "hello-world"),
			Entry("mixed case", "Capture Event", "capture-event"),
			Entry("underscores", "manage_skills", "manage-skills"),
			Entry("multiple spaces", "foo  bar", "foo-bar"),
			Entry("special chars stripped", "foo!@bar", "foobar"),
			Entry("leading/trailing hyphens", "-foo-bar-", "foo-bar"),
		)
	})

	Describe("truncate helper", func() {
		It("returns short strings unchanged", func() {
			Expect(truncate("hello", 10)).To(Equal("hello"))
		})

		It("truncates long strings with ellipsis", func() {
			result := truncate("hello world this is long", 10)
			Expect(result).To(HaveLen(10))
			Expect(result).To(HaveSuffix("..."))
		})

		It("handles max <= 3 without panic", func() {
			result := truncate("hello", 2)
			Expect(len(result)).To(BeNumerically("<=", 2))
		})
	})

	Describe("parseAllScenarios", func() {
		It("returns error for missing features directory", func() {
			var errOut bytes.Buffer
			_, err := parseAllScenarios("/nonexistent/features", "testdata/scenarios/", &errOut)
			Expect(err).To(HaveOccurred())
		})

		It("parses scenarios successfully", func() {
			var errOut bytes.Buffer
			scenarios, err := parseAllScenarios("testdata/features/", "testdata/scenarios/", &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(scenarios).NotTo(BeEmpty())
		})
	})

	Describe("parseGenerateFlags", func() {
		It("returns error when output is missing", func() {
			var errOut bytes.Buffer
			_, err := parseGenerateFlags([]string{"--all"}, &errOut)
			Expect(err).To(HaveOccurred())
		})

		It("returns error when no filter is specified", func() {
			var errOut bytes.Buffer
			_, err := parseGenerateFlags([]string{"--output", "/tmp"}, &errOut)
			Expect(err).To(HaveOccurred())
		})

		It("parses valid flags successfully", func() {
			var errOut bytes.Buffer
			opts, err := parseGenerateFlags([]string{"--output", "/tmp", "--all"}, &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(*opts.outputDir).To(Equal("/tmp"))
			Expect(*opts.generateAll).To(BeTrue())
		})

		It("returns error for unknown flag", func() {
			var errOut bytes.Buffer
			_, err := parseGenerateFlags([]string{"--unknown-flag-xyz"}, &errOut)
			Expect(err).To(HaveOccurred())
		})

		It("treats bare 'all' as --all flag", func() {
			var errOut bytes.Buffer
			opts, err := parseGenerateFlags([]string{"all", "--output", "/tmp"}, &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(*opts.generateAll).To(BeTrue())
			Expect(*opts.outputDir).To(Equal("/tmp"))
		})
	})

	Describe("generateTapes", func() {
		var tmpDir string

		BeforeEach(func() {
			tmpDir = GinkgoT().TempDir()
		})

		Context("verbose mode with untranslatable scenario", func() {
			It("prints skip message and increments warnings", func() {
				var out, errOut bytes.Buffer

				scenario := cukesvhs.ScenarioIR{
					Name:         "Untranslatable",
					Feature:      "Test Feature",
					Source:       cukesvhs.SourceBusiness,
					Translatable: false,
					DemoSteps: []cukesvhs.StepIR{
						{Text: "some step", StepType: "When", Translatable: false, UntranslatableReason: "no match"},
					},
				}

				result := cukesvhs.AnalysisResult{
					ScenarioName: "Untranslatable",
					Feature:      "Test Feature",
					Source:       cukesvhs.SourceBusiness,
					Translatable: false,
				}

				filtered := []scenarioWithResult{{scenario: scenario, result: result}}
				cfg := generateConfig{
					outputDir:    tmpDir,
					configSource: "demos/vhs/config.tape",
					verbose:      true,
					out:          &out,
					errOut:       &errOut,
				}

				stats := generateTapes(filtered, cfg)
				Expect(stats.warnings).To(Equal(1))
				Expect(out.String()).To(ContainSubstring("Skipping"))
			})
		})

		Context("writeScenarioTape error propagation", func() {
			It("prints error and continues when tape write fails", func() {
				var out, errOut bytes.Buffer

				scenario := cukesvhs.ScenarioIR{
					Name:         "Tape Error",
					Feature:      "Error Feature",
					Source:       cukesvhs.SourceBusiness,
					Translatable: true,
					DemoSteps: []cukesvhs.StepIR{
						{
							Text:         "do something",
							StepType:     "When",
							Translatable: true,
							Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}},
						},
					},
				}

				result := cukesvhs.AnalysisResult{
					ScenarioName: "Tape Error",
					Feature:      "Error Feature",
					Source:       cukesvhs.SourceBusiness,
					Translatable: true,
				}

				blockingFile := filepath.Join(tmpDir, "error-feature")
				err := os.WriteFile(blockingFile, []byte("block"), 0o600)
				Expect(err).NotTo(HaveOccurred())

				filtered := []scenarioWithResult{{scenario: scenario, result: result}}
				cfg := generateConfig{
					outputDir:    tmpDir,
					configSource: "demos/vhs/config.tape",
					verbose:      false,
					out:          &out,
					errOut:       &errOut,
				}

				stats := generateTapes(filtered, cfg)
				Expect(stats.total).To(Equal(0))
				Expect(errOut.String()).To(ContainSubstring("Error generating tape"))
			})
		})
	})

	Describe("writeScenarioTape", func() {
		var tmpDir string

		BeforeEach(func() {
			tmpDir = GinkgoT().TempDir()
		})

		Context("VHSOnly source routing", func() {
			It("writes tape to {feature-slug}/ subdirectory matching business layout", func() {
				scenario := cukesvhs.ScenarioIR{
					Name:         "VHS Only Test",
					Feature:      "Vhs Feature",
					Source:       cukesvhs.SourceVHSOnly,
					Translatable: true,
					DemoSteps: []cukesvhs.StepIR{
						{
							Text:         "I select the menu item",
							StepType:     "When",
							Translatable: true,
							Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}},
						},
					},
				}

				outPath, err := writeScenarioTape(scenario, tmpDir, "demos/vhs/config.tape")
				Expect(err).NotTo(HaveOccurred())
				Expect(outPath).To(ContainSubstring(filepath.Join("vhs-feature", "vhs-only-test.tape")))
				Expect(outPath).NotTo(ContainSubstring("scenarios"))
				Expect(outPath).To(HaveSuffix(".tape"))

				_, statErr := os.Stat(outPath)
				Expect(statErr).NotTo(HaveOccurred())
			})
		})

		Context("Business source routing", func() {
			It("writes tape to {feature-slug}/ subdirectory with same layout as VHS-only", func() {
				scenario := cukesvhs.ScenarioIR{
					Name:         "Business Test",
					Feature:      "Business Feature",
					Source:       cukesvhs.SourceBusiness,
					Translatable: true,
					DemoSteps: []cukesvhs.StepIR{
						{
							Text:         "I select the menu item",
							StepType:     "When",
							Translatable: true,
							Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}},
						},
					},
				}

				outPath, err := writeScenarioTape(scenario, tmpDir, "demos/vhs/config.tape")
				Expect(err).NotTo(HaveOccurred())
				Expect(outPath).To(ContainSubstring(filepath.Join("business-feature", "business-test.tape")))
				Expect(outPath).NotTo(ContainSubstring("scenarios"))
				Expect(outPath).To(HaveSuffix(".tape"))

				_, statErr := os.Stat(outPath)
				Expect(statErr).NotTo(HaveOccurred())
			})
		})

		Context("MkdirAll failure", func() {
			It("returns error when output dir cannot be created", func() {
				skipIfWindows("file permission tests not supported on Windows")
				scenario := cukesvhs.ScenarioIR{
					Name:         "MkdirAll Fail",
					Feature:      "Dir Fail",
					Source:       cukesvhs.SourceBusiness,
					Translatable: true,
					DemoSteps: []cukesvhs.StepIR{
						{
							Text:         "do something",
							StepType:     "When",
							Translatable: true,
							Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}},
						},
					},
				}

				_, err := writeScenarioTape(scenario, "/proc/cannot-create-here", "demos/vhs/config.tape")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("creating output directory"))
			})
		})

		Context("WriteFile failure", func() {
			It("returns error when tape file cannot be written to read-only dir", func() {
				skipIfWindows("file permission tests not supported on Windows")
				readOnlyDir := filepath.Join(tmpDir, "dir-fail")
				err := os.MkdirAll(readOnlyDir, 0o750)
				Expect(err).NotTo(HaveOccurred())

				scenario := cukesvhs.ScenarioIR{
					Name:         "Write Fail",
					Feature:      "Dir Fail",
					Source:       cukesvhs.SourceBusiness,
					Translatable: true,
					DemoSteps: []cukesvhs.StepIR{
						{
							Text:         "do something",
							StepType:     "When",
							Translatable: true,
							Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}},
						},
					},
				}

				err = os.Chmod(readOnlyDir, 0o000)
				Expect(err).NotTo(HaveOccurred())
				defer func() { _ = os.Chmod(readOnlyDir, 0o700) }()

				_, err = writeScenarioTape(scenario, tmpDir, "demos/vhs/config.tape")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("filterResults", func() {
		Context("scenario not present in analysis results", func() {
			It("skips scenarios missing from the results map", func() {
				scenarios := []cukesvhs.ScenarioIR{
					{Name: "Present Scenario", Feature: "Feature A", Source: cukesvhs.SourceBusiness},
					{Name: "Missing Scenario", Feature: "Feature A", Source: cukesvhs.SourceBusiness},
				}

				results := []cukesvhs.AnalysisResult{
					{
						ScenarioID:   cukesvhs.BuildScenarioID(cukesvhs.SourceBusiness, "Feature A", "Present Scenario"),
						ScenarioName: "Present Scenario",
						Feature:      "Feature A",
						Source:       cukesvhs.SourceBusiness,
						Translatable: true,
					},
				}

				filtered := filterResults(results, scenarios, true, "", "")
				Expect(filtered).To(HaveLen(1))
				Expect(filtered[0].scenario.Name).To(Equal("Present Scenario"))
			})
		})

		Context("same scenario name in different features", func() {
			It("preserves both scenarios without collision", func() {
				scenarios := []cukesvhs.ScenarioIR{
					{Name: "User logs in", Feature: "Feature A", Source: cukesvhs.SourceBusiness},
					{Name: "User logs in", Feature: "Feature B", Source: cukesvhs.SourceBusiness},
				}

				results := []cukesvhs.AnalysisResult{
					{
						ScenarioID:   cukesvhs.BuildScenarioID(cukesvhs.SourceBusiness, "Feature A", "User logs in"),
						ScenarioName: "User logs in",
						Feature:      "Feature A",
						Source:       cukesvhs.SourceBusiness,
						Translatable: true,
					},
					{
						ScenarioID:   cukesvhs.BuildScenarioID(cukesvhs.SourceBusiness, "Feature B", "User logs in"),
						ScenarioName: "User logs in",
						Feature:      "Feature B",
						Source:       cukesvhs.SourceBusiness,
						Translatable: false,
					},
				}

				filtered := filterResults(results, scenarios, true, "", "")
				Expect(filtered).To(HaveLen(2))
				Expect(filtered[0].scenario.Feature).To(Equal("Feature A"))
				Expect(filtered[0].result.Translatable).To(BeTrue())
				Expect(filtered[1].scenario.Feature).To(Equal("Feature B"))
				Expect(filtered[1].result.Translatable).To(BeFalse())
			})
		})
	})

	Describe("run subcommand", func() {
		var tmpDir string

		BeforeEach(func() {
			tmpDir = GinkgoT().TempDir()
		})

		Context("--output missing", func() {
			It("returns exit code 1 with error message", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"run", "--all", "--features", "testdata/features/"}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("--output is required"))
			})
		})

		Context("no filter flags", func() {
			It("returns exit code 1 requiring --all, --feature, or --scenario", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"run", "--output", tmpDir}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("--all"))
			})
		})

		Context("unknown flag", func() {
			It("returns exit code 1 with error message", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"run", "--unknown-flag-xyz"}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("unknown flag"))
			})
		})

		Context("with non-existent features directory", func() {
			It("returns exit code 1", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{
					"run",
					"--all",
					"--features", "/nonexistent/features/",
					"--scenarios-dir", "/nonexistent/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(code).To(Equal(1))
			})
		})

		Context("--all flag with no translatable scenarios", func() {
			It("exits 0 with empty report", func() {
				emptyFeaturesDir := GinkgoT().TempDir()

				var out, errOut bytes.Buffer
				code := Run([]string{
					"run",
					"--all",
					"--features", emptyFeaturesDir,
					"--scenarios-dir", emptyFeaturesDir,
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				output := out.String()
				Expect(output).To(ContainSubstring("No translatable scenarios found."))
				Expect(output).To(ContainSubstring("Results: 0 PASS, 0 FAIL, 0 NEW"))
			})
		})

		Context("when vhs binary is missing from PATH", func() {
			var origPath string

			BeforeEach(func() {
				origPath = os.Getenv("PATH")
				Expect(os.Setenv("PATH", "")).To(Succeed())
			})

			AfterEach(func() {
				Expect(os.Setenv("PATH", origPath)).To(Succeed())
			})

			It("returns exit code 1 with render error when translatable scenarios exist", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{
					"run",
					"--all",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("Error rendering tapes"))
			})
		})
	})

	Describe("parseRunFlags", func() {
		It("returns error when output is missing", func() {
			var errOut bytes.Buffer
			_, err := parseRunFlags([]string{"--all"}, &errOut)
			Expect(err).To(HaveOccurred())
		})

		It("returns error when no filter is specified", func() {
			var errOut bytes.Buffer
			_, err := parseRunFlags([]string{"--output", "/tmp"}, &errOut)
			Expect(err).To(HaveOccurred())
		})

		It("parses valid --all flag successfully", func() {
			var errOut bytes.Buffer
			opts, err := parseRunFlags([]string{"--output", "/tmp", "--all"}, &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(*opts.outputDir).To(Equal("/tmp"))
			Expect(*opts.runAll).To(BeTrue())
		})

		It("parses --feature flag successfully", func() {
			var errOut bytes.Buffer
			opts, err := parseRunFlags([]string{"--output", "/tmp", "--feature", "Docker Management"}, &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(*opts.featureFilter).To(Equal("Docker Management"))
		})

		It("parses --scenario flag successfully", func() {
			var errOut bytes.Buffer
			opts, err := parseRunFlags([]string{"--output", "/tmp", "--scenario", "My Scenario"}, &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(*opts.scenarioFilter).To(Equal("My Scenario"))
		})

		It("applies default golden dir", func() {
			var errOut bytes.Buffer
			opts, err := parseRunFlags([]string{"--output", "/tmp", "--all"}, &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(*opts.goldenDir).To(Equal("demos/golden/"))
		})

		It("applies default timeout", func() {
			var errOut bytes.Buffer
			opts, err := parseRunFlags([]string{"--output", "/tmp", "--all"}, &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(*opts.timeoutSec).To(Equal(120))
		})

		It("returns error for unknown flag", func() {
			var errOut bytes.Buffer
			_, err := parseRunFlags([]string{"--unknown-flag-xyz"}, &errOut)
			Expect(err).To(HaveOccurred())
		})

		It("treats bare 'all' as --all flag", func() {
			var errOut bytes.Buffer
			opts, err := parseRunFlags([]string{"all", "--output", "/tmp"}, &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(*opts.runAll).To(BeTrue())
			Expect(*opts.outputDir).To(Equal("/tmp"))
		})
	})

	Describe("update-baseline subcommand", func() {
		var tmpDir string

		BeforeEach(func() {
			tmpDir = GinkgoT().TempDir()
		})

		Context("--output missing", func() {
			It("returns exit code 1 with error message", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"update-baseline", "--all"}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("--output is required"))
			})
		})

		Context("no --all and no positional scenario", func() {
			It("returns exit code 1 with error message", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"update-baseline", "--output", tmpDir}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("--all or a scenario name is required"))
			})
		})

		Context("unknown flag", func() {
			It("returns exit code 1 with error message", func() {
				var out, errOut bytes.Buffer
				code := Run([]string{"update-baseline", "--unknown-flag-xyz"}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("unknown flag"))
			})
		})

		Context("--all flag with empty output directory", func() {
			It("exits 0 reporting 0 baselines updated", func() {
				goldenDir := GinkgoT().TempDir()
				outputDir := GinkgoT().TempDir()

				var out, errOut bytes.Buffer
				code := Run([]string{
					"update-baseline",
					"--all",
					"--output", outputDir,
					"--golden", goldenDir,
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				Expect(out.String()).To(ContainSubstring("Updated 0 baselines."))
			})
		})

		Context("--all flag with ascii files present", func() {
			It("updates each baseline and reports count", func() {
				goldenDir := GinkgoT().TempDir()
				outputDir := GinkgoT().TempDir()

				asciiPath := filepath.Join(outputDir, "my-scenario.ascii")
				gifPath := filepath.Join(outputDir, "my-scenario.gif")
				Expect(os.WriteFile(asciiPath, []byte("ascii content"), 0o600)).To(Succeed())
				Expect(os.WriteFile(gifPath, []byte("gif content"), 0o600)).To(Succeed())

				var out, errOut bytes.Buffer
				code := Run([]string{
					"update-baseline",
					"--all",
					"--output", outputDir,
					"--golden", goldenDir,
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				Expect(out.String()).To(ContainSubstring("Updated 1 baselines."))
				Expect(out.String()).To(ContainSubstring("Updated:"))
			})
		})

		Context("positional scenario name", func() {
			It("updates the named scenario baseline", func() {
				goldenDir := GinkgoT().TempDir()
				outputDir := GinkgoT().TempDir()

				asciiPath := filepath.Join(outputDir, "my-scenario.ascii")
				gifPath := filepath.Join(outputDir, "my-scenario.gif")
				Expect(os.WriteFile(asciiPath, []byte("ascii content"), 0o600)).To(Succeed())
				Expect(os.WriteFile(gifPath, []byte("gif content"), 0o600)).To(Succeed())

				var out, errOut bytes.Buffer
				code := Run([]string{
					"update-baseline",
					"--output", outputDir,
					"--golden", goldenDir,
					"my-scenario",
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				Expect(out.String()).To(ContainSubstring("Updated 1 baselines."))
			})
		})

		Context("positional scenario name with nested feature directory", func() {
			It("finds and updates the baseline from a subdirectory", func() {
				goldenDir := GinkgoT().TempDir()
				outputDir := GinkgoT().TempDir()

				featureDir := filepath.Join(outputDir, "user-authentication")
				Expect(os.MkdirAll(featureDir, 0o750)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(featureDir, "user-logs-in.ascii"), []byte("ascii content"), 0o600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(featureDir, "user-logs-in.gif"), []byte("gif content"), 0o600)).To(Succeed())

				var out, errOut bytes.Buffer
				code := Run([]string{
					"update-baseline",
					"--output", outputDir,
					"--golden", goldenDir,
					"User logs in",
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				Expect(out.String()).To(ContainSubstring("Updated 1 baselines."))
			})
		})

		Context("positional scenario not found in output directory", func() {
			It("returns exit code 1 with error message", func() {
				goldenDir := GinkgoT().TempDir()
				outputDir := GinkgoT().TempDir()

				var out, errOut bytes.Buffer
				code := Run([]string{
					"update-baseline",
					"--output", outputDir,
					"--golden", goldenDir,
					"nonexistent-scenario",
				}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("Error finding baseline"))
			})
		})

		Context("positional scenario with ambiguous match across features", func() {
			It("returns exit code 1 with ambiguous error", func() {
				goldenDir := GinkgoT().TempDir()
				outputDir := GinkgoT().TempDir()

				featureA := filepath.Join(outputDir, "feature-a")
				featureB := filepath.Join(outputDir, "feature-b")
				Expect(os.MkdirAll(featureA, 0o750)).To(Succeed())
				Expect(os.MkdirAll(featureB, 0o750)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(featureA, "user-logs-in.ascii"), []byte("a"), 0o600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(featureA, "user-logs-in.gif"), []byte("ga"), 0o600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(featureB, "user-logs-in.ascii"), []byte("b"), 0o600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(featureB, "user-logs-in.gif"), []byte("gb"), 0o600)).To(Succeed())

				var out, errOut bytes.Buffer
				code := Run([]string{
					"update-baseline",
					"--output", outputDir,
					"--golden", goldenDir,
					"User logs in",
				}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("ambiguous"))
			})
		})

		Context("--all flag with non-existent output directory", func() {
			It("exits 0 reporting 0 baselines (missing dir is not an error)", func() {
				goldenDir := GinkgoT().TempDir()

				var out, errOut bytes.Buffer
				code := Run([]string{
					"update-baseline",
					"--all",
					"--output", "/nonexistent/output/dir/",
					"--golden", goldenDir,
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				Expect(out.String()).To(ContainSubstring("Updated 0 baselines."))
			})
		})

		Context("--all flag when UpdateBaseline fails due to read-only golden dir", func() {
			It("returns exit code 1 with error message", func() {
				skipIfWindows("file permission tests not supported on Windows")
				goldenDir := GinkgoT().TempDir()
				outputDir := GinkgoT().TempDir()

				asciiPath := filepath.Join(outputDir, "my-scenario.ascii")
				gifPath := filepath.Join(outputDir, "my-scenario.gif")
				Expect(os.WriteFile(asciiPath, []byte("ascii content"), 0o600)).To(Succeed())
				Expect(os.WriteFile(gifPath, []byte("gif content"), 0o600)).To(Succeed())

				Expect(os.Chmod(goldenDir, 0o000)).To(Succeed())
				defer func() { _ = os.Chmod(goldenDir, 0o700) }()

				var out, errOut bytes.Buffer
				code := Run([]string{
					"update-baseline",
					"--all",
					"--output", outputDir,
					"--golden", goldenDir,
				}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("Error updating baseline"))
			})
		})

		Context("positional scenario when UpdateBaseline fails due to read-only golden dir", func() {
			It("returns exit code 1 with error message", func() {
				skipIfWindows("file permission tests not supported on Windows")
				goldenDir := GinkgoT().TempDir()
				outputDir := GinkgoT().TempDir()

				asciiPath := filepath.Join(outputDir, "my-scenario.ascii")
				gifPath := filepath.Join(outputDir, "my-scenario.gif")
				Expect(os.WriteFile(asciiPath, []byte("ascii content"), 0o600)).To(Succeed())
				Expect(os.WriteFile(gifPath, []byte("gif content"), 0o600)).To(Succeed())

				Expect(os.Chmod(goldenDir, 0o000)).To(Succeed())
				defer func() { _ = os.Chmod(goldenDir, 0o700) }()

				var out, errOut bytes.Buffer
				code := Run([]string{
					"update-baseline",
					"--output", outputDir,
					"--golden", goldenDir,
					"my-scenario",
				}, &out, &errOut)

				Expect(code).To(Equal(1))
				Expect(errOut.String()).To(ContainSubstring("Error updating baseline"))
			})
		})
	})

	Describe("parseUpdateBaselineFlags", func() {
		It("returns error when output is missing", func() {
			var errOut bytes.Buffer
			_, _, err := parseUpdateBaselineFlags([]string{"--all"}, &errOut)
			Expect(err).To(HaveOccurred())
		})

		It("parses --all flag successfully", func() {
			var errOut bytes.Buffer
			opts, _, err := parseUpdateBaselineFlags([]string{"--all", "--output", "/tmp"}, &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(*opts.updateAll).To(BeTrue())
		})

		It("applies default golden dir", func() {
			var errOut bytes.Buffer
			opts, _, err := parseUpdateBaselineFlags([]string{"--all", "--output", "/tmp"}, &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(*opts.goldenDir).To(Equal("demos/golden/"))
		})

		It("captures positional scenario names", func() {
			var errOut bytes.Buffer
			_, positional, err := parseUpdateBaselineFlags(
				[]string{"--output", "/tmp", "scenario-a", "scenario-b"},
				&errOut,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(positional).To(ConsistOf("scenario-a", "scenario-b"))
		})

		It("returns error for unknown flag", func() {
			var errOut bytes.Buffer
			_, _, err := parseUpdateBaselineFlags([]string{"--unknown-flag-xyz"}, &errOut)
			Expect(err).To(HaveOccurred())
		})

		It("treats bare 'all' as --all flag", func() {
			var errOut bytes.Buffer
			opts, _, err := parseUpdateBaselineFlags([]string{"all", "--output", "/tmp"}, &errOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(*opts.updateAll).To(BeTrue())
			Expect(*opts.outputDir).To(Equal("/tmp"))
		})
	})

	Describe("reportPipelineResults", func() {
		It("returns 0 when all results are PASS", func() {
			results := []cukesvhs.ValidationResult{
				{Scenario: "scenario-a", Status: cukesvhs.ValidationPass},
				{Scenario: "scenario-b", Status: cukesvhs.ValidationPass},
			}
			var out bytes.Buffer
			code := reportPipelineResults(&out, results)
			Expect(code).To(Equal(0))
			Expect(out.String()).To(ContainSubstring("[PASS] scenario-a"))
			Expect(out.String()).To(ContainSubstring("[PASS] scenario-b"))
			Expect(out.String()).To(ContainSubstring("Results: 2 PASS, 0 FAIL, 0 NEW"))
		})

		It("returns 1 when any result is FAIL", func() {
			results := []cukesvhs.ValidationResult{
				{Scenario: "scenario-a", Status: cukesvhs.ValidationPass},
				{Scenario: "scenario-b", Status: cukesvhs.ValidationFail},
			}
			var out bytes.Buffer
			code := reportPipelineResults(&out, results)
			Expect(code).To(Equal(1))
			Expect(out.String()).To(ContainSubstring("[FAIL] scenario-b"))
			Expect(out.String()).To(ContainSubstring("Results: 1 PASS, 1 FAIL, 0 NEW"))
		})

		It("returns 0 when results are NEW", func() {
			results := []cukesvhs.ValidationResult{
				{Scenario: "scenario-new", Status: cukesvhs.ValidationNew},
			}
			var out bytes.Buffer
			code := reportPipelineResults(&out, results)
			Expect(code).To(Equal(0))
			Expect(out.String()).To(ContainSubstring("[NEW]  scenario-new"))
			Expect(out.String()).To(ContainSubstring("Results: 0 PASS, 0 FAIL, 1 NEW"))
		})

		It("returns 0 with empty results", func() {
			var out bytes.Buffer
			code := reportPipelineResults(&out, []cukesvhs.ValidationResult{})
			Expect(code).To(Equal(0))
			Expect(out.String()).To(ContainSubstring("Results: 0 PASS, 0 FAIL, 0 NEW"))
		})
	})

	Describe("collectOutputASCIIFiles", func() {
		It("returns empty slice for non-existent directory", func() {
			files, err := collectOutputASCIIFiles("/nonexistent/path/")
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(BeEmpty())
		})

		It("returns empty slice for empty directory", func() {
			dir := GinkgoT().TempDir()
			files, err := collectOutputASCIIFiles(dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(BeEmpty())
		})

		It("returns ascii files in directory", func() {
			dir := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(dir, "a.ascii"), []byte("x"), 0o600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dir, "b.gif"), []byte("x"), 0o600)).To(Succeed())

			files, err := collectOutputASCIIFiles(dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(1))
			Expect(files[0]).To(HaveSuffix("a.ascii"))
		})
	})

	Describe("deriveScenarioName", func() {
		It("strips output dir prefix and .ascii suffix, slugifies", func() {
			outputDir := "/tmp/output"
			asciiPath := "/tmp/output/feature-a/scenario-b.ascii"
			result := deriveScenarioName(outputDir, asciiPath)
			Expect(result).To(Equal("feature-a-scenario-b"))
		})

		It("handles root-level ascii files", func() {
			outputDir := "/tmp/output"
			asciiPath := "/tmp/output/my-scenario.ascii"
			result := deriveScenarioName(outputDir, asciiPath)
			Expect(result).To(Equal("my-scenario"))
		})
	})

	Describe("deriveGIFPath", func() {
		It("replaces .ascii extension with .gif", func() {
			result := deriveGIFPath("/tmp/output/my-scenario.ascii")
			Expect(result).To(Equal("/tmp/output/my-scenario.gif"))
		})
	})

	Describe("findASCIIFileForScenario", func() {
		Context("when ascii file is in a nested feature subdirectory", func() {
			It("finds the file by scenario slug", func() {
				outputDir := GinkgoT().TempDir()
				featureDir := filepath.Join(outputDir, "user-authentication")
				Expect(os.MkdirAll(featureDir, 0o750)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(featureDir, "user-logs-in.ascii"), []byte("content"), 0o600)).To(Succeed())

				result, err := findASCIIFileForScenario(outputDir, "user-logs-in")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(filepath.Join(featureDir, "user-logs-in.ascii")))
			})
		})

		Context("when ascii file is at the root of outputDir", func() {
			It("finds the file", func() {
				outputDir := GinkgoT().TempDir()
				Expect(os.WriteFile(filepath.Join(outputDir, "my-scenario.ascii"), []byte("content"), 0o600)).To(Succeed())

				result, err := findASCIIFileForScenario(outputDir, "my-scenario")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(filepath.Join(outputDir, "my-scenario.ascii")))
			})
		})

		Context("when no matching ascii file exists", func() {
			It("returns an error", func() {
				outputDir := GinkgoT().TempDir()

				_, err := findASCIIFileForScenario(outputDir, "nonexistent-scenario")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no .ascii file found"))
			})
		})

		Context("when the same scenario slug exists in multiple feature directories", func() {
			It("returns an ambiguous error", func() {
				outputDir := GinkgoT().TempDir()
				featureA := filepath.Join(outputDir, "feature-a")
				featureB := filepath.Join(outputDir, "feature-b")
				Expect(os.MkdirAll(featureA, 0o750)).To(Succeed())
				Expect(os.MkdirAll(featureB, 0o750)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(featureA, "user-logs-in.ascii"), []byte("a"), 0o600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(featureB, "user-logs-in.ascii"), []byte("b"), 0o600)).To(Succeed())

				_, err := findASCIIFileForScenario(outputDir, "user-logs-in")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ambiguous"))
			})
		})

		Context("when outputDir does not exist", func() {
			It("returns an error", func() {
				_, err := findASCIIFileForScenario("/nonexistent/output/dir", "scenario")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("pipelineTimeout", func() {
		It("converts seconds to duration", func() {
			d := pipelineTimeout(60)
			Expect(d).To(Equal(60 * time.Second))
		})

		It("handles zero seconds", func() {
			d := pipelineTimeout(0)
			Expect(d).To(Equal(time.Duration(0)))
		})
	})

	Describe("runListCount", func() {
		Context("with empty results", func() {
			It("outputs zero counts when no scenarios exist", func() {
				var out bytes.Buffer
				code := runListCount([]cukesvhs.AnalysisResult{}, &out)
				Expect(code).To(Equal(0))
				Expect(out.String()).To(MatchRegexp(`Business: 0/0 translatable \| VHS-only: 0/0 translatable`))
			})
		})

		Context("with translatable business and vhs-only scenarios", func() {
			It("counts translatable scenarios correctly for both sources", func() {
				results := []cukesvhs.AnalysisResult{
					{ScenarioName: "Scenario A", Feature: "Feature A", Source: cukesvhs.SourceBusiness, Translatable: true},
					{ScenarioName: "Scenario B", Feature: "Feature A", Source: cukesvhs.SourceBusiness, Translatable: false},
					{ScenarioName: "Scenario C", Feature: "Feature B", Source: cukesvhs.SourceVHSOnly, Translatable: true},
				}
				var out bytes.Buffer
				code := runListCount(results, &out)
				Expect(code).To(Equal(0))
				Expect(out.String()).To(ContainSubstring("Business: 1/2 translatable"))
				Expect(out.String()).To(ContainSubstring("VHS-only: 1/1 translatable"))
			})
		})
	})

	Describe("runList ParseFeatureDir error", func() {
		It("returns exit code 1 when features dir has a malformed feature file", func() {
			dir := GinkgoT().TempDir()
			err := os.WriteFile(filepath.Join(dir, "bad.feature"), []byte("this is: not: valid: gherkin:\n  garbage yaml"), 0o600)
			Expect(err).NotTo(HaveOccurred())

			var out, errOut bytes.Buffer
			code := Run([]string{"list", "--features", dir, "--scenarios-dir", "/nonexistent/scenarios/"}, &out, &errOut)

			Expect(code).To(Equal(1))
			Expect(errOut.String()).To(ContainSubstring("Error parsing features dir"))
		})
	})

	Describe("parseAllScenarios ParseFeatureDir errors", func() {
		It("returns error when business features dir has a malformed feature file", func() {
			dir := GinkgoT().TempDir()
			err := os.WriteFile(filepath.Join(dir, "bad.feature"), []byte("not valid gherkin {{{{"), 0o600)
			Expect(err).NotTo(HaveOccurred())

			var errOut bytes.Buffer
			_, err = parseAllScenarios(dir, "/nonexistent/scenarios/", &errOut)
			Expect(err).To(HaveOccurred())
		})

		It("returns error when vhs-only dir has a malformed feature file", func() {
			featuresDir := GinkgoT().TempDir()
			scenariosDir := GinkgoT().TempDir()
			err := os.WriteFile(filepath.Join(scenariosDir, "bad.feature"), []byte("not valid gherkin {{{{"), 0o600)
			Expect(err).NotTo(HaveOccurred())

			var errOut bytes.Buffer
			_, err = parseAllScenarios(featuresDir, scenariosDir, &errOut)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("generateTapes business source counting", func() {
		var tmpDir string

		BeforeEach(func() {
			tmpDir = GinkgoT().TempDir()
		})

		It("increments fromBusiness counter for translatable business scenario", func() {
			var out, errOut bytes.Buffer

			scenario := cukesvhs.ScenarioIR{
				Name:         "Business Translatable",
				Feature:      "Business Feature",
				Source:       cukesvhs.SourceBusiness,
				Translatable: true,
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         "I navigate to the menu",
						StepType:     "When",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Enter}},
					},
				},
			}

			result := cukesvhs.AnalysisResult{
				ScenarioName: "Business Translatable",
				Feature:      "Business Feature",
				Source:       cukesvhs.SourceBusiness,
				Translatable: true,
			}

			filtered := []scenarioWithResult{{scenario: scenario, result: result}}
			cfg := generateConfig{
				outputDir:    tmpDir,
				configSource: "demos/vhs/config.tape",
				verbose:      false,
				out:          &out,
				errOut:       &errOut,
			}

			stats := generateTapes(filtered, cfg)
			Expect(stats.fromBusiness).To(Equal(1))
			Expect(stats.total).To(Equal(1))
		})
	})

	Describe("writeScenarioTape GenerateTape error", func() {
		var tmpDir string

		BeforeEach(func() {
			tmpDir = GinkgoT().TempDir()
		})

		It("returns error when GenerateTape fails due to forbidden pattern", func() {
			scenario := cukesvhs.ScenarioIR{
				Name:         "Forbidden",
				Feature:      "Dangerous",
				Source:       cukesvhs.SourceBusiness,
				Translatable: true,
				DemoSteps: []cukesvhs.StepIR{
					{
						Text:         "dangerous step",
						StepType:     "When",
						Translatable: true,
						Commands:     []cukesvhs.VHSCommand{{Type: cukesvhs.Type, Args: []string{"rm -rf /tmp"}}},
					},
				},
			}

			_, err := writeScenarioTape(scenario, tmpDir, "demos/vhs/config.tape")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("forbidden pattern"))
		})
	})
})

// failWriter is an io.Writer that always returns an error.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("write error") }

var _ = Describe("runInitCmd", func() {
	Context("with valid args and writable directory", func() {
		It("returns 0", func() {
			tmpDir := GinkgoT().TempDir()
			outDir := filepath.Join(tmpDir, "config")

			var out, errOut bytes.Buffer
			code := runInitCmd([]string{"--output", outDir}, &out, &errOut)

			Expect(code).To(Equal(0))
			Expect(errOut.String()).To(BeEmpty())
		})
	})

	Context("with invalid flag", func() {
		It("returns 1", func() {
			var out, errOut bytes.Buffer
			code := runInitCmd([]string{"--nonexistent-flag"}, &out, &errOut)

			Expect(code).To(Equal(1))
		})
	})

	Context("with unwritable directory", func() {
		It("returns 1 with error in errOut", func() {
			if runtime.GOOS == "windows" {
				Skip("chmod not supported on Windows")
			}
			if os.Getuid() == 0 {
				Skip("running as root bypasses permission checks")
			}

			tmpDir := GinkgoT().TempDir()
			readOnlyDir := filepath.Join(tmpDir, "readonly")
			Expect(os.MkdirAll(readOnlyDir, 0o555)).To(Succeed())

			var out, errOut bytes.Buffer
			code := runInitCmd([]string{"--output", filepath.Join(readOnlyDir, "subdir")}, &out, &errOut)

			Expect(code).To(Equal(1))
			Expect(errOut.String()).To(ContainSubstring("Error"))
		})
	})
})

var _ = Describe("collectOutputASCIIFiles walkErr branch", func() {
	It("returns error when directory walk encounters permission error", func() {
		if runtime.GOOS == "windows" {
			Skip("chmod not supported on Windows")
		}
		if os.Getuid() == 0 {
			Skip("running as root bypasses permission checks")
		}

		tmpDir := GinkgoT().TempDir()
		subDir := filepath.Join(tmpDir, "blocked")
		Expect(os.MkdirAll(subDir, 0o750)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(subDir, "test.ascii"), []byte("content"), 0o600)).To(Succeed())
		Expect(os.Chmod(subDir, 0o000)).To(Succeed())
		DeferCleanup(func() { os.Chmod(subDir, 0o750) })

		files, err := collectOutputASCIIFiles(tmpDir)

		Expect(err).To(HaveOccurred())
		Expect(files).To(BeNil())
	})
})

var _ = Describe("runListJSON encode error", func() {
	It("returns 1 and writes error to errOut when writer fails", func() {
		results := []cukesvhs.AnalysisResult{
			{ScenarioName: "test", Feature: "test", Source: cukesvhs.SourceBusiness, Translatable: true},
		}

		var errOut bytes.Buffer
		code := runListJSON(results, failWriter{}, &errOut)

		Expect(code).To(Equal(1))
		Expect(errOut.String()).To(ContainSubstring("Error encoding JSON"))
	})
})

var _ = Describe("updateAllBaselines error path", func() {
	It("returns 1 when UpdateBaseline fails due to unwritable goldenDir", func() {
		if runtime.GOOS == "windows" {
			Skip("chmod not supported on Windows")
		}
		if os.Getuid() == 0 {
			Skip("running as root bypasses permission checks")
		}

		tmpDir := GinkgoT().TempDir()

		outputDir := filepath.Join(tmpDir, "output")
		Expect(os.MkdirAll(outputDir, 0o750)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(outputDir, "test-scenario.ascii"), []byte("content"), 0o600)).To(Succeed())

		goldenDir := filepath.Join(tmpDir, "golden")
		Expect(os.MkdirAll(goldenDir, 0o750)).To(Succeed())
		Expect(os.Chmod(goldenDir, 0o555)).To(Succeed())
		DeferCleanup(func() { os.Chmod(goldenDir, 0o750) })

		var out, errOut bytes.Buffer
		code := updateAllBaselines(goldenDir, outputDir, &out, &errOut)

		Expect(code).To(Equal(1))
		Expect(errOut.String()).To(ContainSubstring("Error updating baseline"))
	})
})

var _ = Describe("renderAndValidate error path", func() {
	It("returns 1 and writes error to errOut when RenderAll fails", func() {
		var out, errOut bytes.Buffer
		code := renderAndValidate(&out, &errOut, "/nonexistent/tape/dir", "/nonexistent/golden", 120, "")

		Expect(code).To(Equal(1))
		Expect(errOut.String()).To(ContainSubstring("Error rendering"))
	})
})

var _ = Describe("renderAndValidate with failed render results", func() {
	It("returns 1 when individual tape renders fail", func() {
		skipIfNoVHS()

		tempDir := GinkgoT().TempDir()
		goldenDir := GinkgoT().TempDir()

		tapePath := filepath.Join(tempDir, "invalid.tape")
		err := os.WriteFile(tapePath, []byte("Output /tmp/invalid.gif\nType invalid command\n"), 0o600)
		Expect(err).NotTo(HaveOccurred())

		var out, errOut bytes.Buffer
		code := renderAndValidate(&out, &errOut, tempDir, goldenDir, 5, "")

		Expect(code).To(Equal(1))
		Expect(errOut.String()).To(ContainSubstring("Error rendering"))
	})
})

var _ = Describe("benchmark subcommand", func() {
	Context("with --all flag and valid features directory", func() {
		It("outputs JSON results to stdout", func() {
			var out, errOut bytes.Buffer
			code := Run([]string{
				"benchmark",
				"--all",
				"--features", "testdata/features/",
				"--scenarios-dir", "testdata/scenarios/",
			}, &out, &errOut)

			Expect(code).To(Equal(0))
			Expect(errOut.String()).To(ContainSubstring("Benchmarked"))

			outputBytes := out.Bytes()
			if len(outputBytes) > 0 {
				var payload map[string]interface{}
				jsonStart := bytes.Index(outputBytes, []byte("{"))
				if jsonStart >= 0 {
					err := json.Unmarshal(outputBytes[jsonStart:], &payload)
					Expect(err).NotTo(HaveOccurred())
				}
			}
		})
	})

	Context("with --output flag", func() {
		It("writes JSON results to the specified file", func() {
			tmpDir := GinkgoT().TempDir()
			outputFile := filepath.Join(tmpDir, "benchmark-results.json")

			var out, errOut bytes.Buffer
			code := Run([]string{
				"benchmark",
				"--all",
				"--features", "testdata/features/",
				"--scenarios-dir", "testdata/scenarios/",
				"--output", outputFile,
			}, &out, &errOut)

			Expect(code).To(Equal(0))
			Expect(errOut.String()).To(ContainSubstring("Benchmarked"))

			_, err := os.Stat(outputFile)
			Expect(err).NotTo(HaveOccurred())

			data, err := os.ReadFile(outputFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).NotTo(BeEmpty())

			var payload map[string]interface{}
			err = json.Unmarshal(data, &payload)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with --runs flag", func() {
		It("accepts a custom number of benchmark runs", func() {
			var out, errOut bytes.Buffer
			code := Run([]string{
				"benchmark",
				"--all",
				"--features", "testdata/features/",
				"--scenarios-dir", "testdata/scenarios/",
				"--runs", "1",
			}, &out, &errOut)

			Expect(code).To(Equal(0))
		})
	})

	Context("without --all flag", func() {
		It("returns exit code 1 requiring --all", func() {
			var out, errOut bytes.Buffer
			code := Run([]string{
				"benchmark",
				"--features", "testdata/features/",
			}, &out, &errOut)

			Expect(code).To(Equal(1))
			Expect(errOut.String()).To(ContainSubstring("--all"))
		})
	})

	Context("with non-existent features directory", func() {
		It("returns exit code 1", func() {
			var out, errOut bytes.Buffer
			code := Run([]string{
				"benchmark",
				"--all",
				"--features", "/nonexistent/features/",
				"--scenarios-dir", "/nonexistent/scenarios/",
			}, &out, &errOut)

			Expect(code).To(Equal(1))
		})
	})

	Context("unknown flag", func() {
		It("returns exit code 1 with error message", func() {
			var out, errOut bytes.Buffer
			code := Run([]string{"benchmark", "--unknown-flag-xyz"}, &out, &errOut)

			Expect(code).To(Equal(1))
			Expect(errOut.String()).To(ContainSubstring("unknown flag"))
		})
	})

	Context("summary output", func() {
		It("prints benchmarked count to stderr", func() {
			var out, errOut bytes.Buffer
			Run([]string{
				"benchmark",
				"--all",
				"--features", "testdata/features/",
				"--scenarios-dir", "testdata/scenarios/",
			}, &out, &errOut)

			Expect(errOut.String()).To(MatchRegexp(`Benchmarked \d+ commands across \d+ scenarios`))
		})
	})

	Context("with empty features directory", func() {
		It("exits 0 with zero benchmarks", func() {
			emptyDir := GinkgoT().TempDir()

			var out, errOut bytes.Buffer
			code := Run([]string{
				"benchmark",
				"--all",
				"--features", emptyDir,
				"--scenarios-dir", emptyDir,
			}, &out, &errOut)

			Expect(code).To(Equal(0))
			Expect(errOut.String()).To(ContainSubstring("Benchmarked 0 commands across 0 scenarios"))
		})
	})
})
