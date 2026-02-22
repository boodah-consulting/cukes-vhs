// Package vhsgen provides intermediate representation types and step-to-VHS translation for Gherkin scenarios.
package vhsgen

import (
	"regexp"
)

// menuOrder maps intent names to their position in the main menu.
// Source of truth: internal/testutil/e2e/helpers.go:610-617.
// Unexported to prevent external mutation; use MenuOrder() for a safe copy.
var menuOrder = map[string]int{
	"capture_event":    0,
	"browse_timeline":  1,
	"manage_skills":    2,
	"generate_cv":      3,
	"configure_system": 4,
	"burst_management": 5,
	"fact_management":  6,
}

// MenuOrder returns a copy of the intent-to-menu-position mapping.
func MenuOrder() map[string]int {
	result := make(map[string]int, len(menuOrder))
	for k, v := range menuOrder {
		result[k] = v
	}
	return result
}

// validIntents lists all valid intent names in menu order.
// Unexported to prevent external mutation; use ValidIntents() for a safe copy.
var validIntents = []string{
	"capture_event",
	"browse_timeline",
	"manage_skills",
	"generate_cv",
	"configure_system",
	"burst_management",
	"fact_management",
}

// ValidIntents returns a copy of the valid intent names in menu order.
func ValidIntents() []string {
	result := make([]string, len(validIntents))
	copy(result, validIntents)
	return result
}

const (
	// CtrlE represents the Ctrl+E keyboard command.
	CtrlE VHSCommandType = "Ctrl+E"
	// CtrlS represents the Ctrl+S keyboard command.
	CtrlS VHSCommandType = "Ctrl+S"
)

type stepMatcher struct {
	pattern      *regexp.Regexp
	translatable bool
	reason       string
	category     string
	stepType     string
	example      string
	params       map[string]ParamConstraint
	translate    func(matches []string) []VHSCommand
}

// matchers holds the compiled step matchers. Initialised once in init() and
// never modified thereafter; safe for concurrent reads.
var matchers []stepMatcher

func init() {
	matchers = buildMatchers()
}

// TranslateStep translates a Gherkin step text into VHS commands.
//
// Expected: stepText is a Gherkin step string, stepType is "Given", "When", or "Then".
// Returns: VHS commands (or nil if intent unknown), translatable flag, and reason if untranslatable.
// Side effects: None.
func TranslateStep(stepText, stepType string) ([]VHSCommand, bool, string) {
	_ = stepType

	for _, m := range matchers {
		matches := m.pattern.FindStringSubmatch(stepText)
		if matches == nil {
			continue
		}

		if !m.translatable {
			return nil, false, m.reason
		}

		cmds := m.translate(matches)

		return cmds, true, ""
	}

	return nil, false, "unknown step: no matching pattern"
}

// ListTranslatablePatterns returns all translatable step patterns for use by the CLI --steps flag and documentation.
//
// Expected: None.
// Returns: Slice of StepPattern describing all translatable step patterns.
// Side effects: None.
func ListTranslatablePatterns() []StepPattern {
	var patterns []StepPattern

	for _, m := range matchers {
		if !m.translatable {
			continue
		}

		patterns = append(patterns, StepPattern{
			Pattern:  m.pattern.String(),
			Type:     m.stepType,
			Category: m.category,
			Params:   m.params,
			Example:  m.example,
		})
	}

	return patterns
}

