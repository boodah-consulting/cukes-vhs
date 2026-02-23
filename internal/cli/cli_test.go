package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukes-vhs/internal/vhsgen"
)

func skipIfNoVHS() {
	if _, err := exec.LookPath("vhs"); err != nil {
		Skip("vhs binary not available")
	}
}

func skipIfWindows() {
	if runtime.GOOS == "windows" {
		Skip("file permission tests not supported on Windows")
	}
}

var _ = Describe("vhsgen CLI", func() {
	Context("no subcommand", func() {
		It("prints usage and returns 0", func() {
			var out, errOut bytes.Buffer
			code := Run([]string{}, &out, &errOut)

			Expect(code).To(Equal(0))
			Expect(out.String()).To(ContainSubstring("vhsgen"))
			Expect(out.String()).To(ContainSubstring("list"))
			Expect(out.String()).To(ContainSubstring("generate"))
		})
	})

	Context("--help flag", func() {
		It("prints usage and returns 0", func() {
			var out, errOut bytes.Buffer
			code := Run([]string{"--help"}, &out, &errOut)

			Expect(code).To(Equal(0))
			Expect(out.String()).To(ContainSubstring("vhsgen"))
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
				Expect(errOut.String()).To(ContainSubstring("Error parsing flags"))
			})
		})
	})

	Describe("generate subcommand", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "vhsgen-test-*")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
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
				skipIfNoVHS()
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
				Expect(output).To(ContainSubstring("Rendering..."))
				Expect(output).To(MatchRegexp(`Generated \d+ tapes`))
			})

			It("runs without error even when all scenarios are untranslatable", func() {
				skipIfNoVHS()
				var out, errOut bytes.Buffer
				code := Run([]string{
					"generate",
					"--all",
					"--features", "testdata/features/",
					"--scenarios-dir", "testdata/scenarios/",
					"--output", tmpDir,
				}, &out, &errOut)

				Expect(code).To(Equal(0))
				Expect(errOut.String()).To(BeEmpty())
			})

			It("reports 'Written: <path>' for each tape file written", func() {
				skipIfNoVHS()
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
				skipIfNoVHS()
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
				skipIfNoVHS()
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
				skipIfNoVHS()
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
				skipIfNoVHS()
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
	})

	Describe("slugify helper", func() {
		DescribeTable("converts names to URL-safe slugs",
			func(input, expected string) {
				Expect(vhsgen.Slugify(input)).To(Equal(expected))
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
			var err error
			tmpDir, err = os.MkdirTemp("", "vhsgen-gentapes-*")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		Context("verbose mode with untranslatable scenario", func() {
			It("prints skip message and increments warnings", func() {
				var out, errOut bytes.Buffer

				scenario := vhsgen.ScenarioIR{
					Name:         "Untranslatable",
					Feature:      "Test Feature",
					Source:       vhsgen.SourceBusiness,
					Translatable: false,
					DemoSteps: []vhsgen.StepIR{
						{Text: "some step", StepType: "When", Translatable: false, UntranslatableReason: "no match"},
					},
				}

				result := vhsgen.AnalysisResult{
					ScenarioName: "Untranslatable",
					Feature:      "Test Feature",
					Source:       vhsgen.SourceBusiness,
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

				scenario := vhsgen.ScenarioIR{
					Name:         "Tape Error",
					Feature:      "Error Feature",
					Source:       vhsgen.SourceBusiness,
					Translatable: true,
					DemoSteps: []vhsgen.StepIR{
						{
							Text:         "do something",
							StepType:     "When",
							Translatable: true,
							Commands:     []vhsgen.VHSCommand{{Type: vhsgen.Enter}},
						},
					},
				}

				result := vhsgen.AnalysisResult{
					ScenarioName: "Tape Error",
					Feature:      "Error Feature",
					Source:       vhsgen.SourceBusiness,
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
			var err error
			tmpDir, err = os.MkdirTemp("", "vhsgen-writetape-*")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		Context("VHSOnly source routing", func() {
			It("writes tape to scenarios/{feature-slug}/ subdirectory", func() {
				scenario := vhsgen.ScenarioIR{
					Name:         "VHS Only Test",
					Feature:      "Vhs Feature",
					Source:       vhsgen.SourceVHSOnly,
					Translatable: true,
					DemoSteps: []vhsgen.StepIR{
						{
							Text:         "I select the menu item",
							StepType:     "When",
							Translatable: true,
							Commands:     []vhsgen.VHSCommand{{Type: vhsgen.Enter}},
						},
					},
				}

				outPath, err := writeScenarioTape(scenario, tmpDir, "demos/vhs/config.tape")
				Expect(err).NotTo(HaveOccurred())
				Expect(outPath).To(ContainSubstring(filepath.Join("scenarios", "vhs-feature")))
				Expect(outPath).To(HaveSuffix(".tape"))

				_, statErr := os.Stat(outPath)
				Expect(statErr).NotTo(HaveOccurred())
			})
		})

		Context("MkdirAll failure", func() {
			It("returns error when output dir cannot be created", func() {
				skipIfWindows()
				scenario := vhsgen.ScenarioIR{
					Name:         "MkdirAll Fail",
					Feature:      "Dir Fail",
					Source:       vhsgen.SourceBusiness,
					Translatable: true,
					DemoSteps: []vhsgen.StepIR{
						{
							Text:         "do something",
							StepType:     "When",
							Translatable: true,
							Commands:     []vhsgen.VHSCommand{{Type: vhsgen.Enter}},
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
				skipIfWindows()
				readOnlyDir := filepath.Join(tmpDir, "dir-fail")
				err := os.MkdirAll(readOnlyDir, 0o755)
				Expect(err).NotTo(HaveOccurred())

				scenario := vhsgen.ScenarioIR{
					Name:         "Write Fail",
					Feature:      "Dir Fail",
					Source:       vhsgen.SourceBusiness,
					Translatable: true,
					DemoSteps: []vhsgen.StepIR{
						{
							Text:         "do something",
							StepType:     "When",
							Translatable: true,
							Commands:     []vhsgen.VHSCommand{{Type: vhsgen.Enter}},
						},
					},
				}

				err = os.Chmod(readOnlyDir, 0o000)
				Expect(err).NotTo(HaveOccurred())
				defer os.Chmod(readOnlyDir, 0o755) //nolint:errcheck

				_, err = writeScenarioTape(scenario, tmpDir, "demos/vhs/config.tape")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("filterResults", func() {
		Context("scenario not present in analysis results", func() {
			It("skips scenarios missing from the results map", func() {
				scenarios := []vhsgen.ScenarioIR{
					{Name: "Present Scenario", Feature: "Feature A", Source: vhsgen.SourceBusiness},
					{Name: "Missing Scenario", Feature: "Feature A", Source: vhsgen.SourceBusiness},
				}

				results := []vhsgen.AnalysisResult{
					{ScenarioName: "Present Scenario", Feature: "Feature A", Source: vhsgen.SourceBusiness, Translatable: true},
				}

				filtered := filterResults(results, scenarios, true, "", "")
				Expect(filtered).To(HaveLen(1))
				Expect(filtered[0].scenario.Name).To(Equal("Present Scenario"))
			})
		})
	})

	Describe("run subcommand", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "vhsgen-run-*")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
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
				Expect(errOut.String()).To(ContainSubstring("Error parsing flags"))
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
			var err error
			tmpDir, err = os.MkdirTemp("", "vhsgen-update-*")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
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
				Expect(errOut.String()).To(ContainSubstring("Error parsing flags"))
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
				skipIfWindows()
				goldenDir := GinkgoT().TempDir()
				outputDir := GinkgoT().TempDir()

				asciiPath := filepath.Join(outputDir, "my-scenario.ascii")
				gifPath := filepath.Join(outputDir, "my-scenario.gif")
				Expect(os.WriteFile(asciiPath, []byte("ascii content"), 0o600)).To(Succeed())
				Expect(os.WriteFile(gifPath, []byte("gif content"), 0o600)).To(Succeed())

				Expect(os.Chmod(goldenDir, 0o000)).To(Succeed())
				defer os.Chmod(goldenDir, 0o755) //nolint:errcheck

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
				skipIfWindows()
				goldenDir := GinkgoT().TempDir()
				outputDir := GinkgoT().TempDir()

				asciiPath := filepath.Join(outputDir, "my-scenario.ascii")
				gifPath := filepath.Join(outputDir, "my-scenario.gif")
				Expect(os.WriteFile(asciiPath, []byte("ascii content"), 0o600)).To(Succeed())
				Expect(os.WriteFile(gifPath, []byte("gif content"), 0o600)).To(Succeed())

				Expect(os.Chmod(goldenDir, 0o000)).To(Succeed())
				defer os.Chmod(goldenDir, 0o755) //nolint:errcheck

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
			results := []vhsgen.ValidationResult{
				{Scenario: "scenario-a", Status: vhsgen.ValidationPass},
				{Scenario: "scenario-b", Status: vhsgen.ValidationPass},
			}
			var out bytes.Buffer
			code := reportPipelineResults(&out, results)
			Expect(code).To(Equal(0))
			Expect(out.String()).To(ContainSubstring("[PASS] scenario-a"))
			Expect(out.String()).To(ContainSubstring("[PASS] scenario-b"))
			Expect(out.String()).To(ContainSubstring("Results: 2 PASS, 0 FAIL, 0 NEW"))
		})

		It("returns 1 when any result is FAIL", func() {
			results := []vhsgen.ValidationResult{
				{Scenario: "scenario-a", Status: vhsgen.ValidationPass},
				{Scenario: "scenario-b", Status: vhsgen.ValidationFail},
			}
			var out bytes.Buffer
			code := reportPipelineResults(&out, results)
			Expect(code).To(Equal(1))
			Expect(out.String()).To(ContainSubstring("[FAIL] scenario-b"))
			Expect(out.String()).To(ContainSubstring("Results: 1 PASS, 1 FAIL, 0 NEW"))
		})

		It("returns 0 when results are NEW", func() {
			results := []vhsgen.ValidationResult{
				{Scenario: "scenario-new", Status: vhsgen.ValidationNew},
			}
			var out bytes.Buffer
			code := reportPipelineResults(&out, results)
			Expect(code).To(Equal(0))
			Expect(out.String()).To(ContainSubstring("[NEW]  scenario-new"))
			Expect(out.String()).To(ContainSubstring("Results: 0 PASS, 0 FAIL, 1 NEW"))
		})

		It("returns 0 with empty results", func() {
			var out bytes.Buffer
			code := reportPipelineResults(&out, []vhsgen.ValidationResult{})
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
				code := runListCount([]vhsgen.AnalysisResult{}, &out)
				Expect(code).To(Equal(0))
				Expect(out.String()).To(MatchRegexp(`Business: 0/0 translatable \| VHS-only: 0/0 translatable`))
			})
		})

		Context("with translatable business and vhs-only scenarios", func() {
			It("counts translatable scenarios correctly for both sources", func() {
				results := []vhsgen.AnalysisResult{
					{ScenarioName: "Scenario A", Feature: "Feature A", Source: vhsgen.SourceBusiness, Translatable: true},
					{ScenarioName: "Scenario B", Feature: "Feature A", Source: vhsgen.SourceBusiness, Translatable: false},
					{ScenarioName: "Scenario C", Feature: "Feature B", Source: vhsgen.SourceVHSOnly, Translatable: true},
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
			var err error
			tmpDir, err = os.MkdirTemp("", "vhsgen-business-*")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		It("increments fromBusiness counter for translatable business scenario", func() {
			var out, errOut bytes.Buffer

			scenario := vhsgen.ScenarioIR{
				Name:         "Business Translatable",
				Feature:      "Business Feature",
				Source:       vhsgen.SourceBusiness,
				Translatable: true,
				DemoSteps: []vhsgen.StepIR{
					{
						Text:         "I navigate to the menu",
						StepType:     "When",
						Translatable: true,
						Commands:     []vhsgen.VHSCommand{{Type: vhsgen.Enter}},
					},
				},
			}

			result := vhsgen.AnalysisResult{
				ScenarioName: "Business Translatable",
				Feature:      "Business Feature",
				Source:       vhsgen.SourceBusiness,
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
			var err error
			tmpDir, err = os.MkdirTemp("", "vhsgen-gentape-err-*")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		It("returns error when GenerateTape fails due to forbidden pattern", func() {
			scenario := vhsgen.ScenarioIR{
				Name:         "Forbidden",
				Feature:      "Dangerous",
				Source:       vhsgen.SourceBusiness,
				Translatable: true,
				DemoSteps: []vhsgen.StepIR{
					{
						Text:         "dangerous step",
						StepType:     "When",
						Translatable: true,
						Commands:     []vhsgen.VHSCommand{{Type: vhsgen.Type, Args: []string{"rm -rf /tmp"}}},
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

func (failWriter) Write([]byte) (int, error) { return 0, fmt.Errorf("write error") }

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
		Expect(os.MkdirAll(subDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(subDir, "test.ascii"), []byte("content"), 0o644)).To(Succeed())
		Expect(os.Chmod(subDir, 0o000)).To(Succeed())
		DeferCleanup(func() { os.Chmod(subDir, 0o755) })

		files, err := collectOutputASCIIFiles(tmpDir)

		Expect(err).To(HaveOccurred())
		Expect(files).To(BeNil())
	})
})

var _ = Describe("runListJSON encode error", func() {
	It("returns 1 and writes error to errOut when writer fails", func() {
		results := []vhsgen.AnalysisResult{
			{ScenarioName: "test", Feature: "test", Source: vhsgen.SourceBusiness, Translatable: true},
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
		Expect(os.MkdirAll(outputDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(outputDir, "test-scenario.ascii"), []byte("content"), 0o644)).To(Succeed())

		goldenDir := filepath.Join(tmpDir, "golden")
		Expect(os.MkdirAll(goldenDir, 0o755)).To(Succeed())
		Expect(os.Chmod(goldenDir, 0o555)).To(Succeed())
		DeferCleanup(func() { os.Chmod(goldenDir, 0o755) })

		var out, errOut bytes.Buffer
		code := updateAllBaselines(goldenDir, outputDir, &out, &errOut)

		Expect(code).To(Equal(1))
		Expect(errOut.String()).To(ContainSubstring("Error updating baseline"))
	})
})

var _ = Describe("renderAndValidate error path", func() {
	It("returns 1 and writes error to errOut when RenderAll fails", func() {
		var out, errOut bytes.Buffer
		code := renderAndValidate(&out, &errOut, "/nonexistent/tape/dir", "/nonexistent/golden", 120)

		Expect(code).To(Equal(1))
		Expect(errOut.String()).To(ContainSubstring("Error rendering"))
	})
})
