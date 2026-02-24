package vhsgen_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukesvhs/internal/vhsgen"
)

var _ = Describe("TranslateStep", func() {
	Describe("menu intent selection", func() {
		menuIntentCases := []struct {
			intent    string
			wantDowns int
		}{
			{"capture_event", 0},
			{"browse_timeline", 1},
			{"manage_skills", 2},
			{"generate_cv", 3},
			{"configure_system", 4},
			{"burst_management", 5},
			{"fact_management", 6},
		}

		for _, tc := range menuIntentCases {
			Context("when selecting "+tc.intent+" from the menu", func() {
				var cmds []vhsgen.VHSCommand

				BeforeEach(func() {
					stepText := `I select "` + tc.intent + `" from the menu`
					var translatable bool
					var reason string
					cmds, translatable, reason = vhsgen.TranslateStep(stepText, "When")
					Expect(translatable).To(BeTrue(), "expected translatable, got reason: %s", reason)
				})

				It("produces the correct number of Down commands", func() {
					downCount := 0
					for _, cmd := range cmds {
						if cmd.Type == vhsgen.Down {
							downCount++
						}
					}
					Expect(downCount).To(Equal(tc.wantDowns))
				})

				It("ends with an Enter command", func() {
					hasEnter := false
					for _, cmd := range cmds {
						if cmd.Type == vhsgen.Enter {
							hasEnter = true
						}
					}
					Expect(hasEnter).To(BeTrue())
				})

				It("has the correct total number of commands", func() {
					Expect(cmds).To(HaveLen(tc.wantDowns + 1))
				})
			})
		}

		Context("capture_event: first menu item", func() {
			It("produces only 1 command (Enter only)", func() {
				cmds, translatable, _ := vhsgen.TranslateStep(`I select "capture_event" from the menu`, "When")
				Expect(translatable).To(BeTrue())
				Expect(cmds).To(HaveLen(1))
				Expect(cmds[0].Type).To(Equal(vhsgen.Enter))
			})
		})

		Context("manage_skills: third menu item", func() {
			It("produces 2 Down commands then Enter", func() {
				cmds, translatable, _ := vhsgen.TranslateStep(`I select "manage_skills" from the menu`, "When")
				Expect(translatable).To(BeTrue())
				Expect(cmds).To(HaveLen(3))
				Expect(cmds[0].Type).To(Equal(vhsgen.Down))
				Expect(cmds[1].Type).To(Equal(vhsgen.Down))
				Expect(cmds[2].Type).To(Equal(vhsgen.Enter))
			})
		})

		Context("fact_management: seventh menu item", func() {
			It("produces 6 Down commands then Enter", func() {
				cmds, translatable, _ := vhsgen.TranslateStep(`I select "fact_management" from the menu`, "When")
				Expect(translatable).To(BeTrue())
				Expect(cmds).To(HaveLen(7))
				for i := range 6 {
					Expect(cmds[i].Type).To(Equal(vhsgen.Down), "command[%d] should be Down", i)
				}
				Expect(cmds[6].Type).To(Equal(vhsgen.Enter))
			})
		})

		Context("unknown intent", func() {
			It("matches the pattern but returns nil commands", func() {
				cmds, translatable, reason := vhsgen.TranslateStep(`I select "nonexistent" from the menu`, "When")
				Expect(translatable).To(BeFalse())
				Expect(cmds).To(BeNil())
				Expect(reason).To(Equal("unrecognised menu intent: nonexistent"))
			})
		})
	})

	Describe("form-bypass steps", func() {
		formBypassSteps := []string{
			"I submit the event",
			"I submit the skill form",
			"I confirm filter",
			"I confirm sort",
			"I accept the suggested burst",
			"I accept all inferred skills",
			"I save the burst edit",
			"I save metadata changes",
			"I confirm the review",
		}

		for _, step := range formBypassSteps {
			Context("step: "+step, func() {
				It("is untranslatable with form-bypass reason", func() {
					_, translatable, reason := vhsgen.TranslateStep(step, "When")
					Expect(translatable).To(BeFalse())
					Expect(reason).To(Equal("form-bypass: use keyboard navigation instead"))
				})
			})
		}
	})

	Describe("unknown steps", func() {
		It("returns untranslatable with 'unknown step: no matching pattern' reason", func() {
			_, translatable, reason := vhsgen.TranslateStep("I do something completely unknown", "When")
			Expect(translatable).To(BeFalse())
			Expect(reason).To(Equal("unknown step: no matching pattern"))
		})
	})

	Describe("navigation primitives", func() {
		navigationCases := []struct {
			step     string
			wantType vhsgen.VHSCommandType
		}{
			{"I press enter", vhsgen.Enter},
			{"I press enter to view event details", vhsgen.Enter},
			{"I press escape", vhsgen.Escape},
			{"I close the modal", vhsgen.Escape},
			{"I cancel", vhsgen.Escape},
			{"I navigate down", vhsgen.Down},
			{`I press "j" to navigate down`, vhsgen.Down},
			{"I navigate up", vhsgen.Up},
			{`I press "k" to navigate up`, vhsgen.Up},
			{"I press tab", vhsgen.Tab},
		}

		for _, tc := range navigationCases {
			Context("step: "+tc.step, func() {
				It("translates to the correct command type", func() {
					cmds, translatable, reason := vhsgen.TranslateStep(tc.step, "When")
					Expect(translatable).To(BeTrue(), "expected translatable, got: %s", reason)
					Expect(cmds).To(HaveLen(1))
					Expect(cmds[0].Type).To(Equal(tc.wantType))
				})
			})
		}
	})

	Describe("key discrepancies", func() {
		Context("press s to view events", func() {
			It("sends Ctrl+E", func() {
				cmds, translatable, _ := vhsgen.TranslateStep(`I press "s" to view events`, "When")
				Expect(translatable).To(BeTrue())
				Expect(cmds).To(HaveLen(1))
				Expect(cmds[0].Type).To(Equal(vhsgen.CtrlE))
			})
		})

		Context("press m to open metadata editor", func() {
			It("sends Type 'e'", func() {
				cmds, translatable, _ := vhsgen.TranslateStep(`I press 'm' to open metadata editor`, "When")
				Expect(translatable).To(BeTrue())
				Expect(cmds).To(HaveLen(1))
				Expect(cmds[0].Type).To(Equal(vhsgen.Type))
				Expect(cmds[0].Args[0]).To(Equal("e"))
			})
		})
	})

	Describe("text input", func() {
		It("translates to a Type command with speed and text args", func() {
			cmds, translatable, _ := vhsgen.TranslateStep(`I enter event description "Built a REST API"`, "When")
			Expect(translatable).To(BeTrue())
			Expect(cmds).To(HaveLen(1))
			Expect(cmds[0].Type).To(Equal(vhsgen.Type))
			Expect(cmds[0].Args).To(HaveLen(2))
			Expect(cmds[0].Args[0]).To(Equal("100ms"))
			Expect(cmds[0].Args[1]).To(Equal("Built a REST API"))
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
			Context("step: "+step, func() {
				It("is translatable with nil commands", func() {
					cmds, translatable, reason := vhsgen.TranslateStep(step, "Given")
					Expect(translatable).To(BeTrue(), "setup step should be translatable, got: %s", reason)
					Expect(cmds).To(BeNil())
				})
			})
		}
	})
})

var _ = Describe("ListTranslatablePatterns", func() {
	var patterns []vhsgen.StepPattern

	BeforeEach(func() {
		patterns = vhsgen.ListTranslatablePatterns()
	})

	It("returns a non-empty list", func() {
		Expect(patterns).NotTo(BeEmpty())
	})

	It("every pattern has a non-empty Pattern field", func() {
		for i, p := range patterns {
			Expect(p.Pattern).NotTo(BeEmpty(), "pattern[%d] has empty Pattern", i)
		}
	})

	It("every pattern has a non-empty Type field", func() {
		for i, p := range patterns {
			Expect(p.Type).NotTo(BeEmpty(), "pattern[%d] (%s) has empty Type", i, p.Pattern)
		}
	})

	It("every pattern has a non-empty Category field", func() {
		for i, p := range patterns {
			Expect(p.Category).NotTo(BeEmpty(), "pattern[%d] (%s) has empty Category", i, p.Pattern)
		}
	})

	It("every pattern has a non-empty Example field", func() {
		for i, p := range patterns {
			Expect(p.Example).NotTo(BeEmpty(), "pattern[%d] (%s) has empty Example", i, p.Pattern)
		}
	})

	It("does not include form-bypass patterns", func() {
		for _, p := range patterns {
			Expect(p.Category).NotTo(Equal("form-bypass"))
		}
	})

	Describe("menu selection pattern", func() {
		var menuPattern vhsgen.StepPattern
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

		It("exists in the list", func() {
			Expect(found).To(BeTrue(), "menu selection pattern not found")
		})

		It("has an 'intent' param of type enum", func() {
			intentParam, ok := menuPattern.Params["intent"]
			Expect(ok).To(BeTrue(), "menu pattern should have 'intent' param")
			Expect(intentParam.Type).To(Equal("enum"))
		})

		It("has exactly 7 valid intent values", func() {
			intentParam := menuPattern.Params["intent"]
			Expect(intentParam.Values).To(HaveLen(7))
		})
	})

	Describe("categories", func() {
		It("includes navigation, input, and setup categories", func() {
			categories := make(map[string]bool)
			for _, p := range patterns {
				categories[p.Category] = true
			}
			Expect(categories).To(HaveKey("navigation"))
			Expect(categories).To(HaveKey("input"))
			Expect(categories).To(HaveKey("setup"))
		})
	})
})
