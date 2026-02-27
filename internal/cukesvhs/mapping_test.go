package cukesvhs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

var _ = Describe("TranslateStep", func() {
	DescribeTable("menu intent selection",
		func(intent string, menuPosition int) {
			stepText := `I select "` + intent + `" from the menu`
			cmds, translatable, reason := cukesvhs.TranslateStep(stepText, "When")

			Expect(translatable).To(BeTrue(), "expected translatable, got reason: %s", reason)
			Expect(cmds).NotTo(BeEmpty(), "expected commands for menu selection")
			Expect(cmds).To(HaveLen(menuPosition+1),
				"menu position %d requires %d navigation steps plus confirmation",
				menuPosition, menuPosition)
		},
		Entry("capture_event", "capture_event", 0),
		Entry("browse_timeline", "browse_timeline", 1),
		Entry("manage_skills", "manage_skills", 2),
		Entry("generate_cv", "generate_cv", 3),
		Entry("configure_system", "configure_system", 4),
		Entry("burst_management", "burst_management", 5),
		Entry("fact_management", "fact_management", 6),
	)

	Context("when selecting an unknown intent", func() {
		It("is not translatable with a descriptive reason", func() {
			cmds, translatable, reason := cukesvhs.TranslateStep(`I select "nonexistent" from the menu`, "When")
			Expect(translatable).To(BeFalse())
			Expect(cmds).To(BeNil())
			Expect(reason).To(ContainSubstring("unrecognised"))
			Expect(reason).To(ContainSubstring("nonexistent"))
		})
	})

	DescribeTable("form-bypass steps",
		func(step string) {
			_, translatable, reason := cukesvhs.TranslateStep(step, "When")
			Expect(translatable).To(BeFalse())
			Expect(reason).To(ContainSubstring("form-bypass"))
		},
		Entry("submit event", "I submit the event"),
		Entry("submit skill form", "I submit the skill form"),
		Entry("confirm filter", "I confirm filter"),
		Entry("confirm sort", "I confirm sort"),
		Entry("accept suggested burst", "I accept the suggested burst"),
		Entry("accept inferred skills", "I accept all inferred skills"),
		Entry("save burst edit", "I save the burst edit"),
		Entry("save metadata changes", "I save metadata changes"),
		Entry("confirm review", "I confirm the review"),
	)

	Describe("unknown steps", func() {
		It("are not translatable with a clear reason", func() {
			_, translatable, reason := cukesvhs.TranslateStep("I do something completely unknown", "When")
			Expect(translatable).To(BeFalse())
			Expect(reason).To(ContainSubstring("unknown"))
			Expect(reason).To(ContainSubstring("no matching pattern"))
		})
	})

	Describe("navigation primitives", func() {
		navigationSteps := []struct {
			step        string
			description string
		}{
			{"I press enter", "confirmation"},
			{"I press enter to view event details", "confirmation with context"},
			{"I press escape", "cancellation"},
			{"I close the modal", "modal dismissal"},
			{"I cancel", "action cancellation"},
			{"I navigate down", "downward navigation"},
			{`I press "j" to navigate down`, "vim-style down navigation"},
			{"I navigate up", "upward navigation"},
			{`I press "k" to navigate up`, "vim-style up navigation"},
			{"I press tab", "field navigation"},
		}

		for _, tc := range navigationSteps {
			Context("when processing "+tc.description, func() {
				It("translates to a single command for "+tc.description, func() {
					cmds, translatable, reason := cukesvhs.TranslateStep(tc.step, "When")
					Expect(translatable).To(BeTrue(), "expected translatable, got: %s", reason)
					Expect(cmds).To(HaveLen(1), "navigation primitives should produce exactly one command")
				})
			})
		}
	})

	Describe("key mapping discrepancies", func() {
		Context("when pressing 's' to view events", func() {
			It("translates to a keyboard shortcut command", func() {
				cmds, translatable, _ := cukesvhs.TranslateStep(`I press "s" to view events`, "When")
				Expect(translatable).To(BeTrue())
				Expect(cmds).To(HaveLen(1))
			})
		})

		Context("when pressing 'm' to open metadata editor", func() {
			It("translates to a text input command", func() {
				cmds, translatable, _ := cukesvhs.TranslateStep(`I press 'm' to open metadata editor`, "When")
				Expect(translatable).To(BeTrue())
				Expect(cmds).To(HaveLen(1))
			})
		})
	})

	Describe("text input", func() {
		It("translates to a command that types the provided text", func() {
			cmds, translatable, _ := cukesvhs.TranslateStep(`I enter event description "Built a REST API"`, "When")
			Expect(translatable).To(BeTrue())
			Expect(cmds).To(HaveLen(1))
			Expect(cmds[0].Args).To(ContainElement("Built a REST API"), "the typed text should be in the command args")
		})
	})

	Describe("setup steps (Given)", func() {
		setupSteps := []string{
			"the database is empty",
			"I am on the main menu",
			"I have 3 skills in my profile",
			`I have a skill "Python"`,
			`I have a skill "Go" with category "backend"`,
			`I have an event "Built API" at company "Acme"`,
			`I have 2 events that use skill "Go"`,
		}

		for _, step := range setupSteps {
			Context("when processing step: "+step, func() {
				It("is translatable but produces no VHS commands (setup is external)", func() {
					cmds, translatable, reason := cukesvhs.TranslateStep(step, "Given")
					Expect(translatable).To(BeTrue(), "setup step should be translatable, got: %s", reason)
					Expect(cmds).To(BeNil(), "setup steps produce no VHS commands")
				})
			})
		}
	})

	Describe("Then steps (observation)", func() {
		DescribeTable("output observation steps",
			func(step string) {
				cmds, translatable, reason := cukesvhs.TranslateStep(step, "Then")
				Expect(translatable).To(BeTrue(), "expected translatable, got reason: %s", reason)
				Expect(cmds).To(HaveLen(1), "observation steps should produce exactly one command")
				Expect(cmds[0].Type).To(Equal(cukesvhs.Sleep))
			},
			Entry("generation output", "I should see the generation output"),
			Entry("run output", "I should see the run output"),
			Entry("verbose output", "I should see the verbose output"),
		)

		Context("when observing command completion", func() {
			It("translates to a sleep command with longer duration", func() {
				cmds, translatable, reason := cukesvhs.TranslateStep("the command should complete successfully", "Then")
				Expect(translatable).To(BeTrue(), "expected translatable, got reason: %s", reason)
				Expect(cmds).To(HaveLen(1))
				Expect(cmds[0].Type).To(Equal(cukesvhs.Sleep))
				Expect(cmds[0].Args).To(ContainElement("3s"))
			})
		})

		Context("when observing specific output text", func() {
			It("translates to a sleep command", func() {
				cmds, translatable, reason := cukesvhs.TranslateStep(`I should see "Written: 3 tapes"`, "Then")
				Expect(translatable).To(BeTrue(), "expected translatable, got reason: %s", reason)
				Expect(cmds).To(HaveLen(1))
				Expect(cmds[0].Type).To(Equal(cukesvhs.Sleep))
			})
		})

		Context("when step type is not Then", func() {
			It("does not match output observation patterns", func() {
				_, translatable, _ := cukesvhs.TranslateStep("I should see the generation output", "When")
				Expect(translatable).To(BeFalse())
			})
		})

		DescribeTable("observation step sleep durations",
			func(step string, expectedDuration string) {
				cmds, translatable, reason := cukesvhs.TranslateStep(step, "Then")
				Expect(translatable).To(BeTrue(), "expected translatable, got reason: %s", reason)
				Expect(cmds[0].Args).To(ContainElement(expectedDuration))
			},
			Entry("output observation uses 2s", "I should see the generation output", "2s"),
			Entry("command completion uses 3s", "the command should complete successfully", "3s"),
			Entry("specific text observation uses 2s", `I should see "some text"`, "2s"),
		)
	})
})

