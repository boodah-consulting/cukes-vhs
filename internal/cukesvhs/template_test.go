package cukesvhs_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

var _ = Describe("RenderTape", func() {
	Describe("rendering a full tape", func() {
		var (
			data   cukesvhs.TapeData
			result string
		)

		BeforeEach(func() {
			data = cukesvhs.TapeData{
				FeatureName:      "User Registration",
				ScenarioName:     "Successful registration",
				GIFPath:          "demos/vhs/features/user-registration/happy-path.gif",
				ConfigSourcePath: "demos/vhs/config.tape",
				SetupCommands: `Type "mkdir -p /tmp/demo"
Enter
Sleep 300ms`,
				DemoCommands: `Type "./kariya --config /tmp/demo/config.yaml"
Enter
Sleep 2s`,
			}

			var err error
			result, err = cukesvhs.RenderTape(data)
			Expect(err).NotTo(HaveOccurred())
		})

		It("contains the feature comment", func() {
			Expect(result).To(ContainSubstring("# Feature: User Registration"))
		})

		It("contains the scenario comment", func() {
			Expect(result).To(ContainSubstring("# Scenario: Successful registration"))
		})

		It("includes the configuration source path", func() {
			Expect(result).To(ContainSubstring(data.ConfigSourcePath))
		})

		It("includes the GIF output path", func() {
			Expect(result).To(ContainSubstring(data.GIFPath))
		})
		It("quotes the GIF output path for VHS compatibility", func() {
			Expect(result).To(ContainSubstring(`Output "` + data.GIFPath + `"`))
		})

		It("quotes the config source path for VHS compatibility", func() {
			Expect(result).To(ContainSubstring(`Source "` + data.ConfigSourcePath + `"`))
		})

		It("does not include an ASCII output path", func() {
			asciiPath := strings.Replace(data.GIFPath, ".gif", ".ascii", 1)
			Expect(result).NotTo(ContainSubstring(asciiPath))
		})

		It("separates setup commands from visible demo commands", func() {
			setupIndex := strings.Index(result, data.SetupCommands)
			demoIndex := strings.Index(result, data.DemoCommands)
			Expect(setupIndex).To(BeNumerically(">", -1), "setup commands should be present")
			Expect(demoIndex).To(BeNumerically(">", -1), "demo commands should be present")
			Expect(setupIndex).To(BeNumerically("<", demoIndex), "setup should appear before demo")
		})

		It("contains setup commands", func() {
			Expect(result).To(ContainSubstring("mkdir -p /tmp/demo"))
		})

		It("contains demo commands", func() {
			Expect(result).To(ContainSubstring("./kariya --config /tmp/demo/config.yaml"))
		})

		It("includes a termination sequence", func() {
			hasTermination := strings.Contains(result, "Ctrl+C") ||
				strings.Contains(result, "Ctrl+D") ||
				strings.Contains(result, "exit")
			Expect(hasTermination).To(BeTrue(), "should have some form of termination")
		})

		It("does not contain forbidden cleanup commands", func() {
			Expect(result).NotTo(ContainSubstring("rm -rf"))
			Expect(result).NotTo(ContainSubstring("DELETE"))
			Expect(result).NotTo(ContainSubstring("DROP"))
		})

		It("includes the GIF output path exactly once", func() {
			Expect(strings.Count(result, data.GIFPath)).To(Equal(1))
		})

		It("does not include an ASCII output path based on count", func() {
			asciiPath := strings.Replace(data.GIFPath, ".gif", ".ascii", 1)
			Expect(strings.Count(result, asciiPath)).To(Equal(0))
		})
	})

	Describe("rendering with minimal data", func() {
		var result string

		BeforeEach(func() {
			data := cukesvhs.TapeData{
				FeatureName:      "Minimal",
				ScenarioName:     "Test",
				GIFPath:          "out.gif",
				ConfigSourcePath: "config.tape",
				SetupCommands:    "",
				DemoCommands:     "",
			}

			var err error
			result, err = cukesvhs.RenderTape(data)
			Expect(err).NotTo(HaveOccurred())
		})

		It("contains the feature name", func() {
			Expect(result).To(ContainSubstring("# Feature: Minimal"))
		})

		It("contains the scenario name", func() {
			Expect(result).To(ContainSubstring("# Scenario: Test"))
		})
	})

	Describe("rendering with special characters", func() {
		It("renders without error and produces non-empty output", func() {
			data := cukesvhs.TapeData{
				FeatureName:      "Feature with \"quotes\" and 'apostrophes'",
				ScenarioName:     "Scenario with special chars: <>&",
				GIFPath:          "path/with spaces/output.gif",
				ConfigSourcePath: "config.tape",
				SetupCommands:    `Type "echo 'hello world'"`,
				DemoCommands:     `Type "curl http://example.com?foo=bar&baz=qux"`,
			}

			result, err := cukesvhs.RenderTape(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeEmpty())
		})
	})
})
