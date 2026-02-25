package cukesvhs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukesvhs/internal/cukesvhs"
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

		It("contains the Source directive", func() {
			Expect(result).To(ContainSubstring("Source demos/vhs/config.tape"))
		})

		It("contains the GIF output directive", func() {
			Expect(result).To(ContainSubstring("Output demos/vhs/features/user-registration/happy-path.gif"))
		})

		It("does not contain an ASCII output directive", func() {
			Expect(result).NotTo(ContainSubstring("Output demos/vhs/features/user-registration/happy-path.ascii"))
		})

		It("contains a Hide block", func() {
			Expect(result).To(ContainSubstring("Hide"))
		})

		It("contains a Show block", func() {
			Expect(result).To(ContainSubstring("Show"))
		})

		It("contains setup commands", func() {
			Expect(result).To(ContainSubstring("mkdir -p /tmp/demo"))
		})

		It("contains demo commands", func() {
			Expect(result).To(ContainSubstring("./kariya --config /tmp/demo/config.yaml"))
		})

		It("contains an exit command", func() {
			Expect(result).To(ContainSubstring("Ctrl+C"))
		})

		It("does not contain forbidden cleanup commands", func() {
			Expect(result).NotTo(ContainSubstring("rm -rf"))
			Expect(result).NotTo(ContainSubstring("DELETE"))
			Expect(result).NotTo(ContainSubstring("DROP"))
		})

		It("contains exactly 1 GIF Output directive", func() {
			Expect(countSubstring(result, "Output demos/vhs/features/user-registration/happy-path.gif")).To(Equal(1))
		})

		It("does not contain an ASCII Output directive", func() {
			Expect(countSubstring(result, "Output demos/vhs/features/user-registration/happy-path.ascii")).To(Equal(0))
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

func countSubstring(s, substr string) int {
	count := 0
	idx := 0
	for {
		i := findSubstring(s[idx:], substr)
		if i < 0 {
			break
		}
		count++
		idx += i + len(substr)
	}
	return count
}

func findSubstring(s, substr string) int {
	if substr == "" {
		return 0
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