var _ = Describe("ListTranslatablePatterns", func() {
	var patterns []cukesvhs.StepPattern

	BeforeEach(func() {
		patterns = cukesvhs.ListTranslatablePatterns()
	})

	It("returns available patterns for documentation", func() {
		Expect(patterns).NotTo(BeEmpty(), "patterns list should not be empty")
	})

	Describe("pattern completeness", func() {
		It("every pattern is documented with required fields", func() {
			for i, p := range patterns {
				Expect(p.Pattern).NotTo(BeEmpty(), "pattern[%d] missing Pattern", i)
				Expect(p.Type).NotTo(BeEmpty(), "pattern[%d] (%s) missing Type", i, p.Pattern)
				Expect(p.Category).NotTo(BeEmpty(), "pattern[%d] (%s) missing Category", i, p.Pattern)
				Expect(p.Example).NotTo(BeEmpty(), "pattern[%d] (%s) missing Example", i, p.Pattern)
			}
		})
	})

	Describe("pattern filtering", func() {
		It("excludes form-bypass patterns that cannot be translated", func() {
			for _, p := range patterns {
				Expect(p.Category).NotTo(Equal("form-bypass"),
					"form-bypass patterns should not appear in translatable list")
			}
		})
	})

	Describe("menu selection pattern", func() {
		var menuPattern cukesvhs.StepPattern
		var found bool

		BeforeEach(func() {
			for _, p := range patterns {
				if p.Pattern == `^I select "([^"]*)" from the menu$` {
					menuPattern = p
					found = true
					break
				}
			}
		})

		It("is available for menu navigation", func() {
			Expect(found).To(BeTrue(), "menu selection pattern not found in available patterns")
		})

		It("documents the valid intent options", func() {
			intentParam, ok := menuPattern.Params["intent"]
			Expect(ok).To(BeTrue(), "menu pattern should document intent options")
			Expect(intentParam.Values).NotTo(BeEmpty(), "intent param should list valid values")
		})
	})

	Describe("category coverage", func() {
		It("covers navigation, input, setup, and observation categories", func() {
			categories := make(map[string]bool)
			for _, p := range patterns {
				categories[p.Category] = true
			}
			Expect(categories).To(HaveKey("navigation"), "navigation category should be present")
			Expect(categories).To(HaveKey("input"), "input category should be present")
			Expect(categories).To(HaveKey("setup"), "setup category should be present")
			Expect(categories).To(HaveKey("observation"), "observation category should be present")
		})
	})
})