//nolint:funlen // matcher registration is long by design
func buildMatchers() []stepMatcher {
	return []stepMatcher{
		formBypassMatcher(`^I submit the event$`),
		formBypassMatcher(`^I submit the skill form$`),
		formBypassMatcher(`^I confirm filter$`),
		formBypassMatcher(`^I confirm sort$`),
		formBypassMatcher(`^I accept the suggested burst$`),
		formBypassMatcher(`^I accept all inferred skills$`),
		formBypassMatcher(`^I save the burst edit$`),
		formBypassMatcher(`^I save metadata changes$`),
		formBypassMatcher(`^I confirm the review$`),

		{
			pattern:      regexp.MustCompile(`^I select "([^"]*)" from the menu$`),
			translatable: true,
			category:     "navigation",
			stepType:     "When",
			example:      `I select "manage_skills" from the menu`,
			params: map[string]ParamConstraint{
				"intent": {Type: "enum", Values: validIntents},
			},
			translate: func(matches []string) []VHSCommand {
				intent := matches[1]
				index, ok := menuOrder[intent]
				if !ok {
					return nil
				}

				var cmds []VHSCommand
				for range index {
					cmds = append(cmds, VHSCommand{Type: Down})
				}
				cmds = append(cmds, VHSCommand{Type: Enter})

				return cmds
			},
		},

		verifiedKeyMatcher(`^I press "s" to view events$`, CtrlE,
			`I press "s" to view events`),
		charKeyMatcher(`^I press 'm' to open metadata editor$`, "e",
			`I press 'm' to open metadata editor`),
		charKeyMatcher(`^I press "a" to add skill$`, "a",
			`I press "a" to add skill`),
		charKeyMatcher(`^I press "d" to delete$`, "d",
			`I press "d" to delete`),
		charKeyMatcher(`^I press "i" to infer skills$`, "i",
			`I press "i" to infer skills`),
		charKeyMatcher(`^I press "f" to filter$`, "f",
			`I press "f" to filter`),
		charKeyMatcher(`^I press "s" to sort$`, "s",
			`I press "s" to sort`),
		charKeyMatcher(`^I press "/" to search$`, "/",
			`I press "/" to search`),
		charKeyMatcher(`^I press 'b' to open bursts editor$`, "b",
			`I press 'b' to open bursts editor`),
		charKeyMatcher(`^I press 'f' to open facts editor$`, "f",
			`I press 'f' to open facts editor`),
		verifiedKeyMatcher(`^I press Ctrl\+S$`, CtrlS,
			`I press Ctrl+S`),

		navMatcher(`^I press enter`, Enter, `I press enter`),
		navMatcher(`^I close the modal$`, Escape, `I close the modal`),
		navMatcher(`^I press escape$`, Escape, `I press escape`),
		navMatcher(`^I cancel$`, Escape, `I cancel`),
		navMatcher(`^I navigate down$`, Down, `I navigate down`),
		navMatcher(`^I press "j" to navigate down$`, Down,
			`I press "j" to navigate down`),
		navMatcher(`^I navigate up$`, Up, `I navigate up`),
		navMatcher(`^I press "k" to navigate up$`, Up,
			`I press "k" to navigate up`),
		navMatcher(`^I press tab$`, Tab, `I press tab`),

		inputMatcher(`^I enter event description "([^"]*)"$`,
			`I enter event description "Built a REST API"`),
		inputMatcher(`^I enter skill name "([^"]*)"$`,
			`I enter skill name "Python"`),
		inputMatcher(`^I enter "([^"]*)" as skill name$`,
			`I enter "Go" as skill name`),
		inputMatcher(`^I enter years of experience "([^"]*)"$`,
			`I enter years of experience "5"`),
		inputMatcher(`^I type "([^"]*)"$`,
			`I type "hello"`),
		inputMatcher(`^I enter "([^"]*)"$`,
			`I enter "test input"`),

		setupMatcher(`^the database is empty$`,
			`the database is empty`),
		setupMatcher(`^I am on the main menu$`,
			`I am on the main menu`),
		setupMatcher(`^I have \d+ skills? in my profile$`,
			`I have 3 skills in my profile`),
		setupMatcher(`^I have a skill "([^"]*)"$`,
			`I have a skill "Python"`),
		setupMatcher(`^I have a skill "([^"]*)" with`,
			`I have a skill "Go" with category "backend"`),
		setupMatcher(`^I have an event`,
			`I have an event "Built API" at company "Acme"`),
		setupMatcher(`^I have \d+ events that use skill`,
			`I have 3 events that use skill "Go"`),
	}
}

func formBypassMatcher(pattern string) stepMatcher {
	return stepMatcher{
		pattern:      regexp.MustCompile(pattern),
		translatable: false,
		reason:       "form-bypass: use keyboard navigation instead",
		category:     "form-bypass",
	}
}

func navMatcher(pattern string, cmdType VHSCommandType, example string) stepMatcher {
	return stepMatcher{
		pattern:      regexp.MustCompile(pattern),
		translatable: true,
		category:     "navigation",
		stepType:     "When",
		example:      example,
		translate: func(_ []string) []VHSCommand {
			return []VHSCommand{{Type: cmdType}}
		},
	}
}

func charKeyMatcher(pattern, key, example string) stepMatcher {
	return stepMatcher{
		pattern:      regexp.MustCompile(pattern),
		translatable: true,
		category:     "navigation",
		stepType:     "When",
		example:      example,
		translate: func(_ []string) []VHSCommand {
			return []VHSCommand{{Type: Type, Args: []string{key}}}
		},
	}
}

func verifiedKeyMatcher(pattern string, cmdType VHSCommandType, example string) stepMatcher {
	return stepMatcher{
		pattern:      regexp.MustCompile(pattern),
		translatable: true,
		category:     "navigation",
		stepType:     "When",
		example:      example,
		translate: func(_ []string) []VHSCommand {
			return []VHSCommand{{Type: cmdType}}
		},
	}
}

func inputMatcher(pattern, example string) stepMatcher {
	return stepMatcher{
		pattern:      regexp.MustCompile(pattern),
		translatable: true,
		category:     "input",
		stepType:     "When",
		example:      example,
		translate: func(matches []string) []VHSCommand {
			return []VHSCommand{{Type: Type, Args: []string{"100ms", matches[1]}}}
		},
	}
}

func setupMatcher(pattern, example string) stepMatcher {
	return stepMatcher{
		pattern:      regexp.MustCompile(pattern),
		translatable: true,
		category:     "setup",
		stepType:     "Given",
		example:      example,
		translate:    func(_ []string) []VHSCommand { return nil },
	}
}
