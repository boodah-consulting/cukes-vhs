package vhsgen_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukesvhs/internal/vhsgen"
)

// writeASCIIFile writes content to a .ascii file at path and returns the path.
func writeASCIIFile(path, content string) {
	Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())
}

var _ = Describe("Validator", func() {
	var (
		goldenDir string
		outputDir string
	)

	BeforeEach(func() {
		goldenDir = GinkgoT().TempDir()
		outputDir = GinkgoT().TempDir()
	})

	Describe("ValidationStatus constants", func() {
		It("defines PASS, FAIL, and NEW statuses", func() {
			Expect(vhsgen.ValidationPass).To(Equal(vhsgen.ValidationStatus("PASS")))
			Expect(vhsgen.ValidationFail).To(Equal(vhsgen.ValidationStatus("FAIL")))
			Expect(vhsgen.ValidationNew).To(Equal(vhsgen.ValidationStatus("NEW")))
		})
	})

	Describe("ValidationResult type", func() {
		It("has all required fields", func() {
			r := vhsgen.ValidationResult{
				Scenario:   "my scenario",
				ASCIIPath:  "/current/output.ascii",
				GoldenPath: "/golden/baseline.txt",
				Status:     vhsgen.ValidationPass,
				Diff:       "",
			}
			Expect(r.Scenario).To(Equal("my scenario"))
			Expect(r.ASCIIPath).To(Equal("/current/output.ascii"))
			Expect(r.GoldenPath).To(Equal("/golden/baseline.txt"))
			Expect(r.Status).To(Equal(vhsgen.ValidationPass))
			Expect(r.Diff).To(BeEmpty())
		})
	})

	Describe("ValidateScenario", func() {
		Context("when no golden baseline exists", func() {
			It("returns ValidationNew status", func() {
				asciiPath := filepath.Join(outputDir, "output.ascii")
				writeASCIIFile(asciiPath, "hello world\n")

				result, err := vhsgen.ValidateScenario(goldenDir, "My Scenario", asciiPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Status).To(Equal(vhsgen.ValidationNew))
			})

			It("populates GoldenPath with the newly saved baseline", func() {
				asciiPath := filepath.Join(outputDir, "output.ascii")
				writeASCIIFile(asciiPath, "hello world\n")

				result, err := vhsgen.ValidateScenario(goldenDir, "My Scenario", asciiPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.GoldenPath).NotTo(BeEmpty())
			})

			It("auto-saves current output as the golden baseline", func() {
				asciiPath := filepath.Join(outputDir, "output.ascii")
				writeASCIIFile(asciiPath, "new baseline content\n")

				_, err := vhsgen.ValidateScenario(goldenDir, "New Scenario", asciiPath)
				Expect(err).NotTo(HaveOccurred())

				_, exists, getErr := vhsgen.GetBaseline(goldenDir, "New Scenario")
				Expect(getErr).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})

			It("sets the Scenario field on the result", func() {
				asciiPath := filepath.Join(outputDir, "output.ascii")
				writeASCIIFile(asciiPath, "content\n")

				result, err := vhsgen.ValidateScenario(goldenDir, "Brand New", asciiPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Scenario).To(Equal("Brand New"))
			})

			It("sets ASCIIPath to the current file", func() {
				asciiPath := filepath.Join(outputDir, "output.ascii")
				writeASCIIFile(asciiPath, "content\n")

				result, err := vhsgen.ValidateScenario(goldenDir, "Brand New", asciiPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ASCIIPath).To(Equal(asciiPath))
			})
		})

		Context("when the golden baseline exists and content matches", func() {
			var asciiPath string

			BeforeEach(func() {
				asciiPath = filepath.Join(outputDir, "output.ascii")
				writeASCIIFile(asciiPath, "line one\nline two\n")

				_, err := vhsgen.ValidateScenario(goldenDir, "Match Scenario", asciiPath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns ValidationPass status", func() {
				result, err := vhsgen.ValidateScenario(goldenDir, "Match Scenario", asciiPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Status).To(Equal(vhsgen.ValidationPass))
			})

			It("returns an empty Diff field", func() {
				result, err := vhsgen.ValidateScenario(goldenDir, "Match Scenario", asciiPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Diff).To(BeEmpty())
			})

			It("populates GoldenPath", func() {
				result, err := vhsgen.ValidateScenario(goldenDir, "Match Scenario", asciiPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.GoldenPath).NotTo(BeEmpty())
			})
		})

		Context("when content matches after ANSI code stripping", func() {
			var (
				cleanPath string
				ansiPath  string
			)

			BeforeEach(func() {
				cleanContent := "hello world\nfoo bar\n"
				ansiContent := "\x1b[32mhello world\x1b[0m\n\x1b[1mfoo bar\x1b[0m\n"

				cleanPath = filepath.Join(outputDir, "clean.ascii")
				ansiPath = filepath.Join(outputDir, "ansi.ascii")

				writeASCIIFile(cleanPath, cleanContent)
				writeASCIIFile(ansiPath, ansiContent)

				_, err := vhsgen.ValidateScenario(goldenDir, "ANSI Scenario", cleanPath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns ValidationPass when ANSI-decorated content matches stripped golden", func() {
				result, err := vhsgen.ValidateScenario(goldenDir, "ANSI Scenario", ansiPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Status).To(Equal(vhsgen.ValidationPass))
			})
		})

		Context("when content matches after trailing whitespace normalisation", func() {
			var asciiPath string

			BeforeEach(func() {
				goldenContent := "line one\nline two\n"
				asciiPath = filepath.Join(outputDir, "trailing.ascii")
				writeASCIIFile(asciiPath, goldenContent)

				_, err := vhsgen.ValidateScenario(goldenDir, "Trailing Scenario", asciiPath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns PASS when trailing spaces differ but trimmed content matches", func() {
				trailingSpacePath := filepath.Join(outputDir, "trailing-spaces.ascii")
				writeASCIIFile(trailingSpacePath, "line one   \nline two  \n")

				result, err := vhsgen.ValidateScenario(goldenDir, "Trailing Scenario", trailingSpacePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Status).To(Equal(vhsgen.ValidationPass))
			})
		})

		Context("when the golden baseline exists and content differs", func() {
			var (
				asciiPath   string
				changedPath string
			)

			BeforeEach(func() {
				asciiPath = filepath.Join(outputDir, "output.ascii")
				writeASCIIFile(asciiPath, "line one\nline two\n")

				_, err := vhsgen.ValidateScenario(goldenDir, "Diff Scenario", asciiPath)
				Expect(err).NotTo(HaveOccurred())

				changedPath = filepath.Join(outputDir, "changed.ascii")
				writeASCIIFile(changedPath, "line one\nline CHANGED\n")
			})

			It("returns ValidationFail status", func() {
				result, err := vhsgen.ValidateScenario(goldenDir, "Diff Scenario", changedPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Status).To(Equal(vhsgen.ValidationFail))
			})

			It("populates the Diff field with a non-empty diff", func() {
				result, err := vhsgen.ValidateScenario(goldenDir, "Diff Scenario", changedPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Diff).NotTo(BeEmpty())
			})

			It("includes --- golden header in the diff", func() {
				result, err := vhsgen.ValidateScenario(goldenDir, "Diff Scenario", changedPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Diff).To(ContainSubstring("--- golden"))
			})

			It("includes +++ current header in the diff", func() {
				result, err := vhsgen.ValidateScenario(goldenDir, "Diff Scenario", changedPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Diff).To(ContainSubstring("+++ current"))
			})

			It("shows removed lines prefixed with -", func() {
				result, err := vhsgen.ValidateScenario(goldenDir, "Diff Scenario", changedPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Diff).To(ContainSubstring("-line two"))
			})

			It("shows added lines prefixed with +", func() {
				result, err := vhsgen.ValidateScenario(goldenDir, "Diff Scenario", changedPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Diff).To(ContainSubstring("+line CHANGED"))
			})
		})

		Context("when the current ASCII file does not exist", func() {
			BeforeEach(func() {
				existingPath := filepath.Join(outputDir, "existing.ascii")
				writeASCIIFile(existingPath, "content\n")

				_, err := vhsgen.ValidateScenario(goldenDir, "Missing Current", existingPath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error when the current file is missing", func() {
				_, err := vhsgen.ValidateScenario(goldenDir, "Missing Current", "/nonexistent/file.ascii")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("reading current ASCII"))
			})
		})

		Context("when the golden baseline file is unreadable after being found", func() {
			It("returns an error mentioning reading golden ASCII", func() {
				asciiPath := filepath.Join(outputDir, "output.ascii")
				writeASCIIFile(asciiPath, "content\n")

				_, err := vhsgen.ValidateScenario(goldenDir, "Unreadable Golden", asciiPath)
				Expect(err).NotTo(HaveOccurred())

				goldenASCII, _, _ := vhsgen.GetBaseline(goldenDir, "Unreadable Golden")
				Expect(os.Chmod(goldenASCII, 0o000)).To(Succeed())
				defer os.Chmod(goldenASCII, 0o644) //nolint:errcheck

				_, err = vhsgen.ValidateScenario(goldenDir, "Unreadable Golden", asciiPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("reading golden ASCII"))
			})
		})

		Context("when goldenDir is not writable (placeholder dir creation fails)", func() {
			It("returns an error mentioning creating placeholder GIF", func() {
				readOnlyGoldenDir := GinkgoT().TempDir()
				Expect(os.Chmod(readOnlyGoldenDir, 0o555)).To(Succeed())
				defer os.Chmod(readOnlyGoldenDir, 0o755) //nolint:errcheck

				asciiPath := filepath.Join(outputDir, "output.ascii")
				writeASCIIFile(asciiPath, "content\n")

				_, err := vhsgen.ValidateScenario(readOnlyGoldenDir, "Locked Scenario", asciiPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("creating placeholder GIF"))
			})
		})

		Context("when placeholder dir exists but is not writable (GIF write fails)", func() {
			It("returns an error mentioning creating placeholder GIF", func() {
				customGoldenDir := GinkgoT().TempDir()

				placeholderDir := filepath.Join(customGoldenDir, ".placeholders")
				Expect(os.MkdirAll(placeholderDir, 0o755)).To(Succeed())
				Expect(os.Chmod(placeholderDir, 0o555)).To(Succeed())
				defer os.Chmod(placeholderDir, 0o755) //nolint:errcheck

				asciiPath := filepath.Join(outputDir, "output.ascii")
				writeASCIIFile(asciiPath, "content\n")

				_, err := vhsgen.ValidateScenario(customGoldenDir, "Write Locked Placeholder", asciiPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("creating placeholder GIF"))
			})
		})

		Context("when content is completely different (no common lines)", func() {
			It("returns ValidationFail with all lines shown as removed and added", func() {
				goldenPath := filepath.Join(outputDir, "golden-src.ascii")
				writeASCIIFile(goldenPath, "alpha\nbeta\ngamma\n")

				_, err := vhsgen.ValidateScenario(goldenDir, "No Common", goldenPath)
				Expect(err).NotTo(HaveOccurred())

				newPath := filepath.Join(outputDir, "new-content.ascii")
				writeASCIIFile(newPath, "one\ntwo\nthree\n")

				result, err := vhsgen.ValidateScenario(goldenDir, "No Common", newPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Status).To(Equal(vhsgen.ValidationFail))
				Expect(result.Diff).To(ContainSubstring("-alpha"))
				Expect(result.Diff).To(ContainSubstring("+one"))
			})
		})

		Context("when lines differ before and after a common anchor", func() {
			It("returns ValidationFail showing context lines around the common anchor", func() {
				goldenPath := filepath.Join(outputDir, "anchor-golden.ascii")
				writeASCIIFile(goldenPath, "aaa\ncommon\nccc\n")

				_, err := vhsgen.ValidateScenario(goldenDir, "Anchor Diff", goldenPath)
				Expect(err).NotTo(HaveOccurred())

				currentPath := filepath.Join(outputDir, "anchor-current.ascii")
				writeASCIIFile(currentPath, "XXX\ncommon\nYYY\n")

				result, err := vhsgen.ValidateScenario(goldenDir, "Anchor Diff", currentPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Status).To(Equal(vhsgen.ValidationFail))
				Expect(result.Diff).To(ContainSubstring("-aaa"))
				Expect(result.Diff).To(ContainSubstring("+XXX"))
				Expect(result.Diff).To(ContainSubstring(" common"))
				Expect(result.Diff).To(ContainSubstring("-ccc"))
				Expect(result.Diff).To(ContainSubstring("+YYY"))
			})
		})

		Context("when current has extra lines beyond the golden content", func() {
			It("returns ValidationFail with the extra lines shown as additions", func() {
				goldenPath := filepath.Join(outputDir, "short-golden.ascii")
				writeASCIIFile(goldenPath, "shared line\nend of golden")

				_, err := vhsgen.ValidateScenario(goldenDir, "Extra Lines Current", goldenPath)
				Expect(err).NotTo(HaveOccurred())

				longerPath := filepath.Join(outputDir, "longer-current.ascii")
				writeASCIIFile(longerPath, "shared line\nextra one\nextra two")

				result, err := vhsgen.ValidateScenario(goldenDir, "Extra Lines Current", longerPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Status).To(Equal(vhsgen.ValidationFail))
				Expect(result.Diff).To(ContainSubstring("+extra one"))
				Expect(result.Diff).To(ContainSubstring("+extra two"))
			})
		})

		Context("when diffs occur far apart (multiple separate hunks)", func() {
			It("returns ValidationFail with separated diff hunks", func() {
				goldenPath := filepath.Join(outputDir, "multi-hunk-golden.ascii")
				goldenContent := "aaa\nbbb\nccc\nddd\neee\nfff\nggg\nhhh\niii\njjj\n"
				writeASCIIFile(goldenPath, goldenContent)

				_, err := vhsgen.ValidateScenario(goldenDir, "Multi Hunk", goldenPath)
				Expect(err).NotTo(HaveOccurred())

				changedPath := filepath.Join(outputDir, "multi-hunk-current.ascii")
				changedContent := "AAA\nbbb\nccc\nddd\neee\nfff\nggg\nhhh\niii\nJJJ\n"
				writeASCIIFile(changedPath, changedContent)

				result, err := vhsgen.ValidateScenario(goldenDir, "Multi Hunk", changedPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Status).To(Equal(vhsgen.ValidationFail))
				Expect(result.Diff).NotTo(BeEmpty())
			})
		})
	})

	Describe("ValidateAll", func() {
		Context("when the output directory is empty", func() {
			It("returns an empty results slice and no error", func() {
				results, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		Context("when the output directory does not exist", func() {
			It("returns an error", func() {
				results, err := vhsgen.ValidateAll(goldenDir, "/nonexistent/output/dir")
				Expect(err).To(HaveOccurred())
				Expect(results).To(BeNil())
			})
		})

		Context("with multiple .ascii files", func() {
			BeforeEach(func() {
				writeASCIIFile(filepath.Join(outputDir, "scene1.ascii"), "scene one content\n")
				writeASCIIFile(filepath.Join(outputDir, "scene2.ascii"), "scene two content\n")
			})

			It("returns one result per .ascii file", func() {
				results, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(2))
			})

			It("marks all as NEW when no baselines exist", func() {
				results, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())

				for _, r := range results {
					Expect(r.Status).To(Equal(vhsgen.ValidationNew))
				}
			})

			It("marks all as PASS after baselines are created", func() {
				_, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())

				results, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())

				for _, r := range results {
					Expect(r.Status).To(Equal(vhsgen.ValidationPass))
				}
			})
		})

		Context("with .ascii files in subdirectories", func() {
			BeforeEach(func() {
				subDir := filepath.Join(outputDir, "feature")
				Expect(os.MkdirAll(subDir, 0o755)).To(Succeed())

				writeASCIIFile(filepath.Join(outputDir, "top.ascii"), "top level\n")
				writeASCIIFile(filepath.Join(subDir, "nested.ascii"), "nested content\n")
			})

			It("discovers .ascii files recursively", func() {
				results, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(2))
			})
		})

		Context("when some files match and some differ", func() {
			var passingPath string

			BeforeEach(func() {
				passingPath = filepath.Join(outputDir, "passing.ascii")
				writeASCIIFile(passingPath, "stable content\n")

				failingPath := filepath.Join(outputDir, "failing.ascii")
				writeASCIIFile(failingPath, "original content\n")

				_, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())

				writeASCIIFile(failingPath, "changed content\n")
			})

			It("returns all results without stopping on FAIL", func() {
				results, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(2))
			})

			It("reports PASS for unchanged file", func() {
				results, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())

				statuses := make(map[string]vhsgen.ValidationStatus)
				for _, r := range results {
					statuses[r.ASCIIPath] = r.Status
				}

				Expect(statuses[passingPath]).To(Equal(vhsgen.ValidationPass))
			})

			It("reports FAIL for changed file", func() {
				results, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())

				var hasFail bool
				for _, r := range results {
					if r.Status == vhsgen.ValidationFail {
						hasFail = true
						break
					}
				}

				Expect(hasFail).To(BeTrue())
			})
		})

		Context("when files do not have .ascii extension", func() {
			BeforeEach(func() {
				writeASCIIFile(filepath.Join(outputDir, "notascii.txt"), "text content\n")
				writeASCIIFile(filepath.Join(outputDir, "output.gif"), "gif content\n")
			})

			It("ignores non-.ascii files", func() {
				results, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		Context("when an .ascii file is unreadable causing ValidateScenario to error", func() {
			It("records a FAIL result with the error message rather than stopping", func() {
				asciiPath := filepath.Join(outputDir, "readable.ascii")
				writeASCIIFile(asciiPath, "content\n")

				_, saveErr := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(saveErr).NotTo(HaveOccurred())

				Expect(os.Chmod(asciiPath, 0o000)).To(Succeed())
				defer os.Chmod(asciiPath, 0o644) //nolint:errcheck

				results, err := vhsgen.ValidateAll(goldenDir, outputDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))
				Expect(results[0].Status).To(Equal(vhsgen.ValidationFail))
				Expect(results[0].Diff).NotTo(BeEmpty())
			})
		})
	})
})
