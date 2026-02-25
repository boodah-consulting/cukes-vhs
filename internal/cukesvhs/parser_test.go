package cukesvhs_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

func featuresDir() string {
	dir := filepath.Join("testdata", "features")
	if _, err := os.Stat(dir); err != nil {
		return ""
	}
	return dir
}

var _ = Describe("ParseFeatureDir", func() {
	Describe("parsing real features directory", func() {
		var dir string

		BeforeEach(func() {
			dir = featuresDir()
		})

		It("returns at least one scenario", func() {
			if dir == "" {
				Skip("testdata/features directory not found")
			}
			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).NotTo(BeEmpty())
		})

		It("returns scenarios from multiple features", func() {
			if dir == "" {
				Skip("testdata/features directory not found")
			}
			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())

			featureNames := make(map[string]bool)
			for _, ir := range results {
				featureNames[ir.Feature] = true
			}
			Expect(len(featureNames)).To(BeNumerically(">=", 2))
		})
	})

	Describe("Source field propagation", func() {
		var dir string

		BeforeEach(func() {
			dir = featuresDir()
		})

		Context("with SourceBusiness", func() {
			It("sets Source to SourceBusiness on all results", func() {
				if dir == "" {
					Expect(dir).NotTo(BeEmpty(), "features directory not found")
				}
				results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
				Expect(err).NotTo(HaveOccurred())
				for _, ir := range results {
					Expect(ir.Source).To(Equal(cukesvhs.SourceBusiness),
						"scenario %q should have SourceBusiness", ir.Name)
				}
			})
		})

		Context("with SourceVHSOnly", func() {
			It("sets Source to SourceVHSOnly on all results", func() {
				results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceVHSOnly)
				Expect(err).NotTo(HaveOccurred())
				for _, ir := range results {
					Expect(ir.Source).To(Equal(cukesvhs.SourceVHSOnly),
						"scenario %q should have SourceVHSOnly", ir.Name)
				}
			})
		})
	})

	Describe("Background steps", func() {
		var dir string

		BeforeEach(func() {
			dir = featuresDir()
		})

		It("includes Background steps in SetupSteps for Git Workflows scenarios", func() {
			if dir == "" {
				Skip("testdata/features directory not found")
			}
			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())

			var skillsScenarios []cukesvhs.ScenarioIR
			for _, ir := range results {
				if ir.Feature == "Git Workflows" {
					skillsScenarios = append(skillsScenarios, ir)
				}
			}
			Expect(skillsScenarios).NotTo(BeEmpty(), "expected scenarios from Git Workflows feature")

			for _, ir := range skillsScenarios {
				Expect(ir.SetupSteps).NotTo(BeEmpty(),
					"scenario %q: expected SetupSteps from Background", ir.Name)

				foundBackground := false
				for _, step := range ir.SetupSteps {
					if step.Text == "I am on the main menu" && step.StepType == "Given" {
						foundBackground = true
						break
					}
				}
				Expect(foundBackground).To(BeTrue(),
					"scenario %q: SetupSteps missing Background step 'I am on the main menu'", ir.Name)
			}
		})

		It("places Background step first in SetupSteps for Git Workflows", func() {
			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())

			for _, ir := range results {
				if ir.Feature != "Git Workflows" || len(ir.SetupSteps) == 0 {
					continue
				}
				Expect(ir.SetupSteps[0].Text).To(Equal("I am on the main menu"),
					"scenario %q: first SetupStep should be from Background", ir.Name)
			}
		})
	})

	Describe("empty and missing directories", func() {
		It("returns zero results for an empty directory", func() {
			dir := GinkgoT().TempDir()
			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(BeEmpty())
		})

		It("returns zero results for a non-existent directory", func() {
			results, err := cukesvhs.ParseFeatureDir("/tmp/nonexistent-cukesvhs-dir-test", cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(BeEmpty())
		})

		It("returns zero results for a directory with no .feature files", func() {
			dir := GinkgoT().TempDir()
			err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not a feature"), 0o600)
			Expect(err).NotTo(HaveOccurred())

			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(BeEmpty())
		})

		It("returns zero results for a .feature file with no Feature block", func() {
			dir := GinkgoT().TempDir()
			content := `# This is just a comment, no feature`
			err := os.WriteFile(filepath.Join(dir, "empty.feature"), []byte(content), 0o600)
			Expect(err).NotTo(HaveOccurred())

			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(BeEmpty())
		})
	})

	Describe("step classification", func() {
		var dir string

		BeforeEach(func() {
			dir = featuresDir()
			Expect(dir).NotTo(BeEmpty(), "features directory must exist")
		})

		It("classifies SetupSteps as Given and DemoSteps as When or Then", func() {
			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())

			for _, ir := range results {
				for _, step := range ir.SetupSteps {
					Expect(step.StepType).To(Equal("Given"),
						"scenario %q: SetupStep %q should be Given", ir.Name, step.Text)
				}
				for _, step := range ir.DemoSteps {
					Expect(step.StepType).To(SatisfyAny(Equal("When"), Equal("Then")),
						"scenario %q: DemoStep %q should be When or Then", ir.Name, step.Text)
				}
			}
		})
	})

	Describe("Scenario Outline", func() {
		Context("with two example rows", func() {
			var dir string
			var results []cukesvhs.ScenarioIR

			BeforeEach(func() {
				dir = GinkgoT().TempDir()
				content := `Feature: Test Outline

  Scenario Outline: Greet user
    Given I am on the main menu
    When I enter "<input>"
    Then I should see "<output>"

    Examples:
      | input | output      |
      | Alice | Hello Alice |
      | Bob   | Hello Bob   |
`
				err := os.WriteFile(filepath.Join(dir, "outline.feature"), []byte(content), 0o600)
				Expect(err).NotTo(HaveOccurred())

				results, err = cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceVHSOnly)
				Expect(err).NotTo(HaveOccurred())
			})

			It("produces exactly 1 scenario (first row only)", func() {
				Expect(results).To(HaveLen(1))
			})

			It("preserves the scenario name", func() {
				Expect(results[0].Name).To(Equal("Greet user"))
			})

			It("sets the correct source", func() {
				Expect(results[0].Source).To(Equal(cukesvhs.SourceVHSOnly))
			})

			It("sets the correct feature name", func() {
				Expect(results[0].Feature).To(Equal("Test Outline"))
			})

			It("substitutes the first row input value", func() {
				foundAlice := false
				for _, step := range results[0].DemoSteps {
					if strings.Contains(step.Text, "Alice") {
						foundAlice = true
					}
				}
				Expect(foundAlice).To(BeTrue(), "expected DemoStep with 'Alice' after substitution")
			})

			It("substitutes the first row output value", func() {
				foundHelloAlice := false
				for _, step := range results[0].DemoSteps {
					if strings.Contains(step.Text, "Hello Alice") {
						foundHelloAlice = true
					}
				}
				Expect(foundHelloAlice).To(BeTrue(), "expected DemoStep with 'Hello Alice' after substitution")
			})

			It("does not contain unsubstituted placeholders", func() {
				allSteps := append(results[0].SetupSteps, results[0].DemoSteps...) //nolint:gocritic
				for _, step := range allSteps {
					Expect(step.Text).NotTo(ContainSubstring("<input>"))
					Expect(step.Text).NotTo(ContainSubstring("<output>"))
				}
			})
		})

		Context("outline preserves setup steps", func() {
			It("correctly separates Given setup from When demo steps", func() {
				dir := GinkgoT().TempDir()
				content := `Feature: Outline With Setup

  Scenario Outline: Test with setup
    Given I am on the main menu
    When I enter "<text>"

    Examples:
      | text  |
      | hello |
      | world |
`
				err := os.WriteFile(filepath.Join(dir, "setup_outline.feature"), []byte(content), 0o600)
				Expect(err).NotTo(HaveOccurred())

				results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))

				ir := results[0]
				Expect(ir.SetupSteps).To(HaveLen(1))
				Expect(ir.SetupSteps[0].Text).To(Equal("I am on the main menu"))
				Expect(ir.DemoSteps).To(HaveLen(1))
				Expect(ir.DemoSteps[0].Text).To(ContainSubstring("hello"))
			})
		})

		Context("outline with multiple substitution columns", func() {
			It("uses the first example row values", func() {
				dir := GinkgoT().TempDir()
				content := `Feature: Substitution Test
  Scenario Outline: Test with examples
    Given I have <count> items
    When I add <item>
    Then I have <total> items

    Examples:
      | count | item | total |
      | 5     | one  | 6     |
      | 10    | two  | 12    |`

				err := os.WriteFile(filepath.Join(dir, "outline.feature"), []byte(content), 0o600)
				Expect(err).NotTo(HaveOccurred())

				results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))

				Expect(results[0].SetupSteps[0].Text).To(ContainSubstring("5"))
				Expect(results[0].DemoSteps[0].Text).To(ContainSubstring("one"))
			})
		})
	})

	Describe("translate step integration", func() {
		var dir string

		BeforeEach(func() {
			dir = featuresDir()
			Expect(dir).NotTo(BeEmpty(), "features directory must exist")
		})

		It("finds both translatable and untranslatable steps across all features", func() {
			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())

			translatableFound := false
			untranslatableFound := false

			for _, ir := range results {
				for _, step := range append(ir.SetupSteps, ir.DemoSteps...) {
					if step.Translatable {
						translatableFound = true
					} else {
						untranslatableFound = true
					}
				}
			}

			Expect(translatableFound).To(BeTrue(), "expected at least one translatable step")
			Expect(untranslatableFound).To(BeTrue(), "expected at least one untranslatable step")
		})
	})

	Describe("tags", func() {
		var dir string

		BeforeEach(func() {
			dir = featuresDir()
			Expect(dir).NotTo(BeEmpty(), "features directory must exist")
		})

		It("parses scenario tags from the real features", func() {
			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())

			taggedCount := 0
			for _, ir := range results {
				if len(ir.Tags) > 0 {
					taggedCount++
				}
			}
			Expect(taggedCount).To(BeNumerically(">", 0), "expected at least some tagged scenarios")
		})
	})

	Describe("multiple feature files", func() {
		It("returns one scenario per feature file", func() {
			dir := GinkgoT().TempDir()

			content1 := `Feature: Feature One
  Scenario: Scenario One
    Given setup
    When action
    Then result`

			content2 := `Feature: Feature Two
  Scenario: Scenario Two
    Given setup
    When action
    Then result`

			err := os.WriteFile(filepath.Join(dir, "feature1.feature"), []byte(content1), 0o600)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(filepath.Join(dir, "feature2.feature"), []byte(content2), 0o600)
			Expect(err).NotTo(HaveOccurred())

			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(2))
		})
	})

	Describe("error paths", func() {
		It("returns error when a .feature file has invalid gherkin syntax", func() {
			dir := GinkgoT().TempDir()
			err := os.WriteFile(filepath.Join(dir, "bad.feature"), []byte("not valid gherkin {{{{"), 0o600)
			Expect(err).NotTo(HaveOccurred())

			_, err = cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("walking directory"))
		})
	})

	Describe("Scenario Outline with Background", func() {
		It("applies substituteExampleValues to background steps", func() {
			dir := GinkgoT().TempDir()
			content := `Feature: Outline With Background

  Background:
    Given I have <count> items

  Scenario Outline: Use outline with background
    When I add <item>
    Then I see <result>

    Examples:
      | count | item | result |
      | 3     | pen  | 4      |
`
			err := os.WriteFile(filepath.Join(dir, "bg_outline.feature"), []byte(content), 0o600)
			Expect(err).NotTo(HaveOccurred())

			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))

			ir := results[0]
			Expect(ir.SetupSteps).To(HaveLen(1))
			Expect(ir.SetupSteps[0].Text).To(ContainSubstring("3"))
			Expect(ir.SetupSteps[0].Text).NotTo(ContainSubstring("<count>"))
		})
	})

	Describe("Scenario Outline with no valid Examples", func() {
		It("produces an IR with unsubstituted placeholders in demo steps when Examples have no table body", func() {
			dir := GinkgoT().TempDir()
			content := `Feature: Empty Examples

  Scenario Outline: No examples
    Given I have something
    When I do <action>
    Then I see <result>

    Examples:
      | action | result |
`
			err := os.WriteFile(filepath.Join(dir, "empty_examples.feature"), []byte(content), 0o600)
			Expect(err).NotTo(HaveOccurred())

			results, err := cukesvhs.ParseFeatureDir(dir, cukesvhs.SourceBusiness)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))

			ir := results[0]
			Expect(ir.DemoSteps).NotTo(BeEmpty())

			foundPlaceholder := false
			for _, step := range ir.DemoSteps {
				if strings.Contains(step.Text, "<") {
					foundPlaceholder = true
				}
			}
			Expect(foundPlaceholder).To(BeTrue(), "expected unsubstituted placeholder when no example rows")
		})
	})
})
